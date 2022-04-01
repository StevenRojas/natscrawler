package sender

import (
	"context"
	"fmt"
	"github.com/StevenRojas/natscrawler/grpcapi/pkg/crawler/pb"
	"github.com/StevenRojas/natscrawler/reader/config"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"sync"
)

type Sender interface {
	Start(ctx context.Context, wg*sync.WaitGroup, rowCh <-chan string)
}

type sender struct {
	grpcConn     *grpc.ClientConn
	grpcCancelFn context.CancelFunc
}

// NewGrpcService returns a GRPC service instance
func NewGrpcService(ctx context.Context, conf config.GrpcServer) (Sender, error) {
	ctx, cancel := context.WithTimeout(ctx, conf.DialTimeout)
	conn, err := grpc.DialContext(
		ctx,
		conf.ListenerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                conf.PingInterval,
			Timeout:             conf.PingTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("grpcapi connection error: %v", err)
	}

	return &sender{
		grpcConn:     conn,
		grpcCancelFn: cancel,
	}, nil
}

// Start listening for urls from the channel and send them to the Crawler service using GRPC
func (s sender) Start(ctx context.Context, wg*sync.WaitGroup, urlCh <-chan string) {
	defer wg.Done()
	client := pb.NewCrawlerServiceClient(s.grpcConn)
	for url := range urlCh {
		select {
		case <-ctx.Done():
			return
		default:
		}
		response, err := client.ProcessUrl(ctx, &pb.UrlRequest{
			RequestId: uuid.New().String(),
			Url: url,
		})
		if err != nil {
			log.Fatalf("error sending URL to Crawler service: %s\n", err.Error())
		}
		switch response.Status {
		case pb.UrlResponse_STATUS_UNAVAILABLE:
			log.Fatalf("the Crawler service is not ablailable to process requests\n")
		case pb.UrlResponse_STATUS_RETRY:
			log.Printf("the Crawler service is asking for retry the current URL: %s\n", url)
		}
	}
	log.Println("Sender is done")
}
