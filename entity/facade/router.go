package facade

import "context"

// Router 路由/链接信息（有状态路由）
type Router interface {
	TrySetLocal(ctx context.Context, id string) (ok bool, prevPod int, err error)
	UnlinkLocal(ctx context.Context, id string) error
	CurrentPod() int
	ResolvePod(ctx context.Context, id string) (pod int, err error)
}
