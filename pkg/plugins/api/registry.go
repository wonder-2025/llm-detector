package api

import (
	"time"

	"llm-detector/pkg/plugins"
)

// RegisterAll 注册所有API探测插件
func RegisterAll(registry *plugins.Registry, timeout time.Duration) {
	registry.RegisterAPI(NewOpenAIPlugin(timeout))
	registry.RegisterAPI(NewOllamaPlugin(timeout))
	registry.RegisterAPI(NewVLLMPlugin(timeout))
	registry.RegisterAPI(NewTGIPlugin(timeout))
	registry.RegisterAPI(NewFastAPIPlugin(timeout))
	registry.RegisterAPI(NewLiteLLMPlugin(timeout))
	registry.RegisterAPI(NewJupyterPlugin(timeout))
	registry.RegisterAPI(NewAttuPlugin(timeout))
	registry.RegisterAPI(NewDifyPlugin(timeout))
	registry.RegisterAPI(NewComfyUIPlugin(timeout))
	registry.RegisterAPI(NewMLflowPlugin(timeout))
	registry.RegisterAPI(NewClickHousePlugin(timeout))
	registry.RegisterAPI(NewOpenWebUIPlugin(timeout))
	registry.RegisterAPI(NewAirflowPlugin(timeout))
	registry.RegisterAPI(NewZenMLPlugin(timeout))
	registry.RegisterAPI(NewHugeGraphPlugin(timeout))
	registry.RegisterAPI(NewGenericPlugin(timeout))
}
