package server

import (
	"crypto/tls"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultMaxMsgSize = 1024 * 1024 * 4
)

type grpcOption func(*grpcOptions)

type grpcOptions struct {
	ServerOptions      []grpc.ServerOption
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	TLSConfig          *tls.Config
}

func (o *grpcOptions) Apply(options ...grpcOption) {
	for _, option := range options {
		option(o)
	}
}

func WithGRPCServerOptions(options ...grpc.ServerOption) grpcOption {
	return func(o *grpcOptions) {
		o.ServerOptions = append(o.ServerOptions, options...)
	}
}

func WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) grpcOption {
	return func(o *grpcOptions) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, interceptor)
	}
}

func WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) grpcOption {
	return func(o *grpcOptions) {
		o.StreamInterceptors = append(o.StreamInterceptors, interceptor)
	}
}

func WithGRPCTLSConfig(tlsConfig *tls.Config) grpcOption {
	return func(o *grpcOptions) {
		o.TLSConfig = tlsConfig
	}
}

func defaultGRPCOptions() *grpcOptions {
	options := &grpcOptions{
		ServerOptions: []grpc.ServerOption{
			grpc.MaxRecvMsgSize(defaultMaxMsgSize),
			grpc.MaxSendMsgSize(defaultMaxMsgSize),
		},
	}
	return options
}

func NewGRPCServer(options ...grpcOption) *grpc.Server {
	opts := defaultGRPCOptions()
	opts.Apply(options...)

	return newGRPCServer(opts)
}

func newGRPCServer(options *grpcOptions) *grpc.Server {
	if options.TLSConfig != nil {
		options.ServerOptions = append(options.ServerOptions, grpc.Creds(credentials.NewTLS(options.TLSConfig)))
	}
	options.ServerOptions = append(options.ServerOptions,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(options.UnaryInterceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(options.StreamInterceptors...)),
	)
	srv := grpc.NewServer(options.ServerOptions...)
	return srv
}
