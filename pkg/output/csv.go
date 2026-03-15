package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"llm-detector/pkg/core"
)

// CSVExporter CSV导出器
type CSVExporter struct {
	includeHeaders bool
}

// NewCSVExporter 创建CSV导出器
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{
		includeHeaders: true,
	}
}

// SetIncludeHeaders 设置是否包含表头
func (e *CSVExporter) SetIncludeHeaders(include bool) {
	e.includeHeaders = include
}

// Export 导出单个结果到CSV
func (e *CSVExporter) Export(result *core.DetectionResult, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if e.includeHeaders {
		headers := e.getHeaders()
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// 写入数据行
	record := e.resultToRecord(result)
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	return nil
}

// ExportBatch 批量导出结果到CSV
func (e *CSVExporter) ExportBatch(results []*core.DetectionResult, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入UTF-8 BOM以支持Excel中文显示
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	// 写入表头
	if e.includeHeaders {
		headers := e.getHeaders()
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// 写入数据行
	for _, result := range results {
		record := e.resultToRecord(result)
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// ExportToString 导出为CSV字符串
func (e *CSVExporter) ExportToString(results []*core.DetectionResult) (string, error) {
	var output string

	// 表头
	if e.includeHeaders {
		headers := e.getHeaders()
		for i, h := range headers {
			if i > 0 {
				output += ","
			}
			output += e.escapeCSV(h)
		}
		output += "\n"
	}

	// 数据行
	for _, result := range results {
		record := e.resultToRecord(result)
		for i, field := range record {
			if i > 0 {
				output += ","
			}
			output += e.escapeCSV(field)
		}
		output += "\n"
	}

	return output, nil
}

// getHeaders 获取CSV表头
func (e *CSVExporter) getHeaders() []string {
	return []string{
		"目标地址",
		"检测时间",
		"检测耗时(ms)",
		"评分模式",
		"置信度阈值",
		"检测到的API",
		"API数量",
		"识别模型",
		"模型提供商",
		"模型类型",
		"模型置信度",
		"模型版本",
		"识别框架",
		"框架置信度",
		"框架版本",
		"部署方式",
		"是否检测到",
		"最高置信度",
	}
}

// resultToRecord 将检测结果转换为CSV记录
func (e *CSVExporter) resultToRecord(result *core.DetectionResult) []string {
	// 收集API信息
	var apiTypes []string
	apiCount := 0
	for _, api := range result.APIResults {
		if api.Available {
			apiTypes = append(apiTypes, api.Type)
			apiCount++
		}
	}
	apiInfo := joinStrings(apiTypes, ";")

	// 模型信息
	var modelName, modelProvider, modelType, modelVersion string
	var modelConfidence float64
	if result.ModelGuess != nil {
		modelName = result.ModelGuess.Name
		modelProvider = result.ModelGuess.Provider
		modelType = result.ModelGuess.Type
		modelVersion = result.ModelGuess.Version
		modelConfidence = result.ModelGuess.Confidence
	}

	// 框架信息
	var frameworkName, frameworkVersion, deployment string
	var frameworkConfidence float64
	if result.ServiceInfo != nil {
		frameworkName = result.ServiceInfo.Framework
		frameworkVersion = result.ServiceInfo.Version
		frameworkConfidence = result.ServiceInfo.Confidence
		deployment = result.ServiceInfo.Deployment
	}

	return []string{
		result.Target,
		result.Timestamp.Format("2006-01-02 15:04:05"),
		strconv.FormatInt(result.Duration.Milliseconds(), 10),
		result.Mode,
		fmt.Sprintf("%.0f%%", result.Threshold*100),
		apiInfo,
		strconv.Itoa(apiCount),
		modelName,
		modelProvider,
		modelType,
		fmt.Sprintf("%.2f", modelConfidence),
		modelVersion,
		frameworkName,
		fmt.Sprintf("%.2f", frameworkConfidence),
		frameworkVersion,
		deployment,
		fmt.Sprintf("%t", result.IsDetected()),
		fmt.Sprintf("%.2f", result.GetConfidence()),
	}
}

// escapeCSV 转义CSV字段
func (e *CSVExporter) escapeCSV(field string) string {
	if field == "" {
		return ""
	}
	// 如果字段包含逗号、引号或换行符，需要用引号包裹
	if containsAny(field, ",\"") {
		return "\"" + field + "\""
	}
	return field
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// containsAny 检查字符串是否包含任意字符
func containsAny(s string, chars string) bool {
	for _, c := range chars {
		if containsRune(s, c) {
			return true
		}
	}
	return false
}

// containsRune 检查字符串是否包含指定字符
func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

// ExportSummary 导出统计摘要CSV
func (e *CSVExporter) ExportSummary(results []*core.DetectionResult, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create summary CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入UTF-8 BOM
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	// 统计信息
	stats := calculateStats(results)

	// 写入统计摘要
	headers := []string{"统计项", "数值"}
	writer.Write(headers)

	rows := [][]string{
		{"总检测数", strconv.Itoa(stats.Total)},
		{"成功检测", strconv.Itoa(stats.Detected)},
		{"检测失败", strconv.Itoa(stats.NotDetected)},
		{"检测成功率", fmt.Sprintf("%.1f%%", stats.DetectionRate*100)},
		{"平均置信度", fmt.Sprintf("%.2f", stats.AvgConfidence)},
		{"最高置信度", fmt.Sprintf("%.2f", stats.MaxConfidence)},
		{"总耗时(ms)", strconv.FormatInt(stats.TotalDuration.Milliseconds(), 10)},
		{"平均耗时(ms)", strconv.FormatInt(stats.AvgDuration.Milliseconds(), 10)},
		{"检测时间", time.Now().Format("2006-01-02 15:04:05")},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write summary row: %w", err)
		}
	}

	return nil
}

// Stats 统计信息
type Stats struct {
	Total         int
	Detected      int
	NotDetected   int
	DetectionRate float64
	AvgConfidence float64
	MaxConfidence float64
	TotalDuration time.Duration
	AvgDuration   time.Duration
}

// calculateStats 计算统计信息
func calculateStats(results []*core.DetectionResult) Stats {
	stats := Stats{
		Total: len(results),
	}

	var totalConfidence float64
	var totalDuration time.Duration

	for _, r := range results {
		if r.IsDetected() {
			stats.Detected++
		} else {
			stats.NotDetected++
		}

		conf := r.GetConfidence()
		if conf > stats.MaxConfidence {
			stats.MaxConfidence = conf
		}
		totalConfidence += conf
		totalDuration += r.Duration
	}

	if stats.Total > 0 {
		stats.DetectionRate = float64(stats.Detected) / float64(stats.Total)
		stats.AvgConfidence = totalConfidence / float64(stats.Total)
		stats.AvgDuration = totalDuration / time.Duration(stats.Total)
	}
	stats.TotalDuration = totalDuration

	return stats
}
