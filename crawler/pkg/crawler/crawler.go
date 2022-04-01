package crawler

import (
	"context"
	"fmt"
	"github.com/StevenRojas/natscrawler/crawler/config"
	"github.com/StevenRojas/natscrawler/crawler/pkg/model"
	"github.com/StevenRojas/natscrawler/crawler/pkg/repository"
	"github.com/StevenRojas/natscrawler/grpcapi/pkg/crawler/pb"
	"github.com/chromedp/chromedp"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Crawler interface {
	Process(ctx context.Context)
}

type crawlerService struct {
	repo       repository.Repository
	natsClient *nats.Conn
	topic      string
	group      string
}

func NewCrawler(repo repository.Repository, conf config.AppConfig) (Crawler, error) {
	nc, err := connectToNats(conf)
	if err != nil {
		return nil, err
	}
	return &crawlerService{
		repo:       repo,
		natsClient: nc,
		topic:      conf.Queue.Topic,
		group:      conf.Queue.Group,
	}, nil
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
			fmt.Println("Got task request on:", message.RequestID, message.Url)
			resultChannel <- doCrawler(ctx, message)
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
					fmt.Printf("error storing URL info in the DB: %s\n", err.Error())
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

func doCrawler(ctx context.Context, urlInfo model.UrlInfo) model.UrlInfo {
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

		err := chromedp.Run(ctx,
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
	fmt.Println("Connected to NATS at:", nc.ConnectedUrl())
	return nc, nil
}
