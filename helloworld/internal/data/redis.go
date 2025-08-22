package data

import (
	"github.com/redis/go-redis/v9"
	"helloworld/internal/conf"
)

func NewRedis(dataCnf *conf.Data) *redis.Client {
	cfg := dataCnf.Redis
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       int(cfg.Db),
	})
}
