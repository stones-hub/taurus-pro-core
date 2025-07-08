package consul

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-consul/pkg/consul"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

var ConsulComponent = types.Component{
	Name:         "consul",
	Package:      "github.com/stones-hub/taurus-pro-consul",
	Version:      "v0.0.1",
	Description:  "consul组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{consulWire},
}

var consulWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-consul/pkg/consul", "time", "log"},
	Name:         "Consul",
	Type:         "*consul.Client",
	ProviderName: "ProvideConsulComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
	if !cfg.GetBool("consul.enable") {
		return nil, func() {}, nil
	}

	options := make([]consul.Option, 0)

	options = append(options, consul.WithAddress(cfg.GetString("consul.client.address")))
	if cfg.GetString("consul.client.token") != "" {
		options = append(options, consul.WithToken(cfg.GetString("consul.client.token")))
	}
	options = append(options, consul.WithTimeout(time.Duration(cfg.GetInt("consul.client.timeout"))*time.Second))
	options = append(options, consul.WithScheme(cfg.GetString("consul.client.scheme")))
	options = append(options, consul.WithDatacenter(cfg.GetString("consul.client.datacenter")))
	options = append(options, consul.WithWaitTime(time.Duration(cfg.GetInt("consul.client.wait_time"))*time.Second))
	options = append(options, consul.WithRetryTime(time.Duration(cfg.GetInt("consul.client.retry_time"))*time.Second))
	options = append(options, consul.WithMaxRetries(cfg.GetInt("consul.client.max_retrys")))
	if cfg.GetString("consul.client.http_basic_auth.username") != "" && cfg.GetString("consul.client.http_basic_auth.password") != "" {
		options = append(options, consul.WithBasicAuth(cfg.GetString("consul.client.http_basic_auth.username"),
			cfg.GetString("consul.client.http_basic_auth.password")))
	}

	client, err := consul.NewClient(options...)
	if err != nil {
		log.Printf("consul.NewClient error: %v", err)
		return nil, func() {}, err
	}

	client.Put("config/"+cfg.GetString("consul.service.name"), []byte(cfg.ToJSONString()))

	meta := cfg.Get("consul.service.meta").(map[string]interface{})
	metaMap := make(map[string]string)
	for k, v := range meta {
		metaMap[k] = v.(string)
	}

	serviceConfig := consul.ServiceConfig{
		Name:    cfg.GetString("consul.service.name"),
		ID:      cfg.GetString("consul.service.id"),
		Tags:    cfg.GetStringSlice("consul.service.tags"),
		Address: cfg.GetString("consul.service.address"),
		Port:    cfg.GetInt("consul.service.port"),
		Meta:    metaMap,
		Checks:  make([]*consul.CheckConfig, 0),
	}

	healths := cfg.Get("consul.service.healths")
	if healths != nil {
		healthsList := healths.([]interface{})
		for _, h := range healthsList {
			health := h.(map[string]interface{})
			httpHeaders := health["http_headers"].(map[string]interface{})
			headers := make(map[string][]string)
			for k, v := range httpHeaders {
				values := v.([]interface{})
				strValues := make([]string, len(values))
				for i, val := range values {
					strValues[i] = val.(string)
				}
				headers[k] = strValues
			}

			serviceConfig.Checks = append(serviceConfig.Checks, &consul.CheckConfig{
				HTTP:            health["http"].(string),
				Method:          health["http_method"].(string),
				Header:          headers,
				TCP:             health["tcp"].(string),
				Interval:        time.Duration(health["interval"].(int)) * time.Second,
				Timeout:         time.Duration(health["timeout"].(int)) * time.Second,
				DeregisterAfter: time.Duration(health["deregister_after"].(int)) * time.Second,
				TLSSkipVerify:   health["tls_skip_verify"].(bool),
			})
		}
	}

	if err := client.RegisterService(&serviceConfig); err != nil {
		log.Printf("consul.RegisterService error: %v", err)
		return nil, func() {}, err
	}

	if err := client.WatchConfig("config/"+cfg.GetString("consul.service.name"), cfg, &consul.WatchOptions{
		WaitTime:  time.Duration(cfg.GetInt("consul.watch.wait_time")) * time.Second,
		RetryTime: time.Duration(cfg.GetInt("consul.watch.retry_time")) * time.Second,
	}); err != nil {
		log.Printf("consul.WatchConfig error: %v", err)
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Initialize consul components successfully. %s\n", "\033[32m", "\033[0m")

	return client, func() {

		client.DeregisterService(cfg.GetString("consul.service.id"))

		client.Close()
		log.Printf("%s🔗 -> Clean up consul components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}`,
}

func ProvideConsulComponent(cfg *config.Config) (*consul.Client, func(), error) {
	if !cfg.GetBool("consul.enable") {
		return nil, func() {}, nil
	}

	options := make([]consul.Option, 0)

	options = append(options, consul.WithAddress(cfg.GetString("consul.client.address")))
	if cfg.GetString("consul.client.token") != "" {
		options = append(options, consul.WithToken(cfg.GetString("consul.client.token")))
	}
	options = append(options, consul.WithTimeout(time.Duration(cfg.GetInt("consul.client.timeout"))*time.Second))
	options = append(options, consul.WithScheme(cfg.GetString("consul.client.scheme")))
	options = append(options, consul.WithDatacenter(cfg.GetString("consul.client.datacenter")))
	options = append(options, consul.WithWaitTime(time.Duration(cfg.GetInt("consul.client.wait_time"))*time.Second))
	options = append(options, consul.WithRetryTime(time.Duration(cfg.GetInt("consul.client.retry_time"))*time.Second))
	options = append(options, consul.WithMaxRetries(cfg.GetInt("consul.client.max_retrys")))
	if cfg.GetString("consul.client.http_basic_auth.username") != "" && cfg.GetString("consul.client.http_basic_auth.password") != "" {
		options = append(options, consul.WithBasicAuth(cfg.GetString("consul.client.http_basic_auth.username"),
			cfg.GetString("consul.client.http_basic_auth.password")))
	}

	client, err := consul.NewClient(options...)
	if err != nil {
		log.Printf("consul.NewClient error: %v", err)
		return nil, func() {}, err
	}

	client.Put("config/"+cfg.GetString("consul.service.name"), []byte(cfg.ToJSONString()))

	meta := cfg.Get("consul.service.meta").(map[string]interface{})
	metaMap := make(map[string]string)
	for k, v := range meta {
		metaMap[k] = v.(string)
	}

	serviceConfig := consul.ServiceConfig{
		Name:    cfg.GetString("consul.service.name"),
		ID:      cfg.GetString("consul.service.id"),
		Tags:    cfg.GetStringSlice("consul.service.tags"),
		Address: cfg.GetString("consul.service.address"),
		Port:    cfg.GetInt("consul.service.port"),
		Meta:    metaMap,
		Checks:  make([]*consul.CheckConfig, 0),
	}

	healths := cfg.Get("consul.service.healths")
	if healths != nil {
		healthsList := healths.([]interface{})
		for _, h := range healthsList {
			health := h.(map[string]interface{})
			httpHeaders := health["http_headers"].(map[string]interface{})
			headers := make(map[string][]string)
			for k, v := range httpHeaders {
				values := v.([]interface{})
				strValues := make([]string, len(values))
				for i, val := range values {
					strValues[i] = val.(string)
				}
				headers[k] = strValues
			}

			serviceConfig.Checks = append(serviceConfig.Checks, &consul.CheckConfig{
				HTTP:            health["http"].(string),
				Method:          health["http_method"].(string),
				Header:          headers,
				TCP:             health["tcp"].(string),
				Interval:        time.Duration(health["interval"].(int)) * time.Second,
				Timeout:         time.Duration(health["timeout"].(int)) * time.Second,
				DeregisterAfter: time.Duration(health["deregister_after"].(int)) * time.Second,
				TLSSkipVerify:   health["tls_skip_verify"].(bool),
			})
		}
	}

	if err := client.RegisterService(&serviceConfig); err != nil {
		log.Printf("consul.RegisterService error: %v", err)
		return nil, func() {}, err
	}

	if err := client.WatchConfig("config/"+cfg.GetString("consul.service.name"), cfg, &consul.WatchOptions{
		WaitTime:  time.Duration(cfg.GetInt("consul.watch.wait_time")) * time.Second,
		RetryTime: time.Duration(cfg.GetInt("consul.watch.retry_time")) * time.Second,
	}); err != nil {
		log.Printf("consul.WatchConfig error: %v", err)
		return nil, func() {}, err
	}

	log.Printf("%s🔗 -> Initialize consul components successfully. %s\n", "\033[32m", "\033[0m")

	return client, func() {
		client.DeregisterService(cfg.GetString("consul.service.id"))
		client.Close()
		log.Printf("%s🔗 -> Clean up consul components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
