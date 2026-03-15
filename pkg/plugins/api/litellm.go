package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// LiteLLMPlugin LiteLLM探测插件
type LiteLLMPlugin struct {
	client *http.Client
}

// NewLiteLLMPlugin 创建LiteLLM插件
func NewLiteLLMPlugin(timeout time.Duration) *LiteLLMPlugin {
	return &LiteLLMPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *LiteLLMPlugin) Name() string {
	return "litellm"
}

// Version 返回插件版本
func (p *LiteLLMPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测LiteLLM API
func (p *LiteLLMPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/v1/models",
		"/health",
		"/health/readiness",
		"/health/liveliness",
		"/model/info",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "litellm"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "litellm",
		Available: false,
		Error:     "LiteLLM API endpoints not found",
	}, nil
}

func (p *LiteLLMPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "litellm",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "litellm",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为LiteLLM特征
	isLiteLLM := p.checkLiteLLMFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "litellm",
		Endpoint:   endpoint,
		Available:  isLiteLLM,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isLiteLLM {
		result.Error = "Not a LiteLLM API"
	}

	return result
}

func (p *LiteLLMPlugin) checkLiteLLMFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	if resp.Header.Get("x-litellm-version") != "" {
		return true
	}

	// 检查响应体特征
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// LiteLLM模型信息
		if _, ok := data["data"]; ok {
			// 检查是否包含litellm相关信息
			if contains(body, "litellm") {
				return true
			}
		}

		// 健康检查端点
		if status, ok := data["status"].(string); ok {
			if status == "healthy" || status == "ok" {
				return true
			}
		}
	}

	return false
}
