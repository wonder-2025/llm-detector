package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// FastAPIPlugin FastAPI探测插件
type FastAPIPlugin struct {
	client *http.Client
}

// NewFastAPIPlugin 创建FastAPI插件
func NewFastAPIPlugin(timeout time.Duration) *FastAPIPlugin {
	return &FastAPIPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *FastAPIPlugin) Name() string {
	return "fastapi"
}

// Version 返回插件版本
func (p *FastAPIPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测FastAPI服务
func (p *FastAPIPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/docs",
		"/openapi.json",
		"/redoc",
		"/health",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "fastapi"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "fastapi",
		Available: false,
		Error:     "FastAPI endpoints not found",
	}, nil
}

func (p *FastAPIPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "fastapi",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "fastapi",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为FastAPI特征
	isFastAPI := p.checkFastAPIFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "fastapi",
		Endpoint:   endpoint,
		Available:  isFastAPI,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isFastAPI {
		result.Error = "Not a FastAPI service"
	}

	return result
}

func (p *FastAPIPlugin) checkFastAPIFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	server := resp.Header.Get("server")
	if contains(server, "uvicorn") || contains(server, "hypercorn") {
		return true
	}

	// 检查OpenAPI文档
	if resp.Request.URL.Path == "/openapi.json" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(body), &data); err == nil {
			if _, ok := data["openapi"]; ok {
				return true
			}
		}
	}

	// 检查Swagger UI
	if resp.Request.URL.Path == "/docs" {
		if contains(body, "Swagger UI") || contains(body, "fastapi") {
			return true
		}
	}

	return false
}
