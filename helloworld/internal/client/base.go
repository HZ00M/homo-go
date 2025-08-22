package client

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
)

type BaseGRPCClient struct {
	conn   *grpc.ClientConn
	Logger *log.Helper
}

type ClientConfig struct {
	ServiceName string
	Timeout     time.Duration
}

// NewBaseGRPCClient 创建一个通用 gRPC 客户端连接
func NewBaseGRPCClient(ctx context.Context, discovery registry.Discovery, cfg ClientConfig, logger log.Logger) (*BaseGRPCClient, error) {
	conn, err := kgrpc.Dial(
		ctx,
		kgrpc.WithEndpoint("discovery:///"+cfg.ServiceName),
		kgrpc.WithDiscovery(discovery),
		kgrpc.WithTimeout(cfg.Timeout),
		kgrpc.WithMiddleware(), // 可按需添加 tracing、logging
	)
	if err != nil {
		return nil, err
	}

	return &BaseGRPCClient{
		conn:   conn,
		Logger: log.NewHelper(logger),
	}, nil
}

func (b *BaseGRPCClient) Conn() *grpc.ClientConn {
	return b.conn
}

func (b *BaseGRPCClient) Close() error {
	return b.conn.Close()
}
