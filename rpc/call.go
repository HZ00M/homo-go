package rpc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
)

// rpc业务服务基础
type CallDispatcher struct {
	name string
	//业务方法
	RpcMethod map[string]*FunInfo
}

func (c *CallDispatcher) Dispatch(ctx context.Context, fun string, req RpcContent) (ret RpcContent, err error) {
	funInfo, exists := c.RpcMethod[fun]
	if !exists {
		return nil, errors.New(404, "METHOD_NOT_FOUND", fmt.Sprintf("Method not found for msgId: %s", fun))
	}
	return funInfo.Packer.Call(ctx, req, funInfo)
}

// NewService 创建新的服务
func NewCallDispatcher(service any) *CallDispatcher {
	serviceValue := reflect.ValueOf(service)
	serviceType := serviceValue.Type()
	serviceName := serviceType.Name()
	c := &CallDispatcher{
		name:      serviceName,
		RpcMethod: make(map[string]*FunInfo),
	}
	// 遍历服务的方法
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		methodValue := serviceValue.Method(i)

		// 跳过私有方法
		if method.IsExported() {
			// 注册方法
			err := c.RegisterMethod(method.Name, methodValue.Interface())
			if err != nil {
				log.Errorf("Failed to register Method %s: %v", method.Name, err)
				continue
			}
			log.Infof("Registered Method: %s.%s ", serviceType, method.Name)
		}

	}
	return c
}

// RegisterMethod 注册RPC方法
func (s *CallDispatcher) RegisterMethod(msgId string, method interface{}) error {
	if method != nil {
		methodValue := reflect.ValueOf(method)
		methodType := methodValue.Type()

		// 检查方法签名
		if methodType.Kind() != reflect.Func {
			return errors.New(400, "INVALID_METHOD", "Method must be a function")
		}

		protoMessageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
		ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
		errorType := reflect.TypeOf((*error)(nil)).Elem()

		// 分析参数
		paramInfos := make([]ParamInfo, 0, methodType.NumIn())
		for i := 0; i < methodType.NumIn(); i++ {
			paramType := methodType.In(i)
			// 跳过 context.Context
			if paramType == ctxType {
				continue
			}
			paramName := paramType.String()
			codec := "json"
			underlying := paramType
			if underlying.Kind() == reflect.Ptr {
				underlying = underlying.Elem()
			}
			// 判定 proto
			if paramType.Implements(protoMessageType) || reflect.PtrTo(underlying).Implements(protoMessageType) {
				codec = "proto"
			} else if underlying.Kind() == reflect.Struct && underlying.Name() != "" {
				codec = "json"
			} else {
				codec = "msgpack"
			}
			paramInfos = append(paramInfos, ParamInfo{
				Name:  paramName,
				Type:  paramType,
				Codec: codec,
			})
		}

		// 分析返回值（忽略 error 类型）
		returnParamInfos := make([]ParamInfo, 0, methodType.NumOut())
		for i := 0; i < methodType.NumOut(); i++ {
			returnType := methodType.Out(i)
			if returnType.Implements(errorType) {
				// 跳过 error 返回
				continue
			}
			paramName := returnType.String()
			codec := "json"
			underlying := returnType
			if underlying.Kind() == reflect.Ptr {
				underlying = underlying.Elem()
			}
			if returnType.Implements(protoMessageType) || reflect.PtrTo(underlying).Implements(protoMessageType) {
				codec = "proto"
			} else if underlying.Kind() == reflect.Struct && underlying.Name() != "" {
				codec = "json"
			} else {
				codec = "msgpack"
			}
			returnParamInfos = append(returnParamInfos, ParamInfo{
				Name:  paramName,
				Type:  returnType,
				Codec: codec,
			})
		}

		// 创建方法信息
		funInfo := &FunInfo{
			Name:        msgId,
			Method:      methodValue,
			Param:       paramInfos,
			ReturnParam: returnParamInfos,
		}
		// 根据方法参数和返回值选择合适的打包器
		packer := SelectPacker(funInfo)
		funInfo.Packer = packer

		s.RpcMethod[msgId] = funInfo
	}
	return nil
}
