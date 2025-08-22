package data

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"helloworld/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewPostgresDB, NewRedis)

// Data .
type Data struct {
	DB    *gorm.DB
	Cache *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, cache *redis.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db, cache}, cleanup, nil
}
