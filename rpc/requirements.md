# RPC模块需求文档

## 1. 项目背景
RPC模块是一个基于反射的通用RPC框架，支持多种序列化格式（Proto、JSON、MessagePack），具有自动打包器选择、服务注册和方法调用功能。该模块包含服务端和客户端两个核心组件，支持gRPC传输协议，旨在提供统一的跨服务通信解决方案。

**业务问题**：
- 传统RPC框架缺乏灵活性，难以适应不同序列化需求
- 服务注册和发现机制复杂，开发效率低
- 缺乏统一的错误处理和并发安全机制

**目标与范围**：
- 构建基于反射的通用RPC框架
- 支持多种序列化格式的自动选择
- 提供完整的服务端和客户端实现
- 集成gRPC传输协议

## 2. 功能目标
- **业务价值**：提供统一的RPC调用接口，降低微服务间通信复杂度
- **可量化指标**：
  - 支持3种序列化格式（Proto、JSON、MessagePack）
  - 自动打包器选择准确率100%
  - 服务注册成功率100%
  - 并发调用支持1000+ QPS

## 3. 功能需求（FR）
| 编号 | 功能描述 | 优先级 | 验收标准 |
|------|----------|--------|----------|
| FR-01 | 提供统一调用接口协议 | 🔴 高 | 所有模块按协议调用无差异 |
| FR-02 | 支持多种序列化格式 | 🔴 高 | Proto、JSON、MessagePack三种格式完全支持 |
| FR-03 | 自动打包器选择机制 | 🔴 高 | 基于方法签名自动选择合适打包器 |
| FR-04 | 服务注册与发现 | 🔴 高 | 支持服务注册和自动方法扫描 |
| FR-05 | gRPC传输协议集成 | 🔴 高 | 完整的gRPC服务端和客户端实现 |
| FR-06 | 并发安全调用 | 🟡 中 | 支持高并发场景，线程安全 |
| FR-07 | 错误处理机制 | 🟡 中 | 统一的错误处理和优雅降级 |
| FR-08 | 流式调用支持 | 🟢 低 | 支持有状态服务的流式通信 |

## 4. 非功能需求（NFR）
- **可扩展性**：支持自定义打包器和编解码器
- **隔离性**：服务间调用完全隔离，不影响其他服务
- **灵活性**：支持匿名结构体和接口实现
- **通用性**：基于反射机制，适用于任意Go结构体
- **性能**：反射结果缓存，减少运行时开销
- **可靠性**：完善的错误处理和超时机制

## 5. 核心设计规则
1. **单一职责**：每个打包器只负责一种序列化格式
2. **接口优先**：基于接口设计，支持多种实现
3. **可扩展性**：支持自定义打包器和编解码器注册
4. **模块边界清晰**：服务端、客户端、打包器职责分离
5. **观测性内建**：支持上下文传递和错误追踪
6. **失败优雅降级**：连接失败时提供降级策略

## 6. 核心设计

### 核心接口

#### RPC服务器接口
```go
type RpcServerDriver interface {
    OnCall(ctx context.Context, srcService string, msgId string, data RpcContent) (RpcContent, error)
    Register(service any) error
}
```
**功能**：处理RPC调用请求，管理服务注册
**参数**：
- `ctx`: 上下文信息
- `srcService`: 源服务名称
- `msgId`: 消息标识
- `data`: RPC内容数据
**返回值**：RPC响应内容和错误信息

#### RPC客户端接口
```go
type RpcClientDriver interface {
    Call(ctx context.Context, msgId string, data RpcContent) (RpcContent, error)
}
```
**功能**：发起RPC调用请求
**参数**：
- `ctx`: 上下文信息
- `msgId`: 消息标识
- `data`: RPC内容数据
**返回值**：RPC响应内容和错误信息

#### RPC内容接口
```go
type RpcContent interface {
    Type() RpcContentType
    Data() any
}
```
**功能**：定义RPC内容的类型和数据
**返回值**：内容类型枚举和任意类型数据

#### 参数打包器接口
```go
type Packer interface {
    Name() string
    Match(fun *FunInfo) bool
    UnSerialize(content RpcContent, fun *FunInfo) ([]interface{}, error)
    Serialize(fun *FunInfo, returnValues []any) (RpcContent, error)
    Call(ctx context.Context, req RpcContent, fun *FunInfo) (RpcContent, error)
}
```
**功能**：序列化和反序列化RPC参数
**参数**：RPC内容和函数信息
**返回值**：序列化结果和错误信息

### 数据模型与协议

#### 内容类型枚举
```go
type RpcContentType int

const (
    RpcContentBytes RpcContentType = iota  // 字节类型
    RpcContentJson                         // JSON类型
    RpcContentFile                         // 文件类型
)
```

#### 泛型内容信息
```go
type ContentInfo[T any] struct {
    ContentType RpcContentType  // 内容类型
    Content     T               // 泛型内容数据
}
```

#### 方法信息结构
```go
type FunInfo struct {
    Name        string           // 方法名
    Method      reflect.Value    // 方法反射值
    Param       []paramInfo      // 参数信息
    ReturnParam []paramInfo      // 返回值信息
    Packer      Packer           // 选中的打包器
}
```

### 使用示例

#### 服务端定义
```go
type UserService struct{}

func (s *UserService) GetUser(req *GetUserRequest) (*GetUserResponse, error) {
    // 业务逻辑
    return &GetUserResponse{...}, nil
}
```

#### 服务端注册
```go
rpcServer := NewRpcServer(grpcType)
userService := &UserService{}
err := rpcServer.Register(userService)
```

#### 客户端调用
```go
// 创建客户端
client, err := NewRpcProxyDirect(ctx, "localhost:8080", "client-service", false, grpc.WithInsecure())

// JSON调用
jsonContent := &ContentInfo[string]{
    ContentType: RpcContentJson,
    Content:     `{"id": 123}`,
}
result, err := client.Call(ctx, "GetUser", jsonContent)
```

## 7. 约束与边界

### 依赖模块
- **transport/grpc**: gRPC传输协议支持
- **encoding/json**: JSON序列化支持
- **encoding/msgpack**: MessagePack序列化支持
- **errors**: 统一错误处理
- **log**: 日志记录
- **metadata**: 元数据管理

### 功能边界
- **提供功能**：
  - RPC服务端和客户端实现
  - 多种序列化格式支持
  - 自动打包器选择
  - 服务注册和发现
  - 错误处理和超时机制

- **不提供功能**：
  - 数据库直接访问
  - HTTP等其他传输协议
  - 非Go语言支持
  - 业务逻辑实现

### 预期支持的功能
- 微服务间通信
- 分布式系统调用
- 服务网格集成
- 云原生应用支持

## 8. 模块目录规划与文件预期

```
rpc/
├── requirements.md          # 需求文档（本文件）
├── README.md               # 模块说明文档
├── server.go               # RPC服务器实现
├── client.go               # RPC客户端实现
├── content.go              # RPC内容类型定义
├── packer.go               # 打包器接口定义
├── json_packer.go          # JSON打包器实现
├── bytes_packer.go         # 字节打包器实现
├── service.go              # gRPC服务实现
├── errors.go               # 错误定义
└── tests/                  # 测试文件目录
    ├── server_test.go
    ├── client_test.go
    └── packer_test.go
```

## 9. 验收标准

### 核心测试列表

#### 单元测试
- [ ] RPC服务器接口测试
- [ ] RPC客户端接口测试
- [ ] 打包器接口测试
- [ ] 内容类型测试
- [ ] 错误处理测试

#### 集成测试
- [ ] 服务端注册测试
- [ ] 客户端调用测试
- [ ] 端到端通信测试
- [ ] 多服务交互测试

#### 性能测试
- [ ] 并发调用测试（1000+ QPS）
- [ ] 序列化性能测试
- [ ] 内存使用测试
- [ ] 响应时间测试

#### 功能验证测试
- [ ] 3种序列化格式支持测试
- [ ] 自动打包器选择测试
- [ ] 错误场景测试
- [ ] 超时机制测试

#### 兼容性测试
- [ ] gRPC协议兼容性
- [ ] 不同protobuf版本兼容性
- [ ] 向后兼容性保证
- [ ] 跨平台兼容性

#### 回归测试
- [ ] 核心功能稳定性测试
- [ ] 边界条件测试
- [ ] 异常情况处理测试
- [ ] 长期运行稳定性测试 