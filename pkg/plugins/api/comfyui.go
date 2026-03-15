package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// ComfyUIPlugin ComfyUI探测插件
type ComfyUIPlugin struct {
	client *http.Client
}

// NewComfyUIPlugin 创建ComfyUI插件
func NewComfyUIPlugin(timeout time.Duration) *ComfyUIPlugin {
	return &ComfyUIPlugin{
		client: NewHTTPClient(timeout),
	}
}

// Name 返回插件名称
func (p *ComfyUIPlugin) Name() string {
	return "comfyui"
}

// Version 返回插件版本
func (p *ComfyUIPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *ComfyUIPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	endpoints := []string{
		"/prompt",
		"/object_info",
		"/history",
		"/",
	}

	for _, endpoint := range endpoints {
		result := p.probeEndpoint(ctx, baseURL, endpoint)
		if result.Available {
			result.Type = "comfyui"
			return result, nil
		}
	}

	return &plugins.APIResult{
		Type:      "comfyui",
		Available: false,
		Error:     "ComfyUI endpoints not found",
	}, nil
}

func (p *ComfyUIPlugin) probeEndpoint(ctx context.Context, baseURL, endpoint string) *plugins.APIResult {
	url := baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &plugins.APIResult{Type: "comfyui", Available: false, Error: err.Error()}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &plugins.APIResult{Type: "comfyui", Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查ComfyUI特征
	isComfyUI := false
	if containsAny(bodyStr, []string{"comfyui", "comfy", "workflow", "node", "prompt"}) {
		isComfyUI = true
	}

	return &plugins.APIResult{
		Type:       "comfyui",
		Endpoint:   endpoint,
		Available:  isComfyUI,
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       bodyStr,
	}
}
