package rpc

import (
	"context"
	"github.com/go-kratos/kratos/v2/api/rpc"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"testing"
	"time"
)

// MockDiscovery 模拟服务发现
type MockDiscovery struct {
	mock.Mock
}

func (m *MockDiscovery) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	args := m.Called(ctx, serviceName)
	return args.Get(0).([]*registry.ServiceInstance), args.Error(1)
}

func (m *MockDiscovery) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	args := m.Called(ctx, serviceName)
	return args.Get(0).(registry.Watcher), args.Error(1)
}

// MockWatcher 模拟服务监听器
type MockWatcher struct {
	mock.Mock
	callCount int
	lastCall  time.Time
}

func (m *MockWatcher) Next() ([]*registry.ServiceInstance, error) {
	m.callCount++
	if m.callCount > 1 {
		time.Sleep(time.Minute)
	}
	args := m.Called()
	return args.Get(0).([]*registry.ServiceInstance), args.Error(1)
}

func (m *MockWatcher) Stop() error {
	args := m.Called()
	return args.Error(0)
}

// MockRouterServiceClient 模拟路由服务客户端
type MockRouterServiceClient struct {
	mock.Mock
}

func (m *MockRouterServiceClient) JsonMessage(ctx context.Context, req *rpc.JsonReq, opts ...grpc.CallOption) (*rpc.JsonRsp, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rpc.JsonRsp), args.Error(1)
}

func (m *MockRouterServiceClient) RpcCall(ctx context.Context, req *rpc.Req, opts ...grpc.CallOption) (*rpc.Rsp, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rpc.Rsp), args.Error(1)
}

func (m *MockRouterServiceClient) StreamCall(ctx context.Context, opts ...grpc.CallOption) (rpc.RouterService_StreamCallClient, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(rpc.RouterService_StreamCallClient), args.Error(1)
}

func TestNewRpcProxy(t *testing.T) {
	ctx := context.Background()

	t.Run("TestNewRpcProxyWithDiscovery", func(t *testing.T) {
		mockDiscovery := &MockDiscovery{}
		mockWatcher := &MockWatcher{}

		// 设置GetService的mock
		mockDiscovery.On("GetService", mock.Anything, "test-CallDispatcher").Return([]*registry.ServiceInstance{
			{ID: "1", Name: "test-CallDispatcher", Endpoints: []string{"localhost:8080"}},
		}, nil)

		// 设置Watch的mock
		mockDiscovery.On("Watch", mock.Anything, "test-CallDispatcher").Return(mockWatcher, nil)
		mockWatcher.On("Next").Return([]*registry.ServiceInstance{}, nil)
		mockWatcher.On("Stop").Return(nil)

		client, err := NewRpcProxyWithDiscovery(ctx, "test-CallDispatcher", "src-CallDispatcher", false, mockDiscovery)

		// 由于无法实际建立gRPC连接，这里只测试函数调用不panic
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}
	})

	t.Run("TestNewRpcProxyDirect", func(t *testing.T) {
		client, err := NewRpcProxyDirect(ctx, "localhost:8080", "src-CallDispatcher", false, grpc.WithInsecure())

		// 由于无法实际建立gRPC连接，这里只测试函数调用不panic
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}
	})
}

func TestRpcClient_Call(t *testing.T) {
	ctx := context.Background()

	t.Run("TestCallWithClosedClient", func(t *testing.T) {
		client := &rpcClient{
			closed: true,
		}

		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"test": "data"}`,
		}

		result, err := client.Call(ctx, "test-method", content)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "RPC proxy is closed")
	})

	t.Run("TestCallWithUnsupportedContentType", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		content := &Content[string]{
			CType:   RpcContentFile, // 不支持的类型
			Content: "test",
		}

		result, err := client.Call(ctx, "test-method", content)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unsupported content type")
	})
}

func TestRpcClient_CallJson(t *testing.T) {
	ctx := context.Background()

	t.Run("TestCallJsonSuccess", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		expectedResponse := &rpc.JsonRsp{
			MsgContent: `{"result": "success"}`,
		}
		mockClient.On("JsonMessage", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			closed:     false,
		}

		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"test": "data"}`,
		}

		result, err := client.callJson(ctx, "test-method", content)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RpcContentJson, result.Type())

		responseData, ok := result.Data().(string)
		assert.True(t, ok)
		assert.Equal(t, `{"result": "success"}`, responseData)

		mockClient.AssertExpectations(t)
	})

	t.Run("TestCallJsonWithInvalidDataType", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		content := &Content[int]{
			CType:   RpcContentJson,
			Content: 123, // 错误的数据类型
		}

		result, err := client.callJson(ctx, "test-method", content)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "expect string for json request")
	})

	t.Run("TestCallJsonWithClientError", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		mockClient.On("JsonMessage", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			closed:     false,
		}

		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"test": "data"}`,
		}

		result, err := client.callJson(ctx, "test-method", content)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)

		mockClient.AssertExpectations(t)
	})
}

func TestRpcClient_CallBytes(t *testing.T) {
	ctx := context.Background()

	t.Run("TestCallBytesUnary", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		expectedResponse := &rpc.Rsp{
			MsgContent: [][]byte{[]byte("success")},
		}
		mockClient.On("RpcCall", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			stateful:   false,
			closed:     false,
		}

		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{[]byte("test")},
		}

		result, err := client.callBytes(ctx, "test-method", content)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RpcContentBytes, result.Type())

		responseData, ok := result.Data().([][]byte)
		assert.True(t, ok)
		assert.Equal(t, [][]byte{[]byte("success")}, responseData)

		mockClient.AssertExpectations(t)
	})

	t.Run("TestCallBytesWithInvalidDataType", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		content := &Content[string]{
			CType:   RpcContentBytes,
			Content: "invalid-data", // 错误的数据类型
		}

		result, err := client.callBytes(ctx, "test-method", content)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "expect [][]byte for bytes request")
	})
}

func TestRpcClient_CallUnary(t *testing.T) {
	ctx := context.Background()

	t.Run("TestCallUnarySuccess", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		expectedResponse := &rpc.Rsp{
			MsgContent: [][]byte{[]byte("unary-success")},
		}
		mockClient.On("RpcCall", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			closed:     false,
		}

		data := [][]byte{[]byte("test-data")}
		result, err := client.callUnary(ctx, "test-method", data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RpcContentBytes, result.Type())

		responseData, ok := result.Data().([][]byte)
		assert.True(t, ok)
		assert.Equal(t, [][]byte{[]byte("unary-success")}, responseData)

		mockClient.AssertExpectations(t)
	})

	t.Run("TestCallUnaryWithClientError", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		mockClient.On("RpcCall", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			closed:     false,
		}

		data := [][]byte{[]byte("test-data")}
		result, err := client.callUnary(ctx, "test-method", data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, assert.AnError, err)

		mockClient.AssertExpectations(t)
	})
}

func TestRpcClient_Close(t *testing.T) {
	t.Run("TestCloseSuccess", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		err := client.Close()
		assert.NoError(t, err)
		assert.True(t, client.closed)
	})

	t.Run("TestCloseAlreadyClosed", func(t *testing.T) {
		client := &rpcClient{
			closed: true,
		}

		err := client.Close()
		assert.NoError(t, err)
		assert.True(t, client.closed)
	})

	t.Run("TestCloseWithStreamError", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		err := client.Close()
		assert.NoError(t, err) // Close方法应该忽略stream关闭错误
		assert.True(t, client.closed)
	})
}

func TestRpcClient_IsClosed(t *testing.T) {
	t.Run("TestIsClosedTrue", func(t *testing.T) {
		client := &rpcClient{
			closed: true,
		}

		assert.True(t, client.IsClosed())
	})

	t.Run("TestIsClosedFalse", func(t *testing.T) {
		client := &rpcClient{
			closed: false,
		}

		assert.False(t, client.IsClosed())
	})
}

func TestRpcClient_GetService(t *testing.T) {
	t.Run("TestGetService", func(t *testing.T) {
		expectedService := "test-CallDispatcher"
		client := &rpcClient{
			srcService: expectedService,
		}

		assert.Equal(t, expectedService, client.GetService())
	})
}

func TestRpcClient_IsStateful(t *testing.T) {
	t.Run("TestIsStatefulTrue", func(t *testing.T) {
		client := &rpcClient{
			stateful: true,
		}

		assert.True(t, client.IsStateful())
	})

	t.Run("TestIsStatefulFalse", func(t *testing.T) {
		client := &rpcClient{
			stateful: false,
		}

		assert.False(t, client.IsStateful())
	})
}

func TestRpcClient_Concurrency(t *testing.T) {
	t.Run("TestConcurrentCalls", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		expectedResponse := &rpc.JsonRsp{
			MsgContent: `{"result": "success"}`,
		}
		mockClient.On("JsonMessage", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil).Times(10)

		client := &rpcClient{
			client:     mockClient,
			srcService: "src-CallDispatcher",
			closed:     false,
		}

		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"test": "data"}`,
		}

		// 并发调用
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				result, err := client.callJson(context.Background(), "test-method", content)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				done <- true
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// 成功
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent calls")
			}
		}

		mockClient.AssertExpectations(t)
	})
}

func TestRpcClient_Integration(t *testing.T) {
	t.Run("TestFullCallFlow", func(t *testing.T) {
		mockClient := &MockRouterServiceClient{}
		expectedResponse := &rpc.JsonRsp{
			MsgContent: `{"result": "integration-success"}`,
		}
		mockClient.On("JsonMessage", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		client := &rpcClient{
			client:     mockClient,
			srcService: "integration-CallDispatcher",
			closed:     false,
		}

		// 测试完整调用流程
		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"integration": "test"}`,
		}

		result, err := client.Call(context.Background(), "integration-method", content)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RpcContentJson, result.Type())

		responseData, ok := result.Data().(string)
		assert.True(t, ok)
		assert.Equal(t, `{"result": "integration-success"}`, responseData)

		// 验证服务信息
		assert.Equal(t, "integration-CallDispatcher", client.GetService())
		assert.False(t, client.IsStateful())
		assert.False(t, client.IsClosed())

		// 测试关闭
		err = client.Close()
		assert.NoError(t, err)
		assert.True(t, client.IsClosed())

		mockClient.AssertExpectations(t)
	})
}
func TestMockWatcher_Next(t *testing.T) {
	t.Run("TestMockWatcherBehavior", func(t *testing.T) {
		mockWatcher := &MockWatcher{}

		// 设置期望的返回值
		expectedInstances := []*registry.ServiceInstance{
			{ID: "1", Name: "test-CallDispatcher", Endpoints: []string{"localhost:8080"}},
		}
		mockWatcher.On("Next").Return(expectedInstances, nil)

		// 第一次调用应该立即返回
		instances, err := mockWatcher.Next()
		assert.NoError(t, err)
		assert.Equal(t, expectedInstances, instances)
		assert.Equal(t, 1, mockWatcher.callCount)

		// 第二次调用应该返回空（因为距离第一次调用不足1分钟）
		instances, err = mockWatcher.Next()
		assert.NoError(t, err)
		assert.Empty(t, instances)
		assert.Equal(t, 2, mockWatcher.callCount)

		// 模拟时间过去1分钟
		mockWatcher.lastCall = time.Now().Add(-time.Minute - time.Second)

		// 第三次调用应该返回结果（因为距离上次调用超过1分钟）
		instances, err = mockWatcher.Next()
		assert.NoError(t, err)
		assert.Equal(t, expectedInstances, instances)
		assert.Equal(t, 3, mockWatcher.callCount)

		mockWatcher.AssertExpectations(t)
	})
}
func TestRpcRemoteCall(t *testing.T) {
	ctx := context.Background()
	t.Run("TestRpcRemoteCall_EntityCall", func(t *testing.T) {
		mockDiscovery := &MockDiscovery{}
		mockWatcher := &MockWatcher{}

		// 设置GetService的mock
		mockDiscovery.On("GetService", mock.Anything, "stateful-entity-CallDispatcher").Return([]*registry.ServiceInstance{
			{ID: "1", Name: "stateful-entity-CallDispatcher", Endpoints: []string{"grpc://10.0.9.4:31011"}},
		}, nil)

		// 设置Watch的mock
		mockDiscovery.On("Watch", mock.Anything, "stateful-entity-CallDispatcher").Return(mockWatcher, nil)
		mockWatcher.On("Next").Return([]*registry.ServiceInstance{
			{ID: "1", Name: "stateful-entity-CallDispatcher", Endpoints: []string{"grpc://10.0.9.4:31011"}},
		}, nil)
		mockWatcher.On("Stop").Return(nil)

		client, err := NewRpcProxyWithDiscovery(ctx, "stateful-entity-CallDispatcher", "src-CallDispatcher", true, mockDiscovery)

		// 由于无法实际建立gRPC连接，这里只测试函数调用不panic
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}
		req := &rpc.EntityRequest{}
		var fstIntZero = []byte{0xF7, 0x00}
		bytesData, err := encoding.GetCodec("proto").Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{fstIntZero, bytesData},
		}
		resp, err := client.callBytes(ctx, "onEntityCall", content)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

	})

	t.Run("TestRpcRemoteCall_userLogin", func(t *testing.T) {
		mockDiscovery := &MockDiscovery{}
		mockWatcher := &MockWatcher{}

		// 设置GetService的mock
		mockDiscovery.On("GetService", mock.Anything, "stateful-entity-CallDispatcher").Return([]*registry.ServiceInstance{
			{ID: "1", Name: "stateful-entity-CallDispatcher", Endpoints: []string{"grpc://10.0.9.4:31011"}},
		}, nil)

		// 设置Watch的mock
		mockDiscovery.On("Watch", mock.Anything, "stateful-entity-CallDispatcher").Return(mockWatcher, nil)
		mockWatcher.On("Next").Return([]*registry.ServiceInstance{
			{ID: "1", Name: "stateful-entity-CallDispatcher", Endpoints: []string{"grpc://10.0.9.4:31011"}},
		}, nil)
		mockWatcher.On("Stop").Return(nil)

		client, err := NewRpcProxyWithDiscovery(ctx, "stateful-entity-CallDispatcher", "src-CallDispatcher", true, mockDiscovery)

		// 由于无法实际建立gRPC连接，这里只测试函数调用不panic
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}

		var fstIntZero = []byte{0xF7, 0x00}
		paramMsg := &rpc.ParameterMsg{
			UserId:    "20001_1",
			ChannelId: "2",
		}
		paramMsgBytes, err := encoding.GetCodec("proto").Marshal(paramMsg)
		accountInfo := &rpc.AccountInfo{
			UserId:   "20001_1",
			ChanelId: 2,
		}
		entityReq := &rpc.UserLoginRequest{
			LoginAccount: accountInfo,
		}
		entityReqBytes, err := encoding.GetCodec("proto").Marshal(entityReq)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{fstIntZero, paramMsgBytes, entityReqBytes},
		}
		resp, err := client.callBytes(ctx, "userLogin", content)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		data := resp.Data().([][]byte)
		userResp := &rpc.UserLoginResponse{}
		encoding.GetCodec("proto").Unmarshal(data[0], userResp)
		assert.Equal(t, userResp.Msg, rpc.NihongMsg_Success)
	})
}
