package storage

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-storage/pkg/redisx"
)

func ProvideRedisComponent(cfg *config.Config) (*redisx.RedisClient, func(), error) {

	enable := cfg.GetBool("redis.enable")
	if !enable {
		return nil, func() {}, nil
	}

	address := cfg.GetStringSlice("redis.address")

	err := redisx.InitRedis(
		redisx.WithAddrs(address...),
		redisx.WithPassword(cfg.GetString("redis.password")),
		redisx.WithDB(cfg.GetInt("redis.db")),
		redisx.WithPoolSize(cfg.GetInt("redis.pool_size")),
		redisx.WithMinIdleConns(cfg.GetInt("redis.min_idle_conns")),
		redisx.WithTimeout(
			time.Duration(cfg.GetInt("redis.dial_timeout"))*time.Second,
			time.Duration(cfg.GetInt("redis.read_timeout"))*time.Second,
			time.Duration(cfg.GetInt("redis.write_timeout"))*time.Second),
		redisx.WithMaxRetries(cfg.GetInt("redis.max_retries")),
	)

	if err != nil {
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Redis all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return redisx.Redis, func() {
		redisx.Redis.Close()
		log.Printf("%s🔗 -> Clean up redis components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}

var redisWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-storage/pkg/redisx", "log"},
	Name:         "Redis",
	Type:         "*redisx.RedisClient",
	ProviderName: "ProvideRedisComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) (*redisx.RedisClient,func(), error) {

	enable := cfg.GetBool("redis.enable")
	if !enable {
		return nil, func() {}, nil
	}

	address := cfg.GetStringSlice("redis.address")

	err := redisx.InitRedis(
		redisx.WithAddrs(address...),
		redisx.WithPassword(cfg.GetString("redis.password")),
		redisx.WithDB(cfg.GetInt("redis.db")),
		redisx.WithPoolSize(cfg.GetInt("redis.pool_size")),
		redisx.WithMinIdleConns(cfg.GetInt("redis.min_idle_conns")),
		redisx.WithTimeout(
			time.Duration(cfg.GetInt("redis.dial_timeout"))*time.Second,
			time.Duration(cfg.GetInt("redis.read_timeout"))*time.Second,
			time.Duration(cfg.GetInt("redis.write_timeout"))*time.Second),
		redisx.WithMaxRetries(cfg.GetInt("redis.max_retries")),
	)

	if err != nil {
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Redis all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return redisx.Redis, func() {
		redisx.Redis.Close()
		log.Printf("%s🔗 -> Clean up redis components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}`,
}
