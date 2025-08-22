package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// ==================== Redis管理器 ====================

// RedisManager Redis管理器 - 直接使用go-redis/v9
type RedisManager struct {
	config *RedisConfig
	logger log.Logger
	client *redis.Client
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addresses    []string      `json:"addresses"`    // Redis地址列表
	Password     string        `json:"password"`     // Redis密码
	DB           int           `json:"db"`           // Redis数据库
	PoolSize     int           `json:"poolSize"`     // 连接池大小
	MinIdle      int           `json:"minIdle"`      // 最小空闲连接数
	MaxRetries   int           `json:"maxRetries"`   // 最大重试次数
	DialTimeout  time.Duration `json:"dialTimeout"`  // 连接超时
	ReadTimeout  time.Duration `json:"readTimeout"`  // 读取超时
	WriteTimeout time.Duration `json:"writeTimeout"` // 写入超时
	PoolTimeout  time.Duration `json:"poolTimeout"`  // 连接池超时
	IdleTimeout  time.Duration `json:"idleTimeout"`  // 空闲超时
}

// ==================== Option模式配置 ====================

// Option Redis管理器配置选项
type Option func(*RedisManager)

// WithAddresses 设置Redis地址
func WithAddresses(addresses ...string) Option {
	return func(m *RedisManager) {
		m.config.Addresses = addresses
	}
}

// WithPassword 设置Redis密码
func WithPassword(password string) Option {
	return func(m *RedisManager) {
		m.config.Password = password
	}
}

// WithDB 设置Redis数据库
func WithDB(db int) Option {
	return func(m *RedisManager) {
		m.config.DB = db
	}
}

// WithPoolSize 设置连接池大小
func WithPoolSize(poolSize int) Option {
	return func(m *RedisManager) {
		m.config.PoolSize = poolSize
	}
}

// WithMinIdle 设置最小空闲连接数
func WithMinIdle(minIdle int) Option {
	return func(m *RedisManager) {
		m.config.MinIdle = minIdle
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(m *RedisManager) {
		m.config.MaxRetries = maxRetries
	}
}

// WithDialTimeout 设置连接超时
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(m *RedisManager) {
		m.config.DialTimeout = dialTimeout
	}
}

// WithReadTimeout 设置读取超时
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(m *RedisManager) {
		m.config.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout 设置写入超时
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(m *RedisManager) {
		m.config.WriteTimeout = writeTimeout
	}
}

// WithPoolTimeout 设置连接池超时
func WithPoolTimeout(poolTimeout time.Duration) Option {
	return func(m *RedisManager) {
		m.config.PoolTimeout = poolTimeout
	}
}

// WithIdleTimeout 设置空闲超时
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(m *RedisManager) {
		m.config.IdleTimeout = idleTimeout
	}
}

// WithLogger 设置日志器
func WithLogger(logger log.Logger) Option {
	return func(m *RedisManager) {
		m.logger = logger
	}
}

// WithConfig 使用完整配置
func WithConfig(config *RedisConfig) Option {
	return func(m *RedisManager) {
		if config != nil {
			m.config = config
		}
	}
}

// ==================== 构造函数 ====================

// NewRedisManager 创建新的Redis管理器
func NewRedisManager(opts ...Option) (*RedisManager, error) {
	// 创建默认配置
	manager := &RedisManager{
		config: &RedisConfig{
			Addresses:    []string{"localhost:6379"},
			Password:     "",
			DB:           0,
			PoolSize:     10,
			MinIdle:      5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
			IdleTimeout:  5 * time.Minute,
		},
		logger: log.DefaultLogger,
	}

	// 应用选项
	for _, opt := range opts {
		opt(manager)
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:            strings.Join(manager.config.Addresses, ","),
		Password:        manager.config.Password,
		DB:              manager.config.DB,
		PoolSize:        manager.config.PoolSize,
		MinIdleConns:    manager.config.MinIdle,
		MaxRetries:      manager.config.MaxRetries,
		DialTimeout:     manager.config.DialTimeout,
		ReadTimeout:     manager.config.ReadTimeout,
		WriteTimeout:    manager.config.WriteTimeout,
		PoolTimeout:     manager.config.PoolTimeout,
		ConnMaxIdleTime: manager.config.IdleTimeout,
	})

	manager.client = client

	// 测试连接
	if err := manager.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("Redis连接测试失败: %w", err)
	}

	log.NewHelper(manager.logger).Info("Redis管理器创建成功")
	return manager, nil
}

// ==================== 基础操作 ====================

// Get 获取键值
func (m *RedisManager) Get(ctx context.Context, key string) (string, error) {
	return m.client.Get(ctx, key).Result()
}

// Set 设置键值
func (m *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除键
func (m *RedisManager) Del(ctx context.Context, keys ...string) error {
	return m.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (m *RedisManager) Exists(ctx context.Context, keys ...string) (int64, error) {
	return m.client.Exists(ctx, keys...).Result()
}

// ==================== Hash操作 ====================

// HGet 获取Hash字段值
func (m *RedisManager) HGet(ctx context.Context, key, field string) (string, error) {
	return m.client.HGet(ctx, key, field).Result()
}

// HSet 设置Hash字段值
func (m *RedisManager) HSet(ctx context.Context, key string, values ...interface{}) error {
	return m.client.HSet(ctx, key, values...).Err()
}

// HGetAll 获取Hash所有字段
func (m *RedisManager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return m.client.HGetAll(ctx, key).Result()
}

// HDel 删除Hash字段
func (m *RedisManager) HDel(ctx context.Context, key string, fields ...string) error {
	return m.client.HDel(ctx, key, fields...).Err()
}

// ==================== 列表操作 ====================

// LPush 左推入列表
func (m *RedisManager) LPush(ctx context.Context, key string, values ...interface{}) error {
	return m.client.LPush(ctx, key, values...).Err()
}

// RPush 右推入列表
func (m *RedisManager) RPush(ctx context.Context, key string, values ...interface{}) error {
	return m.client.RPush(ctx, key, values...).Err()
}

// LPop 左弹出列表
func (m *RedisManager) LPop(ctx context.Context, key string) (string, error) {
	return m.client.LPop(ctx, key).Result()
}

// RPop 右弹出列表
func (m *RedisManager) RPop(ctx context.Context, key string) (string, error) {
	return m.client.RPop(ctx, key).Result()
}

// ==================== 集合操作 ====================

// SAdd 添加集合成员
func (m *RedisManager) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return m.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除集合成员
func (m *RedisManager) SRem(ctx context.Context, key string, members ...interface{}) error {
	return m.client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (m *RedisManager) SMembers(ctx context.Context, key string) ([]string, error) {
	return m.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (m *RedisManager) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return m.client.SIsMember(ctx, key, member).Result()
}

// ==================== 有序集合操作 ====================

// ZAdd 添加有序集合成员
func (m *RedisManager) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return m.client.ZAdd(ctx, key, members...).Err()
}

// ZRem 移除有序集合成员
func (m *RedisManager) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return m.client.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围
func (m *RedisManager) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return m.client.ZRange(ctx, key, start, stop).Result()
}

// ZScore 获取有序集合成员分数
func (m *RedisManager) ZScore(ctx context.Context, key, member string) (float64, error) {
	return m.client.ZScore(ctx, key, member).Result()
}

// ==================== 过期时间操作 ====================

// Expire 设置过期时间
func (m *RedisManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return m.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余生存时间
func (m *RedisManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return m.client.TTL(ctx, key).Result()
}

// ==================== 事务操作 ====================

// TxPipeline 获取事务管道
func (m *RedisManager) TxPipeline() redis.Pipeliner {
	return m.client.TxPipeline()
}

// ==================== 管理操作 ====================

// Close 关闭Redis管理器
func (m *RedisManager) Close() error {
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			log.NewHelper(m.logger).Errorf("关闭Redis客户端失败: %v", err)
			return err
		}
		log.NewHelper(m.logger).Info("Redis管理器已关闭")
	}
	return nil
}

// Ping 测试连接
func (m *RedisManager) Ping(ctx context.Context) error {
	return m.client.Ping(ctx).Err()
}

// PoolStats 获取连接池统计
func (m *RedisManager) PoolStats() *redis.PoolStats {
	return m.client.PoolStats()
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (m *RedisManager) GetClient() *redis.Client {
	return m.client
}
