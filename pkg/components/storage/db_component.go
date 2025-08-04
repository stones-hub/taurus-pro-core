package storage

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-storage/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ProvideDbComponent(cfg *config.Config) (map[string]*gorm.DB, func(), error) {
	enable := cfg.GetBool("databases.enable")

	if !enable {
		return nil, func() {}, nil
	}

	rawList := cfg.Get("databases.list").([]interface{})
	dbOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		dbOptionsList[i] = raw.(map[string]interface{})
	}

	for _, dbOptions := range dbOptionsList {
		// 日志地址
		logPath := dbOptions["log_path"].(string)
		if logPath == "" {
			logPath = "./logs/db/db.log"
		}

		// 日志级别
		logLevel := dbOptions["log_level"].(string)
		level := logger.Info
		switch logLevel {
		case "info":
			level = logger.Info
		case "warn":
			level = logger.Warn
		case "error":
			level = logger.Error
		case "silent":
			level = logger.Silent
		}

		// 日志格式
		logFormatter := dbOptions["log_formatter"].(string)
		formatter := db.DefaultLogFormatter
		switch logFormatter {
		case "json":
			formatter = db.JSONLogFormatter
		case "default":
			formatter = db.DefaultLogFormatter
		}

		err := db.InitDB(db.WithMaxOpenConns(dbOptions["max_open_conns"].(int)),
			db.WithMaxIdleConns(dbOptions["max_idle_conns"].(int)),
			db.WithConnMaxLifetime(time.Duration(dbOptions["conn_max_lifetime"].(int))*time.Second),
			db.WithMaxRetries(dbOptions["max_retries"].(int)),
			db.WithRetryDelay(dbOptions["retry_delay"].(int)),
			db.WithDBName(dbOptions["dbname"].(string)),
			db.WithDBType(dbOptions["dbtype"].(string)),
			db.WithDSN(dbOptions["dsn"].(string)),
			db.WithLogger(db.NewDbLogger(
				db.WithLogFilePath(logPath),
				db.WithLogLevel(level),
				db.WithLogFormatter(formatter))),
		)
		if err != nil {
			return nil, func() {}, err
		}
	}

	log.Printf("%s🔗 -> Database all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return db.DbList(), func() {
		db.CloseDB()
		log.Printf("%s🔗 -> Clean up database components successfully. %s\n", "\033[32m", "\033[0m")

	}, nil
}

var dbWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-storage/pkg/db", "gorm.io/gorm", "gorm.io/gorm/logger", "log", "time"},
	Name:         "DbList",
	Type:         "map[string]*gorm.DB",
	ProviderName: "ProvideDbComponent",
	Provider: `
	func {{.ProviderName}}(cfg *config.Config) (map[string]*gorm.DB, func(), error) {

		enable := cfg.GetBool("databases.enable")

	if !enable {
		return nil, func() {}, nil
	}

	rawList := cfg.Get("databases.list").([]interface{})
	dbOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		dbOptionsList[i] = raw.(map[string]interface{})
	}

	for _, dbOptions := range dbOptionsList {
		// 日志地址
		logPath := dbOptions["log_path"].(string)
		if logPath == "" {
			logPath = "./logs/db/db.log"
		}

		// 日志级别
		logLevel := dbOptions["log_level"].(string)
		level := logger.Info
		switch logLevel {
		case "info":
			level = logger.Info
		case "warn":
			level = logger.Warn
		case "error":
			level = logger.Error
		case "silent":
			level = logger.Silent
		}

		// 日志格式
		logFormatter := dbOptions["log_formatter"].(string)
		formatter := db.DefaultLogFormatter
		switch logFormatter {
		case "json":
			formatter = db.JSONLogFormatter
		case "default":
			formatter = db.DefaultLogFormatter
		}

		err := db.InitDB(db.WithMaxOpenConns(dbOptions["max_open_conns"].(int)),
			db.WithMaxIdleConns(dbOptions["max_idle_conns"].(int)),
			db.WithConnMaxLifetime(time.Duration(dbOptions["conn_max_lifetime"].(int))*time.Second),
			db.WithMaxRetries(dbOptions["max_retries"].(int)),
			db.WithRetryDelay(dbOptions["retry_delay"].(int)),
			db.WithDBName(dbOptions["dbname"].(string)),
			db.WithDBType(dbOptions["dbtype"].(string)),
			db.WithDSN(dbOptions["dsn"].(string)),
			db.WithLogger(db.NewDbLogger(
				db.WithLogFilePath(logPath),
				db.WithLogLevel(level),
				db.WithLogFormatter(formatter))),
		)
		if err != nil {
			return nil, func() {}, err
		}
	}

	log.Printf("%s🔗 -> Database all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return db.DbList(), func() {
		db.CloseDB()
		log.Printf("%s🔗 -> Clean up database components successfully. %s\n", "\033[32m", "\033[0m")

	}, nil
	
}
`,
}
