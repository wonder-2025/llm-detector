package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"llm-detector/pkg/fingerprints"
	"llm-detector/pkg/plugins"
)

// Engine 探测引擎
type Engine struct {
	registry   *plugins.Registry
	loader     *fingerprints.Loader
	scorer     *Scorer
	timeout    time.Duration
	workers    int
	mode       ScoringMode
}

// NewEngine 创建探测引擎
func NewEngine(registry *plugins.Registry, loader *fingerprints.Loader, timeout time.Duration) *Engine {
	weights := DefaultWeights()
	scorer, _ := NewScorer(weights, ModeStrict)

	return &Engine{
		registry: registry,
		loader:   loader,
		scorer:   scorer,
		timeout:  timeout,
		workers:  5,
		mode:     ModeStrict,
	}
}

// NewEngineWithWorkers 创建指定并发数的探测引擎
func NewEngineWithWorkers(registry *plugins.Registry, loader *fingerprints.Loader, timeout time.Duration, workers int) *Engine {
	weights := DefaultWeights()
	scorer, _ := NewScorer(weights, ModeStrict)

	return &Engine{
		registry: registry,
		loader:   loader,
		scorer:   scorer,
		timeout:  timeout,
		workers:  workers,
		mode:     ModeStrict,
	}
}

// SetWorkers 设置并发数
func (e *Engine) SetWorkers(workers int) {
	if workers > 0 {
		e.workers = workers
	}
}

// GetWorkers 获取当前并发数
func (e *Engine) GetWorkers() int {
	return e.workers
}

// SetMode 设置评分模式
func (e *Engine) SetMode(mode ScoringMode) {
	e.mode = mode
	weights := e.scorer.weights
	scorer, _ := NewScorer(weights, mode)
	e.scorer = scorer
}

// GetMode 获取当前评分模式
func (e *Engine) GetMode() ScoringMode {
	return e.mode
}

// SetThreshold 设置置信度阈值
func (e *Engine) SetThreshold(threshold float64) {
	e.scorer.SetThreshold(threshold)
}

// GetThreshold 获取当前阈值
func (e *Engine) GetThreshold() float64 {
	return e.scorer.GetThreshold()
}

// Detect 执行探测
func (e *Engine) Detect(ctx context.Context, target *Target) (*DetectionResult, error) {
	startTime := time.Now()

	result := &DetectionResult{
		Target:    target.String(),
		Timestamp: startTime,
		Mode:      e.mode.String(),
		Threshold: e.scorer.GetThreshold(),
	}

	// 1. API探测
	apiResults := e.detectAPIs(ctx, target)
	result.APIResults = apiResults

	// 2. 主动探测（如果API探测失败）
	if len(apiResults) == 0 || !hasAvailableAPI(apiResults) {
		prober := NewActiveProber(e.timeout)
		probeResults := prober.ProbeTarget(ctx, target)
		
		// 将主动探测结果转换为API结果
		for _, pr := range probeResults {
			if pr.Available {
				apiResults = append(apiResults, APIResult{
					Type:       "active_probe",
					Endpoint:   pr.Endpoint,
					Available:  true,
					StatusCode: pr.StatusCode,
					Headers:    pr.Headers,
					Body:       pr.Body,
				})
			}
		}
		result.APIResults = apiResults
	}

	// 3. 智能探测（如果主动探测也失败）- 针对自定义端点
	if len(apiResults) == 0 || !hasAvailableAPI(apiResults) {
		smartProber := NewSmartProber(e.timeout)
		_, smartResults := smartProber.SmartDiscovery(ctx, target)
		
		// 合并智能探测结果
		apiResults = append(apiResults, smartResults...)
		result.APIResults = apiResults
	}

	// 4. 模型指纹识别（使用新的评分系统）
	if len(apiResults) > 0 {
		modelGuess := e.fingerprintModel(ctx, target, apiResults)
		result.ModelGuess = modelGuess
	}

	// 5. 服务指纹识别（使用新的评分系统）
	serviceInfo := e.fingerprintService(ctx, target, apiResults)
	result.ServiceInfo = serviceInfo

	result.Duration = time.Since(startTime)

	return result, nil
}

func hasAvailableAPI(results []APIResult) bool {
	for _, r := range results {
		if r.Available {
			return true
		}
	}
	return false
}

// detectAPIs 探测API端点
func (e *Engine) detectAPIs(ctx context.Context, target *Target) []APIResult {
	var results []APIResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	apiPlugins := e.registry.AllAPIs()
	semaphore := make(chan struct{}, e.workers)

	for _, plugin := range apiPlugins {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(p plugins.Plugin) {
			defer wg.Done()
			defer func() { <-semaphore }()

			ctx, cancel := context.WithTimeout(ctx, e.timeout)
			defer cancel()

			apiResult, err := p.Detect(ctx, target)
			if err != nil {
				apiResult = &plugins.APIResult{
					Type:      p.Name(),
					Available: false,
					Error:     err.Error(),
				}
			}

			mu.Lock()
			results = append(results, APIResult{
				Type:       apiResult.Type,
				Endpoint:   apiResult.Endpoint,
				Available:  apiResult.Available,
				StatusCode: apiResult.StatusCode,
				Headers:    apiResult.Headers,
				Body:       apiResult.Body,
				Error:      apiResult.Error,
			})
			mu.Unlock()
		}(plugin)
	}

	wg.Wait()
	return results
}

// fingerprintModel 模型指纹识别（使用评分引擎）
func (e *Engine) fingerprintModel(ctx context.Context, target *Target, apiResults []APIResult) *ModelGuess {
	// 获取所有模型指纹
	modelFPs := e.loader.AllModels()

	var bestMatch *ModelGuess
	bestScore := 0.0
	var allMatches []AlternativeModel

	for _, fp := range modelFPs {
		scoringResult := e.scorer.ScoreModel(ctx, fp, apiResults)
		
		if scoringResult.Score > bestScore {
			bestScore = scoringResult.Score
			bestMatch = &ModelGuess{
				Name:        fp.Name,
				Provider:    fp.Provider,
				Type:        fp.Type,
				Confidence:  scoringResult.Score,
				Features:    scoringResult.MatchedRules,
				Version:     scoringResult.Version,
				ScoringDetails: &scoringResult.Details,
			}
		}

		// 收集备选模型（置信度>0.3）
		if scoringResult.Score > 0.3 && scoringResult.Score < bestScore {
			allMatches = append(allMatches, AlternativeModel{
				Name:       fp.Name,
				Confidence: scoringResult.Score,
			})
		}
	}

	// 添加备选模型
	if bestMatch != nil && len(allMatches) > 0 {
		// 只保留前3个备选
		if len(allMatches) > 3 {
			allMatches = allMatches[:3]
		}
		bestMatch.Alternative = allMatches
	}

	if bestMatch != nil && bestMatch.Confidence >= e.scorer.GetThreshold() {
		return bestMatch
	}

	return nil
}

// fingerprintService 服务指纹识别（使用评分引擎）
func (e *Engine) fingerprintService(ctx context.Context, target *Target, apiResults []APIResult) *ServiceInfo {
	info := &ServiceInfo{
		Headers: make(map[string]string),
	}

	// 获取所有框架指纹
	frameworkFPs := e.loader.AllFrameworks()

	var bestMatch *fingerprints.FrameworkFingerprint
	bestScore := 0.0
	var bestScoringResult *ScoringResult

	for _, fp := range frameworkFPs {
		scoringResult := e.scorer.ScoreFramework(fp, apiResults)
		
		if scoringResult.Score > bestScore {
			bestScore = scoringResult.Score
			bestMatch = fp
			bestScoringResult = scoringResult
		}
	}

	if bestMatch != nil && bestScore >= e.scorer.GetThreshold() {
		info.Framework = bestMatch.Name
		info.Confidence = bestScore
		info.ScoringDetails = &bestScoringResult.Details
		
		if bestMatch.Deployment.DefaultPort != "" {
			info.Deployment = fmt.Sprintf("Port %s", bestMatch.Deployment.DefaultPort)
		}
		
		// 设置版本信息
		if bestScoringResult.Version != "" {
			info.Version = bestScoringResult.Version
		}
	}

	// 收集响应头信息
	for _, apiResult := range apiResults {
		for key, value := range apiResult.Headers {
			info.Headers[key] = value
		}
	}

	return info
}

// String 返回评分模式的字符串表示
func (m ScoringMode) String() string {
	switch m {
	case ModeStrict:
		return "strict"
	case ModeLoose:
		return "loose"
	default:
		return "unknown"
	}
}
