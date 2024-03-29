// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.17.3
// source: crawler.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CrawlerServiceClient is the client API for CrawlerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerServiceClient interface {
	// Process the URL and returns an status
	ProcessUrl(ctx context.Context, in *UrlRequest, opts ...grpc.CallOption) (*UrlResponse, error)
}

type crawlerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerServiceClient(cc grpc.ClientConnInterface) CrawlerServiceClient {
	return &crawlerServiceClient{cc}
}

func (c *crawlerServiceClient) ProcessUrl(ctx context.Context, in *UrlRequest, opts ...grpc.CallOption) (*UrlResponse, error) {
	out := new(UrlResponse)
	err := c.cc.Invoke(ctx, "/crawler.CrawlerService/ProcessUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CrawlerServiceServer is the server API for CrawlerService service.
// All implementations must embed UnimplementedCrawlerServiceServer
// for forward compatibility
type CrawlerServiceServer interface {
	// Process the URL and returns an status
	ProcessUrl(context.Context, *UrlRequest) (*UrlResponse, error)
	mustEmbedUnimplementedCrawlerServiceServer()
}

// UnimplementedCrawlerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCrawlerServiceServer struct {
}

func (UnimplementedCrawlerServiceServer) ProcessUrl(context.Context, *UrlRequest) (*UrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessUrl not implemented")
}
func (UnimplementedCrawlerServiceServer) mustEmbedUnimplementedCrawlerServiceServer() {}

// UnsafeCrawlerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerServiceServer will
// result in compilation errors.
type UnsafeCrawlerServiceServer interface {
	mustEmbedUnimplementedCrawlerServiceServer()
}

func RegisterCrawlerServiceServer(s grpc.ServiceRegistrar, srv CrawlerServiceServer) {
	s.RegisterService(&CrawlerService_ServiceDesc, srv)
}

func _CrawlerService_ProcessUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerServiceServer).ProcessUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crawler.CrawlerService/ProcessUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerServiceServer).ProcessUrl(ctx, req.(*UrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CrawlerService_ServiceDesc is the grpc.ServiceDesc for CrawlerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CrawlerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "crawler.CrawlerService",
	HandlerType: (*CrawlerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessUrl",
			Handler:    _CrawlerService_ProcessUrl_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "crawler.proto",
}
