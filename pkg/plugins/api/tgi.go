package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// TGIPlugin TGI (Text Generation Inference) 探测插件
type TGIPlugin struct {
	client *http.Client
}

// NewTGIPlugin 创建TGI插件
func NewTGIPlugin(timeout time.Duration) *TGIPlugin {
	return &TGIPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *TGIPlugin) Name() string {
	return "tgi"
}

// Version 返回插件版本
func (p *TGIPlugin) Version() string {
	return "1.0.0"
}

// Detect 探测TGI API
func (p *TGIPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	endpoints := []string{
		"/info",
		"/health",
		"/generate",
		"/generate_stream",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, target, endpoint)
		if result.Available {
			result.Type = "tgi"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "tgi",
		Available: false,
		Error:     "TGI API endpoints not found",
	}, nil
}

func (p *TGIPlugin) probeEndpoint(ctx context.Context, target plugins.Target, endpoint string) *plugins.APIResult {
	url := target.BaseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{
			Type:      "tgi",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{
			Type:      "tgi",
			Endpoint:  endpoint,
			Available: false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 检查是否为TGI特征
	isTGI := p.checkTGIFeatures(resp, string(body))

	result := &plugins.APIResult{
		Type:       "tgi",
		Endpoint:   endpoint,
		Available:  isTGI,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	if !isTGI {
		result.Error = "Not a TGI API"
	}

	return result
}

func (p *TGIPlugin) checkTGIFeatures(resp *http.Response, body string) bool {
	// 检查响应头特征
	if resp.Header.Get("x-tgi-version") != "" {
		return true
	}

	// 检查响应体特征
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// TGI info端点返回模型信息
		if _, ok := data["model_id"]; ok {
			return true
		}
		if _, ok := data["model_pipeline_tag"]; ok {
			return true
		}
		if _, ok := data["sha"]; ok {
			return true
		}
	}

	return false
}
