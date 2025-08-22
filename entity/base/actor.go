package base

import (
	"context"
	"sync"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

// Actor 串行执行模型：单 goroutine + chan
// 仅用于最小验证顺序执行与溢出错误

type Actor struct {
	ch     chan func()
	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex
}

// NewActor 创建 Actor，queueSize 决定通道容量
func NewActor(queueSize int) *Actor {
	if queueSize <= 0 {
		queueSize = 1
	}
	a := &Actor{ch: make(chan func(), queueSize)}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for f := range a.ch {
			func() {
				defer func() { _ = recover() }()
				f()
			}()
		}
	}()
	return a
}

// Enqueue 将任务入队；若队列已满则返回 ErrQueueOverflow
func (a *Actor) Enqueue(ctx context.Context, f func()) error {
	a.mu.Lock()
	closed := a.closed
	a.mu.Unlock()
	if closed {
		return context.Canceled
	}
	select {
	case a.ch <- f:
		return nil
	default:
		return facade.ErrQueueOverflow
	}
}

// QueueLen 返回当前队列长度（近似）
func (a *Actor) QueueLen() int { return len(a.ch) }

// Close 关闭队列并等待消费完成
func (a *Actor) Close() {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return
	}
	a.closed = true
	close(a.ch)
	a.mu.Unlock()
	a.wg.Wait()
}
