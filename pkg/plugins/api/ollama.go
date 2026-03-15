package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// OllamaPlugin Ollama API探测插件
type OllamaPlugin struct {
	client *http.Client
}

// NewOllamaPlugin 创建Ollama插件
func NewOllamaPlugin(timeout time.Duration) *OllamaPlugin {
	return &OllamaPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *OllamaPlugin) Name() string {
	return "ollama"
}

// Version 返回插件版本
func (p *OllamaPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测Ollama API
func (p *OllamaPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/api/tags",
		"/api/version",
		"/api/generate",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "ollama"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "ollama",
		Available: false,
		Error:     "Ollama API endpoints not found",
	}, nil
}

func (p *OllamaPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "ollama",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "ollama",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为Ollama API特征
	isOllama := p.checkOllamaFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "ollama",
		Endpoint:   endpoint,
		Available:  isOllama,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isOllama {
		result.Error = "Not an Ollama API"
	}

	return result
}

func (p *OllamaPlugin) checkOllamaFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	if resp.Header.Get("Ollama-Version") != "" {
		return true
	}

	// 检查响应体特征
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// /api/tags 返回模型列表
		if models, ok := data["models"].([]interface{}); ok {
			if len(models) > 0 {
				return true
			}
		}

		// /api/version 返回版本信息
		if version, ok := data["version"].(string); ok && version != "" {
			return true
		}
	}

	return false
}

// GetModels 获取Ollama模型列表
func (p *OllamaPlugin) GetModels(ctx context.Context, target plugins.Target) ([]string, error) {
	url := target.BaseURL() + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range data.Models {
		models = append(models, m.Name)
	}

	return models, nil
}


