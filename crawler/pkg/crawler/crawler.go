package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/StevenRojas/natscrawler/crawler/config"
	"github.com/StevenRojas/natscrawler/crawler/pkg/model"
	"github.com/StevenRojas/natscrawler/crawler/pkg/repository"
	"github.com/StevenRojas/natscrawler/grpcapi/pkg/crawler/pb"
	"github.com/chromedp/chromedp"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Crawler interface {
	Process(ctx context.Context)
}

type chromeService struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	UserAgent            string `json:"User-Agent"`
	V8Version            string `json:"V8-Version"`
	WebKitVersion        string `json:"WebKit-Version"`
	WebSocketDebuggerUrl string `json:"webSocketDebuggerUrl"`
}

type crawlerService struct {
	repo       repository.Repository
	natsClient *nats.Conn
	topic      string
	group      string
	wsUrl      string
	useAPI     bool
}

func NewCrawler(repo repository.Repository, conf config.AppConfig) (Crawler, error) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lmsgprefix)

	var webSocketDebuggerUrl string
	if !conf.Crawler.UseAPI {
		cs, err := getChromeInfo()
		if err != nil {
			return nil, err
		}
		webSocketDebuggerUrl = cs.WebSocketDebuggerUrl
	}

	nc, err := connectToNats(conf)
	if err != nil {
		return nil, err
	}
	return &crawlerService{
		repo:       repo,
		natsClient: nc,
		topic:      conf.Queue.Topic,
		group:      conf.Queue.Group,
		wsUrl:      webSocketDebuggerUrl,
		useAPI:     conf.Crawler.UseAPI,
	}, nil
}

func getChromeInfo() (*chromeService, error) {
	req, err := http.NewRequest("GET", "http://chrome:9222/json/version", nil)
	if err != nil {
		return nil, err
	}

	req.Host = "localhost"
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to get Chrome service information: %s", resp.Status)
	}
	var cs chromeService
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &cs)
	if err != nil {
		return nil, err
	}
	cs.WebSocketDebuggerUrl = strings.Replace(cs.WebSocketDebuggerUrl, "localhost", "chrome:9222", 1)
	log.Print("\033[H\033[2J")
	log.Printf("Connecting to chrome at %s", cs.WebSocketDebuggerUrl)
	return &cs, nil
}

// Process get URL from queue, collect information and store it, using Fan-In Fan-Out pattern
func (c *crawlerService) Process(ctx context.Context) {
	messages := make(chan model.UrlInfo)
	numCollectors := runtime.NumCPU()
	// Fan-Out the work to multiple collectors
	collectors := make([]<-chan model.UrlInfo, numCollectors)
	for i := 0; i < numCollectors; i++ {
		collectors[i] = c.collect(ctx, messages)
	}
	// Fan-In the results and store them
	go c.storeResults(ctx, collectors...)

	// Listen for queue and produce messages for collectors
	sub, err := c.natsClient.QueueSubscribe(c.topic, c.group, func(m *nats.Msg) {
		var request pb.UrlRequest
		err := proto.Unmarshal(m.Data, &request)
		if err != nil {
			log.Printf("unable to parse incomming message %s\n", err)
		} else {
			select {
			case <-ctx.Done():
				return
			case messages <- model.UrlInfo{
				RequestID: request.RequestId,
				Url:       request.Url,
				Stats: model.Stats{
					Waiting: model.Times{StartAt: time.Now().UTC()},
				},
			}:
			}
		}
	})
	if err != nil {
		log.Fatal("Error establishing connection to NATS:", err)
	}

	defer func(sub *nats.Subscription) {
		err := sub.Unsubscribe()
		if err != nil {
			log.Fatal("unable to unsubscribe from NATS:", err)
		}
		close(messages)
	}(sub)
	<-ctx.Done()
}

func (c *crawlerService) collect(ctx context.Context, messages <-chan model.UrlInfo) <-chan model.UrlInfo {
	resultChannel := make(chan model.UrlInfo)
	go func() {
		for message := range messages {
			log.Println("Got task request on:", message.RequestID, message.Url)
			if c.useAPI {
				resultChannel <- doAPI(ctx, message)
			} else {
				resultChannel <- c.doCrawler(ctx, message)
			}
		}
		close(resultChannel)
	}()
	return resultChannel
}

func (c *crawlerService) storeResults(ctx context.Context, collectors ...<-chan model.UrlInfo) {
	var wg sync.WaitGroup
	wg.Add(len(collectors))
	multiplex := func(responses <-chan model.UrlInfo, collectorID int) {
		defer wg.Done()
		for urlInfo := range responses {
			select {
			case <-ctx.Done():
				return
			default:
				urlInfo.Stats.CollectorID = collectorID
				urlInfo.Stats.Collector.EndAt = time.Now().UTC()
				urlInfo.Stats.Collector.Duration = time.Since(urlInfo.Stats.Collector.StartAt).Milliseconds()
				err := c.repo.AddURL(ctx, urlInfo)
				if err != nil {
					log.Printf("error storing URL info in the DB: %s\n", err.Error())
				}
			}
		}
	}
	for id, collector := range collectors {
		go multiplex(collector, id)
	}
	go func() {
		wg.Wait()
	}()
}

func (c *crawlerService) doCrawler(ctx context.Context, urlInfo model.UrlInfo) model.UrlInfo {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	urlInfo.Stats.Waiting.EndAt = time.Now().UTC()
	urlInfo.Stats.Waiting.Duration = time.Since(urlInfo.Stats.Waiting.StartAt).Milliseconds()
	urlInfo.Stats.Collector.StartAt = time.Now().UTC()
	urlInfo.Success = false
	select {
	case <-ctx.Done():
		urlInfo.LastError = ctx.Err().Error()
		return urlInfo
	default:
		var name string
		var ratingValue string
		var ratingCount string
		ctx, _ = chromedp.NewRemoteAllocator(ctx, c.wsUrl)
		ctxt, _ := chromedp.NewContext(ctx)
		err := chromedp.Run(ctxt,
			chromedp.Navigate(urlInfo.Url),
			chromedp.WaitReady(`.Roku-User-Channels`),
			chromedp.Text(`h1[itemprop="name"]`, &name, chromedp.NodeVisible),
			chromedp.Text(`span[itemprop="averageRating"]`, &ratingValue, chromedp.NodeVisible),
			chromedp.Text(`small[itemprop="starRating"]`, &ratingCount, chromedp.NodeVisible),
		)

		if err != nil {
			urlInfo.LastError = err.Error()
			return urlInfo
		}
		rating, err := strconv.ParseFloat(ratingValue, 64)
		if err != nil {
			urlInfo.LastError = err.Error()
			return urlInfo
		}

		temp := strings.Replace(ratingCount, ratingValue, "", 1)
		temp = strings.Trim(strings.Replace(temp, "ratings", "", 1), " \n")
		count, err := strconv.Atoi(temp)
		if err != nil {
			urlInfo.LastError = err.Error()
			return urlInfo
		}
		urlInfo.AppName = name
		urlInfo.Rating = rating
		urlInfo.RatingCount = count
		urlInfo.Success = true
		return urlInfo
	}
}

func doAPI(ctx context.Context, urlInfo model.UrlInfo) model.UrlInfo {
	urlInfo.Stats.Waiting.EndAt = time.Now().UTC()
	urlInfo.Stats.Waiting.Duration = time.Since(urlInfo.Stats.Waiting.StartAt).Milliseconds()
	urlInfo.Stats.Collector.StartAt = time.Now().UTC()
	urlInfo.Success = false
	parsed, _ := url.Parse(urlInfo.Url)
	parts := strings.Split(parsed.Path, "/")
	var uri string
	if len(parts) == 4 {
		uri = fmt.Sprintf("https://%s/api/v6/channels/detailsunion/%s", parsed.Host, parts[2])
	} else {
		locale := strings.Split(parts[1], "-")
		uri = fmt.Sprintf("https://%s/api/v6/channels/detailsunion/%s?country=%s&language=%s", parsed.Host, parts[3], locale[1], locale[0])
	}
	response, err := http.Get(uri)
	if err != nil {
		fmt.Printf("URL error %+v\n", uri)
		urlInfo.LastError = err.Error()
		return urlInfo
	}
	if response.StatusCode != 200 {
		fmt.Printf("URL response != 200 %+v\n", uri)
		urlInfo.LastError = fmt.Sprintf("unable to get API information: %s", response.Status)
		return urlInfo
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		urlInfo.LastError = err.Error()
		return urlInfo
	}
	var apiResponse map[string]interface{}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		urlInfo.LastError = err.Error()
		return urlInfo
	}
	feed := apiResponse["feedChannel"].(map[string]interface{})
	rating := (feed["starRating"].(float64) * 5) / 100
	urlInfo.AppName = feed["name"].(string)
	urlInfo.Rating = math.Round(rating*100) / 100
	if feed["starRatingCount"] != nil {
		urlInfo.RatingCount = int(feed["starRatingCount"].(float64))
	}
	return urlInfo
}

// connectToNats connect to NATS server
func connectToNats(conf config.AppConfig) (*nats.Conn, error) {
	opts := nats.Options{
		Url:            conf.Nats.Host,
		AllowReconnect: conf.Nats.AllowReconnect,
		MaxReconnect:   conf.Nats.MaxReconnectAttempts,
		ReconnectWait:  conf.Nats.ReconnectWait,
		Timeout:        conf.Nats.Timeout,
	}

	nc, err := opts.Connect()
	if err != nil {
		return nil, err
	}
	log.Println("Connected to NATS at:", nc.ConnectedUrl())
	return nc, nil
}
