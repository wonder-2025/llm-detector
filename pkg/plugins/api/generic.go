package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// GenericPlugin 通用API探测插件
type GenericPlugin struct {
	client *http.Client
}

// NewGenericPlugin 创建通用插件
func NewGenericPlugin(timeout time.Duration) *GenericPlugin {
	return &GenericPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *GenericPlugin) Name() string {
	return "generic"
}

// Version 返回插件版本
func (p *GenericPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测通用API端点
func (p *GenericPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	// 常见大模型API端点
	endpoints := []struct {
		path string
		name string
	}{
		{"/v1/chat/completions", "openai-compatible"},
		{"/api/generate", "ollama-compatible"},
		{"/generate", "tgi-compatible"},
		{"/v1/completions", "openai-legacy"},
		{"/api/chat", "custom-chat"},
		{"/chat/completions", "custom-openai"},
		{"/predict", "custom-predict"},
		{"/inference", "custom-inference"},
		{"/api/v1/generate", "custom-v1"},
		{"/v2/chat/completions", "openai-v2"},
	}

	for _, ep := range endpoints {
		result := p.probeEndpoint(ctx, target, ep.path, ep.name)
		if result.Available {
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "generic",
		Available: false,
		Error:     "No LLM API endpoints found",
	}, nil
}

func (p *GenericPlugin) probeEndpoint(ctx context.Context, target plugins.Target, path, apiType string) *plugins.APIResult {
	url := target.BaseURL() + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      apiType,
			Endpoint:  path,
			Available: false,
			Error:     err.Error(),
		}
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      apiType,
			Endpoint:  path,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否可能是LLM API
	isLLM := p.checkLLMFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       apiType,
		Endpoint:   path,
		Available:  isLLM,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isLLM {
		result.Error = "Not an LLM API endpoint"
	}

	return result
}

func (p *GenericPlugin) checkLLMFeatures(resp *http.Response, body string) bool {
	// 检查常见LLM API特征
	llmIndicators := []string{
		"model", "choices", "messages", "completion",
		"generated_text", "text", "content", "role",
		"assistant", "user", "system", "prompt",
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// 检查JSON中是否包含LLM相关字段
		for _, indicator := range llmIndicators {
			if _, ok := data[indicator]; ok {
				return true
			}
		}

		// 检查错误消息
		if errData, ok := data["error"].(map[string]interface{}); ok {
			if msg, ok := errData["message"].(string); ok {
				// 常见LLM API错误
				llmErrors := []string{
					"model", "authorization", "api key",
					"rate limit", "token", "context",
				}
				for _, err := range llmErrors {
					if contains(msg, err) {
						return true
					}
				}
			}
		}
	}

	// 检查响应头
	if resp.Header.Get("x-ratelimit-limit") != "" {
		return true
	}
	if resp.Header.Get("x-request-id") != "" {
		return true
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
