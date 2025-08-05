package common

import (
	"context"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
	"github.com/stones-hub/taurus-pro-common/pkg/cron"
	"github.com/stones-hub/taurus-pro-common/pkg/hook"
	"github.com/stones-hub/taurus-pro-common/pkg/logx"
	"github.com/stones-hub/taurus-pro-common/pkg/templates"
	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

var CommonComponent = types.Component{
	Name:         "common",
	Package:      "github.com/stones-hub/taurus-pro-common",
	Version:      "v0.1.13",
	Description:  "通用基础组件，包含定时任务、日志、模板工具等",
	IsCustom:     true,
	Required:     true,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{cronWire, loggerWire, templateWire, hookWire, cmdWire},
}

var cronWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/cron", "time", "log"},
	Name:         "Cron",
	Type:         "*cron.CronManager",
	ProviderName: "ProvideCronComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
enable := cfg.GetBool("cron.enable")
if !enable {
return nil, func() {}, nil
}

location, err := time.LoadLocation(cfg.GetString("cron.location"))
if err != nil {
location, err = time.LoadLocation("Asia/Shanghai")
if err != nil {
	return nil, func() {}, err
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
cm.GracefulStop(time.Second * 3)
log.Printf("%s🔗 -> Clean up cron components successfully. %s\n", "\033[32m", "\033[0m")
}, nil
}`,
}

func ProvideCronComponent(cfg *config.Config) (*cron.CronManager, func(), error) {
	enable := cfg.GetBool("cron.enable")
	if !enable {
		return nil, func() {}, nil
	}

	location, err := time.LoadLocation(cfg.GetString("cron.location"))
	if err != nil {
		location, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return nil, func() {}, err
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
		cm.GracefulStop(time.Second * 3)
		log.Printf("%s🔗 -> Clean up cron components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}

var loggerWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/logx"},
	Name:         "Logger",
	Type:         "*logx.Manager",
	ProviderName: "ProvideLoggerComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {

	rawList := cfg.Get("loggers").([]interface{})
	loggerOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		loggerOptionsList[i] = raw.(map[string]interface{})
	}

	options := make([]logx.LoggerOptions, 0)
	for _, opts := range loggerOptionsList {
		options = append(options, logx.LoggerOptions{
			Name:       opts["name"].(string),
			Prefix:     opts["prefix"].(string),
			FilePath:   opts["log_file_path"].(string),
			MaxSize:    opts["max_size"].(int),
			MaxBackups: opts["max_backups"].(int),
			MaxAge:     opts["max_age"].(int),
			Compress:   opts["compress"].(bool),
			Formatter:  opts["formatter"].(string),
			Level:      logx.Level(opts["log_level"].(int)),
			Output:     logx.OutputType(opts["output_type"].(string)),
		})
	}

	manager, cleanup, err := logx.BuildManager(options...)
	if err != nil {
		log.Printf("%s🔗 -> Log all initialized failed. %s\n", "\033[31m", "\033[0m")
	} else {
		log.Printf("%s🔗 -> Log all initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

	return manager, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up log components successfully. %s\n", "\033[32m", "\033[0m")
	}, err
}
`,
}

func ProvideLoggerComponent(cfg *config.Config) (*logx.Manager, func(), error) {

	rawList := cfg.Get("loggers").([]interface{})
	loggerOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		loggerOptionsList[i] = raw.(map[string]interface{})
	}

	options := make([]logx.LoggerOptions, 0)
	for _, opts := range loggerOptionsList {
		options = append(options, logx.LoggerOptions{
			Name:       opts["name"].(string),
			Prefix:     opts["prefix"].(string),
			FilePath:   opts["log_file_path"].(string),
			MaxSize:    opts["max_size"].(int),
			MaxBackups: opts["max_backups"].(int),
			MaxAge:     opts["max_age"].(int),
			Compress:   opts["compress"].(bool),
			Formatter:  opts["formatter"].(string),
			Level:      logx.Level(opts["log_level"].(int)),
			Output:     logx.OutputType(opts["output_type"].(string)),
		})
	}

	manager, cleanup, err := logx.BuildManager(options...)
	if err != nil {
		log.Printf("%s🔗 -> Log all initialized failed. %s\n", "\033[31m", "\033[0m")
	} else {
		log.Printf("%s🔗 -> Log all initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

	return manager, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up log components successfully. %s\n", "\033[32m", "\033[0m")
	}, err
}

var templateWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/templates"},
	Name:         "Templates",
	Type:         "*templates.Manager",
	ProviderName: "ProvideTemplateComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
	enable := cfg.GetBool("templates.enable")
	if !enable {
		return nil, func() {}, nil
	}
	rawList := cfg.Get("templates.list").([]interface{})
	templateOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		templateOptionsList[i] = raw.(map[string]interface{})
	}
	options := make([]templates.TemplateOptions, 0)
	for _, opts := range templateOptionsList {
		options = append(options, templates.TemplateOptions{
			Name: opts["name"].(string),
			Path: opts["path"].(string),
		})
	}
	manager, cleanup, err := templates.New(options...)
	if err != nil {
		log.Printf("%s🔗 -> Templates all initialized failed. %s\n", "\033[31m", "\033[0m")
	} else {
		log.Printf("%s🔗 -> Templates all initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

	return manager, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up templates components successfully. %s\n", "\033[32m", "\033[0m")
	}, err
}`,
}

func ProvideTemplateComponent(cfg *config.Config) (*templates.Manager, func(), error) {
	enable := cfg.GetBool("templates.enable")
	if !enable {
		return nil, func() {}, nil
	}
	rawList := cfg.Get("templates.list").([]interface{})
	templateOptionsList := make([]map[string]interface{}, len(rawList))
	for i, raw := range rawList {
		templateOptionsList[i] = raw.(map[string]interface{})
	}
	options := make([]templates.TemplateOptions, 0)
	for _, opts := range templateOptionsList {
		options = append(options, templates.TemplateOptions{
			Name: opts["name"].(string),
			Path: opts["path"].(string),
		})
	}
	manager, cleanup, err := templates.New(options...)
	if err != nil {
		log.Printf("%s🔗 -> Templates all initialized failed. %s\n", "\033[31m", "\033[0m")
	} else {
		log.Printf("%s🔗 -> Templates all initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

	return manager, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up templates components successfully. %s\n", "\033[32m", "\033[0m")
	}, err
}

var hookWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/hook", "context", "time", "log"},
	Name:         "Hook",
	Type:         "*hook.HookManager",
	ProviderName: "ProvideHookComponent",
	Provider: `func {{.ProviderName}}() ({{.Type}}, func(), error) {
	hook := hook.NewHookManager()
	log.Printf("%s🔗 -> Hook all initialized successfully. %s\n", "\033[32m", "\033[0m")
	return hook, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := hook.Stop(ctx)
		if err != nil {
			log.Printf("%s🔗 -> Clean up hook components failed, error: %v %s\n", "\033[31m", err, "\033[0m")
		} else {
			log.Printf("%s🔗 -> Clean up hook components successfully. %s\n", "\033[32m", "\033[0m")
		}
	}, nil
}`,
}

func ProvideHookComponent() (*hook.HookManager, func(), error) {
	hook := hook.NewHookManager()
	log.Printf("%s🔗 -> Hook all initialized successfully. %s\n", "\033[32m", "\033[0m")
	return hook, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := hook.Stop(ctx)
		if err != nil {
			log.Printf("%s🔗 -> Clean up hook components failed, error: %v %s\n", "\033[31m", err, "\033[0m")
		} else {
			log.Printf("%s🔗 -> Clean up hook components successfully. %s\n", "\033[32m", "\033[0m")
		}
	}, nil
}

var cmdWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-common/pkg/cmd", "log"},
	Name:         "Command",
	Type:         "*cmd.Manager",
	ProviderName: "ProvideCmdComponent",
	Provider: `func {{.ProviderName}}() ({{.Type}}, func(), error) {
	command := cmd.NewManager()
	log.Printf("%s🔗 -> Command manager all initialized successfully. %s\n", "\033[32m", "\033[0m")
	return command, func() {
		command.Clear()
		log.Printf("%s🔗 -> Clean up command manager successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}`,
}

func ProvideCmdComponent() (*cmd.Manager, func(), error) {
	command := cmd.NewManager()
	log.Printf("%s🔗 -> Command manager all initialized successfully. %s\n", "\033[32m", "\033[0m")
	return command, func() {
		command.Clear()
		log.Printf("%s🔗 -> Clean up command manager successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
