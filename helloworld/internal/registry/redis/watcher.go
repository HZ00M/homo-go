package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/redis/go-redis/v9"
)

type redisWatcher struct {
	client      *redis.Client
	serviceName string
	ticker      *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
}

func newRedisWatcher(ctx context.Context, client *redis.Client, serviceName string) registry.Watcher {
	c, cancel := context.WithCancel(ctx)
	return &redisWatcher{
		client:      client,
		serviceName: serviceName,
		ticker:      time.NewTicker(2 * time.Second),
		ctx:         c,
		cancel:      cancel,
	}
}

func (w *redisWatcher) Next() ([]*registry.ServiceInstance, error) {
	for {
		select {
		case <-w.ctx.Done():
			return nil, w.ctx.Err()
		case <-w.ticker.C:
			keys, err := w.client.Keys(w.ctx, redisPrefix+":"+w.serviceName+":*").Result()
			if err != nil {
				return nil, err
			}
			var services []*registry.ServiceInstance
			for _, key := range keys {
				val, err := w.client.Get(w.ctx, key).Result()
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
	}
}

func (w *redisWatcher) Stop() error {
	w.cancel()
	w.ticker.Stop()
	return nil
}
