package common

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

func ProvideCronComponent(cfg *config.Config) (*cron.CronManager, func(), error) {
	enable := cfg.GetBool("cron.enable")
	if !enable {
		return nil, nil, nil
	}

	location, err := time.LoadLocation(cfg.GetString("cron.location"))
	if err != nil {
		location, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return nil, nil, err
		}
	}

	cronOptions := []cron.Option{cron.WithLocation(location)}

	if cfg.GetBool("cron.enable_seconds") {
		cronOptions = append(cronOptions, cron.WithSeconds())
	}

	concurrencyMode := cron.ConcurrencyMode(cfg.GetInt("cron.concurrency_mode"))
	cronOptions = append(cronOptions, cron.WithConcurrencyMode(concurrencyMode))
	cm := cron.New(cronOptions...)

	log.Printf("%s🔗 -> Cron all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return cm, func() {
		cm.Stop()
		log.Printf("%s🔗 -> Clean up cron components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}

var cronWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/cron", "time", "log"},
	Name:         "Cron",
	Type:         "*cron.CronManager",
	ProviderName: "ProvideCronComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
enable := cfg.GetBool("cron.enable")
if !enable {
return nil, nil, nil
}

location, err := time.LoadLocation(cfg.GetString("cron.location"))
if err != nil {
location, err = time.LoadLocation("Asia/Shanghai")
if err != nil {
	return nil, nil, err
}
}

cronOptions := []cron.Option{cron.WithLocation(location)}

if cfg.GetBool("cron.enable_seconds") {
cronOptions = append(cronOptions, cron.WithSeconds())
}

concurrencyMode := cron.ConcurrencyMode(cfg.GetInt("cron.concurrency_mode"))
cronOptions = append(cronOptions, cron.WithConcurrencyMode(concurrencyMode))
cm := cron.New(cronOptions...)

log.Printf("%s🔗 -> Cron all initialized successfully. %s\n", "\033[32m", "\033[0m")

return cm, func() {
cm.Stop()
log.Printf("%s🔗 -> Clean up cron components successfully. %s\n", "\033[32m", "\033[0m")
}, nil
}`,
}

var CommonComponent = types.Component{
	Name:         "common",
	Package:      "github.com/stones-hub/taurus-pro-common",
	Version:      "v0.0.2",
	Description:  "通用组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{cronWire},
}
