package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// MLflowPlugin MLflow探测插件
type MLflowPlugin struct {
	client *http.Client
}

// NewMLflowPlugin 创建MLflow插件
func NewMLflowPlugin(timeout time.Duration) *MLflowPlugin {
	return &MLflowPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *MLflowPlugin) Name() string {
	return "mlflow"
}

// Version 返回插件版本
func (p *MLflowPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *MLflowPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/api/2.0/mlflow/experiments/list",
		"/api/2.0/mlflow/runs/search",
		"/#/experiments",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "mlflow"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "mlflow",
		Available: false,
		Error:     "MLflow endpoints not found",
	}, nil
}

func (p *MLflowPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "mlflow", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "mlflow", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查MLflow特征
	isMLflow := false
	if resp.Header.Get("mlflow-version") != "" {
		isMLflow = true
	}
	if containsAny(bodyStr, []string{"mlflow", "experiment", "run", "artifact"}) {
		isMLflow = true
	}

	return &plugins.APIResult{
		Type:       "mlflow",
		Endpoint:   endpoint,
		Available:  isMLflow,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
