package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type (
	Config struct {
		Addr        string        `yaml:"addr"`
		Password    string        `yaml:"password"`
		User        string        `yaml:"user"`
		DB          int           `yaml:"db"`
		MaxRetries  int           `yaml:"max_retries"`
		DialTimeout time.Duration `yaml:"dial_timeout"`
		Timeout     time.Duration `yaml:"timeout"`
	}
)

func NewClient(ctx context.Context, cnf Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cnf.Addr,
		Password:     cnf.Password,
		DB:           cnf.DB,
		Username:     cnf.User,
		MaxRetries:   cnf.MaxRetries,
		DialTimeout:  cnf.DialTimeout,
		ReadTimeout:  cnf.Timeout,
		WriteTimeout: cnf.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server %s/n", err.Error())
		return nil, err
	}

	return db, nil
}
