package rpc

import "context"

type Packer interface {
	Name() string
	Match(fun *FunInfo) bool
	UnSerialize(content RpcContent, fun *FunInfo) ([]interface{}, error)
	Serialize(fun *FunInfo, returnValues []any) (RpcContent, error)
	Call(ctx context.Context, req RpcContent, fun *FunInfo) (RpcContent, error)
}

type RpcContent interface {
	Type() RpcContentType
	Data() any
}

type RpcClientDriver interface {
	Call(ctx context.Context, msgId string, data RpcContent) (RpcContent, error)
}

type RpcServerDriver interface {
	OnCall(ctx context.Context, srcService string, msgId string, data RpcContent) (RpcContent, error)
	Register(service any) error
}
