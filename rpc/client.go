package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/api/rpc"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

// rpcClient RPC代理客户端
type rpcClient struct {
	conn       *grpc.ClientConn
	client     rpc.RouterServiceClient
	stream     rpc.RouterService_StreamCallClient
	stateful   bool
	mu         sync.Mutex
	srcService string
	reqIdGen   func() string
	closed     bool
}

// NewRpcProxy 创建新的RPC代理
func NewRpcProxy(ctx context.Context, targetService string, srcServiceName string, stateful bool, r registry.Discovery, opts ...grpc.DialOption) (*rpcClient, error) {
	var conn *grpc.ClientConn
	var err error

	// 如果提供了服务发现，使用服务发现
	if r != nil {
		conn, err = kgrpc.DialInsecure(
			ctx,
			kgrpc.WithEndpoint("discovery:///"+targetService),
			kgrpc.WithDiscovery(r),
			kgrpc.WithTimeout(360*time.Second),
		)
	} else {
		// 否则直接连接目标地址
		conn, err = grpc.DialContext(ctx, targetService, opts...)
	}

	if err != nil {
		return nil, err
	}

	client := rpc.NewRouterServiceClient(conn)
	rc := &rpcClient{
		conn:       conn,
		client:     client,
		stateful:   stateful,
		srcService: srcServiceName,
		reqIdGen:   func() string { return uuid.New().String() },
		closed:     false,
	}

	// 如果是有状态服务，建立stream
	if stateful {
		stream, err := client.StreamCall(ctx)
		if err != nil {
			conn.Close()
			return nil, err
		}
		rc.stream = stream
	}

	return rc, nil
}

// NewRpcProxyWithDiscovery 使用服务发现创建RPC代理
func NewRpcProxyWithDiscovery(ctx context.Context, targetService string, serviceName string, stateful bool, r registry.Discovery) (*rpcClient, error) {
	return NewRpcProxy(ctx, targetService, serviceName, stateful, r)
}

// NewRpcProxyDirect 直接连接创建RPC代理
func NewRpcProxyDirect(ctx context.Context, targetAddress string, serviceName string, stateful bool, opts ...grpc.DialOption) (*rpcClient, error) {
	return NewRpcProxy(ctx, targetAddress, serviceName, stateful, nil, opts...)
}

// Call 执行RPC调用
func (c *rpcClient) Call(ctx context.Context, msgId string, msgContent RpcContent) (RpcContent, error) {
	if c.closed {
		return nil, errors.New(500, "PROXY_CLOSED", "RPC proxy is closed")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	contentType := msgContent.Type()
	switch contentType {
	case RpcContentBytes:
		return c.callBytes(ctx, msgId, msgContent)
	case RpcContentJson:
		return c.callJson(ctx, msgId, msgContent)
	default:
		return nil, errors.New(400, "UNSUPPORTED_CONTENT_TYPE", "unsupported content type")
	}
}

// callBytes 处理字节类型的调用
func (c *rpcClient) callBytes(ctx context.Context, msgId string, msgContent RpcContent) (RpcContent, error) {
	data, ok := msgContent.Data().([][]byte)
	if !ok {
		return nil, errors.New(400, "INVALID_DATA_TYPE", "expect [][]byte for bytes request")
	}

	if c.stateful {
		return c.callStream(ctx, msgId, data)
	} else {
		return c.callUnary(ctx, msgId, data)
	}
}

// callJson 处理JSON类型的调用
func (c *rpcClient) callJson(ctx context.Context, msgId string, msgContent RpcContent) (RpcContent, error) {
	data, ok := msgContent.Data().(string)
	if !ok {
		return nil, errors.New(400, "INVALID_DATA_TYPE", "expect string for json request")
	}

	rsp, err := c.client.JsonMessage(ctx, &rpc.JsonReq{
		SrcService: c.srcService,
		MsgId:      msgId,
		MsgContent: data,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("JSON response: %s", rsp.MsgContent)

	// 返回JSON格式的响应
	return &Content[string]{
		CType: RpcContentJson,
		Dt:    rsp.MsgContent,
	}, nil
}

// callStream 处理流式调用
func (c *rpcClient) callStream(ctx context.Context, msgId string, data [][]byte) (RpcContent, error) {
	reqId := c.reqIdGen()
	trace := &rpc.TraceInfo{
		Traceid: time.Now().UnixNano(),
		Spanid:  1,
		Sample:  true,
	}

	err := c.stream.Send(&rpc.StreamReq{
		SrcService: c.srcService,
		MsgId:      msgId,
		MsgContent: data,
		ReqId:      reqId,
		Traceinfo:  trace,
	})
	if err != nil {
		return nil, err
	}

	// 等待响应
	for {
		rsp, err := c.stream.Recv()
		if err != nil {
			return nil, err
		}
		if rsp.ReqId == reqId {
			// 返回字节格式的响应
			return &Content[[][]byte]{
				CType: RpcContentBytes,
				Dt:    rsp.MsgContent,
			}, nil
		}
		// 继续等待匹配的响应
	}
}

// callUnary 处理一元调用
func (c *rpcClient) callUnary(ctx context.Context, msgId string, data [][]byte) (RpcContent, error) {
	rsp, err := c.client.RpcCall(ctx, &rpc.Req{
		SrcService: c.srcService,
		MsgId:      msgId,
		MsgContent: data,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("Unary response: %v", rsp.MsgContent)

	// 返回字节格式的响应
	return &Content[[][]byte]{
		CType: RpcContentBytes,
		Dt:    rsp.MsgContent,
	}, nil
}

// Close 关闭连接
func (c *rpcClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.stream != nil {
		if err := c.stream.CloseSend(); err != nil {
			log.Errorf("Failed to close stream: %v", err)
		}
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// IsClosed 检查是否已关闭
func (c *rpcClient) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// GetService 获取服务名
func (c *rpcClient) GetService() string {
	return c.srcService
}

// IsStateful 检查是否为有状态服务
func (c *rpcClient) IsStateful() bool {
	return c.stateful
}
