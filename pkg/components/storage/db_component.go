package storage

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-storage/pkg/db"
	"gorm.io/gorm"
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

	options := make([]db.DbOptions, 0)
	for _, dbOptions := range dbOptionsList {
		options = append(options, db.NewDbOptions(
			dbOptions["dbname"].(string),
			dbOptions["dbtype"].(string),
			dbOptions["dsn"].(string),
			db.WithMaxOpenConns(dbOptions["max_open_conns"].(int)),
			db.WithMaxIdleConns(dbOptions["max_idle_conns"].(int)),
			db.WithConnMaxLifetime(time.Duration(dbOptions["conn_max_lifetime"].(int))*time.Second),
			db.WithMaxRetries(dbOptions["max_retries"].(int)),
			db.WithRetryDelay(dbOptions["retry_delay"].(int)),
			db.WithLoggerName(dbOptions["logger_name"].(string)),
		))
	}

	err := db.InitDB(options...)
	if err != nil {
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Database all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return db.DbList(), func() {
		db.CloseDB()
		log.Printf("%s🔗 -> Clean up database components successfully. %s\n", "\033[32m", "\033[0m")

	}, nil
}

var dbWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-storage/pkg/db", "gorm.io/gorm", "log"},
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

	options := make([]db.DbOptions, 0)
	for _, dbOptions := range dbOptionsList {
		options = append(options, db.NewDbOptions(
			dbOptions["dbname"].(string),
			dbOptions["dbtype"].(string),
			dbOptions["dsn"].(string),
			db.WithMaxOpenConns(dbOptions["max_open_conns"].(int)),
			db.WithMaxIdleConns(dbOptions["max_idle_conns"].(int)),
			db.WithConnMaxLifetime(time.Duration(dbOptions["conn_max_lifetime"].(int))*time.Second),
			db.WithMaxRetries(dbOptions["max_retries"].(int)),
			db.WithRetryDelay(dbOptions["retry_delay"].(int)),
			db.WithLoggerName(dbOptions["logger_name"].(string)),
		))
	}

	err := db.InitDB(options...)
	if err != nil {
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Database all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return db.DbList(), func() {
		db.CloseDB()
		log.Printf("%s🔗 -> Clean up database components successfully. %s\n", "\033[32m", "\033[0m")

	}, nil
}
`,
}
