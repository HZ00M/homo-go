package rpc

import (
	"reflect"
)

type RpcType int

const (
	grpcType RpcType = iota
	httpType
)

// 参数信息结构
type ParamInfo struct {
	Name  string
	Type  reflect.Type
	Codec string
}

// rpc方法信息
type FunInfo struct {
	Name        string
	Method      reflect.Value
	Param       []ParamInfo
	ReturnParam []ParamInfo
	Packer      Packer
}

// ParamTypes 返回方法参数类型列表
func (f *FunInfo) ParamTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(f.Param))
	for _, param := range f.Param {
		types = append(types, param.Type)
	}
	return types
}

// ReturnTypes 返回方法返回值类型列表
func (f *FunInfo) ReturnTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(f.ReturnParam))
	for _, param := range f.ReturnParam {
		types = append(types, param.Type)
	}
	return types
}

type RpcContentType int

const (
	RpcContentBytes RpcContentType = iota
	RpcContentJson
	RpcContentFile
)

// 泛型实现结构体
type Content[T any] struct {
	CType RpcContentType
	Dt    T
}

func (r *Content[T]) Type() RpcContentType {
	return r.CType
}

func (r *Content[T]) Data() any {
	return r.Dt
}
