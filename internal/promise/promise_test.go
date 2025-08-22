package promise

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"testing"
	"time"
)

// 模拟一个会失败几次后成功的异步任务
func flakyTask(maxFail int) func() *Promise[int] {
	attempts := 0
	return func() *Promise[int] {
		return NewPromise(func(resolve func(int), reject func(error)) {
			attempts++
			if attempts <= maxFail {
				fmt.Printf("Attempt %d: fail\n", attempts)
				reject(errors.New("fail"))
			} else {
				fmt.Printf("Attempt %d: success\n", attempts)
				resolve(attempts)
			}
		})
	}
}

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()

	// 失败两次，第三次成功
	task := flakyTask(2)

	p := Retry(task, 5, 100*time.Millisecond)
	val, err := p.Await(ctx)
	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if val != 3 {
		t.Fatalf("Expected success on 3rd attempt, got %d", val)
	}
	t.Logf("Retry succeeded with value: %d", val)
}

func TestRetry_Fail(t *testing.T) {
	ctx := context.Background()

	// 失败次数大于最大重试次数，必然失败
	task := flakyTask(10)

	p := Retry(task, 5, 50*time.Millisecond)
	_, err := p.Await(ctx)
	if err == nil {
		t.Fatal("Expected error due to exceeding max retries, got nil")
	} else {
		t.Logf("Retry failed as expected: %v", err)
	}
}

func TestThenCatch(t *testing.T) {
	ctx := context.Background()

	// 成功的 Promise
	p := Resolve(42)

	// Then 转换 int->string
	p2 := Then(p, func(v int) *Promise[string] {
		log.Info("Then called with value:", v)
		return Resolve(fmt.Sprintf("number:%d", v))
	})

	// Catch 不应该被调用
	p3 := Catch(p2, func(err error) *Promise[string] {
		log.Info("Catch called with error:", err)
		return Resolve("error")
	})

	val, err := p3.Await(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := "number:42"
	if val != expected {
		t.Fatalf("Expected %q, got %q", expected, val)
	}
	t.Logf("Then-Catch success with value: %s", val)
}

func TestCatchOnReject(t *testing.T) {
	ctx := context.Background()

	// 创建一个拒绝的 Promise
	p := Reject[int](errors.New("fail"))

	// Catch 捕获错误，返回默认值
	p2 := Catch(p, func(err error) *Promise[int] {
		return Resolve(123)
	})

	val, err := p2.Await(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != 123 {
		t.Fatalf("Expected 123, got %d", val)
	}
	t.Logf("Catch recovered with value: %d", val)
}
