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

// RegisterService 注册服务（内部方法）
func (s *rpcServer) registerService(service *CallDispatcher) {
	s.services[service.name] = service

	// 将服务的方法注册到funToService映射中
	for msgId, funInfo := range service.RpcMethod {
		s.funToService[msgId] = service
		log.Infof("Registered CallDispatcher Method: %s -> %s.%s", msgId, service.name, funInfo.Name)
	}
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
