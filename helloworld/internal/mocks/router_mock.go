package mocks

import (
	"context"
	"errors"
	"io"
	"sync"

	"google.golang.org/grpc"
	v1 "helloworld/api/helloworld/v1"
)

// MockRouterServiceClient 手写的 RouterServiceClient mock
type MockRouterServiceClient struct {
	mu              sync.Mutex
	JsonMessageFunc func(ctx context.Context, in *v1.JsonReq, opts ...grpc.CallOption) (*v1.JsonRsp, error)
	RpcCallFunc     func(ctx context.Context, in *v1.Req, opts ...grpc.CallOption) (*v1.Rsp, error)
	StreamCallFunc  func(ctx context.Context, opts ...grpc.CallOption) (v1.RouterService_StreamCallClient, error)
}

func (m *MockRouterServiceClient) JsonMessage(ctx context.Context, in *v1.JsonReq, opts ...grpc.CallOption) (*v1.JsonRsp, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.JsonMessageFunc != nil {
		return m.JsonMessageFunc(ctx, in, opts...)
	}
	return nil, errors.New("JsonMessageFunc not implemented")
}

func (m *MockRouterServiceClient) RpcCall(ctx context.Context, in *v1.Req, opts ...grpc.CallOption) (*v1.Rsp, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.RpcCallFunc != nil {
		return m.RpcCallFunc(ctx, in, opts...)
	}
	return nil, errors.New("RpcCallFunc not implemented")
}

func (m *MockRouterServiceClient) StreamCall(ctx context.Context, opts ...grpc.CallOption) (v1.RouterService_StreamCallClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.StreamCallFunc != nil {
		return m.StreamCallFunc(ctx, opts...)
	}
	return nil, errors.New("StreamCallFunc not implemented")
}

// MockRouterService_StreamCallClient 流式客户端的 mock
type MockRouterService_StreamCallClient struct {
	SendFunc      func(req *v1.StreamReq) error
	RecvFunc      func() (*v1.StreamRsp, error)
	CloseSendFunc func() error
	ContextFunc   func() context.Context
}

func (m *MockRouterService_StreamCallClient) Send(req *v1.StreamReq) error {
	if m.SendFunc != nil {
		return m.SendFunc(req)
	}
	return errors.New("SendFunc not implemented")
}

func (m *MockRouterService_StreamCallClient) Recv() (*v1.StreamRsp, error) {
	if m.RecvFunc != nil {
		return m.RecvFunc()
	}
	return nil, io.EOF
}

func (m *MockRouterService_StreamCallClient) CloseSend() error {
	if m.CloseSendFunc != nil {
		return m.CloseSendFunc()
	}
	return errors.New("CloseSendFunc not implemented")
}

func (m *MockRouterService_StreamCallClient) Context() context.Context {
	if m.ContextFunc != nil {
		return m.ContextFunc()
	}
	return context.Background()
}
