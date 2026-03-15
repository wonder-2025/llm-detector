package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"llm-detector/pkg/plugins"
)

// JupyterPlugin Jupyter Notebook/JupyterHub探测插件
type JupyterPlugin struct {
	client  *http.Client
	timeout time.Duration
}

// NewJupyterPlugin 创建Jupyter插件
func NewJupyterPlugin(timeout time.Duration) *JupyterPlugin {
	return &JupyterPlugin{
		client: NewHTTPClientWithRedirect(timeout, 10),
	}
}

// Name 返回插件名称
func (p *JupyterPlugin) Name() string {
	return "jupyter"
}

// Version 返回插件版本
func (p *JupyterPlugin) Version() string {
	return "1.0.0"
}

// Detect 执行探测
func (p *JupyterPlugin) Detect(ctx context.Context, target plugins.Target) (*plugins.APIResult, error) {
	baseURL := target.BaseURL()

	// Jupyter常见端点 - 扩展更多路径
	endpoints := []struct {
		path   string
		method string
	}{
		{"/", "GET"},
		{"/api", "GET"},
		{"/api/status", "GET"},
		{"/api/sessions", "GET"},
		{"/api/kernels", "GET"},
		{"/api/contents", "GET"},
		{"/api/kernelspecs", "GET"},
		{"/tree", "GET"},
		{"/lab", "GET"},
		{"/lab/api/workspaces", "GET"},
		{"/hub/", "GET"},
		{"/hub/api", "GET"},
		{"/hub/api/users", "GET"},
		{"/hub/api/status", "GET"},
		{"/login", "GET"},
		{"/static/favicons/favicon.ico", "GET"},
	}

	for _, ep := range endpoints {
		url := baseURL + ep.path
		
		req, err := http.NewRequestWithContext(ctx, ep.method, url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "LLM-Detector/1.0")

		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 检查Jupyter特征
		if p.isJupyterResponse(resp, string(body)) {
			headers := make(map[string]string)
			for key, values := range resp.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			// 提取版本信息
			version := p.extractVersion(string(body))
			if version != "" {
				headers["jupyter-version"] = version
			}

			return &plugins.APIResult{
				Type:       "jupyter",
				Endpoint:   url,
				Available:  true,
				StatusCode: resp.StatusCode,
				Headers:    headers,
				Body:       string(body),
			}, nil
		}
	}

	return nil, fmt.Errorf("Jupyter endpoints not found")
}

// isJupyterResponse 检查是否是Jupyter响应
func (p *JupyterPlugin) isJupyterResponse(resp *http.Response, body string) bool {
	// 检查Server头
	server := resp.Header.Get("Server")
	if server != "" && (containsLower(server, "jupyter") || 
		containsLower(server, "tornado") ||
		containsLower(server, "ipython")) {
		return true
	}

	// 检查Set-Cookie中的jupyter token
	cookies := resp.Header.Values("Set-Cookie")
	for _, cookie := range cookies {
		if containsLower(cookie, "jupyter") || containsLower(cookie, "jupyterhub") ||
			containsLower(cookie, "jupyter-session") || containsLower(cookie, "_xsrf") {
			return true
		}
	}

	// 检查XSRF token头 (Jupyter特征)
	if resp.Header.Get("X-Xsrftoken") != "" || resp.Header.Get("X-Xsrf-Token") != "" {
		return true
	}

	// 检查响应体
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// Jupyter API特征字段
		jupyterFields := []string{
			"version", "notebook_version", "kernelspecs",
			"sessions", "kernels", "content", "name", "path", "type",
			"workspace", "items", "last_modified", "created",
		}
		
		for _, field := range jupyterFields {
			if _, ok := data[field]; ok {
				return true
			}
		}

		// 检查是否是kernelspecs
		if _, ok := data["kernelspecs"]; ok {
			return true
		}

		// 检查是否有kernels列表
		if kernels, ok := data["kernels"].([]interface{}); ok && len(kernels) > 0 {
			return true
		}
		
		// 检查是否有error字段包含jupyter特征
		if errData, ok := data["error"].(string); ok {
			if containsLower(errData, "jupyter") || containsLower(errData, "ipython") {
				return true
			}
		}
		if errMap, ok := data["error"].(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				if containsLower(msg, "jupyter") || containsLower(msg, "ipython") ||
					containsLower(msg, "xsrf") || containsLower(msg, "token") {
					return true
				}
			}
		}
	}

	// 检查HTML内容 (更宽松的匹配)
	bodyLower := toLowerStr(body)
	if containsStr(bodyLower, "jupyter") || 
		containsStr(bodyLower, "jupyterlab") ||
		containsStr(bodyLower, "kernelspec") ||
		containsStr(bodyLower, "ipython") ||
		containsStr(bodyLower, "notebook") ||
		containsStr(bodyLower, "_xsrf") ||
		containsStr(bodyLower, "jupyter-config-data") ||
		containsStr(bodyLower, "jupyterlab-workspaces") ||
		containsStr(bodyLower, "hub/login") ||
		containsStr(bodyLower, "singleuser") {
		return true
	}
	
	// 检查特定的HTML title
	if containsStr(bodyLower, "<title>jupyter") ||
		containsStr(bodyLower, "<title>jupyterhub") ||
		containsStr(bodyLower, "<title>server") {
		return true
	}

	return false
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && containsHelperStr(s, substr)
}

func containsHelperStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractVersion 提取Jupyter版本
func (p *JupyterPlugin) extractVersion(body string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		if version, ok := data["version"].(string); ok {
			return version
		}
		if version, ok := data["notebook_version"].(string); ok {
			return version
		}
	}
	return ""
}

func containsLower(s, substr string) bool {
	return len(s) >= len(substr) && containsHelperLower(s, substr)
}

func containsHelperLower(s, substr string) bool {
	sLower := toLowerStr(s)
	subLower := toLowerStr(substr)
	for i := 0; i <= len(sLower)-len(subLower); i++ {
		if sLower[i:i+len(subLower)] == subLower {
			return true
		}
	}
	return false
}

func toLowerStr(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
