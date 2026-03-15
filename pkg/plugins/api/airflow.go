package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// AirflowPlugin Apache Airflow探测插件
type AirflowPlugin struct {
	client *http.Client
}

// NewAirflowPlugin 创建Airflow插件
func NewAirflowPlugin(timeout time.Duration) *AirflowPlugin {
	return &AirflowPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *AirflowPlugin) Name() string {
	return "airflow"
}

// Version 返回插件版本
func (p *AirflowPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *AirflowPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/api/v1/dags",
		"/api/v1/dagRuns",
		"/home",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "airflow"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "airflow",
		Available: false,
		Error:     "Airflow endpoints not found",
	}, nil
}

func (p *AirflowPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "airflow", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "airflow", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查Airflow特征
	isAirflow := false
	if resp.Header.Get("airflow-version") != "" {
		isAirflow = true
	}
	if containsAny(bodyStr, []string{"airflow", "dag", "task", "scheduler"}) {
		isAirflow = true
	}

	return &plugins.APIResult{
		Type:       "airflow",
		Endpoint:   endpoint,
		Available:  isAirflow,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
