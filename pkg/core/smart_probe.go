package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SmartProber 智能探测
type SmartProber struct {
	client  *http.Client
	timeout time.Duration
}

// NewSmartProber 创建智能探测器
func NewSmartProber(timeout time.Duration) *SmartProber {
	return &SmartProber{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// DiscoveryResult 发现结果
type DiscoveryResult struct {
	Found       bool
	Type        string
	Endpoint    string
	Method      string
	Confidence  float64
	Indicators  []string
}

// DiscoverEndpoints 智能发现端点
func (p *SmartProber) DiscoverEndpoints(ctx context.Context, baseURL string) []*DiscoveryResult {
	var results []*DiscoveryResult

	// 1. 尝试根路径获取信息
	if r := p.probeRoot(ctx, baseURL); r != nil {
		results = append(results, r)
	}

	// 2. 尝试常见API前缀
	prefixes := []string{"/api", "/v1", "/v2", "/llm", "/ai", "/model", "/inference"}
	for _, prefix := range prefixes {
		if r := p.probePrefix(ctx, baseURL, prefix); r != nil {
			results = append(results, r)
		}
	}

	// 3. 尝试发送测试请求触发错误信息
	if r := p.probeWithTestRequest(ctx, baseURL); r != nil {
		results = append(results, r)
	}

	// 4. 尝试OPTIONS请求发现端点
	if r := p.probeOptions(ctx, baseURL); r != nil {
		results = append(results, r)
	}

	return results
}

// probeRoot 探测根路径
func (p *SmartProber) probeRoot(ctx context.Context, baseURL string) *DiscoveryResult {
	url := baseURL + "/"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 分析响应识别服务类型
	indicators := p.analyzeResponse(resp, bodyStr)
	if len(indicators) > 0 {
		return &DiscoveryResult{
			Found:      true,
			Type:       p.detectTypeFromIndicators(indicators),
			Endpoint:   url,
			Method:     "GET",
			Confidence: float64(len(indicators)) * 0.2,
			Indicators: indicators,
		}
	}

	return nil
}

// probePrefix 探测API前缀
func (p *SmartProber) probePrefix(ctx context.Context, baseURL, prefix string) *DiscoveryResult {
	// 尝试在前缀下找常见端点
	subPaths := []string{"", "/", "/chat", "/generate", "/complete", "/models", "/health"}
	
	for _, sub := range subPaths {
		url := baseURL + prefix + sub
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		indicators := p.analyzeResponse(resp, string(body))
		if len(indicators) > 0 {
			return &DiscoveryResult{
				Found:      true,
				Type:       p.detectTypeFromIndicators(indicators),
				Endpoint:   url,
				Method:     "GET",
				Confidence: float64(len(indicators)) * 0.2,
				Indicators: indicators,
			}
		}
	}

	return nil
}

// probeWithTestRequest 发送测试请求触发错误
func (p *SmartProber) probeWithTestRequest(ctx context.Context, baseURL string) *DiscoveryResult {
	// 尝试向根路径发送POST请求（通常会返回错误，但错误信息可能暴露服务类型）
	testPayloads := []struct {
		path    string
		payload string
	}{
		{"/", `{"test": true}`},
		{"/api", `{"messages": []}`},
		{"/chat", `{"prompt": "hi"}`},
	}

	for _, test := range testPayloads {
		url := baseURL + test.path
		
		req, err := http.NewRequestWithContext(ctx, "POST", url, 
			bytes.NewBufferString(test.payload))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		indicators := p.analyzeErrorResponse(resp, string(body))
		if len(indicators) > 0 {
			return &DiscoveryResult{
				Found:      true,
				Type:       p.detectTypeFromIndicators(indicators),
				Endpoint:   url,
				Method:     "POST",
				Confidence: float64(len(indicators)) * 0.25,
				Indicators: indicators,
			}
		}
	}

	return nil
}

// probeOptions 使用OPTIONS请求
func (p *SmartProber) probeOptions(ctx context.Context, baseURL string) *DiscoveryResult {
	url := baseURL + "/"
	
	req, err := http.NewRequestWithContext(ctx, "OPTIONS", url, nil)
	if err != nil {
		return nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	// 检查Allow头
	allow := resp.Header.Get("Allow")
	if allow != "" {
		// 如果支持POST，可能是API端点
		if strings.Contains(allow, "POST") {
			return &DiscoveryResult{
				Found:      true,
				Type:       "unknown_api",
				Endpoint:   url,
				Method:     "OPTIONS",
				Confidence: 0.3,
				Indicators: []string{"supports POST"},
			}
		}
	}

	return nil
}

// analyzeResponse 分析响应特征
func (p *SmartProber) analyzeResponse(resp *http.Response, body string) []string {
	var indicators []string

	// 检查响应头
	headerChecks := map[string]string{
		"openai-model":       "openai",
		"x-request-id":       "openai",
		"anthropic-ratelimit": "anthropic",
		"ollama-version":     "ollama",
		"x-vllm-executor":    "vllm",
		"x-litellm-version":  "litellm",
		"x-tgi-version":      "tgi",
		"server":             "server_type",
	}

	for header, indicator := range headerChecks {
		if resp.Header.Get(header) != "" {
			indicators = append(indicators, indicator+"_header")
		}
	}

	// 检查响应体JSON字段
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		fieldChecks := map[string]string{
			"model":           "model_field",
			"choices":         "choices_field",
			"messages":        "messages_field",
			"completion":      "completion_field",
			"generated_text":  "generated_text_field",
			"response":        "response_field",
			"data":            "data_array",
			"models":          "models_array",
		}

		for field, indicator := range fieldChecks {
			if _, ok := data[field]; ok {
				indicators = append(indicators, indicator)
			}
		}
	}

	// 检查HTML内容（可能是文档页面）
	if strings.Contains(body, "swagger") || strings.Contains(body, "Swagger") {
		indicators = append(indicators, "swagger_ui")
	}
	if strings.Contains(body, "redoc") || strings.Contains(body, "ReDoc") {
		indicators = append(indicators, "redoc")
	}
	if strings.Contains(body, "fastapi") || strings.Contains(body, "FastAPI") {
		indicators = append(indicators, "fastapi")
	}

	return indicators
}

// analyzeErrorResponse 分析错误响应
func (p *SmartProber) analyzeErrorResponse(resp *http.Response, body string) []string {
	var indicators []string

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// 检查错误结构
		if errData, ok := data["error"].(map[string]interface{}); ok {
			if msg, ok := errData["message"].(string); ok {
				// 根据错误消息判断类型
				if strings.Contains(msg, "model") {
					indicators = append(indicators, "model_error")
				}
				if strings.Contains(msg, "api key") || strings.Contains(msg, "authorization") {
					indicators = append(indicators, "auth_error")
				}
				if strings.Contains(msg, "rate limit") {
					indicators = append(indicators, "rate_limit")
				}
				if strings.Contains(msg, "token") {
					indicators = append(indicators, "token_error")
				}
			}

			// 检查错误类型
			if errType, ok := errData["type"].(string); ok {
				if errType == "invalid_request_error" {
					indicators = append(indicators, "openai_error_format")
				}
			}
		}

		// 检查detail字段（FastAPI风格）
		if detail, ok := data["detail"].(string); ok && detail != "" {
			indicators = append(indicators, "fastapi_error_format")
		}
	}

	return indicators
}

// detectTypeFromIndicators 根据指标判断类型
func (p *SmartProber) detectTypeFromIndicators(indicators []string) string {
	typeScores := map[string]int{
		"openai":    0,
		"ollama":    0,
		"vllm":      0,
		"tgi":       0,
		"litellm":   0,
		"fastapi":   0,
		"anthropic": 0,
	}

	for _, ind := range indicators {
		switch {
		case strings.Contains(ind, "openai"):
			typeScores["openai"]++
		case strings.Contains(ind, "ollama"):
			typeScores["ollama"]++
		case strings.Contains(ind, "vllm"):
			typeScores["vllm"]++
		case strings.Contains(ind, "tgi"):
			typeScores["tgi"]++
		case strings.Contains(ind, "litellm"):
			typeScores["litellm"]++
		case strings.Contains(ind, "fastapi"):
			typeScores["fastapi"]++
		case strings.Contains(ind, "anthropic"):
			typeScores["anthropic"]++
		case strings.Contains(ind, "model") || strings.Contains(ind, "choices"):
			// 通用LLM特征，给所有类型加分
			for k := range typeScores {
				typeScores[k]++
			}
		}
	}

	// 找出得分最高的类型
	maxScore := 0
	bestType := "unknown_llm"
	for t, s := range typeScores {
		if s > maxScore {
			maxScore = s
			bestType = t
		}
	}

	if maxScore == 0 {
		return "unknown_api"
	}

	return bestType
}

// SmartDiscovery 智能发现入口
func (p *SmartProber) SmartDiscovery(ctx context.Context, target *Target) ([]*DiscoveryResult, []APIResult) {
	baseURL := target.BaseURL()
	
	// 如果是纯IP，尝试多个端口
	var allResults []*DiscoveryResult
	var apiResults []APIResult

	if target.Type == TargetIP {
		ports := []int{8080, 8000, 3000, 5000, 11434, 4000}
		for _, port := range ports {
			url := fmt.Sprintf("http://%s:%d", target.Host, port)
			results := p.DiscoverEndpoints(ctx, url)
			allResults = append(allResults, results...)
			
			// 转换为APIResult
			for _, r := range results {
				if r.Found {
					apiResults = append(apiResults, APIResult{
						Type:       r.Type,
						Endpoint:   r.Endpoint,
						Available:  true,
						Confidence: r.Confidence,
					})
				}
			}
		}
	} else {
		results := p.DiscoverEndpoints(ctx, baseURL)
		allResults = results
		
		for _, r := range results {
			if r.Found {
				apiResults = append(apiResults, APIResult{
					Type:       r.Type,
					Endpoint:   r.Endpoint,
					Available:  true,
					Confidence: r.Confidence,
				})
			}
		}
	}

	return allResults, apiResults
}
