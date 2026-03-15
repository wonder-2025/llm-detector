package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// DifyPlugin Dify探测插件
type DifyPlugin struct {
	client *http.Client
}

// NewDifyPlugin 创建Dify插件
func NewDifyPlugin(timeout time.Duration) *DifyPlugin {
	return &DifyPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *DifyPlugin) Name() string {
	return "dify"
}

// Version 返回插件版本
func (p *DifyPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *DifyPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/api/apps",
		"/api/conversations",
		"/console/api/apps",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "dify"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "dify",
		Available: false,
		Error:     "Dify endpoints not found",
	}, nil
}

func (p *DifyPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "dify", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "dify", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查Dify特征
	isDify := false
	if resp.Header.Get("dify-version") != "" {
		isDify = true
	}
	if containsAny(bodyStr, []string{"dify", "conversation", "app", "workflow"}) {
		isDify = true
	}

	return &plugins.APIResult{
		Type:       "dify",
		Endpoint:   endpoint,
		Available:  isDify,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
