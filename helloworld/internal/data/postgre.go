package data

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"helloworld/internal/conf"
)

// NewDB initializes a PostgreSQL connection via GORM
func NewPostgresDB(c *conf.Data) (*gorm.DB, error) {
	database := c.Database
	if database.Driver != "postgres" {
		return nil, fmt.Errorf("unsupported driver: %s", database.Driver)
	}

	db, err := gorm.Open(postgres.Open(database.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
