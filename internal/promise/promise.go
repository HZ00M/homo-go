package promise

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Promise[T any] struct {
	value T
	err   error
	done  chan struct{}
	once  sync.Once
}

func NewPromise[T any](fn func(resolve func(T), reject func(error))) *Promise[T] {
	p := &Promise[T]{done: make(chan struct{})}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				p.reject(errors.New("panic"))
			}
		}()
		fn(p.resolve, p.reject)
	}()
	return p
}

func (p *Promise[T]) resolve(val T) {
	p.once.Do(func() {
		p.value = val
		close(p.done)
	})
}

func (p *Promise[T]) reject(err error) {
	p.once.Do(func() {
		p.err = err
		close(p.done)
	})
}

func (p *Promise[T]) Await(ctx context.Context) (T, error) {
	select {
	case <-p.done:
		return p.value, p.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

func Resolve[T any](val T) *Promise[T] {
	return NewPromise(func(resolve func(T), reject func(error)) {
		resolve(val)
	})
}

func Reject[T any](err error) *Promise[T] {
	return NewPromise(func(resolve func(T), reject func(error)) {
		reject(err)
	})
}

func Retry[T any](fn func() *Promise[T], max int, delay time.Duration) *Promise[T] {
	return NewPromise(func(resolve func(T), reject func(error)) {
		var attempt int
		var try func()
		try = func() {
			attempt++
			p := fn()
			Then(p, func(val T) *Promise[T] {
				resolve(val)
				return Resolve(val)
			})
			// 注意这里用 Catch 函数，不是方法
			Catch(p, func(err error) *Promise[T] {
				if attempt < max {
					// 非阻塞的延迟重试
					go func() {
						time.Sleep(delay)
						try()
					}()
					// 返回一个 dummy Promise 继续链式调用
					return Resolve(*new(T))
				} else {
					reject(err)
					return Reject[T](err)
				}
			})
		}
		try()
	})
}

// ------------------ 改成普通泛型函数的 Then -------------------

func Then[T, R any](p *Promise[T], fn func(T) *Promise[R]) *Promise[R] {
	return NewPromise(func(resolve func(R), reject func(error)) {
		<-p.done
		if p.err != nil {
			reject(p.err)
			return
		}
		next := fn(p.value)
		<-next.done
		if next.err != nil {
			reject(next.err)
		} else {
			resolve(next.value)
		}
	})
}

// ------------------ 改成普通泛型函数的 Catch -------------------

func Catch[T any](p *Promise[T], fn func(error) *Promise[T]) *Promise[T] {
	return NewPromise(func(resolve func(T), reject func(error)) {
		<-p.done
		if p.err != nil {
			next := fn(p.err)
			<-next.done
			if next.err != nil {
				reject(next.err)
			} else {
				resolve(next.value)
			}
		} else {
			resolve(p.value)
		}
	})
}
