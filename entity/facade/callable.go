package facade

import (
	"context"
	"github.com/go-kratos/kratos/v2/api/entity"
)

type CallSystem interface {
	Call(ctx context.Context, srcName string, funName string, entityRequest *entity.EntityRequest) ([][]byte, error)

	LocalCall(ctx context.Context, srcName string, funName string, params []any) ([]any, error)
}
