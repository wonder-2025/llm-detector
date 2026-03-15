package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// AttuPlugin Attu (Milvus GUI)探测插件
type AttuPlugin struct {
	client  *http.Client
	timeout time.Duration
}

// NewAttuPlugin 创建Attu插件
func NewAttuPlugin(timeout time.Duration) *AttuPlugin {
	return &AttuPlugin{
		client: NewHTTPClientWithRedirect(timeout, 10),
	}
}

// Name 返回插件名称
func (p *AttuPlugin) Name() string {
	return "attu"
}

// Version 返回插件版本
func (p *AttuPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *AttuPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	// Attu常见端点
	endpoints := []struct {
		path   string
		method string
	}{
		{"/", "GET"},
		{"/#/", "GET"},
		{"/connect", "GET"},
		{"/api/v1/milvus/connect", "GET"},
		{"/api/v1/milvus/check", "GET"},
		{"/api/v1/collections", "GET"},
		{"/static", "GET"},
		{"/index.html", "GET"},
	}

	for _, ep := range endpoints {
		url := baseURL + ep.path
		
		req, err := http.NewRequestWithContext(ctx, ep.method, url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "text/html,application/json")
		req.Header.Set("User-Agent", "LLM-Detector/1.0")

		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 检查Attu特征
		if p.isAttuResponse(resp, string(body)) {
			headers := make(map[string]string)
			for key, values := range resp.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			return &plugins.APIResult{
				Type:       "attu",
				Endpoint:   url,
				Available:  true,
				StatusCode: resp.StatusCode,
				Headers:    headers,
				Body:       string(body),
			}, nil
		}
	}

	return nil, fmt.Errorf("Attu endpoints not found")
}

// isAttuResponse 检查是否是Attu响应
func (p *AttuPlugin) isAttuResponse(resp *http.Response, body string) bool {
	// 检查HTML内容特征
	bodyLower := toLowerStr(body)
	
	attuIndicators := []string{
		"attu",
		"milvus",
		"zilliz",
		"vector database",
		"attu-ui",
		"attu-logo",
		"milvus-admin",
	}
	
	for _, indicator := range attuIndicators {
		if containsStr(bodyLower, indicator) {
			return true
		}
	}
	
	// 检查特定的HTML title
	if containsStr(bodyLower, "<title>attu</title>") ||
		containsStr(bodyLower, "<title>attu - milvus") ||
		containsStr(bodyLower, "<title>milvus admin") {
		return true
	}
	
	// 检查JS/CSS文件路径
	if containsStr(bodyLower, "/static/js/attu") ||
		containsStr(bodyLower, "/static/css/attu") ||
		containsStr(bodyLower, "attu.min.js") ||
		containsStr(bodyLower, "attu.min.css") {
		return true
	}
	
	// 检查API响应
	if containsStr(bodyLower, `"code":`) && 
		(containsStr(bodyLower, `"data":`) || containsStr(bodyLower, `"message":`)) {
		// 可能是Milvus API响应
		if containsStr(bodyLower, `"collections"`) ||
			containsStr(bodyLower, `"connect"`) ||
			containsStr(bodyLower, `"milvus"`) {
			return true
		}
	}

	return false
}
