package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OutputFormat 输出格式
type OutputFormat int

const (
	FormatJSON OutputFormat = iota
	FormatJSONL
	FormatCSV
	FormatHTML
)

// OutputWriter 输出写入器接口
type OutputWriter interface {
	Write(result *DetectionResult) error
	Close() error
}

// JSONWriter JSON格式写入器
type JSONWriter struct {
	file   *os.File
	first  bool
	indent bool
}

// NewJSONWriter 创建JSON写入器
func NewJSONWriter(path string, indent bool) (*JSONWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON file: %w", err)
	}

	writer := &JSONWriter{
		file:   file,
		first:  true,
		indent: indent,
	}

	// 写入JSON数组开始
	if _, err := file.WriteString("[\n"); err != nil {
		file.Close()
		return nil, err
	}

	return writer, nil
}

// Write 写入结果
func (w *JSONWriter) Write(result *DetectionResult) error {
	if !w.first {
		if _, err := w.file.WriteString(",\n"); err != nil {
			return err
		}
	}
	w.first = false

	var data []byte
	var err error
	if w.indent {
		data, err = json.MarshalIndent(result, "  ", "  ")
	} else {
		data, err = json.Marshal(result)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = w.file.Write(data)
	return err
}

// Close 关闭写入器
func (w *JSONWriter) Close() error {
	if _, err := w.file.WriteString("\n]"); err != nil {
		return err
	}
	return w.file.Close()
}

// JSONLWriter JSON Lines格式写入器
type JSONLWriter struct {
	file *os.File
}

// NewJSONLWriter 创建JSONL写入器
func NewJSONLWriter(path string) (*JSONLWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSONL file: %w", err)
	}

	return &JSONLWriter{file: file}, nil
}

// Write 写入结果
func (w *JSONLWriter) Write(result *DetectionResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = w.file.Write(data)
	if err != nil {
		return err
	}

	_, err = w.file.WriteString("\n")
	return err
}

// Close 关闭写入器
func (w *JSONLWriter) Close() error {
	return w.file.Close()
}

// CSVWriter CSV格式写入器
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
	header bool
}

// NewCSVWriter 创建CSV写入器
func NewCSVWriter(path string) (*CSVWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// 写入表头
	headers := []string{
		"target",
		"timestamp",
		"duration_ms",
		"detected",
		"framework",
		"framework_confidence",
		"model",
		"model_confidence",
		"provider",
		"model_type",
		"version",
		"components",
		"error",
	}

	if err := writer.Write(headers); err != nil {
		file.Close()
		return nil, err
	}
	writer.Flush()

	return &CSVWriter{
		file:   file,
		writer: writer,
		header: true,
	}, nil
}

// Write 写入结果
func (w *CSVWriter) Write(result *DetectionResult) error {
	// 收集检测到的组件
	var components []string
	for _, api := range result.APIResults {
		if api.Available {
			components = append(components, api.Type)
		}
	}

	// 构建记录
	record := []string{
		result.Target,
		result.Timestamp.Format(time.RFC3339),
		fmt.Sprintf("%d", result.Duration.Milliseconds()),
		fmt.Sprintf("%t", result.IsDetected()),
	}

	// 框架信息
	if result.ServiceInfo != nil && result.ServiceInfo.Framework != "" {
		record = append(record,
			result.ServiceInfo.Framework,
			fmt.Sprintf("%.2f", result.ServiceInfo.Confidence),
		)
	} else {
		record = append(record, "", "")
	}

	// 模型信息
	if result.ModelGuess != nil {
		record = append(record,
			result.ModelGuess.Name,
			fmt.Sprintf("%.2f", result.ModelGuess.Confidence),
			result.ModelGuess.Provider,
			result.ModelGuess.Type,
			result.ModelGuess.Version,
		)
	} else {
		record = append(record, "", "", "", "", "")
	}

	// 组件列表
	record = append(record, strings.Join(components, ";"))

	// 错误信息
	var errors []string
	for _, api := range result.APIResults {
		if api.Error != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", api.Type, api.Error))
		}
	}
	record = append(record, strings.Join(errors, "; "))

	if err := w.writer.Write(record); err != nil {
		return err
	}

	w.writer.Flush()
	return w.writer.Error()
}

// Close 关闭写入器
func (w *CSVWriter) Close() error {
	w.writer.Flush()
	return w.file.Close()
}

// HTMLWriter HTML报告写入器
type HTMLWriter struct {
	file    *os.File
	results []*DetectionResult
	stats   *ScanStatistics
}

// NewHTMLWriter 创建HTML写入器
func NewHTMLWriter(path string) (*HTMLWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTML file: %w", err)
	}

	return &HTMLWriter{
		file:    file,
		results: make([]*DetectionResult, 0),
	}, nil
}

// Write 写入结果
func (w *HTMLWriter) Write(result *DetectionResult) error {
	w.results = append(w.results, result)
	return nil
}

// SetStatistics 设置统计信息
func (w *HTMLWriter) SetStatistics(stats *ScanStatistics) {
	w.stats = stats
}

// Close 关闭并生成HTML报告
func (w *HTMLWriter) Close() error {
	if w.file == nil {
		return nil
	}
	defer w.file.Close()

	// 确保stats不为nil
	if w.stats == nil {
		w.stats = &ScanStatistics{
			ComponentDist: make(map[string]int),
			PortDist:      make(map[int]int),
		}
	}

	tmpl := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LLM Detector 扫描报告</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            background: white;
            border-radius: 16px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
        }
        .header h1 {
            color: #333;
            font-size: 2em;
            margin-bottom: 10px;
        }
        .header .subtitle {
            color: #666;
            font-size: 1.1em;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
            transition: transform 0.2s;
        }
        .stat-card:hover {
            transform: translateY(-4px);
        }
        .stat-card .label {
            color: #888;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .stat-card .value {
            font-size: 2em;
            font-weight: bold;
            color: #333;
            margin-top: 8px;
        }
        .stat-card.success .value { color: #10b981; }
        .stat-card.error .value { color: #ef4444; }
        .stat-card.warning .value { color: #f59e0b; }
        .stat-card.info .value { color: #3b82f6; }
        .section {
            background: white;
            border-radius: 16px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.1);
        }
        .section h2 {
            color: #333;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f0f0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }
        th {
            background: #f8fafc;
            font-weight: 600;
            color: #555;
            text-transform: uppercase;
            font-size: 0.85em;
            letter-spacing: 0.5px;
        }
        tr:hover {
            background: #f8fafc;
        }
        .badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 500;
        }
        .badge-success {
            background: #d1fae5;
            color: #065f46;
        }
        .badge-error {
            background: #fee2e2;
            color: #991b1b;
        }
        .badge-warning {
            background: #fef3c7;
            color: #92400e;
        }
        .badge-info {
            background: #dbeafe;
            color: #1e40af;
        }
        .confidence-bar {
            width: 100%;
            height: 8px;
            background: #e5e7eb;
            border-radius: 4px;
            overflow: hidden;
        }
        .confidence-fill {
            height: 100%;
            border-radius: 4px;
            transition: width 0.3s;
        }
        .confidence-high { background: linear-gradient(90deg, #10b981, #34d399); }
        .confidence-medium { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
        .confidence-low { background: linear-gradient(90deg, #ef4444, #f87171); }
        .components-list {
            display: flex;
            flex-wrap: wrap;
            gap: 6px;
        }
        .component-tag {
            background: #f3f4f6;
            padding: 4px 10px;
            border-radius: 6px;
            font-size: 0.85em;
            color: #4b5563;
        }
        .footer {
            text-align: center;
            color: rgba(255,255,255,0.8);
            margin-top: 40px;
            padding: 20px;
        }
        @media (max-width: 768px) {
            .stats-grid {
                grid-template-columns: repeat(2, 1fr);
            }
            table {
                font-size: 0.9em;
            }
            th, td {
                padding: 8px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 LLM Detector 扫描报告</h1>
            <p class="subtitle">生成时间: {{.GeneratedAt}}</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card info">
                <div class="label">总目标数</div>
                <div class="value">{{.Stats.TotalTargets}}</div>
            </div>
            <div class="stat-card success">
                <div class="label">成功检测</div>
                <div class="value">{{.Stats.SuccessCount}}</div>
            </div>
            <div class="stat-card error">
                <div class="label">检测失败</div>
                <div class="value">{{.Stats.FailCount}}</div>
            </div>
            <div class="stat-card warning">
                <div class="label">跳过</div>
                <div class="value">{{.Stats.SkippedCount}}</div>
            </div>
        </div>

        <div class="section">
            <h2>📊 扫描统计</h2>
            <table>
                <tr>
                    <th>指标</th>
                    <th>数值</th>
                </tr>
                <tr>
                    <td>扫描耗时</td>
                    <td>{{.Duration}}</td>
                </tr>
                <tr>
                    <td>检测状态</td>
                    <td>{{if gt .Stats.SuccessCount 0}}✅ 已检测到 {{.Stats.SuccessCount}} 个组件{{else}}❌ 未检测到组件{{end}}</td>
                </tr>
            </table>
        </div>

        {{if .ComponentDist}}
        <div class="section">
            <h2>🔧 组件分布</h2>
            <table>
                <tr>
                    <th>组件</th>
                    <th>数量</th>
                    <th>占比</th>
                </tr>
                {{range $component, $count := .ComponentDist}}
                <tr>
                    <td>{{$component}}</td>
                    <td>{{$count}}</td>
                    <td>{{printf "%.1f" (div $count $.Stats.SuccessCount 100)}}%</td>
                </tr>
                {{end}}
            </table>
        </div>
        {{end}}

        {{if .PortDist}}
        <div class="section">
            <h2>🌐 端口分布</h2>
            <table>
                <tr>
                    <th>端口</th>
                    <th>数量</th>
                    <th>占比</th>
                </tr>
                {{range $port, $count := .PortDist}}
                <tr>
                    <td>{{$port}}</td>
                    <td>{{$count}}</td>
                    <td>{{printf "%.1f" (div $count $.Stats.SuccessCount 100)}}%</td>
                </tr>
                {{end}}
            </table>
        </div>
        {{end}}

        <div class="section">
            <h2>📋 详细结果</h2>
            <table>
                <tr>
                    <th>目标</th>
                    <th>状态</th>
                    <th>框架</th>
                    <th>置信度</th>
                    <th>模型</th>
                    <th>组件</th>
                </tr>
                {{range .Results}}
                <tr>
                    <td>{{.Target}}</td>
                    <td>
                        {{if .IsDetected}}
                        <span class="badge badge-success">已检测</span>
                        {{else}}
                        <span class="badge badge-error">未检测</span>
                        {{end}}
                    </td>
                    <td>
                        {{if .ServiceInfo}}
                        {{.ServiceInfo.Framework}}
                        {{else}}-{{end}}
                    </td>
                    <td>
                        {{if .ServiceInfo}}
                        <div class="confidence-bar">
                            <div class="confidence-fill {{if ge .ServiceInfo.Confidence 0.7}}confidence-high{{else if ge .ServiceInfo.Confidence 0.4}}confidence-medium{{else}}confidence-low{{end}}" 
                                 style="width: {{printf "%.0f" (mul .ServiceInfo.Confidence 100)}}%"></div>
                        </div>
                        <small>{{printf "%.0f%%" (mul .ServiceInfo.Confidence 100)}}</small>
                        {{else}}-{{end}}
                    </td>
                    <td>
                        {{if .ModelGuess}}
                        {{.ModelGuess.Name}}
                        {{else}}-{{end}}
                    </td>
                    <td>
                        <div class="components-list">
                            {{range .APIResults}}
                            {{if .Available}}
                            <span class="component-tag">{{.Type}}</span>
                            {{end}}
                            {{end}}
                        </div>
                    </td>
                </tr>
                {{end}}
            </table>
        </div>

        <div class="footer">
            <p>Generated by LLM Detector v2.1.0</p>
        </div>
    </div>
</body>
</html>
`

	// 准备模板数据
	data := struct {
		GeneratedAt    string
		Stats          *ScanStatistics
		Results        []*DetectionResult
		ComponentDist  map[string]int
		PortDist       map[int]int
		Duration       string
	}{
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Stats:         w.stats,
		Results:       w.results,
		ComponentDist: w.stats.ComponentDist,
		PortDist:      w.stats.PortDist,
		Duration:      w.stats.Duration().Round(time.Second).String(),
	}

	// 添加辅助函数
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b, c int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b) * float64(c)
		},
	}

	t := template.Must(template.New("report").Funcs(funcMap).Parse(tmpl))
	return t.Execute(w.file, data)
}

// MultiOutputWriter 多输出写入器
type MultiOutputWriter struct {
	writers []OutputWriter
}

// NewMultiOutputWriter 创建多输出写入器
func NewMultiOutputWriter(writers ...OutputWriter) *MultiOutputWriter {
	return &MultiOutputWriter{writers: writers}
}

// Write 写入所有输出
func (m *MultiOutputWriter) Write(result *DetectionResult) error {
	for _, w := range m.writers {
		if err := w.Write(result); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭所有写入器
func (m *MultiOutputWriter) Close() error {
	if m == nil {
		return nil
	}
	var lastErr error
	for _, w := range m.writers {
		if w == nil {
			continue
		}
		if err := w.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// CreateOutputWriter 根据格式创建输出写入器
func CreateOutputWriter(format OutputFormat, path string) (OutputWriter, error) {
	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	switch format {
	case FormatJSON:
		return NewJSONWriter(path, true)
	case FormatJSONL:
		return NewJSONLWriter(path)
	case FormatCSV:
		return NewCSVWriter(path)
	case FormatHTML:
		return NewHTMLWriter(path)
	default:
		return nil, fmt.Errorf("unsupported output format: %v", format)
	}
}

// DetectOutputFormat 根据文件扩展名检测输出格式
func DetectOutputFormat(path string) OutputFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FormatJSON
	case ".jsonl", ".ndjson":
		return FormatJSONL
	case ".csv":
		return FormatCSV
	case ".html", ".htm":
		return FormatHTML
	default:
		return FormatJSON
	}
}
