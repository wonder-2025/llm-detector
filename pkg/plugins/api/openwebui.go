package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// OpenWebUIPlugin OpenWebUI探测插件
type OpenWebUIPlugin struct {
	client *http.Client
}

// NewOpenWebUIPlugin 创建OpenWebUI插件
func NewOpenWebUIPlugin(timeout time.Duration) *OpenWebUIPlugin {
	return &OpenWebUIPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *OpenWebUIPlugin) Name() string {
	return "openwebui"
}

// Version 返回插件版本
func (p *OpenWebUIPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *OpenWebUIPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/api/models",
		"/api/chats",
		"/api/users",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "openwebui"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "openwebui",
		Available: false,
		Error:     "OpenWebUI endpoints not found",
	}, nil
}

func (p *OpenWebUIPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "openwebui", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "openwebui", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查OpenWebUI特征
	isOpenWebUI := false
	if containsAny(bodyStr, []string{"openwebui", "webui", "ollama-webui"}) {
		isOpenWebUI = true
	}

	return &plugins.APIResult{
		Type:       "openwebui",
		Endpoint:   endpoint,
		Available:  isOpenWebUI,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
