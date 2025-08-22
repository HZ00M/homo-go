package executor

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
	"github.com/redis/go-redis/v9"
)

// TestInterfaceImplementation 测试StatefulExecutorImpl是否正确实现了StatefulExecutor接口
func TestInterfaceImplementation(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	defer client.Close()

	// 创建日志记录器
	logger := log.DefaultLogger

	// 创建执行器实例
	executor := NewStatefulExecutor(client, logger)

	// 编译时检查：确保类型转换成功
	var _ route.StatefulExecutor = executor

	// 运行时检查：确保不是nil
	if executor == nil {
		t.Fatal("executor should not be nil")
	}

	t.Log("StatefulExecutorImpl successfully implements StatefulExecutor interface")
}

// TestInterfaceMethods 测试接口方法是否存在
func TestInterfaceMethods(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	defer client.Close()

	// 创建日志记录器
	logger := log.DefaultLogger

	// 创建执行器实例
	var executor route.StatefulExecutor = NewStatefulExecutor(client, logger)

	// 验证所有必需的方法都存在
	if executor == nil {
		t.Fatal("executor should not be nil")
	}

	// 这里只是验证接口实现，不执行实际的Redis操作
	t.Log("All required interface methods are implemented")
}
