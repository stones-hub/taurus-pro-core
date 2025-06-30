package common

import (
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

func ProvideCronComponent(cfg *config.Config) (*cron.CronManager, error) {
	enable := cfg.GetBool("cron.enable")
	if !enable {
		return nil, nil
	}

	location, err := time.LoadLocation(cfg.GetString("cron.location"))
	if err != nil {
		location, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return nil, err
		}
	}

	cronOptions := []cron.Option{cron.WithLocation(location)}

	if cfg.GetBool("cron.enable_seconds") {
		cronOptions = append(cronOptions, cron.WithSeconds())
	}

	concurrencyMode := cron.ConcurrencyMode(cfg.GetInt("cron.concurrency_mode"))
	cronOptions = append(cronOptions, cron.WithConcurrencyMode(concurrencyMode))

	return cron.New(cronOptions...), nil
}

var CommonComponent = types.Component{
	Name:         "common",
	Package:      "github.com/stones-hub/taurus-pro-common",
	Version:      "v0.0.1",
	Description:  "通用组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire: &types.Wire{
		RequirePath:  "github.com/stones-hub/taurus-pro-common/pkg/cron",
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

return cm, func() {
	cm.Stop()
}, nil
}`,
	},
}
