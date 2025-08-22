package executor

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func TestNewRedisManager(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "默认配置",
			opts:    []Option{},
			wantErr: false,
		},
		{
			name: "自定义地址",
			opts: []Option{
				WithAddresses("custom-redis:6379"),
			},
			wantErr: false,
		},
		{
			name: "完整配置",
			opts: []Option{
				WithAddresses("redis-server:6379"),
				WithPassword("mypassword"),
				WithDB(1),
				WithPoolSize(20),
				WithMinIdle(10),
				WithMaxRetries(5),
				WithDialTimeout(10 * time.Second),
				WithReadTimeout(5 * time.Second),
				WithWriteTimeout(5 * time.Second),
				WithPoolTimeout(8 * time.Second),
				WithIdleTimeout(10 * time.Minute),
				WithLogger(log.DefaultLogger),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewRedisManager(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				defer manager.Close()

				// 测试基本连接
				if err := manager.Ping(context.Background()); err != nil {
					t.Errorf("Ping() failed: %v", err)
				}
			}
		})
	}
}

func TestRedisManagerOptions(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	// 测试单个选项
	manager, err := NewRedisManager(
		WithAddresses("test:6379"),
		WithPassword("testpass"),
		WithDB(2),
	)
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	// 验证配置是否正确应用
	if manager.config.Addresses[0] != "test:6379" {
		t.Errorf("WithAddresses not applied correctly")
	}
	if manager.config.Password != "testpass" {
		t.Errorf("WithPassword not applied correctly")
	}
	if manager.config.DB != 2 {
		t.Errorf("WithDB not applied correctly")
	}
}

func TestRedisManagerDefaultConfig(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	// 验证默认配置
	if len(manager.config.Addresses) == 0 || manager.config.Addresses[0] != "localhost:6379" {
		t.Error("默认地址配置错误")
	}
	if manager.config.PoolSize != 10 {
		t.Error("默认连接池大小配置错误")
	}
	if manager.config.MinIdle != 5 {
		t.Error("默认最小空闲连接数配置错误")
	}
}

func TestRedisManagerWithConfig(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	customConfig := &RedisConfig{
		Addresses:    []string{"custom:6379"},
		Password:     "custompass",
		DB:           5,
		PoolSize:     25,
		MinIdle:      12,
		MaxRetries:   7,
		DialTimeout:  8 * time.Second,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		PoolTimeout:  6 * time.Second,
		IdleTimeout:  8 * time.Minute,
	}

	manager, err := NewRedisManager(WithConfig(customConfig))
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	// 验证配置是否正确应用
	if manager.config.Addresses[0] != "custom:6379" {
		t.Error("自定义配置地址未正确应用")
	}
	if manager.config.Password != "custompass" {
		t.Error("自定义配置密码未正确应用")
	}
	if manager.config.DB != 5 {
		t.Error("自定义配置数据库未正确应用")
	}
}

func TestRedisManagerClose(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}

	// 测试关闭
	if err := manager.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestRedisManagerPoolStats(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	// 获取连接池统计
	stats := manager.PoolStats()
	if stats == nil {
		t.Error("PoolStats() returned nil")
	}

	// 验证统计字段存在
	_ = stats.TotalConns
	_ = stats.IdleConns
	_ = stats.StaleConns
}

func TestRedisManagerBasicOperations(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// 测试Set
	if err := manager.Set(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Set() failed: %v", err)
	}

	// 测试Get
	result, err := manager.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if result != value {
		t.Errorf("Get() returned %s, want %s", result, value)
	}

	// 测试Exists
	exists, err := manager.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if exists != 1 {
		t.Errorf("Exists() returned %d, want 1", exists)
	}

	// 测试Del
	if err := manager.Del(ctx, key); err != nil {
		t.Errorf("Del() failed: %v", err)
	}

	// 验证删除成功
	exists, err = manager.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists() after delete failed: %v", err)
	}
	if err == nil && exists != 0 {
		t.Errorf("Key still exists after delete")
	}
}

func TestRedisManagerHashOperations(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()
	key := "test_hash"
	field := "test_field"
	value := "test_value"

	// 测试HSet
	if err := manager.HSet(ctx, key, field, value); err != nil {
		t.Errorf("HSet() failed: %v", err)
	}

	// 测试HGet
	result, err := manager.HGet(ctx, key, field)
	if err != nil {
		t.Errorf("HGet() failed: %v", err)
	}
	if result != value {
		t.Errorf("HGet() returned %s, want %s", result, value)
	}

	// 测试HGetAll
	all, err := manager.HGetAll(ctx, key)
	if err != nil {
		t.Errorf("HGetAll() failed: %v", err)
	}
	if all[field] != value {
		t.Errorf("HGetAll() returned %s, want %s", all[field], value)
	}

	// 清理
	manager.Del(ctx, key)
}

func TestRedisManagerGetClient(t *testing.T) {
	// 跳过需要Redis连接的测试
	t.Skip("跳过需要Redis连接的测试")

	manager, err := NewRedisManager()
	if err != nil {
		t.Fatalf("NewRedisManager() failed: %v", err)
	}
	defer manager.Close()

	// 获取原始客户端
	client := manager.GetClient()
	if client == nil {
		t.Error("GetClient() returned nil")
	}

	// 验证客户端可用
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Errorf("Raw client Ping() failed: %v", err)
	}
}

// TestRedisManagerConfigValidation 测试配置验证逻辑
func TestRedisManagerConfigValidation(t *testing.T) {
	// 测试配置选项函数
	config := &RedisConfig{}

	// 测试WithAddresses
	WithAddresses("test1:6379", "test2:6379")(&RedisManager{config: config})
	if len(config.Addresses) != 2 || config.Addresses[0] != "test1:6379" {
		t.Error("WithAddresses not working correctly")
	}

	// 测试WithPassword
	WithPassword("testpass")(&RedisManager{config: config})
	if config.Password != "testpass" {
		t.Error("WithPassword not working correctly")
	}

	// 测试WithDB
	WithDB(5)(&RedisManager{config: config})
	if config.DB != 5 {
		t.Error("WithDB not working correctly")
	}

	// 测试WithPoolSize
	WithPoolSize(25)(&RedisManager{config: config})
	if config.PoolSize != 25 {
		t.Error("WithPoolSize not working correctly")
	}

	// 测试WithMinIdle
	WithMinIdle(15)(&RedisManager{config: config})
	if config.MinIdle != 15 {
		t.Error("WithMinIdle not working correctly")
	}

	// 测试WithMaxRetries
	WithMaxRetries(7)(&RedisManager{config: config})
	if config.MaxRetries != 7 {
		t.Error("WithMaxRetries not working correctly")
	}

	// 测试WithDialTimeout
	dialTimeout := 10 * time.Second
	WithDialTimeout(dialTimeout)(&RedisManager{config: config})
	if config.DialTimeout != dialTimeout {
		t.Error("WithDialTimeout not working correctly")
	}

	// 测试WithReadTimeout
	readTimeout := 5 * time.Second
	WithReadTimeout(readTimeout)(&RedisManager{config: config})
	if config.ReadTimeout != readTimeout {
		t.Error("WithReadTimeout not working correctly")
	}

	// 测试WithWriteTimeout
	writeTimeout := 5 * time.Second
	WithWriteTimeout(writeTimeout)(&RedisManager{config: config})
	if config.WriteTimeout != writeTimeout {
		t.Error("WithWriteTimeout not working correctly")
	}

	// 测试WithPoolTimeout
	poolTimeout := 8 * time.Second
	WithPoolTimeout(poolTimeout)(&RedisManager{config: config})
	if config.PoolTimeout != poolTimeout {
		t.Error("WithPoolTimeout not working correctly")
	}

	// 测试WithIdleTimeout
	idleTimeout := 10 * time.Minute
	WithIdleTimeout(idleTimeout)(&RedisManager{config: config})
	if config.IdleTimeout != idleTimeout {
		t.Error("WithIdleTimeout not working correctly")
	}

	// 测试WithLogger
	logger := log.DefaultLogger
	WithLogger(logger)(&RedisManager{logger: nil})
	// 这里我们只是测试函数调用，不验证结果

	// 测试WithConfig
	newConfig := &RedisConfig{Addresses: []string{"new:6379"}}
	WithConfig(newConfig)(&RedisManager{config: config})
	if config.Addresses[0] != "new:6379" {
		t.Error("WithConfig not working correctly")
	}
}
