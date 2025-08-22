package base

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestActor_Order(t *testing.T) {
	a := NewActor(128)
	defer a.Close()
	var seq int64
	for i := 0; i < 100; i++ {
		want := int64(i + 1)
		if err := a.Enqueue(context.Background(), func() {
			got := atomic.AddInt64(&seq, 1)
			if got != want {
				t.Errorf("order broken: want %d got %d", want, got)
			}
		}); err != nil {
			t.Fatalf("enqueue err: %v", err)
		}
	}
	// 等待执行
	time.Sleep(50 * time.Millisecond)
}

func TestActor_Overflow(t *testing.T) {
	a := NewActor(1)
	defer a.Close()
	block := make(chan struct{})
	// 塞满一个任务且阻塞消费
	if err := a.Enqueue(context.Background(), func() { <-block }); err != nil {
		t.Fatalf("first enqueue err: %v", err)
	}
	// 第二个任务因为队列容量1，应当溢出
	if err := a.Enqueue(context.Background(), func() {}); err == nil {
		t.Fatalf("expect overflow error")
	}
	close(block)
}
