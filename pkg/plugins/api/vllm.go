package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// VLLMPlugin vLLM API探测插件
type VLLMPlugin struct {
	client *http.Client
}

// NewVLLMPlugin 创建vLLM插件
func NewVLLMPlugin(timeout time.Duration) *VLLMPlugin {
	return &VLLMPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *VLLMPlugin) Name() string {
	return "vllm"
}

// Version 返回插件版本
func (p *VLLMPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测vLLM API
func (p *VLLMPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/v1/models",
		"/health",
		"/metrics",
		"/v1/chat/completions",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "vllm"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "vllm",
		Available: false,
		Error:     "vLLM API endpoints not found",
	}, nil
}

func (p *VLLMPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "vllm",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "vllm",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为vLLM特征
	isVLLM := p.checkVLLMFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "vllm",
		Endpoint:   endpoint,
		Available:  isVLLM,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isVLLM {
		result.Error = "Not a vLLM API"
	}

	return result
}

func (p *VLLMPlugin) checkVLLMFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	if resp.Header.Get("x-vllm-executor") != "" {
		return true
	}

	// 检查响应体特征
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// vLLM模型列表
		if data, ok := data["data"].([]interface{}); ok && len(data) > 0 {
			return true
		}

		// Prometheus metrics
		if len(body) > 100 && contains(body, "vllm") {
			return true
		}
	}

	return false
}
