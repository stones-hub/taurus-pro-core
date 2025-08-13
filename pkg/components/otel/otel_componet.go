package otel

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-opentelemetry/pkg/otelemetry"
)

var OtelComponent = types.Component{
	Name:         "otel",
	Package:      "github.com/stones-hub/taurus-pro-opentelemetry",
	Version:      "v0.0.2",
	Description:  "otel_tarce组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{otelWire},
}

var otelWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-opentelemetry/pkg/otelemetry"},
	Name:         "OtelProvider",
	Type:         "*otelemetry.OTelProvider",
	ProviderName: "ProvideOtelComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {

	enable := cfg.GetBool("otel.enable")
	if !enable {
		return nil, func() {}, nil
	}

	timeout, err := time.ParseDuration(cfg.GetString("otel.export.timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	batchTimeout, err := time.ParseDuration(cfg.GetString("otel.batch.timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	exportTimeout, err := time.ParseDuration(cfg.GetString("otel.batch.export_timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	provider, cleanup, err := otelemetry.NewOTelProvider(
		otelemetry.WithServiceName(cfg.GetString("otel.service.name")),
		otelemetry.WithServiceVersion(cfg.GetString("otel.service.version")),
		otelemetry.WithEnvironment(cfg.GetString("otel.service.environment")), // 环境

		otelemetry.WithExportProtocol(otelemetry.ExportProtocol(cfg.GetString("otel.export.protocol"))),
		otelemetry.WithEndpoint(cfg.GetString("otel.export.endpoint")),
		otelemetry.WithInsecure(cfg.GetBool("otel.export.insecure")),
		otelemetry.WithTimeout(timeout),

		otelemetry.WithSamplingRatio(cfg.GetFloat64("otel.sampling.ratio")),

		otelemetry.WithBatchTimeout(batchTimeout),
		otelemetry.WithMaxExportBatchSize(cfg.GetInt("otel.batch.max_size")),
		otelemetry.WithMaxQueueSize(cfg.GetInt("otel.batch.max_queue_size")),
		otelemetry.WithExportTimeout(exportTimeout),
	)

	if err != nil {
		log.Printf("%s🔗 -> Initialize otel components failed. %s\n", "\033[31m", "\033[0m")
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Initialize otel components successfully. %s\n", "\033[32m", "\033[0m")

	// 添加配置的tracer
	tracers := cfg.GetStringSlice("otel.tracers")
	for _, tracer := range tracers {
		otelemetry.RegisterTracer(tracer, provider.Tracer(tracer))
	}

	return provider, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up otel components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}`,
}

func ProvideOtelComponent(cfg *config.Config) (*otelemetry.OTelProvider, func(), error) {

	enable := cfg.GetBool("otel.enable")
	if !enable {
		return nil, func() {}, nil
	}

	timeout, err := time.ParseDuration(cfg.GetString("otel.export.timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	batchTimeout, err := time.ParseDuration(cfg.GetString("otel.batch.timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	exportTimeout, err := time.ParseDuration(cfg.GetString("otel.batch.export_timeout"))
	if err != nil {
		return nil, func() {}, err
	}

	provider, cleanup, err := otelemetry.NewOTelProvider(
		otelemetry.WithServiceName(cfg.GetString("otel.service.name")),
		otelemetry.WithServiceVersion(cfg.GetString("otel.service.version")),
		otelemetry.WithEnvironment(cfg.GetString("otel.service.environment")), // 环境

		otelemetry.WithExportProtocol(otelemetry.ExportProtocol(cfg.GetString("otel.export.protocol"))),
		otelemetry.WithEndpoint(cfg.GetString("otel.export.endpoint")),
		otelemetry.WithInsecure(cfg.GetBool("otel.export.insecure")),
		otelemetry.WithTimeout(timeout),

		otelemetry.WithSamplingRatio(cfg.GetFloat64("otel.sampling.ratio")),

		otelemetry.WithBatchTimeout(batchTimeout),
		otelemetry.WithMaxExportBatchSize(cfg.GetInt("otel.batch.max_size")),
		otelemetry.WithMaxQueueSize(cfg.GetInt("otel.batch.max_queue_size")),
		otelemetry.WithExportTimeout(exportTimeout),
	)

	if err != nil {
		log.Printf("%s🔗 -> Initialize otel components failed. %s\n", "\033[31m", "\033[0m")
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Initialize otel components successfully. %s\n", "\033[32m", "\033[0m")

	// 添加配置的tracer
	tracers := cfg.GetStringSlice("otel.tracers")
	for _, tracer := range tracers {
		otelemetry.RegisterTracer(tracer, provider.Tracer(tracer))
	}

	return provider, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up otel components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
