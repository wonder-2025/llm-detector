package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// ZenMLPlugin ZenML探测插件
type ZenMLPlugin struct {
	client *http.Client
}

// NewZenMLPlugin 创建ZenML插件
func NewZenMLPlugin(timeout time.Duration) *ZenMLPlugin {
	return &ZenMLPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *ZenMLPlugin) Name() string {
	return "zenml"
}

// Version 返回插件版本
func (p *ZenMLPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *ZenMLPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/api/v1/pipelines",
		"/api/v1/runs",
		"/api/v1/stacks",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "zenml"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "zenml",
		Available: false,
		Error:     "ZenML endpoints not found",
	}, nil
}

func (p *ZenMLPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "zenml", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "zenml", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查ZenML特征
	isZenML := false
	if containsAny(bodyStr, []string{"zenml", "pipeline", "step", "stack"}) {
		isZenML = true
	}

	return &plugins.APIResult{
		Type:       "zenml",
		Endpoint:   endpoint,
		Available:  isZenML,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
