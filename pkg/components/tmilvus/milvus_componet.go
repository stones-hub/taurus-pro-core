package tmilvus

import (
	"log"
	"math"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-milvus/pkg/milvus"
	mclient "github.com/stones-hub/taurus-pro-milvus/pkg/milvus/client"
)

var MilvusComponent = types.Component{
	Name:         "milvus",
	Package:      "github.com/stones-hub/taurus-pro-milvus",
	Version:      "v0.0.7",
	Description:  "milvus向量数据库组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{milvusWire},
}

var milvusWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-milvus/pkg/milvus", "mclient@github.com/stones-hub/taurus-pro-milvus/pkg/milvus/client", "log", "math", "time"},
	Name:         "Milvus",
	Type:         "milvus.Pool",
	ProviderName: "ProvideMilvusComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
	// 检查是否启用 Milvus
	if !cfg.GetBool("milvus.enable") {
		return nil, func() {}, nil
	}

	// 获取配置列表
	rawList := cfg.Get("milvus.list").([]interface{})
	if len(rawList) == 0 {
		return nil, func() {}, nil
	}

	// 创建连接池
	pool := milvus.NewPool()

	// 遍历配置列表，为每个配置创建客户端
	for _, raw := range rawList {
		configMap := raw.(map[string]interface{})
		name, _ := configMap["name"].(string)
		if name == "" {
			log.Printf("%s🔗 -> Milvus name is empty. %s\n", "\033[31m", "\033[0m")
			continue
		}

		var opts []mclient.Option

		// 基础连接配置
		if address, ok := configMap["address"].(string); ok && address != "" {
			opts = append(opts, mclient.WithAddress(address))
		}

		// 认证配置 - 优先使用 API Key，否则使用用户名密码
		if apiKey, ok := configMap["api_key"].(string); ok && apiKey != "" {
			opts = append(opts, mclient.WithAPIKey(apiKey))
		} else if username, _ := configMap["username"].(string); username != "" || configMap["password"] != "" {
			password, _ := configMap["password"].(string)
			opts = append(opts, mclient.WithAuth(username, password))
		}

		// 数据库名称
		if dbName, ok := configMap["db_name"].(string); ok && dbName != "" {
			opts = append(opts, mclient.WithDatabase(dbName))
		}

		// TLS 配置
		if enableTLS, ok := configMap["enable_tls_auth"].(bool); ok && enableTLS {
			opts = append(opts, mclient.WithTLS())
		}

		// 重试配置
		maxRetry := uint(3)
		if retry, ok := configMap["max_retry"].(int); ok && retry > 0 {
			maxRetry = uint(retry)
		}
		maxRetryBackoff := 30 * time.Second
		if backoffStr, ok := configMap["max_retry_backoff"].(string); ok && backoffStr != "" {
			if backoff, err := time.ParseDuration(backoffStr); err == nil {
				maxRetryBackoff = backoff
			}
		}
		opts = append(opts, mclient.WithRetry(maxRetry, maxRetryBackoff))

		// GRPC连接配置 - 使用新的 WithGrpcOpts 方法
		keepAliveTime := 30 * time.Second
		if timeStr, ok := configMap["keepalive_time"].(string); ok && timeStr != "" {
			if t, err := time.ParseDuration(timeStr); err == nil {
				keepAliveTime = t
			}
		}
		keepAliveTimeout := 10 * time.Second
		if timeoutStr, ok := configMap["keepalive_timeout"].(string); ok && timeoutStr != "" {
			if t, err := time.ParseDuration(timeoutStr); err == nil {
				keepAliveTimeout = t
			}
		}

		// 其他GRPC配置
		permitWithoutStream := true
		if permit, ok := configMap["permit_without_stream"].(bool); ok {
			permitWithoutStream = permit
		}

		baseDelay := 1 * time.Second
		if delayStr, ok := configMap["base_delay"].(string); ok && delayStr != "" {
			if delay, err := time.ParseDuration(delayStr); err == nil {
				baseDelay = delay
			}
		}

		multiplier := 1.6
		if mult, ok := configMap["multiplier"].(float64); ok && mult > 0 {
			multiplier = mult
		}

		jitter := 0.2
		if jit, ok := configMap["jitter"].(float64); ok && jit >= 0 {
			jitter = jit
		}

		maxDelay := 120 * time.Second
		if maxDelayStr, ok := configMap["max_delay"].(string); ok && maxDelayStr != "" {
			if delay, err := time.ParseDuration(maxDelayStr); err == nil {
				maxDelay = delay
			}
		}

		minConnectTimeout := 20 * time.Second
		if minTimeoutStr, ok := configMap["min_connect_timeout"].(string); ok && minTimeoutStr != "" {
			if timeout, err := time.ParseDuration(minTimeoutStr); err == nil {
				minConnectTimeout = timeout
			}
		}

		maxRecvMsgSize := math.MaxInt32
		if size, ok := configMap["max_recv_msg_size"].(int); ok && size > 0 {
			maxRecvMsgSize = size
		}

		// 应用GRPC配置
		opts = append(opts, mclient.WithGrpcOpts(
			keepAliveTime,
			keepAliveTimeout,
			permitWithoutStream,
			baseDelay,
			multiplier,
			jitter,
			maxDelay,
			minConnectTimeout,
			maxRecvMsgSize,
		))

		// 禁用连接握手配置
		if disableConn, ok := configMap["disable_conn"].(bool); ok && disableConn {
			opts = append(opts, mclient.WithDisableConn(disableConn))
		}

		// 添加客户端到连接池
		if err := pool.Add(name, opts...); err != nil {
			log.Printf("%s🔗 -> Milvus add client failed, error: %s, name: %s. %s\n", "\033[31m", err, name, "\033[0m")
		}
	}

	log.Printf("%s🔗 -> Milvus all initialized successfully. %s\n", "\033[32m", "\033[0m")

	// 返回连接池和清理函数
	return pool, func() {
		pool.Close()
		log.Printf("%s🔗 -> Clean up milvus components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}`,
}

func ProvideMilvusComponent(cfg *config.Config) (milvus.Pool, func(), error) {
	// 检查是否启用 Milvus
	if !cfg.GetBool("milvus.enable") {
		return nil, func() {}, nil
	}

	// 获取配置列表
	rawList := cfg.Get("milvus.list").([]interface{})
	if len(rawList) == 0 {
		return nil, func() {}, nil
	}

	// 创建连接池
	pool := milvus.NewPool()

	// 遍历配置列表，为每个配置创建客户端
	for _, raw := range rawList {
		configMap := raw.(map[string]interface{})
		name, _ := configMap["name"].(string)
		if name == "" {
			log.Printf("%s🔗 -> Milvus name is empty. %s\n", "\033[31m", "\033[0m")
			continue
		}

		var opts []mclient.Option

		// 基础连接配置
		if address, ok := configMap["address"].(string); ok && address != "" {
			opts = append(opts, mclient.WithAddress(address))
		}

		// 认证配置 - 优先使用 API Key，否则使用用户名密码
		if apiKey, ok := configMap["api_key"].(string); ok && apiKey != "" {
			opts = append(opts, mclient.WithAPIKey(apiKey))
		} else if username, _ := configMap["username"].(string); username != "" || configMap["password"] != "" {
			password, _ := configMap["password"].(string)
			opts = append(opts, mclient.WithAuth(username, password))
		}

		// 数据库名称
		if dbName, ok := configMap["db_name"].(string); ok && dbName != "" {
			opts = append(opts, mclient.WithDatabase(dbName))
		}

		// TLS 配置
		if enableTLS, ok := configMap["enable_tls_auth"].(bool); ok && enableTLS {
			opts = append(opts, mclient.WithTLS())
		}

		// 重试配置
		maxRetry := uint(3)
		if retry, ok := configMap["max_retry"].(int); ok && retry > 0 {
			maxRetry = uint(retry)
		}
		maxRetryBackoff := 30 * time.Second
		if backoffStr, ok := configMap["max_retry_backoff"].(string); ok && backoffStr != "" {
			if backoff, err := time.ParseDuration(backoffStr); err == nil {
				maxRetryBackoff = backoff
			}
		}
		opts = append(opts, mclient.WithRetry(maxRetry, maxRetryBackoff))

		// GRPC连接配置 - 使用新的 WithGrpcOpts 方法
		keepAliveTime := 30 * time.Second
		if timeStr, ok := configMap["keepalive_time"].(string); ok && timeStr != "" {
			if t, err := time.ParseDuration(timeStr); err == nil {
				keepAliveTime = t
			}
		}
		keepAliveTimeout := 10 * time.Second
		if timeoutStr, ok := configMap["keepalive_timeout"].(string); ok && timeoutStr != "" {
			if t, err := time.ParseDuration(timeoutStr); err == nil {
				keepAliveTimeout = t
			}
		}

		// 其他GRPC配置
		permitWithoutStream := true
		if permit, ok := configMap["permit_without_stream"].(bool); ok {
			permitWithoutStream = permit
		}

		baseDelay := 1 * time.Second
		if delayStr, ok := configMap["base_delay"].(string); ok && delayStr != "" {
			if delay, err := time.ParseDuration(delayStr); err == nil {
				baseDelay = delay
			}
		}

		multiplier := 1.6
		if mult, ok := configMap["multiplier"].(float64); ok && mult > 0 {
			multiplier = mult
		}

		jitter := 0.2
		if jit, ok := configMap["jitter"].(float64); ok && jit >= 0 {
			jitter = jit
		}

		maxDelay := 120 * time.Second
		if maxDelayStr, ok := configMap["max_delay"].(string); ok && maxDelayStr != "" {
			if delay, err := time.ParseDuration(maxDelayStr); err == nil {
				maxDelay = delay
			}
		}

		minConnectTimeout := 20 * time.Second
		if minTimeoutStr, ok := configMap["min_connect_timeout"].(string); ok && minTimeoutStr != "" {
			if timeout, err := time.ParseDuration(minTimeoutStr); err == nil {
				minConnectTimeout = timeout
			}
		}

		maxRecvMsgSize := math.MaxInt32
		if size, ok := configMap["max_recv_msg_size"].(int); ok && size > 0 {
			maxRecvMsgSize = size
		}

		// 应用GRPC配置
		opts = append(opts, mclient.WithGrpcOpts(
			keepAliveTime,
			keepAliveTimeout,
			permitWithoutStream,
			baseDelay,
			multiplier,
			jitter,
			maxDelay,
			minConnectTimeout,
			maxRecvMsgSize,
		))

		// 禁用连接握手配置
		if disableConn, ok := configMap["disable_conn"].(bool); ok && disableConn {
			opts = append(opts, mclient.WithDisableConn(disableConn))
		}

		// 添加客户端到连接池
		if err := pool.Add(name, opts...); err != nil {
			log.Printf("%s🔗 -> Milvus add client failed, error: %s, name: %s. %s\n", "\033[31m", err, name, "\033[0m")
		}
	}

	log.Printf("%s🔗 -> Milvus all initialized successfully. %s\n", "\033[32m", "\033[0m")

	// 返回连接池和清理函数
	return pool, func() {
		pool.Close()
		log.Printf("%s🔗 -> Clean up milvus components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
