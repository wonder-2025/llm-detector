package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// HugeGraphPlugin Apache HugeGraph探测插件
type HugeGraphPlugin struct {
	client *http.Client
}

// NewHugeGraphPlugin 创建HugeGraph插件
func NewHugeGraphPlugin(timeout time.Duration) *HugeGraphPlugin {
	return &HugeGraphPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *HugeGraphPlugin) Name() string {
	return "hugegraph"
}

// Version 返回插件版本
func (p *HugeGraphPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *HugeGraphPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/graphs",
		"/versions",
		"/api/v1/graphs",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "hugegraph"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "hugegraph",
		Available: false,
		Error:     "HugeGraph endpoints not found",
	}, nil
}

func (p *HugeGraphPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "hugegraph", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "hugegraph", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查HugeGraph特征
	isHugeGraph := false
	if containsAny(bodyStr, []string{"hugegraph", "graph", "vertex", "edge"}) {
		isHugeGraph = true
	}

	return &plugins.APIResult{
		Type:       "hugegraph",
		Endpoint:   endpoint,
		Available:  isHugeGraph,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
