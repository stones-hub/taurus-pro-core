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

	levelStr := cfg.GetString("redis.logger_level")
	level := redisx.LogLevelInfo
	switch levelStr {
	case "debug":
		level = redisx.LogLevelDebug
	case "info":
		level = redisx.LogLevelInfo
	case "warn":
		level = redisx.LogLevelWarn
	case "error":
		level = redisx.LogLevelError
	default:
		level = redisx.LogLevelInfo
	}

	formatterStr := cfg.GetString("redis.logger_fomatter")
	formatter := redisx.JSONLogFormatter
	switch formatterStr {
	case "default":
		formatter = redisx.DefaultLogFormatter
	case "json":
		formatter = redisx.JSONLogFormatter
	default:
		formatter = redisx.DefaultLogFormatter
	}

	logger, err := redisx.NewRedisLogger(
		redisx.WithLogFilePath(cfg.GetString("redis.logger_path")),
		redisx.WithLogLevel(level),
		redisx.WithLogFormatter(formatter),
		redisx.WithLogMaxSize(cfg.GetInt("redis.logger_max_size")),
		redisx.WithLogMaxBackups(cfg.GetInt("redis.logger_max_backups")),
		redisx.WithLogMaxAge(cfg.GetInt("redis.logger_max_age")),
	)
	if err != nil {
		return nil, func() {}, err
	}

	err = redisx.InitRedis(
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
		redisx.WithLogging(logger),
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
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-storage/pkg/redisx", "log", "time"},
	Name:         "Redis",
	Type:         "*redisx.RedisClient",
	ProviderName: "ProvideRedisComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {

		enable := cfg.GetBool("redis.enable")
	if !enable {
		return nil, func() {}, nil
	}

	address := cfg.GetStringSlice("redis.address")

	levelStr := cfg.GetString("redis.logger_level")
	level := redisx.LogLevelInfo
	switch levelStr {
	case "debug":
		level = redisx.LogLevelDebug
	case "info":
		level = redisx.LogLevelInfo
	case "warn":
		level = redisx.LogLevelWarn
	case "error":
		level = redisx.LogLevelError
	default:
		level = redisx.LogLevelInfo
	}

	formatterStr := cfg.GetString("redis.logger_fomatter")
	formatter := redisx.JSONLogFormatter
	switch formatterStr {
	case "default":
		formatter = redisx.DefaultLogFormatter
	case "json":
		formatter = redisx.JSONLogFormatter
	default:
		formatter = redisx.DefaultLogFormatter
	}

	logger, err := redisx.NewRedisLogger(
		redisx.WithLogFilePath(cfg.GetString("redis.logger_path")),
		redisx.WithLogLevel(level),
		redisx.WithLogFormatter(formatter),
		redisx.WithLogMaxSize(cfg.GetInt("redis.logger_max_size")),
		redisx.WithLogMaxBackups(cfg.GetInt("redis.logger_max_backups")),
		redisx.WithLogMaxAge(cfg.GetInt("redis.logger_max_age")),
	)
	if err != nil {
		return nil, func() {}, err
	}

	err = redisx.InitRedis(
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
		redisx.WithLogging(logger),
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
