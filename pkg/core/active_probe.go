package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ActiveProber 主动探测
type ActiveProber struct {
	client  *http.Client
	timeout time.Duration
}

// NewActiveProber 创建主动探测器
func NewActiveProber(timeout time.Duration) *ActiveProber {
	return &ActiveProber{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProbeResult 探测结果
type ProbeResult struct {
	Available   bool
	Type        string
	Endpoint    string
	StatusCode  int
	Headers     map[string]string
	Body        string
	Error       string
}

// ProbeTarget 主动探测目标
func (p *ActiveProber) ProbeTarget(ctx context.Context, target *Target) []*ProbeResult {
	var results []*ProbeResult

	// 常见LLM端口
	ports := []int{11434, 8080, 8000, 3000, 5000, 5001, 8443, 9443, 9090, 4000}

	// 常见API路径
	paths := []string{
		"/v1/chat/completions",
		"/v1/models",
		"/api/tags",
		"/api/generate",
		"/generate",
		"/health",
		"/docs",
		"/",
	}

	baseURL := target.BaseURL()

	// 如果目标是纯IP，尝试多个端口
	if target.Type == TargetIP {
		for _, port := range ports {
			url := fmt.Sprintf("http://%s:%d", target.Host, port)
			for _, path := range paths {
				result := p.probeURL(ctx, url+path)
				if result.Available {
					results = append(results, result)
				}
			}
		}
	} else {
		// 否则只探测给定的URL
		for _, path := range paths {
			result := p.probeURL(ctx, baseURL+path)
			if result.Available {
				results = append(results, result)
			}
		}
	}

	return results
}

func (p *ActiveProber) probeURL(ctx context.Context, url string) *ProbeResult {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &ProbeResult{Available: false, Error: err.Error()}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "LLM-Detector/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return &ProbeResult{Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否是LLM相关响应
	isLLM := p.isLLMResponse(resp, string(body))

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &ProbeResult{
		Available:  isLLM,
		Endpoint:   url,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
	}
}

func (p *ActiveProber) isLLMResponse(resp *http.Response, body string) bool {
	// 检查响应头
	llmHeaders := []string{
		"openai-model",
		"x-request-id",
		"anthropic-ratelimit",
		"ollama-version",
		"x-vllm-executor",
		"x-litellm-version",
		"x-tgi-version",
	}

	for _, header := range llmHeaders {
		if resp.Header.Get(header) != "" {
			return true
		}
	}

	// 检查响应体
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// 常见LLM字段
		llmFields := []string{"model", "choices", "messages", "completion", "generated_text"}
		for _, field := range llmFields {
			if _, ok := data[field]; ok {
				return true
			}
		}

		// 错误消息
		if errData, ok := data["error"].(map[string]interface{}); ok {
			if msg, ok := errData["message"].(string); ok {
				llmErrors := []string{"model", "api key", "rate limit", "token", "authorization"}
				for _, err := range llmErrors {
					if containsLower(msg, err) {
						return true
					}
				}
			}
		}
	}

	return false
}

// TestModel 测试模型响应
func (p *ActiveProber) TestModel(ctx context.Context, endpoint string, apiType string) (*ProbeResult, error) {
	var payload map[string]interface{}

	switch apiType {
	case "openai", "vllm", "litellm":
		payload = map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]string{
				{"role": "user", "content": "Hi"},
			},
			"max_tokens": 5,
		}
	case "ollama":
		payload = map[string]interface{}{
			"model":  "llama2",
			"prompt": "Hi",
			"stream": false,
		}
	default:
		payload = map[string]interface{}{
			"prompt": "Hi",
		}
	}

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &ProbeResult{
		Available:  true,
		Endpoint:   endpoint,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func containsLower(s, substr string) bool {
	return len(s) >= len(substr) && containsHelperLower(s, substr)
}

func containsHelperLower(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(subLower); i++ {
		if sLower[i:i+len(subLower)] == subLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
