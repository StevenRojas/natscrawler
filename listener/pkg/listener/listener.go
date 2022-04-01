package listener

import (
	"context"
	"fmt"
	"github.com/StevenRojas/natscrawler/grpcapi/pkg/crawler/pb"
	"github.com/StevenRojas/natscrawler/listener/config"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
)

type crawlerServer struct {
	pb.UnimplementedCrawlerServiceServer
	natsClient *nats.Conn
	topic string
}

// ProcessUrl get a URL and send it to NATS queue to be processed by the crawler services
func (c *crawlerServer) ProcessUrl(ctx context.Context, request *pb.UrlRequest) (*pb.UrlResponse, error) {
	// Fake a rejected URL
	if rand.Float32() > 0.9999 {
		return &pb.UrlResponse{
			RequestId: request.RequestId,
			Status: pb.UrlResponse_STATUS_RETRY,
		}, nil
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	// publish encoded proto request to NATS
	err = c.natsClient.Publish(c.topic, encoded)
	if err != nil {
		return nil, err
	}

	return &pb.UrlResponse{
		RequestId: request.RequestId,
		Status: pb.UrlResponse_STATUS_ACCEPTED,
	}, nil
}

// NewListener creates a new listener instance
func NewListener(conf config.AppConfig) error {
	nc, err := connectToNats(conf)
	if err != nil {
		return err
	}
	crawler := &crawlerServer{
		natsClient: nc,
		topic: conf.Queue.Topic,
	}
	err = startGRPCServer(conf, crawler)
	if err != nil {
		return err
	}
	return nil
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

// startGRPCServer starts the GRPC server
func startGRPCServer(conf config.AppConfig, crawler *crawlerServer) error {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", conf.Grpc.Address)
	if err != nil {
		return err
	}

	pb.RegisterCrawlerServiceServer(server, crawler)
	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Panicf("unable to start GRPC server: %s\n", err.Error())
		}
	}()
	fmt.Printf("Starting GRPC server at %s\n", conf.Grpc.Address)
	return nil
}
