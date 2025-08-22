package base

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/api/entity"
	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

// EntityServiceTemplate 为 Entity 提供的服务端适配桩
// 后续将对接 kratos transport grpc/http
// 最小实现：
// - OnEntityCall: Router.TrySetLocal -> CallSystem.Call -> EntityResponse
// 说明：串行保障由 CallSystem 内部的 per-entity Actor 实现；此处不再重复排队

type EntityServiceTemplate struct {
	eMgr    facade.EntityMgr
	callSys facade.CallSystem
	Router  facade.Router
}

func (s *EntityServiceTemplate) OnEntityCall(ctx context.Context, req *entity.EntityRequest) (*entity.EntityResponse, error) {
	// 路由绑定尝试（可选）
	if s.Router != nil && req.Id != "" {
		ok, prevPod, err := s.Router.TrySetLocal(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%w: targetPod=%d", facade.ErrAlreadyInOtherPod, prevPod)
		}
	}
	//这里可做鉴权，限流等

	// 分发到实体方法，由 CallSystem 保证串行
	resp, err := s.callSys.Call(ctx, "src", req.FunName, req)
	if err != nil {
		return nil, err
	}
	// 返回结果
	var res = &entity.EntityResponse{
		Content:  resp,
		RespCode: entity.EntityResponse_OK,
	}
	return res, nil
}

func (s *EntityServiceTemplate) OnBroadcastEntityCall(ctx context.Context, req *entity.BroadcastEntityRequest) (*facade.Empty, error) {
	return &facade.Empty{}, nil
}

func (s *EntityServiceTemplate) Ping(ctx context.Context, req *entity.PingRequest) (*facade.BytesReply, error) {
	return &facade.BytesReply{Payload: []byte("pong")}, nil
}
