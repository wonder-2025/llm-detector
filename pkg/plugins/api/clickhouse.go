package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// ClickHousePlugin ClickHouse探测插件
type ClickHousePlugin struct {
	client *http.Client
}

// NewClickHousePlugin 创建ClickHouse插件
func NewClickHousePlugin(timeout time.Duration) *ClickHousePlugin {
	return &ClickHousePlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *ClickHousePlugin) Name() string {
	return "clickhouse"
}

// Version 返回插件版本
func (p *ClickHousePlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *ClickHousePlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/ping",
		"/play",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "clickhouse"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "clickhouse",
		Available: false,
		Error:     "ClickHouse endpoints not found",
	}, nil
}

func (p *ClickHousePlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "clickhouse", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "clickhouse", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查ClickHouse特征
	isClickHouse := false
	if resp.Header.Get("clickhouse-server") != "" {
		isClickHouse = true
	}
	if resp.Header.Get("x-clickhouse-summary") != "" {
		isClickHouse = true
	}
	if containsAny(bodyStr, []string{"clickhouse", "clickhouse-server"}) {
		isClickHouse = true
	}

	return &plugins.APIResult{
		Type:       "clickhouse",
		Endpoint:   endpoint,
		Available:  isClickHouse,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
