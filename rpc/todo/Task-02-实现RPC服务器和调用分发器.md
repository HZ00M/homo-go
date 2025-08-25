## 6A 任务卡：实现RPC服务器和调用分发器

- 编号: Task-02
- 模块: rpc
- 责任人: 待分配
- 优先级: 🔴
- 状态: ✅ 已完成
- 预计完成时间: 2025-01-27
- 实际完成时间: 2025-01-27

### A1 目标（Aim）
实现RPC服务器和调用分发器，提供基于反射的服务注册、方法扫描、调用分发等功能，支持任意Go结构体的自动注册和动态方法调用。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：RPC服务器已在 `rpc/server.go` 中实现
  - ✅ 已实现：调用分发器已在 `rpc/call.go` 中实现
  - ✅ 已实现：所有核心功能已实现
- **差距**：无
- **约束**：需要遵循Go语言最佳实践，支持反射和动态调用
- **风险**：
  - 技术风险：无
  - 业务风险：无
  - 依赖风险：无

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`RpcServerDriver` - RPC服务器驱动接口
  - **核心方法**：
    - 服务注册：`Register(service any) error`
    - 调用处理：`OnCall(ctx, srcService, msgId, data) (RpcContent, error)`
  - **输入输出参数及错误码**：
    - 使用Go标准库的`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装和类型断言
    - 使用Go的多返回值特性返回结果和错误

- **架构设计**：
  - 采用反射机制，支持任意Go结构体的自动注册
  - 使用映射表管理服务和方法的关系
  - 支持动态方法扫描和打包器选择
  - 使用组合模式，将调用分发器集成到服务器中

- **核心功能模块**：
  - `rpcServer`: RPC服务器实现
  - `CallDispatcher`: 调用分发器
  - 服务管理：服务注册、方法扫描、映射管理
  - 调用分发：方法查找、参数分发、结果返回

- **极小任务拆分**：
  - T02-01：实现RPC服务器基础结构
  - T02-02：实现服务注册和方法扫描
  - T02-03：实现调用分发器
  - T02-04：实现调用处理和错误管理

### A4 行动（Act）
#### T02-01：实现RPC服务器基础结构
```go
// rpc/server.go
package rpc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// RpcServer默认实现
type rpcServer struct {
	rpcType      RpcType
	funToService map[string]*CallDispatcher //通过msgId找到对应的service
	services     map[string]*CallDispatcher //通过serviceName找到对应的service
}

// NewRpcServer 创建新的RPC服务器
func NewRpcServer(rpcType RpcType) RpcServerDriver {
	return &rpcServer{
		rpcType:      rpcType,
		funToService: make(map[string]*CallDispatcher),
		services:     make(map[string]*CallDispatcher),
	}
}
```

#### T02-02：实现服务注册和方法扫描
```go
// Register 注册服务（支持扩展）
func (s *rpcServer) Register(service any) error {
	// 检查nil值
	if service == nil {
		return fmt.Errorf("cannot register nil CallDispatcher")
	}

	serviceValue := reflect.ValueOf(service)
	serviceType := serviceValue.Type()
	serviceName := serviceType.Name()
	// 如果类型名为空，可能是接口或匿名类型
	if serviceName == "" {
		// 如果是接口，尝试获取其实现的结构体名称
		if serviceType.Kind() == reflect.Interface || serviceValue.Kind() == reflect.Ptr {
			implType := serviceValue.Elem().Type()
			serviceName = implType.Name()

			if serviceName == "" {
				return fmt.Errorf("cannot register anonymous interface implementation")
			}
		} else {
			return fmt.Errorf("unsupported or anonymous CallDispatcher type: %v", serviceType.String())
		}
	}

	log.Infof("Registering CallDispatcher: %s", serviceName)

	// 创建服务实例
	callDispatcher := NewCallDispatcher(serviceName)
	// 遍历服务的方法
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		methodValue := serviceValue.Method(i)

		// 跳过私有方法
		if method.IsExported() {
			// 注册方法
			err := callDispatcher.RegisterMethod(method.Name, methodValue.Interface())
			if err != nil {
				log.Errorf("Failed to register Method %s: %v", method.Name, err)
				continue
			}
			log.Infof("Registered Method: %s.%s ", serviceName, method.Name)
		}
	}

	// 注册服务到RPC服务器
	s.registerService(callDispatcher)
	return nil
}
```

#### T02-03：实现调用分发器
```go
// rpc/call.go
package rpc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// CallDispatcher 调用分发器
type CallDispatcher struct {
	name      string
	RpcMethod map[string]*FunInfo
}

// NewCallDispatcher 创建新的调用分发器
func NewCallDispatcher(name string) *CallDispatcher {
	return &CallDispatcher{
		name:      name,
		RpcMethod: make(map[string]*FunInfo),
	}
}

// RegisterMethod 注册方法
func (c *CallDispatcher) RegisterMethod(name string, method interface{}) error {
	methodValue := reflect.ValueOf(method)
	methodType := methodValue.Type()

	// 创建方法信息
	funInfo := &FunInfo{
		Name:   name,
		Method: methodValue,
	}

	// 分析参数信息
	paramCount := methodType.NumIn()
	for i := 0; i < paramCount; i++ {
		paramType := methodType.In(i)
		paramInfo := ParamInfo{
			Name: fmt.Sprintf("param%d", i),
			Type: paramType,
		}
		funInfo.Param = append(funInfo.Param, paramInfo)
	}

	// 分析返回值信息
	returnCount := methodType.NumOut()
	for i := 0; i < returnCount; i++ {
		returnType := methodType.Out(i)
		returnInfo := ParamInfo{
			Name: fmt.Sprintf("return%d", i),
			Type: returnType,
		}
		funInfo.ReturnParam = append(funInfo.ReturnParam, returnInfo)
	}

	// 自动选择打包器
	funInfo.Packer = selectPacker(funInfo)

	// 注册方法
	c.RpcMethod[name] = funInfo
	return nil
}
```

#### T02-04：实现调用处理和错误管理
```go
// Dispatch 分发调用
func (c *CallDispatcher) Dispatch(ctx context.Context, msgId string, data RpcContent) (RpcContent, error) {
	// 查找方法
	funInfo, exists := c.RpcMethod[msgId]
	if !exists {
		return nil, errors.New(404, "METHOD_NOT_FOUND", fmt.Sprintf("method %s not found in service %s", msgId, c.name))
	}

	// 使用打包器处理调用
	return funInfo.Packer.Call(ctx, data, funInfo)
}

// OnCall 处理RPC调用
func (s *rpcServer) OnCall(ctx context.Context, srcService string, msgId string, data RpcContent) (RpcContent, error) {
	log.Infof("RPC call from %s, msgId: %s", srcService, msgId)

	// 查找对应的服务
	callDispatcher, exists := s.funToService[msgId]
	if !exists {
		return nil, errors.New(404, "SERVICE_NOT_FOUND", fmt.Sprintf("CallDispatcher not found for msgId: %s", msgId))
	}
	return callDispatcher.Dispatch(ctx, msgId, data)
}
```

### A5 验证（Assure）
- **测试用例**：RPC服务器和调用分发器实现已完成，包含完整的单元测试
- **性能验证**：使用反射和映射表优化方法查找，支持高并发调用
- **回归测试**：与现有模块集成测试通过
- **测试结果**：✅ 通过

### A6 迭代（Advance）
- 性能优化：已使用反射和映射表优化方法查找，支持高并发调用
- 功能扩展：支持任意Go结构体的自动注册和方法扫描
- 观测性增强：完整的日志记录和错误处理
- 下一步任务链接：[Task-03](./Task-03-实现RPC客户端和连接管理.md)

### 📋 质量检查
- [x] 代码质量检查完成
- [x] 文档质量检查完成
- [x] 测试质量检查完成

### 📋 完成总结
Task-02已完成，成功实现了RPC服务器和调用分发器。该实现提供了完整的服务注册、方法扫描、调用分发等功能，支持任意Go结构体的自动注册和动态方法调用。实现遵循Go语言最佳实践，使用反射机制实现动态方法调用，使用映射表管理服务和方法的关系，支持智能的打包器选择。所有核心功能已实现并通过测试。
