package core

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"llm-detector/pkg/fingerprints"
)

// ScoringMode 评分模式
type ScoringMode int

const (
	ModeStrict ScoringMode = iota // 严格模式 - 高置信度阈值
	ModeLoose                     // 宽松模式 - 低置信度阈值
)

// ScoringWeights 评分权重配置
type ScoringWeights struct {
	HeaderMatch   float64 `yaml:"header_match" json:"header_match"`     // 响应头匹配权重 (默认30%)
	BodyKeywords  float64 `yaml:"body_keywords" json:"body_keywords"`   // 响应体关键词权重 (默认40%)
	JSONStructure float64 `yaml:"json_structure" json:"json_structure"` // JSON结构匹配权重 (默认30%)
}

// DefaultWeights 返回默认权重
func DefaultWeights() ScoringWeights {
	return ScoringWeights{
		HeaderMatch:   0.30,
		BodyKeywords:  0.40,
		JSONStructure: 0.30,
	}
}

// Validate 验证权重总和为1
func (w ScoringWeights) Validate() error {
	total := w.HeaderMatch + w.BodyKeywords + w.JSONStructure
	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("scoring weights must sum to 1.0, got %.2f", total)
	}
	return nil
}

// Scorer 评分引擎
type Scorer struct {
	weights      ScoringWeights
	mode         ScoringMode
	threshold    float64
	regexCache   *RegexCache
}

// ScoringResult 评分结果
type ScoringResult struct {
	Score         float64                `json:"score"`          // 总体置信度 (0-1)
	Matched       bool                   `json:"matched"`        // 是否匹配成功
	Version       string                 `json:"version"`        // 检测到的版本
	Details       ScoringDetails         `json:"details"`        // 详细评分
	MatchedRules  []string               `json:"matched_rules"`  // 匹配的规则列表
	ExtractedData map[string]interface{} `json:"extracted_data"` // 提取的数据
}

// ScoringDetails 详细评分
type ScoringDetails struct {
	HeaderScore   float64 `json:"header_score"`   // 响应头匹配得分
	BodyScore     float64 `json:"body_score"`     // 响应体匹配得分
	JSONScore     float64 `json:"json_score"`     // JSON结构匹配得分
	HeaderWeight  float64 `json:"header_weight"`  // 响应头权重
	BodyWeight    float64 `json:"body_weight"`    // 响应体权重
	JSONWeight    float64 `json:"json_weight"`    // JSON权重
}

// NewScorer 创建评分引擎
func NewScorer(weights ScoringWeights, mode ScoringMode) (*Scorer, error) {
	if err := weights.Validate(); err != nil {
		return nil, err
	}

	threshold := 0.7 // 默认阈值
	if mode == ModeLoose {
		threshold = 0.5
	}

	return &Scorer{
		weights:    weights,
		mode:       mode,
		threshold:  threshold,
		regexCache: NewRegexCache(),
	}, nil
}

// SetThreshold 设置置信度阈值
func (s *Scorer) SetThreshold(threshold float64) {
	s.threshold = threshold
}

// GetThreshold 获取当前阈值
func (s *Scorer) GetThreshold() float64 {
	return s.threshold
}

// ScoreFramework 对框架指纹进行评分
func (s *Scorer) ScoreFramework(fp *fingerprints.FrameworkFingerprint, apiResults []APIResult) *ScoringResult {
	result := &ScoringResult{
		Score:         0,
		Matched:       false,
		MatchedRules:  []string{},
		ExtractedData: make(map[string]interface{}),
	}

	if len(apiResults) == 0 {
		return result
	}

	var totalHeaderScore, totalBodyScore, totalJSONScore float64
	var headerCount, bodyCount, jsonCount int

	for _, apiResult := range apiResults {
		if !apiResult.Available {
			continue
		}

		// 1. 响应头匹配评分
		headerScore, headerMatches := s.scoreHeaders(fp, apiResult)
		if headerScore > 0 {
			totalHeaderScore += headerScore
			headerCount++
			result.MatchedRules = append(result.MatchedRules, headerMatches...)
		}

		// 2. 响应体关键词评分
		bodyScore, bodyMatches := s.scoreBodyPatterns(fp, apiResult)
		if bodyScore > 0 {
			totalBodyScore += bodyScore
			bodyCount++
			result.MatchedRules = append(result.MatchedRules, bodyMatches...)
		}

		// 3. JSON结构匹配评分
		jsonScore, jsonMatches, extractedData := s.scoreJSONStructure(fp, apiResult)
		if jsonScore > 0 {
			totalJSONScore += jsonScore
			jsonCount++
			result.MatchedRules = append(result.MatchedRules, jsonMatches...)
			// 合并提取的数据
			for k, v := range extractedData {
				result.ExtractedData[k] = v
			}
		}

		// 4. 版本识别
		version := s.extractVersion(fp, apiResult)
		if version != "" {
			result.Version = version
		}
	}

	// 计算平均分
	var finalHeaderScore, finalBodyScore, finalJSONScore float64
	if headerCount > 0 {
		finalHeaderScore = totalHeaderScore / float64(headerCount)
	}
	if bodyCount > 0 {
		finalBodyScore = totalBodyScore / float64(bodyCount)
	}
	if jsonCount > 0 {
		finalJSONScore = totalJSONScore / float64(jsonCount)
	}

	// 计算加权总分
	result.Score = finalHeaderScore*s.weights.HeaderMatch +
		finalBodyScore*s.weights.BodyKeywords +
		finalJSONScore*s.weights.JSONStructure

	result.Details = ScoringDetails{
		HeaderScore:  finalHeaderScore,
		BodyScore:    finalBodyScore,
		JSONScore:    finalJSONScore,
		HeaderWeight: s.weights.HeaderMatch,
		BodyWeight:   s.weights.BodyKeywords,
		JSONWeight:   s.weights.JSONStructure,
	}

	result.Matched = result.Score >= s.threshold

	return result
}

// ScoreModel 对模型指纹进行评分
func (s *Scorer) ScoreModel(ctx context.Context, fp *fingerprints.ModelFingerprint, apiResults []APIResult) *ScoringResult {
	result := &ScoringResult{
		Score:         0,
		Matched:       false,
		MatchedRules:  []string{},
		ExtractedData: make(map[string]interface{}),
	}

	if len(apiResults) == 0 {
		return result
	}

	var totalHeaderScore, totalBodyScore, totalJSONScore float64
	var headerCount, bodyCount, jsonCount int

	for _, apiResult := range apiResults {
		if !apiResult.Available {
			continue
		}

		// 1. 响应头匹配评分
		headerScore, headerMatches := s.scoreModelHeaders(fp, apiResult)
		if headerScore > 0 {
			totalHeaderScore += headerScore
			headerCount++
			result.MatchedRules = append(result.MatchedRules, headerMatches...)
		}

		// 2. 响应体关键词评分
		bodyScore, bodyMatches := s.scoreModelBodyPatterns(fp, apiResult)
		if bodyScore > 0 {
			totalBodyScore += bodyScore
			bodyCount++
			result.MatchedRules = append(result.MatchedRules, bodyMatches...)
		}

		// 3. JSON结构匹配评分
		jsonScore, jsonMatches, extractedData := s.scoreModelJSONStructure(fp, apiResult)
		if jsonScore > 0 {
			totalJSONScore += jsonScore
			jsonCount++
			result.MatchedRules = append(result.MatchedRules, jsonMatches...)
			for k, v := range extractedData {
				result.ExtractedData[k] = v
			}
		}
	}

	// 计算平均分
	var finalHeaderScore, finalBodyScore, finalJSONScore float64
	if headerCount > 0 {
		finalHeaderScore = totalHeaderScore / float64(headerCount)
	}
	if bodyCount > 0 {
		finalBodyScore = totalBodyScore / float64(bodyCount)
	}
	if jsonCount > 0 {
		finalJSONScore = totalJSONScore / float64(jsonCount)
	}

	// 计算加权总分
	result.Score = finalHeaderScore*s.weights.HeaderMatch +
		finalBodyScore*s.weights.BodyKeywords +
		finalJSONScore*s.weights.JSONStructure

	result.Details = ScoringDetails{
		HeaderScore:  finalHeaderScore,
		BodyScore:    finalBodyScore,
		JSONScore:    finalJSONScore,
		HeaderWeight: s.weights.HeaderMatch,
		BodyWeight:   s.weights.BodyKeywords,
		JSONWeight:   s.weights.JSONStructure,
	}

	result.Matched = result.Score >= s.threshold

	return result
}

// scoreHeaders 评分响应头匹配
func (s *Scorer) scoreHeaders(fp *fingerprints.FrameworkFingerprint, apiResult APIResult) (float64, []string) {
	if len(fp.Headers) == 0 {
		return 0, nil
	}

	var score float64
	var matches []string
	matchedCount := 0
	requiredCount := 0

	for _, header := range fp.Headers {
		if header.Required {
			requiredCount++
		}

		value, exists := apiResult.Headers[header.Name]
		if !exists {
			// 尝试不区分大小写匹配
			for k, v := range apiResult.Headers {
				if strings.EqualFold(k, header.Name) {
					value = v
					exists = true
					break
				}
			}
		}

		if !exists {
			if header.Required {
				// 必需头缺失，大幅扣分
				score -= 0.3
			}
			continue
		}

		// 检查值匹配
		if header.Pattern != "" {
			if s.regexCache.Match(header.Pattern, value) {
				score += 0.25
				matchedCount++
				matches = append(matches, fmt.Sprintf("header:%s=~%s", header.Name, header.Pattern))
			} else if header.Required {
				score -= 0.1
			}
		} else if header.Value != "" {
			if strings.EqualFold(value, header.Value) {
				score += 0.25
				matchedCount++
				matches = append(matches, fmt.Sprintf("header:%s=%s", header.Name, header.Value))
			} else if header.Required {
				score -= 0.1
			}
		} else {
			// 只要求存在
			score += 0.15
			matchedCount++
			matches = append(matches, fmt.Sprintf("header:%s", header.Name))
		}
	}

	// 归一化分数
	if len(fp.Headers) > 0 {
		score = score / float64(len(fp.Headers))
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
	}

	return score, matches
}

// scoreBodyPatterns 评分响应体模式匹配
func (s *Scorer) scoreBodyPatterns(fp *fingerprints.FrameworkFingerprint, apiResult APIResult) (float64, []string) {
	if len(fp.BodyPatterns) == 0 {
		return 0, nil
	}

	var score float64
	var matches []string
	body := strings.ToLower(apiResult.Body)

	for _, pattern := range fp.BodyPatterns {
		if pattern.Field != "" {
			// 检查JSON字段存在性
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(apiResult.Body), &data); err == nil {
				if _, exists := data[pattern.Field]; exists {
					score += 0.25
					matches = append(matches, fmt.Sprintf("body_field:%s", pattern.Field))
				} else if pattern.Required {
					score -= 0.15
				}
			}
		}

		if pattern.Pattern != "" {
			if s.regexCache.Match(pattern.Pattern, apiResult.Body) {
				score += 0.25
				matches = append(matches, fmt.Sprintf("body_pattern:%s", pattern.Pattern))
			}
		}

		if pattern.Value != "" {
			if strings.Contains(body, strings.ToLower(pattern.Value)) {
				score += 0.25
				matches = append(matches, fmt.Sprintf("body_value:%s", pattern.Value))
			}
		}
	}

	// 归一化
	if len(fp.BodyPatterns) > 0 {
		score = score / float64(len(fp.BodyPatterns))
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
	}

	return score, matches
}

// scoreJSONStructure 评分JSON结构匹配
func (s *Scorer) scoreJSONStructure(fp *fingerprints.FrameworkFingerprint, apiResult APIResult) (float64, []string, map[string]interface{}) {
	matches := []string{}
	extractedData := make(map[string]interface{})

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(apiResult.Body), &data); err != nil {
		return 0, nil, nil
	}

	var score float64

	// 检查端点匹配
	for _, endpoint := range fp.Endpoints {
		if apiResult.Endpoint == endpoint.Path {
			score += 0.4
			matches = append(matches, fmt.Sprintf("endpoint:%s", endpoint.Path))
		}
	}

	// 检查错误模式
	for _, errPattern := range fp.ErrorPatterns {
		if s.regexCache.Match(errPattern.Pattern, apiResult.Body) {
			score += 0.2
			matches = append(matches, fmt.Sprintf("error_pattern:%s", errPattern.Type))
		}
	}

	// 检查版本信息
	if fp.Versions != nil {
		for _, version := range fp.Versions {
			if s.regexCache.Match(version.Pattern, apiResult.Body) {
				score += 0.3
				matches = append(matches, fmt.Sprintf("version_pattern:%s", version.Pattern))
				// 提取特征
				for _, feature := range version.Features {
					extractedData[feature] = true
				}
			}
		}
	}

	// 归一化
	checkCount := len(fp.Endpoints) + len(fp.ErrorPatterns) + len(fp.Versions)
	if checkCount > 0 {
		score = score / float64(checkCount)
		if score > 1 {
			score = 1
		}
	}

	return score, matches, extractedData
}

// scoreModelHeaders 评分模型响应头
func (s *Scorer) scoreModelHeaders(fp *fingerprints.ModelFingerprint, apiResult APIResult) (float64, []string) {
	if len(fp.Response.Headers) == 0 {
		return 0, nil
	}

	var score float64
	var matches []string

	for _, header := range fp.Response.Headers {
		value, exists := apiResult.Headers[header.Name]
		if !exists {
			// 尝试不区分大小写匹配
			for k, v := range apiResult.Headers {
				if strings.EqualFold(k, header.Name) {
					value = v
					exists = true
					break
				}
			}
		}

		if !exists {
			if header.Required {
				score -= 0.3
			}
			continue
		}

		if header.Pattern != "" {
			if s.regexCache.Match(header.Pattern, value) {
				score += 0.25
				matches = append(matches, fmt.Sprintf("model_header:%s", header.Name))
			}
		} else {
			score += 0.15
			matches = append(matches, fmt.Sprintf("model_header:%s", header.Name))
		}
	}

	if len(fp.Response.Headers) > 0 {
		score = score / float64(len(fp.Response.Headers))
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
	}

	return score, matches
}

// scoreModelBodyPatterns 评分模型响应体模式
func (s *Scorer) scoreModelBodyPatterns(fp *fingerprints.ModelFingerprint, apiResult APIResult) (float64, []string) {
	if len(fp.Response.BodyPatterns) == 0 {
		return 0, nil
	}

	var score float64
	var matches []string
	body := strings.ToLower(apiResult.Body)

	for _, pattern := range fp.Response.BodyPatterns {
		if pattern.Field != "" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(apiResult.Body), &data); err == nil {
				if _, exists := data[pattern.Field]; exists {
					score += 0.25
					matches = append(matches, fmt.Sprintf("model_field:%s", pattern.Field))
				}
			}
		}

		if pattern.Pattern != "" {
			if s.regexCache.Match(pattern.Pattern, apiResult.Body) {
				score += 0.25
				matches = append(matches, fmt.Sprintf("model_pattern:%s", pattern.Pattern))
			}
		}

		if pattern.Value != "" {
			if strings.Contains(body, strings.ToLower(pattern.Value)) {
				score += 0.25
				matches = append(matches, fmt.Sprintf("model_value:%s", pattern.Value))
			}
		}
	}

	if len(fp.Response.BodyPatterns) > 0 {
		score = score / float64(len(fp.Response.BodyPatterns))
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
	}

	return score, matches
}

// scoreModelJSONStructure 评分模型JSON结构
func (s *Scorer) scoreModelJSONStructure(fp *fingerprints.ModelFingerprint, apiResult APIResult) (float64, []string, map[string]interface{}) {
	matches := []string{}
	extractedData := make(map[string]interface{})

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(apiResult.Body), &data); err != nil {
		return 0, nil, nil
	}

	var score float64

	// 检查测试指纹
	for _, test := range fp.Fingerprints {
		testScore := s.evaluateTestFingerprint(test, apiResult.Body)
		if testScore > 0 {
			score += testScore * test.Weight
			matches = append(matches, fmt.Sprintf("test:%s", test.Name))
		}
	}

	// 检查变体
	for _, variant := range fp.Variants {
		for _, feature := range variant.Features {
			if s.regexCache.Match(variant.Pattern, apiResult.Body) {
				extractedData[feature] = true
			}
		}
	}

	// 归一化
	if len(fp.Fingerprints) > 0 {
		score = score / float64(len(fp.Fingerprints))
		if score > 1 {
			score = 1
		}
	}

	return score, matches, extractedData
}

// evaluateTestFingerprint 评估测试指纹
func (s *Scorer) evaluateTestFingerprint(test fingerprints.TestFingerprint, body string) float64 {
	var score float64
	bodyLower := strings.ToLower(body)

	// 检查期望关键词
	for _, keyword := range test.ExpectedKeywords {
		if strings.Contains(bodyLower, strings.ToLower(keyword)) {
			score += 0.2
		}
	}

	// 检查期望模式
	for _, pattern := range test.ExpectedPatterns {
		if s.regexCache.Match(pattern, body) {
			score += 0.25
		}
	}

	// 检查禁用关键词
	for _, keyword := range test.ForbiddenKeywords {
		if strings.Contains(bodyLower, strings.ToLower(keyword)) {
			score -= 0.3
		}
	}

	// 检查禁用模式
	for _, pattern := range test.ForbiddenPatterns {
		if s.regexCache.Match(pattern, body) {
			score -= 0.3
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// extractVersion 提取版本信息
func (s *Scorer) extractVersion(fp *fingerprints.FrameworkFingerprint, apiResult APIResult) string {
	// 从响应头中提取
	for _, header := range fp.Headers {
		if value, exists := apiResult.Headers[header.Name]; exists {
			if header.Pattern != "" {
				re := s.regexCache.Get(header.Pattern)
				if re != nil {
					matches := re.FindStringSubmatch(value)
					if len(matches) > 1 {
						return matches[1] // 返回捕获组
					}
					if len(matches) > 0 {
						return matches[0]
					}
				}
			}
		}
	}

	// 从响应体中提取
	for _, version := range fp.Versions {
		re := s.regexCache.Get(version.Pattern)
		if re != nil {
			matches := re.FindStringSubmatch(apiResult.Body)
			if len(matches) > 1 {
				return matches[1]
			}
			if len(matches) > 0 {
				return matches[0]
			}
		}
	}

	return ""
}

// compilePattern 编译正则表达式
func compilePattern(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
