// Package serverinfo 提供服务器信息管理功能
package serverinfo

import (
	"context"
	"fmt"
	"time"
)

// serverInfoKey 用于 Context 中存储 ServerInfo 的键
type serverInfoKey struct{}

// WithServerInfo 将 ServerInfo 注入到 context
func WithServerInfo(ctx context.Context, info *ServerInfo) context.Context {
	if info == nil {
		return ctx
	}
	return context.WithValue(ctx, serverInfoKey{}, info)
}

// FromServerInfo 从 context 中获取 ServerInfo
func FromServerInfo(ctx context.Context) (*ServerInfo, bool) {
	if ctx == nil {
		return nil, false
	}
	info, ok := ctx.Value(serverInfoKey{}).(*ServerInfo)
	return info, ok
}

// InjectIntoContext 注入 ServerInfo 到 Context
func InjectIntoContext(ctx context.Context, info *ServerInfo) context.Context {
	return WithServerInfo(ctx, info)
}

// ExtractFromContext 从 Context 中提取 ServerInfo
func ExtractFromContext(ctx context.Context) (*ServerInfo, error) {
	info, ok := FromServerInfo(ctx)
	if !ok {
		return nil, fmt.Errorf("ServerInfo not found in context")
	}
	return info, nil
}

// MustExtractFromContext 从 Context 中提取 ServerInfo，如果不存在则 panic
func MustExtractFromContext(ctx context.Context) *ServerInfo {
	info, err := ExtractFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to extract ServerInfo from context: %v", err))
	}
	return info
}

// HasServerInfo 检查 Context 中是否包含 RuntimeInfo
func HasServerInfo(ctx context.Context) bool {
	_, ok := FromServerInfo(ctx)
	return ok
}

// GetServerInfoOrDefault 从 Context 中获取 ServerInfo，如果不存在则返回默认值
func GetServerInfoOrDefault(ctx context.Context, defaultInfo *ServerInfo) *ServerInfo {
	if info, ok := FromServerInfo(ctx); ok {
		return info
	}
	return defaultInfo
}

// ContextWithServerInfo 创建包含 ServerInfo 的新 Context
func ContextWithServerInfo(info *ServerInfo) context.Context {
	return context.WithValue(context.Background(), serverInfoKey{}, info)
}

// ContextWithServerInfoAndCancel 创建包含 ServerInfo 的可取消 Context
func ContextWithServerInfoAndCancel(info *ServerInfo) (context.Context, context.CancelFunc) {
	ctx := ContextWithServerInfo(info)
	return context.WithCancel(ctx)
}

// ContextWithServerInfoAndTimeout 创建包含 ServerInfo 的超时 Context
func ContextWithServerInfoAndTimeout(info *ServerInfo, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := ContextWithServerInfo(info)
	return context.WithTimeout(ctx, timeout)
}

// ContextWithServerInfoAndDeadline 创建包含 ServerInfo 的截止时间 Context
func ContextWithServerInfoAndDeadline(info *ServerInfo, deadline time.Time) (context.Context, context.CancelFunc) {
	ctx := ContextWithServerInfo(info)
	return context.WithDeadline(ctx, deadline)
}

// MergeContexts 合并多个 Context，优先使用包含 ServerInfo 的 Context
func MergeContexts(ctxs ...context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.Background()
	}

	// 查找包含 ServerInfo 的 Context
	for _, ctx := range ctxs {
		if HasServerInfo(ctx) {
			return ctx
		}
	}

	// 如果没有找到包含 ServerInfo 的 Context，返回第一个
	return ctxs[0]
}

// CloneContextWithServerInfo 克隆 Context 并注入新的 ServerInfo
func CloneContextWithServerInfo(ctx context.Context, info *ServerInfo) context.Context {
	if ctx == nil {
		return ContextWithServerInfo(info)
	}

	// 创建新的 Context，保留原有的其他值
	newCtx := context.Background()

	// 复制原有的值（除了 serverInfoKey）
	// 这里简化实现，实际可能需要更复杂的复制逻辑
	return WithServerInfo(newCtx, info)
}

// ContextMiddleware 中间件接口，用于在请求处理过程中注入 ServerInfo
type ContextMiddleware interface {
	// InjectRuntimeInfo 注入 ServerInfo 到 Context
	InjectRuntimeInfo(ctx context.Context, info *ServerInfo) context.Context

	// ExtractRuntimeInfo 从 Context 中提取 ServerInfo
	ExtractRuntimeInfo(ctx context.Context) (*ServerInfo, error)
}

// DefaultContextMiddleware 默认的 Context 中间件实现
type DefaultContextMiddleware struct{}

// NewDefaultContextMiddleware 创建新的默认 Context 中间件
func NewDefaultContextMiddleware() *DefaultContextMiddleware {
	return &DefaultContextMiddleware{}
}

// InjectRuntimeInfo 注入 ServerInfo 到 Context
func (dcm *DefaultContextMiddleware) InjectRuntimeInfo(ctx context.Context, info *ServerInfo) context.Context {
	return WithServerInfo(ctx, info)
}

// ExtractRuntimeInfo 从 Context 中提取 ServerInfo
func (dcm *DefaultContextMiddleware) ExtractRuntimeInfo(ctx context.Context) (*ServerInfo, error) {
	return ExtractFromContext(ctx)
}

// ContextBuilder Context 构建器，用于构建包含 ServerInfo 的 Context
type ContextBuilder struct {
	baseCtx context.Context
	info    *ServerInfo
}

// NewContextBuilder 创建新的 Context 构建器
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		baseCtx: context.Background(),
	}
}

// WithBaseContext 设置基础 Context
func (cb *ContextBuilder) WithBaseContext(baseCtx context.Context) *ContextBuilder {
	cb.baseCtx = baseCtx
	return cb
}

// WithRuntimeInfo 设置 ServerInfo
func (cb *ContextBuilder) WithRuntimeInfo(info *ServerInfo) *ContextBuilder {
	cb.info = info
	return cb
}

// Build 构建最终的 Context
func (cb *ContextBuilder) Build() context.Context {
	if cb.info == nil {
		return cb.baseCtx
	}
	return WithServerInfo(cb.baseCtx, cb.info)
}

// BuildWithTimeout 构建带超时的 Context
func (cb *ContextBuilder) BuildWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := cb.Build()
	return context.WithTimeout(ctx, timeout)
}

// BuildWithDeadline 构建带截止时间的 Context
func (cb *ContextBuilder) BuildWithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	ctx := cb.Build()
	return context.WithDeadline(ctx, deadline)
}
