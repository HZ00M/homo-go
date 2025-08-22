# Entity 技术需求文档

## 1. 项目背景
基于 Kratos 在 Go 中实现 Entity 运行时与调用框架，保持 Java 版核心语义（实体/能力/存储/路由/生命周期），简化为同步调用模型，预留异步。当前需要完善基础运行时、存储能力、调用与传输等核心功能模块。

## 2. 功能目标
- **业务价值**：提供统一的实体管理框架，支持分布式环境下的实体生命周期管理、能力挂载、串行执行保障
- **可量化指标**：
  - 实体调用延迟 < 10ms（本地）
  - 支持并发实体数 > 10000
  - 实体方法串行执行100%保证
  - 存储能力支持自动脏标管理与周期保存

## 3. 功能需求（FR）

| 编号 | 功能描述 | 优先级 | 验收标准 |
|------|----------|--------|----------|
| FR-01 | 提供统一调用接口协议 | 🔴 高 | 所有模块按协议调用无差异，支持EntityRequest/EntityResponse标准格式 |
| FR-02 | 支持实体生命周期管理 | 🔴 高 | Create/Get/Remove/Destroy等操作完整实现，支持钩子机制 |
| FR-03 | 支持能力系统挂载 | 🔴 高 | Ability可动态Attach/Detach，支持横向能力扩展 |
| FR-04 | 支持串行执行保障 | 🔴 高 | 同实体方法调用按per-entity Actor排队，100%串行执行 |
| FR-05 | 支持存储能力管理 | 🔴 高 | Saveable接口实现，支持脏标管理与自动保存 |
| FR-06 | 支持路由绑定机制 | 🟡 中 | Router.TrySetLocal保证实体唯一性，支持跨Pod路由 |
| FR-07 | 支持协议版本化 | 🟡 中 | 新版本发布不影响旧版本调用，支持Schema版本管理 |
| FR-08 | 支持参数验证与错误码标准化 | 🔴 高 | 所有接口在输入校验失败时返回统一错误码 |
| FR-09 | 支持广播调用机制 | 🟢 低 | OnBroadcastEntityCall支持并发调用与错误聚合 |

## 4. 非功能需求（NFR）
- **可扩展性**：支持能力动态挂载，支持横向能力扩展
- **隔离性**：模块边界清晰，接口优先设计
- **灵活性**：支持配置化，支持多种存储后端
- **通用性**：支持多种实体类型，支持多种调用方式

## 5. 核心设计规则
1. **单一职责**：每个模块只负责一个核心功能
2. **接口优先**：定义清晰的接口契约，实现与接口分离
3. **可扩展性**：支持能力动态挂载，支持横向能力扩展
4. **模块边界清晰**：facade/transport/ability/base等模块职责明确
5. **观测性内建**：支持日志、metrics、tracing等可观测性
6. **失败优雅降级**：支持错误处理、重试机制、优雅降级

## 6. 核心设计

### 核心接口

#### Entity 有状态实体的最小契约
```go
type Entity interface {
    Type() string
    ID() string
    SetID(id string)
    Init(ctx context.Context) error
    Destroy(ctx context.Context) error
    LoadTime() int64
    GetAbility(name string) Ability
    AddAbility(a Ability)
}
```
**功能**：定义实体的基本行为和属性
**参数**：
- `ctx`: 上下文信息
- `id`: 实体唯一标识
- `a`: 能力对象
**返回值**：实体类型、ID、能力对象等

#### EntityMgr 实体生命周期与缓存管理
```go
type EntityMgr interface {
    Exists(ctx context.Context, entityType, id string) (bool, error)
    Create(ctx context.Context, entityType, id string, ctor func() Entity) (Entity, error)
    Get(ctx context.Context, entityType, id string) (Entity, error)
    Remove(ctx context.Context, entityType, id string) error
    // ... 其他方法
}
```
**功能**：管理实体的创建、获取、删除等生命周期操作
**参数**：
- `ctx`: 上下文信息
- `entityType`: 实体类型
- `id`: 实体ID
- `ctor`: 实体构造函数
**返回值**：实体对象、操作结果和错误信息

#### Ability 能力最小契约
```go
type Ability interface {
    Name() string
    Attach(ctx context.Context, owner Entity) error
    Detach(ctx context.Context) error
}
```
**功能**：定义能力的挂载和卸载行为
**参数**：
- `ctx`: 上下文信息
- `owner`: 拥有该能力的实体
**返回值**：能力名称和操作结果

#### CallSystem 统一调用入口
```go
type CallSystem interface {
    Call(ctx context.Context, srcName string, funName string, entityRequest *entity.EntityRequest) ([][]byte, error)
    LocalCall(ctx context.Context, srcName string, funName string, params []any) ([]any, error)
}
```
**功能**：提供统一的实体方法调用入口
**参数**：
- `ctx`: 上下文信息
- `srcName`: 源服务名称
- `funName`: 方法名称
- `entityRequest`: 实体请求
- `params`: 调用参数
**返回值**：调用结果和错误信息

### 数据模型与协议

#### EntityRequest 实体调用请求
```go
type EntityRequest struct {
    Type     string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`           // 实体类型
    Id       string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`               // 实体ID
    FunName  string `protobuf:"bytes,3,opt,name=fun_name,json=funName,proto3" json:"fun_name,omitempty"` // 方法名称
    Params   []byte `protobuf:"bytes,4,opt,name=params,proto3" json:"params,omitempty"`       // 方法参数
}
```

#### EntityResponse 实体调用响应
```go
type EntityResponse struct {
    Content  [][]byte `protobuf:"bytes,1,rep,name=content,proto3" json:"content,omitempty"` // 响应内容
    RespCode int32    `protobuf:"varint,2,opt,name=resp_code,json=respCode,proto3" json:"resp_code,omitempty"` // 响应码
}
```

### 使用示例

#### 场景1：实体创建与能力挂载
```go
// 创建用户实体
user := &UserEntity{}
user.SetID("user_001")
user.Init(ctx)

// 挂载存储能力
storageAbility := &StorageAbility{}
storageAbility.Attach(ctx, user)

// 挂载定时能力
timerAbility := &TimerAbility{}
timerAbility.Attach(ctx, user)
```

#### 场景2：实体方法调用
```go
// 客户端调用
req := &entity.EntityRequest{
    Type:    "User",
    Id:      "user_001",
    FunName: "UpdateProfile",
    Params:  []byte(`{"name":"张三","age":25}`),
}

resp, err := entityService.OnEntityCall(ctx, req)
```

#### 场景3：生命周期管理
```go
// 获取实体（自动加载或创建）
user, err := entityMgr.GetOrCreate(ctx, "User", "user_001", func() Entity {
    return &UserEntity{}
})

// 实体销毁时自动保存
defer user.Destroy(ctx)
```

## 7. 约束与边界

### 依赖模块
- **rpc**: RPC调用框架支持
- **storage**: 存储系统支持
- **log**: 日志记录
- **errors**: 统一错误处理
- **metadata**: 元数据管理
- **transport/grpc**: gRPC传输协议支持

### 功能边界
- **提供功能**：
  - 实体生命周期管理
  - 能力系统挂载
  - 串行执行保障
  - 存储能力管理
  - 路由绑定机制
  - 统一调用接口

- **不提供功能**：
  - 数据库直接访问
  - 业务逻辑实现
  - 网络传输协议
  - 序列化格式处理

### 预期支持的功能
- 分布式实体管理
- 微服务架构支持
- 云原生应用
- 游戏服务器
- 实时计算系统

## 8. 模块目录规划与文件预期

```
entity/
├── facade/           # 接口契约层
│   ├── entity.go     # Entity、EntityMgr、Ability等核心接口
│   ├── ability.go    # Ability、AbilitySystem接口
│   ├── storage.go    # Storage、SaveObject接口
│   ├── router.go     # Router接口
│   ├── call.go       # CallSystem接口
│   ├── service.go    # EntityService接口
│   ├── errors.go     # 标准错误定义
│   └── config.go     # 配置结构定义
├── base/             # 基础实现层
│   ├── entity.go     # BaseEntity实现
│   ├── ability.go    # BaseAbility实现
│   └── manager.go    # MemoryManager实现
├── ability/          # 能力实现层
│   ├── base.go       # 基础能力实现
│   ├── call/         # 调用能力
│   │   ├── ability.go
│   │   └── system.go
│   └── storage/      # 存储能力
│       ├── ability.go
│       └── system.go
└── transport/        # 传输适配层
    └── server.go     # EntityServiceTemplate实现
```

**文件预期**：
- 每个接口文件包含完整的接口定义和注释
- 每个实现文件包含完整的实现逻辑和单元测试
- 配置文件包含合理的默认值和配置项
- 错误文件包含标准化的错误码和描述

## 9. 验收标准

### 核心测试列表

#### 单元测试
- [ ] Entity接口测试
- [ ] EntityMgr接口测试
- [ ] Ability接口测试
- [ ] CallSystem接口测试
- [ ] 错误处理测试

#### 集成测试
- [ ] 实体生命周期管理测试
- [ ] 能力挂载卸载测试
- [ ] 串行执行保障测试
- [ ] 存储能力测试
- [ ] 路由绑定测试

#### 性能测试
- [ ] 实体调用延迟测试（< 10ms）
- [ ] 并发实体数测试（> 10000）
- [ ] 串行执行性能测试
- [ ] 内存使用测试
- [ ] 存储性能测试

#### 功能验证测试
- [ ] 实体创建、获取、删除测试
- [ ] 能力动态挂载测试
- [ ] 串行执行100%保证测试
- [ ] 脏标管理与自动保存测试
- [ ] 跨Pod路由测试

#### 兼容性测试
- [ ] 协议版本兼容性测试
- [ ] 向后兼容性保证
- [ ] 不同存储后端兼容性
- [ ] 跨平台兼容性

#### 回归测试
- [ ] 核心功能稳定性测试
- [ ] 边界条件测试
- [ ] 异常情况处理测试
- [ ] 长期运行稳定性测试
- [ ] 错误场景处理测试

#### 可观测性测试
- [ ] 日志记录完整性测试
- [ ] Metrics指标准确性测试
- [ ] Tracing链路追踪测试
- [ ] 错误码标准化测试