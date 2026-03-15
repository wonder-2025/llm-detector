package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Matcher 通用匹配器
type Matcher struct {
	regexCache *RegexCache
}

// NewMatcher 创建匹配器
func NewMatcher() *Matcher {
	return &Matcher{
		regexCache: NewRegexCache(),
	}
}

// NewMatcherWithCache 使用指定缓存创建匹配器
func NewMatcherWithCache(cache *RegexCache) *Matcher {
	return &Matcher{
		regexCache: cache,
	}
}

// MatchResult 匹配结果
type MatchResult struct {
	Matched   bool                   `json:"matched"`
	Score     float64                `json:"score"`
	Matches   []string               `json:"matches"`
	Data      map[string]interface{} `json:"data"`
}

// MatchHeader 匹配HTTP头
func (m *Matcher) MatchHeader(headers map[string]string, name, pattern string, required bool) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	// 获取头值（不区分大小写）
	var value string
	var exists bool
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			value = v
			exists = true
			break
		}
	}

	if !exists {
		if required {
			result.Score = -0.3
		}
		return result
	}

	result.Data["header_name"] = name
	result.Data["header_value"] = value

	// 执行匹配
	if pattern != "" {
		if m.regexCache.Match(pattern, value) {
			result.Matched = true
			result.Score = 1.0
			result.Matches = append(result.Matches, fmt.Sprintf("%s: %s", name, value))

			// 提取捕获组
			submatches := m.regexCache.FindStringSubmatch(pattern, value)
			if len(submatches) > 1 {
				result.Data["capture_groups"] = submatches[1:]
			}
		} else {
			result.Score = 0.1 // 头存在但模式不匹配
		}
	} else {
		// 只检查存在性
		result.Matched = true
		result.Score = 0.7
		result.Matches = append(result.Matches, name)
	}

	return result
}

// MatchBody 匹配响应体
func (m *Matcher) MatchBody(body string, pattern string, isRegex bool) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	if pattern == "" {
		return result
	}

	if isRegex {
		if m.regexCache.Match(pattern, body) {
			result.Matched = true
			result.Score = 1.0
			matches := m.regexCache.FindAllString(pattern, body, -1)
			result.Matches = append(result.Matches, matches...)

			// 提取子匹配
			submatches := m.regexCache.FindStringSubmatch(pattern, body)
			if len(submatches) > 1 {
				result.Data["capture_groups"] = submatches[1:]
			}
		}
	} else {
		// 简单字符串匹配
		if strings.Contains(body, pattern) {
			result.Matched = true
			result.Score = 0.8
			result.Matches = append(result.Matches, pattern)
		}
	}

	return result
}

// MatchJSONPath 匹配JSON路径
func (m *Matcher) MatchJSONPath(body string, path string) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return result
	}

	// 简单的路径解析 (支持 a.b.c 或 a/b/c 格式)
	path = strings.ReplaceAll(path, "/", ".")
	parts := strings.Split(path, ".")

	current := interface{}(data)
	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return result
			}
		case []interface{}:
			// 尝试作为索引解析
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return result
			}
		default:
			return result
		}
	}

	result.Matched = true
	result.Score = 1.0
	result.Data["value"] = current
	result.Data["path"] = path

	// 转换为字符串
	switch v := current.(type) {
	case string:
		result.Matches = append(result.Matches, v)
	case float64:
		result.Matches = append(result.Matches, fmt.Sprintf("%v", v))
	case bool:
		result.Matches = append(result.Matches, fmt.Sprintf("%v", v))
	default:
		if jsonBytes, err := json.Marshal(v); err == nil {
			result.Matches = append(result.Matches, string(jsonBytes))
		}
	}

	return result
}

// MatchKeywords 关键词匹配
func (m *Matcher) MatchKeywords(body string, keywords []string, matchAll bool) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	if len(keywords) == 0 {
		return result
	}

	bodyLower := strings.ToLower(body)
	matchedCount := 0

	for _, keyword := range keywords {
		if strings.Contains(bodyLower, strings.ToLower(keyword)) {
			matchedCount++
			result.Matches = append(result.Matches, keyword)
		}
	}

	if matchAll {
		result.Matched = matchedCount == len(keywords)
	} else {
		result.Matched = matchedCount > 0
	}

	if len(keywords) > 0 {
		result.Score = float64(matchedCount) / float64(len(keywords))
	}

	result.Data["total_keywords"] = len(keywords)
	result.Data["matched_count"] = matchedCount

	return result
}

// MatchPatterns 多模式匹配
func (m *Matcher) MatchPatterns(body string, patterns []string) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	if len(patterns) == 0 {
		return result
	}

	matchedCount := 0
	var allMatches []string

	for _, pattern := range patterns {
		if m.regexCache.Match(pattern, body) {
			matchedCount++
			matches := m.regexCache.FindAllString(pattern, body, -1)
			allMatches = append(allMatches, matches...)
		}
	}

	result.Matched = matchedCount > 0
	result.Matches = allMatches

	if len(patterns) > 0 {
		result.Score = float64(matchedCount) / float64(len(patterns))
	}

	result.Data["total_patterns"] = len(patterns)
	result.Data["matched_count"] = matchedCount

	return result
}

// MatchEndpoint 端点匹配
func (m *Matcher) MatchEndpoint(endpoint string, patterns []string) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	if len(patterns) == 0 {
		return result
	}

	for _, pattern := range patterns {
		// 支持精确匹配和通配符
		if pattern == endpoint {
			result.Matched = true
			result.Score = 1.0
			result.Matches = append(result.Matches, pattern)
			return result
		}

		// 通配符匹配
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(endpoint, prefix) {
				result.Matched = true
				result.Score = 0.9
				result.Matches = append(result.Matches, pattern)
				return result
			}
		}

		// 正则匹配
		if m.regexCache.Match(pattern, endpoint) {
			result.Matched = true
			result.Score = 0.95
			result.Matches = append(result.Matches, pattern)
			return result
		}
	}

	return result
}

// ExtractVersion 提取版本号
func (m *Matcher) ExtractVersion(text string, patterns []string) string {
	for _, pattern := range patterns {
		re := m.regexCache.Get(pattern)
		if re == nil {
			continue
		}

		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1] // 返回第一个捕获组
		}
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

// ExtractField 提取JSON字段
func (m *Matcher) ExtractField(body string, field string) (interface{}, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, false
	}

	val, exists := data[field]
	return val, exists
}

// ValidateJSON 验证JSON结构
func (m *Matcher) ValidateJSON(body string, requiredFields []string) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		result.Data["error"] = err.Error()
		return result
	}

	result.Matched = true
	result.Score = 1.0

	// 检查必需字段
	missingFields := []string{}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			missingFields = append(missingFields, field)
			result.Score -= 0.1
		} else {
			result.Matches = append(result.Matches, field)
		}
	}

	if len(missingFields) > 0 {
		result.Data["missing_fields"] = missingFields
		result.Matched = false
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result
}

// MatchComposite 复合匹配（多种条件组合）
func (m *Matcher) MatchComposite(body string, headers map[string]string, conditions []MatchCondition) *MatchResult {
	result := &MatchResult{
		Matched: false,
		Score:   0,
		Matches: []string{},
		Data:    make(map[string]interface{}),
	}

	if len(conditions) == 0 {
		return result
	}

	matchedCount := 0
	var totalWeight float64

	for _, cond := range conditions {
		subResult := m.matchCondition(body, headers, cond)
		totalWeight += cond.Weight

		if subResult.Matched {
			matchedCount++
			result.Score += subResult.Score * cond.Weight
			result.Matches = append(result.Matches, subResult.Matches...)
		}
	}

	if totalWeight > 0 {
		result.Score = result.Score / totalWeight
	}

	// 所有条件都匹配才算成功
	result.Matched = matchedCount == len(conditions)

	return result
}

// MatchCondition 匹配条件
type MatchCondition struct {
	Type     string  `json:"type"`     // header, body, json_path, keyword
	Target   string  `json:"target"`   // 目标字段/路径
	Pattern  string  `json:"pattern"`  // 匹配模式
	Required bool    `json:"required"` // 是否必需
	Weight   float64 `json:"weight"`   // 权重
}

// matchCondition 匹配单个条件
func (m *Matcher) matchCondition(body string, headers map[string]string, cond MatchCondition) *MatchResult {
	switch cond.Type {
	case "header":
		return m.MatchHeader(headers, cond.Target, cond.Pattern, cond.Required)
	case "body":
		isRegex := true
		return m.MatchBody(body, cond.Pattern, isRegex)
	case "json_path":
		return m.MatchJSONPath(body, cond.Target)
	case "keyword":
		keywords := []string{cond.Pattern}
		return m.MatchKeywords(body, keywords, false)
	default:
		return &MatchResult{Matched: false, Score: 0}
	}
}

// CompilePatterns 预编译正则模式
func (m *Matcher) CompilePatterns(patterns []string) error {
	for _, pattern := range patterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		// 触发缓存
		m.regexCache.Get(pattern)
	}
	return nil
}

// GetCacheStats 获取缓存统计
func (m *Matcher) GetCacheStats() CacheStats {
	return m.regexCache.GetStats()
}
