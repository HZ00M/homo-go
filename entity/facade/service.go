package facade

import (
	"context"
	"github.com/go-kratos/kratos/v2/api/entity"
)

// EntityService 统一 RPC 入口（服务端）
type EntityService interface {
	// OnEntityCall: 单实体调用入口
	OnEntityCall(ctx context.Context, req *entity.EntityRequest) (*entity.EntityResponse, error)
	// OnBroadcastEntityCall: 广播实体调用入口
	OnBroadcastEntityCall(ctx context.Context, req *entity.BroadcastEntityRequest) (*Empty, error)
	// Ping: 心跳/延迟测量（可选）
	Ping(ctx context.Context, req *entity.PingRequest) (*BytesReply, error)
}

type BytesReply struct{ Payload []byte }

type PingRequest struct{}

type Empty struct{}
