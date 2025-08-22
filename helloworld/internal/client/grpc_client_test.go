package client

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	v1 "helloworld/api/helloworld/v1"
	"helloworld/mocks"
	"sync"
	"testing"
	"time"
)

// 模拟回调实现
type testCallback struct {
	mu            sync.Mutex
	successCalled bool
	errorCalled   bool
	resp          any
	err           error
}

func (t *testCallback) OnSuccess(funName string, rsp any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.successCalled = true
	t.resp = rsp
}

func (t *testCallback) OnError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorCalled = true
	t.err = err
}

func TestGRPCClient_AsyncCall_JsonMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建 mock
	mockClient := mocks.NewMockRouterServiceClient(ctrl)

	// 设置预期调用
	mockClient.EXPECT().
		JsonMessage(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *v1.JsonReq, _ ...grpc.CallOption) (*v1.JsonRsp, error) {
			return &v1.JsonRsp{
				MsgId:      req.MsgId,
				MsgContent: "mocked",
			}, nil
		}).Times(1)

	// 构造客户端
	client := &GRPCClient{
		client:     mockClient,
		timeoutMap: map[string]time.Duration{"mock-service": time.Second},
		targetName: "mock-service",
	}

	cb := &testCallback{}

	content := &RpcContent{
		Type: ContentTypeJSON,
		JSON: `{"key":"value"}`,
	}

	client.AsyncCall(context.Background(), "JsonMessage", content, cb)

	// 等待 goroutine 结束
	time.Sleep(100000 * time.Millisecond)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	assert.True(t, cb.successCalled, "OnSuccess should be called")
	assert.False(t, cb.errorCalled, "OnError should not be called")

	resp, ok := cb.resp.(*v1.JsonRsp)
	assert.True(t, ok)
	assert.Equal(t, "JsonMessage", resp.MsgId)
	assert.Equal(t, "mocked", resp.MsgContent)
	log.Info("Test passed")
}
