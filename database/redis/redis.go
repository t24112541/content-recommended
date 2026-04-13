package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASS"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

func NewRedisConnection() (conf RedisConfig, err error) {
	err = env.Parse(&conf)
	return
}

func (conf *RedisConfig) Connect() (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
