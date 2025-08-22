package rpc

import (
	"context"
)

var Packers = map[string]Packer{}

func RegisterPacker(packer Packer) {
	Packers[packer.Name()] = packer
}

type Packer interface {
	Name() string
	Match(fun *FunInfo) bool
	UnSerialize(content RpcContent, fun *FunInfo) ([]interface{}, error)
	Serialize(fun *FunInfo, returnValues []any) (RpcContent, error)
	Call(ctx context.Context, req RpcContent, fun *FunInfo) (RpcContent, error)
}

func SelectPacker(fun *FunInfo) Packer {
	// 选择逻辑
	for _, packer := range Packers {
		if packer.Match(fun) {
			return packer
		}
	}
	return nil
}
