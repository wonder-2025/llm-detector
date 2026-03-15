package output

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"llm-detector/pkg/core"
)

// Version 工具版本信息
const (
	Version     = "1.1.0"
	VersionName = "Enhanced Output Edition"
)

// EnhancedResult 增强版检测结果
type EnhancedResult struct {
	// 基础检测信息
	Target    string        `json:"target"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`

	// 版本信息
	DetectorInfo DetectorInfo `json:"detector_info"`

	// 检测配置
	Config DetectionConfig `json:"config"`

	// 检测结果
	Detection DetectionData `json:"detection"`

	// 统计信息
	Statistics Statistics `json:"statistics"`
}

// DetectorInfo 检测器信息
type DetectorInfo struct {
	Version     string    `json:"version"`
	VersionName string    `json:"version_name"`
	BuildTime   string    `json:"build_time"`
	GoVersion   string    `json:"go_version"`
	Platform    string    `json:"platform"`
}

// DetectionConfig 检测配置
type DetectionConfig struct {
	Mode      string        `json:"mode"`
	Threshold float64       `json:"threshold"`
	Timeout   time.Duration `json:"timeout"`
}

// DetectionData 检测数据
type DetectionData struct {
	Detected       bool            `json:"detected"`
	Components     []ComponentInfo `json:"components"`
	ModelGuess     *ModelInfo      `json:"model_guess,omitempty"`
	ServiceInfo    *ServiceDetails `json:"service_info,omitempty"`
	APIResults     []APIResultInfo `json:"api_results"`
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Category   string  `json:"category"` // api, model, framework, deployment
	Confidence float64 `json:"confidence"`
	Version    string  `json:"version,omitempty"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name           string             `json:"name"`
	Provider       string             `json:"provider"`
	Type           string             `json:"type"`
	Confidence     float64            `json:"confidence"`
	ConfidenceLevel string           `json:"confidence_level"` // high, medium, low
	Version        string             `json:"version,omitempty"`
	Features       []string           `json:"features"`
	Alternatives   []AlternativeModel `json:"alternatives,omitempty"`
}

// AlternativeModel 备选模型
type AlternativeModel struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// ServiceDetails 服务详情
type ServiceDetails struct {
	Framework       string  `json:"framework,omitempty"`
	FrameworkConfidence float64 `json:"framework_confidence,omitempty"`
	Version         string  `json:"version,omitempty"`
	Deployment      string  `json:"deployment,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// APIResultInfo API结果信息
type APIResultInfo struct {
	Type       string            `json:"type"`
	Endpoint   string            `json:"endpoint"`
	Available  bool              `json:"available"`
	StatusCode int               `json:"status_code,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Error      string            `json:"error,omitempty"`
	Confidence float64           `json:"confidence,omitempty"`
}

// Statistics 统计信息
type Statistics struct {
	TotalAPIs      int     `json:"total_apis"`
	AvailableAPIs  int     `json:"available_apis"`
	FailedAPIs     int     `json:"failed_apis"`
	MaxConfidence  float64 `json:"max_confidence"`
	AvgConfidence  float64 `json:"avg_confidence"`
}

// EnhancedJSONExporter 增强JSON导出器
type EnhancedJSONExporter struct {
	includeRaw    bool
	prettyPrint   bool
}

// NewEnhancedJSONExporter 创建增强JSON导出器
func NewEnhancedJSONExporter() *EnhancedJSONExporter {
	return &EnhancedJSONExporter{
		includeRaw:  false,
		prettyPrint: true,
	}
}

// SetIncludeRaw 设置是否包含原始数据
func (e *EnhancedJSONExporter) SetIncludeRaw(include bool) {
	e.includeRaw = include
}

// SetPrettyPrint 设置是否美化输出
func (e *EnhancedJSONExporter) SetPrettyPrint(pretty bool) {
	e.prettyPrint = pretty
}

// Export 导出增强JSON
func (e *EnhancedJSONExporter) Export(result *core.DetectionResult) (string, error) {
	enhanced := e.convertToEnhanced(result)

	var data []byte
	var err error

	if e.prettyPrint {
		data, err = json.MarshalIndent(enhanced, "", "  ")
	} else {
		data, err = json.Marshal(enhanced)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal enhanced JSON: %w", err)
	}

	return string(data), nil
}

// ExportBatch 批量导出增强JSON
func (e *EnhancedJSONExporter) ExportBatch(results []*core.DetectionResult) (string, error) {
	enhancedResults := make([]*EnhancedResult, 0, len(results))

	for _, r := range results {
		enhancedResults = append(enhancedResults, e.convertToEnhanced(r))
	}

	batchData := struct {
		ReportInfo struct {
			GeneratedAt time.Time `json:"generated_at"`
			TotalCount  int       `json:"total_count"`
			Version     string    `json:"version"`
		} `json:"report_info"`
		Results []*EnhancedResult `json:"results"`
		Summary Statistics        `json:"summary"`
	}{
		Results: enhancedResults,
	}

	batchData.ReportInfo.GeneratedAt = time.Now()
	batchData.ReportInfo.TotalCount = len(results)
	batchData.ReportInfo.Version = Version

	// 计算汇总统计
	batchData.Summary = e.calculateBatchStats(results)

	var data []byte
	var err error

	if e.prettyPrint {
		data, err = json.MarshalIndent(batchData, "", "  ")
	} else {
		data, err = json.Marshal(batchData)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal batch JSON: %w", err)
	}

	return string(data), nil
}

// convertToEnhanced 转换为增强格式
func (e *EnhancedJSONExporter) convertToEnhanced(r *core.DetectionResult) *EnhancedResult {
	enhanced := &EnhancedResult{
		Target:    r.Target,
		Timestamp: r.Timestamp,
		Duration:  r.Duration,
		DetectorInfo: DetectorInfo{
			Version:     Version,
			VersionName: VersionName,
			BuildTime:   time.Now().Format("2006-01-02"),
			GoVersion:   runtime.Version(),
			Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Config: DetectionConfig{
			Mode:      r.Mode,
			Threshold: r.Threshold,
		},
		Detection: DetectionData{
			Detected:   r.IsDetected(),
			Components: e.extractComponents(r),
			APIResults: e.convertAPIResults(r.APIResults),
		},
		Statistics: e.calculateStats(r),
	}

	// 转换模型信息
	if r.ModelGuess != nil {
		enhanced.Detection.ModelGuess = &ModelInfo{
			Name:            r.ModelGuess.Name,
			Provider:        r.ModelGuess.Provider,
			Type:            r.ModelGuess.Type,
			Confidence:      r.ModelGuess.Confidence,
			ConfidenceLevel: e.getConfidenceLevel(r.ModelGuess.Confidence),
			Version:         r.ModelGuess.Version,
			Features:        r.ModelGuess.Features,
		}

		// 转换备选模型
		if len(r.ModelGuess.Alternative) > 0 {
			for _, alt := range r.ModelGuess.Alternative {
				enhanced.Detection.ModelGuess.Alternatives = append(
					enhanced.Detection.ModelGuess.Alternatives,
					AlternativeModel{
						Name:       alt.Name,
						Confidence: alt.Confidence,
					},
				)
			}
		}
	}

	// 转换服务信息
	if r.ServiceInfo != nil {
		enhanced.Detection.ServiceInfo = &ServiceDetails{
			Framework:           r.ServiceInfo.Framework,
			FrameworkConfidence: r.ServiceInfo.Confidence,
			Version:             r.ServiceInfo.Version,
			Deployment:          r.ServiceInfo.Deployment,
			Headers:             r.ServiceInfo.Headers,
		}
	}

	return enhanced
}

// extractComponents 提取组件信息
func (e *EnhancedJSONExporter) extractComponents(r *core.DetectionResult) []ComponentInfo {
	var components []ComponentInfo

	// 提取API组件
	for _, api := range r.APIResults {
		if api.Available {
			components = append(components, ComponentInfo{
				Type:       api.Type,
				Name:       api.Type,
				Category:   "api",
				Confidence: api.Confidence,
			})
		}
	}

	// 提取模型组件
	if r.ModelGuess != nil && r.ModelGuess.Name != "" {
		components = append(components, ComponentInfo{
			Type:       r.ModelGuess.Type,
			Name:       r.ModelGuess.Name,
			Category:   "model",
			Confidence: r.ModelGuess.Confidence,
			Version:    r.ModelGuess.Version,
		})
	}

	// 提取框架组件
	if r.ServiceInfo != nil && r.ServiceInfo.Framework != "" {
		components = append(components, ComponentInfo{
			Type:       r.ServiceInfo.Framework,
			Name:       r.ServiceInfo.Framework,
			Category:   "framework",
			Confidence: r.ServiceInfo.Confidence,
			Version:    r.ServiceInfo.Version,
		})
	}

	// 提取部署方式
	if r.ServiceInfo != nil && r.ServiceInfo.Deployment != "" {
		components = append(components, ComponentInfo{
			Type:     r.ServiceInfo.Deployment,
			Name:     r.ServiceInfo.Deployment,
			Category: "deployment",
		})
	}

	return components
}

// convertAPIResults 转换API结果
func (e *EnhancedJSONExporter) convertAPIResults(results []core.APIResult) []APIResultInfo {
	var converted []APIResultInfo
	for _, r := range results {
		converted = append(converted, APIResultInfo{
			Type:       r.Type,
			Endpoint:   r.Endpoint,
			Available:  r.Available,
			StatusCode: r.StatusCode,
			Headers:    r.Headers,
			Error:      r.Error,
			Confidence: r.Confidence,
		})
	}
	return converted
}

// calculateStats 计算统计信息
func (e *EnhancedJSONExporter) calculateStats(r *core.DetectionResult) Statistics {
	stats := Statistics{
		TotalAPIs: len(r.APIResults),
	}

	var totalConfidence float64
	confidenceCount := 0

	for _, api := range r.APIResults {
		if api.Available {
			stats.AvailableAPIs++
			if api.Confidence > 0 {
				totalConfidence += api.Confidence
				confidenceCount++
			}
		} else {
			stats.FailedAPIs++
		}
	}

	// 计算最大置信度
	if r.ModelGuess != nil && r.ModelGuess.Confidence > stats.MaxConfidence {
		stats.MaxConfidence = r.ModelGuess.Confidence
	}
	if r.ServiceInfo != nil && r.ServiceInfo.Confidence > stats.MaxConfidence {
		stats.MaxConfidence = r.ServiceInfo.Confidence
	}

	// 计算平均置信度
	if confidenceCount > 0 {
		stats.AvgConfidence = totalConfidence / float64(confidenceCount)
	}

	return stats
}

// calculateBatchStats 计算批量统计
func (e *EnhancedJSONExporter) calculateBatchStats(results []*core.DetectionResult) Statistics {
	stats := Statistics{
		TotalAPIs: len(results),
	}

	var totalConfidence float64
	confidenceCount := 0

	for _, r := range results {
		if r.IsDetected() {
			stats.AvailableAPIs++
		}

		conf := r.GetConfidence()
		if conf > stats.MaxConfidence {
			stats.MaxConfidence = conf
		}
		if conf > 0 {
			totalConfidence += conf
			confidenceCount++
		}
	}

	stats.FailedAPIs = stats.TotalAPIs - stats.AvailableAPIs
	if confidenceCount > 0 {
		stats.AvgConfidence = totalConfidence / float64(confidenceCount)
	}

	return stats
}

// getConfidenceLevel 获取置信度级别
func (e *EnhancedJSONExporter) getConfidenceLevel(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "high"
	case confidence >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

// WriteToFile 写入文件
func (e *EnhancedJSONExporter) WriteToFile(result *core.DetectionResult, filepath string) error {
	jsonStr, err := e.Export(result)
	if err != nil {
		return err
	}

	return writeStringToFile(filepath, jsonStr)
}

// WriteBatchToFile 批量写入文件
func (e *EnhancedJSONExporter) WriteBatchToFile(results []*core.DetectionResult, filepath string) error {
	jsonStr, err := e.ExportBatch(results)
	if err != nil {
		return err
	}

	return writeStringToFile(filepath, jsonStr)
}

// writeStringToFile 写入字符串到文件
func writeStringToFile(filepath string, content string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}
