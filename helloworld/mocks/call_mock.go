package mocks

import (
	"context"
	"google.golang.org/grpc"
	v1 "helloworld/api/helloworld/v1"
	"io"

	"github.com/golang/mock/gomock"
	"reflect"
)

// MockRouterServiceClient 是使用 gomock 生成的 mock，符合 RouterServiceClient 接口
type MockRouterServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockRouterServiceClientMockRecorder
}

type MockRouterServiceClientMockRecorder struct {
	mock *MockRouterServiceClient
}

func NewMockRouterServiceClient(ctrl *gomock.Controller) *MockRouterServiceClient {
	mock := &MockRouterServiceClient{ctrl: ctrl}
	mock.recorder = &MockRouterServiceClientMockRecorder{mock}
	return mock
}

func (m *MockRouterServiceClient) EXPECT() *MockRouterServiceClientMockRecorder {
	return m.recorder
}

// JsonMessage mock 方法
func (m *MockRouterServiceClient) JsonMessage(ctx context.Context, in *v1.JsonReq, opts ...grpc.CallOption) (*v1.JsonRsp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, o := range opts {
		varargs = append(varargs, o)
	}
	ret := m.ctrl.Call(m, "JsonMessage", varargs...)
	ret0, _ := ret[0].(*v1.JsonRsp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRouterServiceClientMockRecorder) JsonMessage(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	varargs = append(varargs, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JsonMessage", reflect.TypeOf((*MockRouterServiceClient)(nil).JsonMessage), varargs...)
}

// RpcCall mock 方法
func (m *MockRouterServiceClient) RpcCall(ctx context.Context, in *v1.Req, opts ...grpc.CallOption) (*v1.Rsp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, o := range opts {
		varargs = append(varargs, o)
	}
	ret := m.ctrl.Call(m, "RpcCall", varargs...)
	ret0, _ := ret[0].(*v1.Rsp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRouterServiceClientMockRecorder) RpcCall(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	varargs = append(varargs, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RpcCall", reflect.TypeOf((*MockRouterServiceClient)(nil).RpcCall), varargs...)
}

// StreamCall 简单手写 mock，返回自定义的 MockStreamClient
func (m *MockRouterServiceClient) StreamCall(ctx context.Context, opts ...grpc.CallOption) (v1.RouterService_StreamCallClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, o := range opts {
		varargs = append(varargs, o)
	}
	ret := m.ctrl.Call(m, "StreamCall", varargs...)
	ret0, _ := ret[0].(v1.RouterService_StreamCallClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRouterServiceClientMockRecorder) StreamCall(ctx interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	varargs = append(varargs, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamCall", reflect.TypeOf((*MockRouterServiceClient)(nil).StreamCall), varargs...)
}

// MockRouterService_StreamCallClient 是 StreamCall 返回的流式客户端的 mock
type MockRouterService_StreamCallClient struct {
	gomock.Controller
	grpc.ClientStream

	sendFunc      func(*v1.StreamReq) error
	recvFunc      func() (*v1.StreamRsp, error)
	closeSendFunc func() error
	contextFunc   func() context.Context
}

func NewMockRouterService_StreamCallClient(ctrl *gomock.Controller) *MockRouterService_StreamCallClient {
	return &MockRouterService_StreamCallClient{
		Controller: *ctrl,
	}
}

func (m *MockRouterService_StreamCallClient) Send(req *v1.StreamReq) error {
	if m.sendFunc != nil {
		return m.sendFunc(req)
	}
	return nil
}

func (m *MockRouterService_StreamCallClient) Recv() (*v1.StreamRsp, error) {
	if m.recvFunc != nil {
		return m.recvFunc()
	}
	return nil, io.EOF
}

func (m *MockRouterService_StreamCallClient) CloseSend() error {
	if m.closeSendFunc != nil {
		return m.closeSendFunc()
	}
	return nil
}

func (m *MockRouterService_StreamCallClient) Context() context.Context {
	if m.contextFunc != nil {
		return m.contextFunc()
	}
	return context.Background()
}
