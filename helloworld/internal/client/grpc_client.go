package client

import (
	"context"
	"errors"
	"sync"
	"time"

	v1 "helloworld/api/helloworld/v1"

	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
)

// RpcContentType 表示请求内容类型
type RpcContentType int

const (
	ContentTypeJSON RpcContentType = iota
	ContentTypeBytes
)

// RpcContent 是通用的请求结构
type RpcContent struct {
	Type RpcContentType
	JSON string
	Data [][]byte
}

// RpcCallback 是异步回调接口
type RpcCallback interface {
	OnSuccess(funName string, rsp any)
	OnError(err error)
}

// StreamCallback 流式回调接口
type StreamCallback interface {
	OnSend(stream v1.RouterService_StreamCallClient) error
	OnReceive(resp *v1.StreamRsp) error
	OnError(err error)
	OnClose()
}

// GRPCClient 通用 gRPC 客户端封装
type GRPCClient struct {
	conn       *grpc.ClientConn
	client     v1.RouterServiceClient
	targetName string
	timeoutMap map[string]time.Duration
	lock       sync.Mutex
}

// NewGRPCClient 创建一个新客户端
func NewGRPCClient(r registry.Discovery, serviceName string, timeoutCfg map[string]time.Duration) (*GRPCClient, error) {
	conn, err := kgrpc.DialInsecure(
		context.Background(),
		kgrpc.WithEndpoint("discovery:///"+serviceName),
		kgrpc.WithDiscovery(r),
	)
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:       conn,
		client:     v1.NewRouterServiceClient(conn),
		targetName: serviceName,
		timeoutMap: timeoutCfg,
	}, nil
}

// AsyncCall 通用异步调用方法（支持 JSON/Bytes）
func (c *GRPCClient) AsyncCall(ctx context.Context, funName string, content *RpcContent, cb RpcCallback) {
	switch content.Type {
	case ContentTypeJSON:
		go c.jsonCall(ctx, funName, content.JSON, cb)
	case ContentTypeBytes:
		go c.bytesCall(ctx, funName, content.Data, cb)
	default:
		cb.OnError(errors.New("unsupported content type"))
	}
}

func (c *GRPCClient) jsonCall(ctx context.Context, funName, jsonData string, cb RpcCallback) {
	req := &v1.JsonReq{
		SrcService: "client",
		MsgId:      funName,
		MsgContent: jsonData,
	}
	timeout := c.getTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rsp, err := c.client.JsonMessage(ctx, req)
	if err != nil {
		cb.OnError(err)
		return
	}
	cb.OnSuccess(funName, rsp)
}

func (c *GRPCClient) bytesCall(ctx context.Context, funName string, data [][]byte, cb RpcCallback) {
	req := &v1.Req{
		SrcService: "client",
		MsgId:      funName,
	}
	for _, d := range data {
		req.MsgContent = append(req.MsgContent, d)
	}
	timeout := c.getTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rsp, err := c.client.RpcCall(ctx, req)
	if err != nil {
		cb.OnError(err)
		return
	}
	cb.OnSuccess(funName, rsp)
}

// StreamCall 流式调用，支持回调处理
func (c *GRPCClient) StreamCall(ctx context.Context, callback StreamCallback) error {
	stream, err := c.client.StreamCall(ctx)
	if err != nil {
		callback.OnError(err)
		return err
	}

	// 启动接收协程
	recvDone := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				recvDone <- err
				return
			}
			if err := callback.OnReceive(resp); err != nil {
				recvDone <- err
				return
			}
		}
	}()

	// 执行发送逻辑
	if err := callback.OnSend(stream); err != nil {
		callback.OnError(err)
		_ = stream.CloseSend()
		return err
	}

	// 等待接收结束或出错
	err = <-recvDone
	if !errors.Is(err, context.Canceled) {
		callback.OnError(err)
	}
	callback.OnClose()
	return err
}

// getTimeout 获取目标服务的超时时间
func (c *GRPCClient) getTimeout() time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()
	if t, ok := c.timeoutMap[c.targetName]; ok {
		return t
	}
	return 10 * time.Second
}
