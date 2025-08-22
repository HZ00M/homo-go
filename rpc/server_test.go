package rpc

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/api/rpc"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/vmihailenco/msgpack/v5"
)

// TestService 测试服务
type TestService struct {
	name string
}

func NewTestService(name string) *TestService {
	return &TestService{name: name}
}

func (s *TestService) Hello(name string) (string, error) {
	return "Hello, " + name + " from " + s.name, nil
}

func (s *TestService) Add(a, b int) (int, error) {
	return a + b, nil
}

func (s *TestService) ProcessData(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["processed"] = true
	result["timestamp"] = time.Now().Unix()
	result["data"] = data
	result["srcService"] = s.name
	return result, nil
}

// ProtoService 用于测试ProtoPacker的服务
type ProtoService struct{}

func (s *ProtoService) ProtoMethod(req *rpc.TestRequest) (*rpc.TestResponse, error) {
	return &rpc.TestResponse{
		Success: true,
		Message: "Processed: " + req.Name,
	}, nil
}

// BasicTypeService 基础类型服务（用于测试MessagePacker）
type BasicTypeService struct{}

func (s *BasicTypeService) BasicMethod(name string, age int) (string, error) {
	return "Basic: " + name + " age " + string(rune(age)), nil
}

// JsonStructService JSON结构体服务（用于测试FastJsonPacker）
type JsonStructService struct{}

type Person struct {
	Name string `json:"Name"`
	Age  int    `json:"age"`
}

func (s *JsonStructService) JsonMethod(person Person) (Person, error) {
	person.Age++
	return person, nil
}

// 添加一个没有错误返回值的方法来测试FastJsonPacker
func (s *JsonStructService) JsonMethodNoError(person Person) Person {
	person.Age++
	return person
}

// MessagePackService MessagePack服务（用于测试MessagePacker）
type MessagePackService struct{}

func (s *MessagePackService) ProcessMessagePack(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["success"] = true
	result["processed"] = data
	result["timestamp"] = time.Now().Unix()
	return result, nil
}

func TestRpcServer(t *testing.T) {
	// 创建RPC服务器
	rpcServer := NewRpcServer(grpcType)

	// 创建测试服务
	testService := NewTestService("test-CallDispatcher")

	// 使用新的Register方法注册服务
	err := rpcServer.Register(testService)
	if err != nil {
		t.Fatalf("Failed to register CallDispatcher: %v", err)
	}

	// 测试Hello方法
	t.Run("TestHello", func(t *testing.T) {
		ctx := context.Background()

		// 准备测试数据 - 基础类型使用MessagePacker，应该分别序列化每个参数
		param0Bytes, err := msgpack.Marshal("World")
		if err != nil {
			t.Errorf("Failed to marshal param0: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes},
		}

		// 调用RPC服务器
		result, err := rpcServer.OnCall(ctx, "client-CallDispatcher", "Hello", content)
		if err != nil {
			t.Errorf("RPC call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// MessagePacker返回RpcContentBytes类型
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 使用msgpack解析数据 - 返回的是分别序列化的字节数组
		var response string
		if err := msgpack.Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to parse response: %v", err)
			return
		}

		if response != "Hello, World from test-CallDispatcher" {
			t.Errorf("Expected 'Hello, World from test-CallDispatcher', got %v", response)
		}
	})

	// 测试Add方法
	t.Run("TestAdd", func(t *testing.T) {
		ctx := context.Background()

		// 准备测试数据 - 基础类型使用MessagePacker，应该分别序列化每个参数
		param0Bytes, err := msgpack.Marshal(10)
		if err != nil {
			t.Errorf("Failed to marshal param0: %v", err)
		}
		param1Bytes, err := msgpack.Marshal(20)
		if err != nil {
			t.Errorf("Failed to marshal param1: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes, param1Bytes},
		}

		// 调用RPC服务器
		result, err := rpcServer.OnCall(ctx, "client-CallDispatcher", "Add", content)
		if err != nil {
			t.Errorf("RPC call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// MessagePacker返回RpcContentBytes类型
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 使用msgpack解析数据 - 返回的是分别序列化的字节数组
		var response interface{}
		if err := msgpack.Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to parse response: %v", err)
			return
		}

		// 检查返回值 - msgpack 可能返回不同的整数类型
		expectedValue := 30
		actualValue := response

		// 转换为 int 进行比较
		var actualInt int
		switch v := actualValue.(type) {
		case int:
			actualInt = v
		case int8:
			actualInt = int(v)
		case int16:
			actualInt = int(v)
		case int32:
			actualInt = int(v)
		case int64:
			actualInt = int(v)
		case uint:
			actualInt = int(v)
		case uint8:
			actualInt = int(v)
		case uint16:
			actualInt = int(v)
		case uint32:
			actualInt = int(v)
		case uint64:
			actualInt = int(v)
		case float32:
			actualInt = int(v)
		case float64:
			actualInt = int(v)
		default:
			t.Errorf("Unexpected type for return value: %T, value: %v", actualValue, actualValue)
			return
		}

		if actualInt != expectedValue {
			t.Errorf("Expected %d, got %d (original type: %T)", expectedValue, actualInt, actualValue)
		}
	})

	// 测试ProcessData方法
	t.Run("TestProcessData", func(t *testing.T) {
		ctx := context.Background()

		// 准备测试数据 - 基础类型使用MessagePacker，应该分别序列化每个参数
		param0Data := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		param0Bytes, err := msgpack.Marshal(param0Data)
		if err != nil {
			t.Errorf("Failed to marshal param0: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes},
		}

		// 调用RPC服务器
		result, err := rpcServer.OnCall(ctx, "client-CallDispatcher", "ProcessData", content)
		if err != nil {
			t.Errorf("RPC call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// MessagePacker返回RpcContentBytes类型
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 使用msgpack解析数据 - 返回的是分别序列化的字节数组
		var response map[string]interface{}
		if err := msgpack.Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to parse response: %v", err)
			return
		}

		// 检查返回值
		if response == nil {
			t.Error("Expected processed data, got nil")
			return
		}

		if response["processed"] != true {
			t.Error("Expected processed=true")
		}
		if response["srcService"] != "test-CallDispatcher" {
			t.Errorf("Expected srcService=test-CallDispatcher, got %v", response["srcService"])
		}
	})
}

func TestContentInfo(t *testing.T) {
	// 测试JSON内容
	t.Run("TestJsonContent", func(t *testing.T) {
		content := &Content[string]{
			CType:   RpcContentJson,
			Content: `{"key": "value"}`,
		}

		if content.Type() != RpcContentJson {
			t.Errorf("Expected JSON content type, got %v", content.Type())
		}

		data := content.Data().(string)
		if data != `{"key": "value"}` {
			t.Errorf("Expected '{\"key\": \"value\"}', got %s", data)
		}
	})

	// 测试字节内容
	t.Run("TestBytesContent", func(t *testing.T) {
		bytesData := [][]byte{
			[]byte("Hello"),
			[]byte("World"),
		}

		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: bytesData,
		}

		if content.Type() != RpcContentBytes {
			t.Errorf("Expected bytes content type, got %v", content.Type())
		}

		data := content.Data().([][]byte)
		if len(data) != 2 {
			t.Errorf("Expected 2 bytes arrays, got %d", len(data))
		}
		if string(data[0]) != "Hello" {
			t.Errorf("Expected 'Hello', got %s", string(data[0]))
		}
		if string(data[1]) != "World" {
			t.Errorf("Expected 'World', got %s", string(data[1]))
		}
	})
}

// TestPackerSelection 测试打包器选择逻辑
func TestPackerSelection(t *testing.T) {
	rpcServer := NewRpcServer(grpcType)

	// 测试基础类型服务（应该使用MessagePacker）
	t.Run("TestBasicTypes", func(t *testing.T) {
		basicService := &BasicTypeService{}
		err := rpcServer.Register(basicService)
		if err != nil {
			t.Fatalf("Failed to register basic CallDispatcher: %v", err)
		}

		// 测试基础类型方法
		ctx := context.Background()
		param0Bytes, err := msgpack.Marshal("test")
		if err != nil {
			t.Errorf("Failed to marshal param0: %v", err)
		}
		param1Bytes, err := msgpack.Marshal(123)
		if err != nil {
			t.Errorf("Failed to marshal param1: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes, param1Bytes},
		}

		result, err := rpcServer.OnCall(ctx, "client", "BasicMethod", content)
		if err != nil {
			t.Errorf("Basic Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
		}
	})
	// 测试MessagePacker打包器
	t.Run("TestMessagePacker", func(t *testing.T) {
		// 创建一个使用MessagePacker的服务
		msgPackService := &MessagePackService{}
		err := rpcServer.Register(msgPackService)
		if err != nil {
			t.Fatalf("Failed to register MessagePack CallDispatcher: %v", err)
		}

		// 测试MessagePacker方法
		ctx := context.Background()
		// 准备测试数据
		param0Data := map[string]interface{}{
			"id":     123,
			"name":   "MessagePack Test",
			"active": true,
		}

		// 使用MessagePack格式调用
		param0Bytes, err := msgpack.Marshal(param0Data)
		if err != nil {
			t.Fatalf("Failed to marshal msgpack data: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes},
		}

		result, err := rpcServer.OnCall(ctx, "client", "ProcessMessagePack", content)
		if err != nil {
			t.Errorf("MessagePack Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// 验证返回类型是字节
		if result.Type() != RpcContentBytes {
			t.Errorf("Expected bytes content type, got %v", result.Type())
			return
		}

		// 解析返回的MessagePack数据
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 反序列化MessagePack数据 - 返回的是分别序列化的字节数组
		var response map[string]interface{}
		if err := msgpack.Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to unmarshal MessagePack response: %v", err)
			return
		}

		// 验证返回结果
		if response["success"] != true {
			t.Errorf("Expected success=true, got %v (type: %T)", response["success"], response["success"])
			t.Logf("Full response: %+v", response)
		}
	})
	t.Run("TestJsonStructs", func(t *testing.T) {
		jsonService := &JsonStructService{}
		err := rpcServer.Register(jsonService)
		if err != nil {
			t.Fatalf("Failed to register JSON CallDispatcher: %v", err)
		}

		// 测试JSON结构体方法 - 使用JSON格式数据
		ctx := context.Background()
		jsonData := `{
			"Name": "test",
			"age": 25
		}`
		content := &Content[string]{
			CType:   RpcContentJson,
			Content: jsonData,
		}

		result, err := rpcServer.OnCall(ctx, "client", "JsonMethod", content)
		if err != nil {
			t.Errorf("JSON Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
		}
		log.Info("Result: ", result.Data())
	})
}

// TestPackerComparison 同时测试三种打包器
func TestPackerComparison(t *testing.T) {
	rpcServer := NewRpcServer(grpcType)

	// 创建测试服务
	protoService := &ProtoService{}
	jsonService := &JsonStructService{}
	msgpackService := &MessagePackService{}

	// 注册服务
	if err := rpcServer.Register(protoService); err != nil {
		t.Fatalf("Failed to register proto CallDispatcher: %v", err)
	}
	if err := rpcServer.Register(jsonService); err != nil {
		t.Fatalf("Failed to register json CallDispatcher: %v", err)
	}
	if err := rpcServer.Register(msgpackService); err != nil {
		t.Fatalf("Failed to register msgpack CallDispatcher: %v", err)
	}

	// 测试ProtoPacker
	t.Run("TestProtoPacker", func(t *testing.T) {
		ctx := context.Background()
		// 创建Proto请求
		req := &rpc.TestRequest{
			Id:   123,
			Name: "Proto Test",
		}

		bytesData, err := encoding.GetCodec("proto").Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{bytesData},
		}

		result, err := rpcServer.OnCall(ctx, "client", "ProtoMethod", content)
		if err != nil {
			t.Errorf("Proto Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// 验证返回类型
		if result.Type() != RpcContentBytes {
			t.Errorf("Expected bytes content type, got %v", result.Type())
			return
		}

		// 解析返回数据
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 使用 msgpack 反序列化响应
		var response rpc.TestResponse
		if err := encoding.GetCodec("proto").Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to unmarshal proto response: %v", err)
			return
		}

		// 验证结果
		if response.Success != true {
			t.Error("Expected success=true")
		}
		if response.Message != "Processed: Proto Test" {
			t.Errorf("Expected message 'Processed: Proto Test', got '%s'", response.Message)
		}
	})

	// 测试FastJsonPacker
	t.Run("TestJsonPacker", func(t *testing.T) {
		jsonService := &JsonStructService{}
		err := rpcServer.Register(jsonService)
		if err != nil {
			t.Fatalf("Failed to register JSON CallDispatcher: %v", err)
		}

		// 测试没有错误返回值的JSON结构体方法 - 应该使用FastJsonPacker
		ctx := context.Background()
		jsonData := `{
			"Name": "FastJson Test",
			"age": 30
		}`
		content := &Content[string]{
			CType:   RpcContentJson,
			Content: jsonData,
		}

		result, err := rpcServer.OnCall(ctx, "client", "JsonMethodNoError", content)
		if err != nil {
			t.Errorf("FastJson Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// FastJsonPacker返回RpcContentJson类型
		if result.Type() != RpcContentJson {
			t.Errorf("Expected JSON content type, got %v", result.Type())
			return
		}

		// 解析返回数据
		responseData := result.Data().(string)
		var response Person
		if err := encoding.GetCodec("json").Unmarshal([]byte(responseData), &response); err != nil {
			t.Errorf("Failed to unmarshal JSON response: %v", err)
			return
		}
		log.Info("Response: ", response)
	})

	// 测试MessagePacker
	t.Run("TestMessagePacker", func(t *testing.T) {
		ctx := context.Background()
		// 准备测试数据
		param0Data := map[string]interface{}{
			"id":     456,
			"name":   "MessagePack Test",
			"active": true,
		}

		// 序列化请求为MessagePack格式
		param0Bytes, err := msgpack.Marshal(param0Data)
		if err != nil {
			t.Fatalf("Failed to marshal msgpack data: %v", err)
		}
		content := &Content[[][]byte]{
			CType:   RpcContentBytes,
			Content: [][]byte{param0Bytes},
		}

		result, err := rpcServer.OnCall(ctx, "client", "ProcessMessagePack", content)
		if err != nil {
			t.Errorf("MessagePack Method call failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Expected result, got nil")
			return
		}

		// 验证返回类型
		if result.Type() != RpcContentBytes {
			t.Errorf("Expected bytes content type, got %v", result.Type())
			return
		}

		// 解析返回数据
		responseData := result.Data().([][]byte)
		if len(responseData) == 0 {
			t.Error("Expected non-empty response data")
			return
		}

		// 反序列化MessagePack响应 - 返回的是分别序列化的字节数组
		var response map[string]interface{}
		if err := msgpack.Unmarshal(responseData[0], &response); err != nil {
			t.Errorf("Failed to unmarshal MessagePack response: %v", err)
			return
		}

		// 验证结果
		if response["success"] != true {
			t.Error("Expected success=true")
		}
		if processed, ok := response["processed"].(map[string]interface{}); ok {
			// processed 就是传入的 param0Data，直接检查其内容
			if processed["name"] != "MessagePack Test" {
				t.Errorf("Expected name 'MessagePack Test', got '%v'", processed["name"])
			}
			// msgpack 可能将整数反序列化为不同的类型
			expectedId := 456
			actualId := processed["id"]
			var actualIdInt int
			switch v := actualId.(type) {
			case int:
				actualIdInt = v
			case int8:
				actualIdInt = int(v)
			case int16:
				actualIdInt = int(v)
			case int32:
				actualIdInt = int(v)
			case int64:
				actualIdInt = int(v)
			case uint:
				actualIdInt = int(v)
			case uint8:
				actualIdInt = int(v)
			case uint16:
				actualIdInt = int(v)
			case uint32:
				actualIdInt = int(v)
			case uint64:
				actualIdInt = int(v)
			case float32:
				actualIdInt = int(v)
			case float64:
				actualIdInt = int(v)
			default:
				t.Errorf("Unexpected type for id: %T, value: %v", actualId, actualId)
				return
			}
			if actualIdInt != expectedId {
				t.Errorf("Expected id %d, got %d (original type: %T)", expectedId, actualIdInt, actualId)
			}
			if processed["active"] != true {
				t.Errorf("Expected active true, got %v", processed["active"])
			}
		} else {
			t.Error("Expected processed data in response")
		}
	})
}

// TestInterfaceRegistration 测试接口类型注册
func TestInterfaceRegistration(t *testing.T) {
	rpcServer := NewRpcServer(grpcType)

	// 测试匿名结构体服务注册（应该失败）
	t.Run("TestAnonymousStructServiceShouldFail", func(t *testing.T) {
		// 创建一个匿名结构体服务
		anonymousService := &struct {
			name string
		}{
			name: "anonymous-test",
		}

		// 验证这是匿名结构体
		serviceType := reflect.TypeOf(anonymousService)
		if serviceType.Name() != "" {
			t.Errorf("Expected empty Name for anonymous struct, got %s", serviceType.Name())
		}

		// 注册服务应该失败
		err := rpcServer.Register(anonymousService)
		if err == nil {
			t.Error("Expected error when registering anonymous struct CallDispatcher, got nil")
		} else {
			// 这是预期的错误，测试通过
			t.Logf("Successfully caught expected error when registering anonymous struct: %v", err)
		}
	})

	// 测试字符串类型（应该成功，因为字符串有类型名）
	t.Run("TestStringTypeShouldSucceed", func(t *testing.T) {
		// 创建一个字符串
		var stringValue interface{} = "just a string"

		// 注册服务应该成功（字符串有类型名）
		err := rpcServer.Register(stringValue)
		if err != nil {
			t.Errorf("Expected success when registering string type, got error: %v", err)
		} else {
			t.Log("Successfully registered string type CallDispatcher")
		}
	})

	// 测试nil值（应该失败）
	t.Run("TestNilValueShouldFail", func(t *testing.T) {
		// 注册nil应该失败
		err := rpcServer.Register(nil)
		if err == nil {
			t.Error("Expected error when registering nil, got nil")
		} else {
			// 这是预期的错误，测试通过
			t.Logf("Successfully caught expected error when registering nil: %v", err)
		}
	})
}
