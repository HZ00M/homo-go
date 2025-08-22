package executor

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// ExampleDirectV9Usage 展示如何直接使用go-redis/v9客户端
func ExampleDirectV9Usage() {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxRetries:      3,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	defer client.Close()

	// 创建有状态执行器
	logger := log.DefaultLogger
	executor := NewStatefulExecutor(client, logger)

	ctx := context.Background()

	// 使用示例
	// 1. 设置服务状态
	err := executor.SetServiceState(ctx, "default", "my-service", 1, "Running")
	if err != nil {
		log.NewHelper(logger).Errorf("设置服务状态失败: %v", err)
		return
	}

	// 2. 获取服务状态
	states, err := executor.GetServiceState(ctx, "default", "my-service")
	if err != nil {
		log.NewHelper(logger).Errorf("获取服务状态失败: %v", err)
		return
	}
	log.NewHelper(logger).Infof("服务状态: %v", states)

	// 3. 设置工作负载状态
	err = executor.SetWorkloadState(ctx, "default", "my-service", "Active")
	if err != nil {
		log.NewHelper(logger).Errorf("设置工作负载状态失败: %v", err)
		return
	}

	// 4. 获取工作负载状态
	workloadState, err := executor.GetWorkloadState(ctx, "default", "my-service")
	if err != nil {
		log.NewHelper(logger).Errorf("获取工作负载状态失败: %v", err)
		return
	}
	log.NewHelper(logger).Infof("工作负载状态: %s", workloadState)

	// 5. 设置Pod链接
	podID, err := executor.SetLinkedPod(ctx, "default", "user123", "my-service", 1, 3600)
	if err != nil {
		log.NewHelper(logger).Errorf("设置Pod链接失败: %v", err)
		return
	}
	log.NewHelper(logger).Infof("Pod链接设置成功，Pod ID: %d", podID)

	// 6. 获取链接服务信息
	linkInfo, err := executor.GetLinkService(ctx, "default", "user123")
	if err != nil {
		log.NewHelper(logger).Errorf("获取链接服务失败: %v", err)
		return
	}
	log.NewHelper(logger).Infof("链接服务信息: %+v", linkInfo)

	// 7. 尝试设置Pod链接
	success, currentPodID, err := executor.TrySetLinkedPod(ctx, "default", "user123", "my-service", 2, 3600)
	if err != nil {
		log.NewHelper(logger).Errorf("尝试设置Pod链接失败: %v", err)
		return
	}
	log.NewHelper(logger).Infof("尝试设置Pod链接结果: success=%v, currentPodID=%d", success, currentPodID)

	log.NewHelper(logger).Info("所有操作完成")
}

// ExampleClusterUsage 展示如何使用Redis集群
func ExampleClusterUsage() {
	// 注意：当前StatefulExecutor只支持redis.Client，不支持redis.ClusterClient
	// 如果需要支持集群模式，需要修改StatefulExecutor的接口设计

	// 创建Redis集群客户端
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
		Password:        "",
		PoolSize:        10,
		MinIdleConns:    5,
		MaxRetries:      3,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	defer clusterClient.Close()

	logger := log.DefaultLogger
	ctx := context.Background()

	// 直接使用集群客户端进行操作
	err := clusterClient.Set(ctx, "cluster:key", "cluster:value", time.Hour).Err()
	if err != nil {
		log.NewHelper(logger).Errorf("集群模式设置键值失败: %v", err)
		return
	}

	value, err := clusterClient.Get(ctx, "cluster:key").Result()
	if err != nil {
		log.NewHelper(logger).Errorf("集群模式获取键值失败: %v", err)
		return
	}

	log.NewHelper(logger).Infof("集群模式操作成功，值: %s", value)
}

// ExampleAdvancedRedisOperations 展示高级Redis操作
func ExampleAdvancedRedisOperations() {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	ctx := context.Background()

	// 直接使用Redis客户端进行高级操作

	// 1. 使用管道操作
	pipe := client.TxPipeline()
	pipe.Set(ctx, "key1", "value1", time.Hour)
	pipe.Set(ctx, "key2", "value2", time.Hour)
	pipe.Expire(ctx, "key1", time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("管道操作失败: %v", err)
		return
	}

	// 2. 使用Lua脚本
	luaScript := `
		local key = KEYS[1]
		local value = ARGV[1]
		local ttl = ARGV[2]
		redis.call('set', key, value)
		redis.call('expire', key, ttl)
		return redis.call('get', key)
	`

	result, err := client.Eval(ctx, luaScript, []string{"script_key"}, "script_value", 3600).Result()
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("Lua脚本执行失败: %v", err)
		return
	}
	log.NewHelper(log.DefaultLogger).Infof("Lua脚本结果: %v", result)

	// 3. 使用发布订阅
	pubsub := client.Subscribe(ctx, "my_channel")
	defer pubsub.Close()

	// 发布消息
	err = client.Publish(ctx, "my_channel", "Hello Redis!").Err()
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("发布消息失败: %v", err)
		return
	}

	// 接收消息（这里只是示例，实际使用时需要在goroutine中处理）
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("接收消息失败: %v", err)
		return
	}
	log.NewHelper(log.DefaultLogger).Infof("收到消息: %s", msg.Payload)

	log.NewHelper(log.DefaultLogger).Info("高级Redis操作完成")
}

// ExampleErrorHandling 展示错误处理
func ExampleErrorHandling() {
	// 创建Redis客户端（故意使用错误的地址）
	client := redis.NewClient(&redis.Options{
		Addr:        "invalid:6379",
		Password:    "",
		DB:          0,
		DialTimeout: 1 * time.Second, // 短超时时间快速失败
	})
	defer client.Close()

	logger := log.DefaultLogger
	executor := NewStatefulExecutor(client, logger)

	ctx := context.Background()

	// 尝试操作，会因为连接失败而出错
	err := executor.SetServiceState(ctx, "default", "my-service", 1, "Running")
	if err != nil {
		log.NewHelper(logger).Errorf("预期的错误: %v", err)

		// 错误处理逻辑
		// 1. 记录错误
		// 2. 重试逻辑
		// 3. 降级处理
		// 4. 告警通知

		log.NewHelper(logger).Info("执行降级处理...")
		return
	}

	log.NewHelper(logger).Info("错误处理示例完成")
}
