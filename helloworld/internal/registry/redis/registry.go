package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/redis/go-redis/v9"
)

const redisPrefix = "kratos:service"

var (
	_ registry.Registrar = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

type Registry struct {
	client  *redis.Client
	timeout time.Duration
}

func NewRegistry(client *redis.Client) *Registry {
	return &Registry{
		client:  client,
		timeout: 5 * time.Second,
	}
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	data, err := json.Marshal(service)
	if err != nil {
		return err
	}
	key := serviceKey(service)
	return r.client.Set(ctx, key, data, 10*time.Second).Err()
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	key := serviceKey(service)
	return r.client.Del(ctx, key).Err()
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	keys, err := r.client.Keys(ctx, fmt.Sprintf("%s:%s:*", redisPrefix, serviceName)).Result()
	if err != nil {
		return nil, err
	}

	var services []*registry.ServiceInstance
	for _, key := range keys {
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		var inst registry.ServiceInstance
		if err := json.Unmarshal([]byte(val), &inst); err != nil {
			continue
		}
		services = append(services, &inst)
	}
	return services, nil
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newRedisWatcher(ctx, r.client, serviceName), nil
}

func serviceKey(service *registry.ServiceInstance) string {
	return fmt.Sprintf("%s:%s:%s", redisPrefix, service.Name, service.ID)
}
