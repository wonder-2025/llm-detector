package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// targetAdapter 适配器，将core.Target适配为plugins.Target
type targetAdapter interface {
	String() string
	BaseURL() string
}

// OpenAIPlugin OpenAI API探测插件
type OpenAIPlugin struct {
	client *http.Client
}

// NewOpenAIPlugin 创建OpenAI插件
func NewOpenAIPlugin(timeout time.Duration) *OpenAIPlugin {
	return &OpenAIPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *OpenAIPlugin) Name() string {
	return "openai"
}

// Version 返回插件版本
func (p *OpenAIPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测OpenAI API
func (p *OpenAIPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/v1/chat/completions",
		"/v1/models",
		"/v1/completions",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "openai"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "openai",
		Available: false,
		Error:     "OpenAI API endpoints not found",
	}, nil
}

func (p *OpenAIPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "openai",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	// 添加常见的API Key头（用于检测响应）
	req.Header.Set("Authorization", "Bearer test")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "openai",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为OpenAI API特征
	isOpenAI := p.checkOpenAIFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "openai",
		Endpoint:   endpoint,
		Available:  isOpenAI,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isOpenAI {
		result.Error = "Not an OpenAI-compatible API"
	}

	return result
}

func (p *OpenAIPlugin) checkOpenAIFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	if resp.Header.Get("openai-model") != "" {
		return true
	}
	if resp.Header.Get("x-request-id") != "" {
		return true
	}

	// 检查响应体特征
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// OpenAI错误格式
		if _, ok := data["error"].(map[string]interface{}); ok {
			if errObj, ok := data["error"].(map[string]interface{}); ok {
				if errType, ok := errObj["type"].(string); ok {
					if errType == "invalid_request_error" || errType == "authentication_error" {
						return true
					}
				}
			}
		}
	}

	return false
}

// TestCompletion 测试 completions API
func (p *OpenAIPlugin) TestCompletion(ctx context.Context, target plugins.Target, apiKey string) (*http.Response, error) {
	url := target.BaseURL() + "/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
		"max_tokens": 10,
	}

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	return p.client.Do(req)
}


