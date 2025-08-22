package executor

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// ExampleUsage 展示如何使用option构建方式创建Redis管理器
func ExampleUsage() {
	// 方式1：使用默认配置
	manager1, err := NewRedisManager()
	if err != nil {
		panic(err)
	}
	defer manager1.Close()

	// 方式2：使用单个选项
	manager2, err := NewRedisManager(
		WithAddresses("redis-server:6379"),
		WithPassword("mypassword"),
	)
	if err != nil {
		panic(err)
	}
	defer manager2.Close()

	// 方式3：使用多个选项
	manager3, err := NewRedisManager(
		WithAddresses("redis-cluster:6379"),
		WithPassword("clusterpass"),
		WithDB(1),
		WithPoolSize(20),
		WithMinIdle(10),
	)
	if err != nil {
		panic(err)
	}
	defer manager3.Close()
}

// ExampleWithCustomLogger 展示如何使用自定义日志器
func ExampleWithCustomLogger() {
	// 创建自定义日志器
	customLogger := log.DefaultLogger

	// 使用自定义日志器创建管理器
	manager, err := NewRedisManager(
		WithAddresses("redis:6379"),
		WithLogger(customLogger),
	)
	if err != nil {
		panic(err)
	}
	defer manager.Close()
}

// ExampleWithMinimalConfig 展示最小配置
func ExampleWithMinimalConfig() {
	// 只设置必要的配置
	manager, err := NewRedisManager(
		WithAddresses("localhost:6379"),
		WithDB(0),
	)
	if err != nil {
		panic(err)
	}
	defer manager.Close()
}

// ExampleWithHighPerformance 展示高性能配置
func ExampleWithHighPerformance() {
	// 高性能配置
	manager, err := NewRedisManager(
		WithAddresses("redis:6379"),
		WithPoolSize(50),
		WithMinIdle(20),
		WithMaxRetries(5),
		WithDialTimeout(2*time.Second),
		WithReadTimeout(1*time.Second),
		WithWriteTimeout(1*time.Second),
		WithPoolTimeout(3*time.Second),
		WithIdleTimeout(10*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer manager.Close()
}

// ExampleBasicOperations 展示基本操作
func ExampleBasicOperations() {
	manager, err := NewRedisManager()
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	ctx := context.Background()

	// 基本键值操作
	err = manager.Set(ctx, "key1", "value1", time.Hour)
	if err != nil {
		panic(err)
	}

	value, err := manager.Get(ctx, "key1")
	if err != nil {
		panic(err)
	}
	_ = value // 使用value

	// Hash操作
	err = manager.HSet(ctx, "hash1", "field1", "value1")
	if err != nil {
		panic(err)
	}

	hashValue, err := manager.HGet(ctx, "hash1", "field1")
	if err != nil {
		panic(err)
	}
	_ = hashValue // 使用hashValue

	// 列表操作
	err = manager.LPush(ctx, "list1", "item1", "item2")
	if err != nil {
		panic(err)
	}

	item, err := manager.LPop(ctx, "list1")
	if err != nil {
		panic(err)
	}
	_ = item // 使用item
}

// ExampleTransaction 展示事务操作
func ExampleTransaction() {
	manager, err := NewRedisManager()
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	ctx := context.Background()

	// 获取事务管道
	pipe := manager.TxPipeline()

	// 添加命令到管道
	pipe.Set(ctx, "key1", "value1", time.Hour)
	pipe.Set(ctx, "key2", "value2", time.Hour)
	pipe.Expire(ctx, "key1", time.Hour)

	// 执行事务
	_, err = pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}
}

// ExampleAdvancedUsage 展示高级用法
func ExampleAdvancedUsage() {
	manager, err := NewRedisManager(
		WithAddresses("redis:6379"),
		WithPassword("password"),
		WithDB(1),
	)
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	ctx := context.Background()

	// 获取原始客户端进行高级操作
	client := manager.GetClient()
	if client == nil {
		panic("client is nil")
	}

	// 使用原始客户端进行复杂操作
	err = client.Eval(ctx, "return redis.call('ping')", []string{}, []interface{}{}).Err()
	if err != nil {
		panic(err)
	}

	// 获取连接池统计
	stats := manager.PoolStats()
	_ = stats.TotalConns
	_ = stats.IdleConns
	_ = stats.StaleConns
}
