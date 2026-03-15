package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"llm-detector/pkg/core"
)

// HTMLExporter HTML报告导出器
type HTMLExporter struct {
	theme       string // "light" 或 "dark"
	chartJSData string // 内嵌的Chart.js代码
}

// NewHTMLExporter 创建HTML导出器
func NewHTMLExporter() *HTMLExporter {
	return &HTMLExporter{
		theme: "light",
	}
}

// SetTheme 设置主题
func (e *HTMLExporter) SetTheme(theme string) {
	if theme == "dark" || theme == "light" {
		e.theme = theme
	}
}

// Export 导出单个结果为HTML报告
func (e *HTMLExporter) Export(result *core.DetectionResult, filepath string) error {
	results := []*core.DetectionResult{result}
	return e.ExportBatch(results, filepath)
}

// ExportBatch 批量导出结果为HTML报告
func (e *HTMLExporter) ExportBatch(results []*core.DetectionResult, filepath string) error {
	html := e.generateHTML(results)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(html); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}

	return nil
}

// generateHTML 生成完整HTML报告
func (e *HTMLExporter) generateHTML(results []*core.DetectionResult) string {
	stats := calculateStats(results)
	chartData := e.generateChartData(results)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LLM Detector 检测报告</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
    <style>
        :root {
            --bg-primary: %s;
            --bg-secondary: %s;
            --bg-card: %s;
            --text-primary: %s;
            --text-secondary: %s;
            --border-color: %s;
            --accent-color: #3b82f6;
            --accent-hover: #2563eb;
            --success-color: #10b981;
            --warning-color: #f59e0b;
            --danger-color: #ef4444;
            --info-color: #06b6d4;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background-color: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            min-height: 100vh;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }
        
        header {
            background: linear-gradient(135deg, var(--accent-color), var(--accent-hover));
            color: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        
        header h1 {
            font-size: 2rem;
            margin-bottom: 10px;
        }
        
        header .meta {
            opacity: 0.9;
            font-size: 0.9rem;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .stat-card {
            background: var(--bg-card);
            border-radius: 12px;
            padding: 20px;
            border: 1px solid var(--border-color);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        
        .stat-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 16px rgba(0,0,0,0.1);
        }
        
        .stat-card .label {
            font-size: 0.85rem;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
        }
        
        .stat-card .value {
            font-size: 2rem;
            font-weight: 700;
            color: var(--accent-color);
        }
        
        .stat-card .sub-value {
            font-size: 0.9rem;
            color: var(--text-secondary);
            margin-top: 4px;
        }
        
        .charts-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .chart-card {
            background: var(--bg-card);
            border-radius: 12px;
            padding: 20px;
            border: 1px solid var(--border-color);
        }
        
        .chart-card h3 {
            margin-bottom: 15px;
            font-size: 1.1rem;
            color: var(--text-primary);
        }
        
        .chart-container {
            position: relative;
            height: 300px;
        }
        
        .results-section {
            background: var(--bg-card);
            border-radius: 12px;
            padding: 20px;
            border: 1px solid var(--border-color);
            margin-bottom: 20px;
        }
        
        .results-section h2 {
            margin-bottom: 20px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .filter-bar {
            display: flex;
            gap: 15px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        
        .filter-bar input, .filter-bar select {
            padding: 10px 15px;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            background: var(--bg-secondary);
            color: var(--text-primary);
            font-size: 0.9rem;
        }
        
        .filter-bar input:focus, .filter-bar select:focus {
            outline: none;
            border-color: var(--accent-color);
        }
        
        .results-table {
            width: 100%%;
            border-collapse: collapse;
            font-size: 0.9rem;
        }
        
        .results-table th,
        .results-table td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }
        
        .results-table th {
            background: var(--bg-secondary);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.8rem;
            letter-spacing: 0.5px;
        }
        
        .results-table tr:hover {
            background: var(--bg-secondary);
        }
        
        .badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 20px;
            font-size: 0.75rem;
            font-weight: 600;
        }
        
        .badge-success {
            background: rgba(16, 185, 129, 0.2);
            color: var(--success-color);
        }
        
        .badge-warning {
            background: rgba(245, 158, 11, 0.2);
            color: var(--warning-color);
        }
        
        .badge-danger {
            background: rgba(239, 68, 68, 0.2);
            color: var(--danger-color);
        }
        
        .badge-info {
            background: rgba(6, 182, 212, 0.2);
            color: var(--info-color);
        }
        
        .confidence-bar {
            width: 100%%;
            height: 8px;
            background: var(--bg-secondary);
            border-radius: 4px;
            overflow: hidden;
        }
        
        .confidence-bar .fill {
            height: 100%%;
            background: linear-gradient(90deg, var(--accent-color), var(--success-color));
            transition: width 0.3s ease;
        }
        
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            background: var(--accent-color);
            color: white;
            font-size: 0.9rem;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.2s;
            text-decoration: none;
        }
        
        .btn:hover {
            background: var(--accent-hover);
        }
        
        .btn-secondary {
            background: var(--bg-secondary);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
        }
        
        .btn-secondary:hover {
            background: var(--border-color);
        }
        
        .actions {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        
        .theme-toggle {
            position: fixed;
            top: 20px;
            right: 20px;
            background: var(--bg-card);
            border: 1px solid var(--border-color);
            border-radius: 50%%;
            width: 45px;
            height: 45px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            font-size: 1.2rem;
            transition: transform 0.2s;
            z-index: 1000;
        }
        
        .theme-toggle:hover {
            transform: scale(1.1);
        }
        
        @media (max-width: 768px) {
            .charts-grid {
                grid-template-columns: 1fr;
            }
            
            .stats-grid {
                grid-template-columns: repeat(2, 1fr);
            }
            
            .results-table {
                font-size: 0.8rem;
            }
            
            .results-table th,
            .results-table td {
                padding: 8px;
            }
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: var(--text-secondary);
        }
        
        .empty-state svg {
            width: 80px;
            height: 80px;
            margin-bottom: 20px;
            opacity: 0.5;
        }
    </style>
</head>
<body>
    <button class="theme-toggle" onclick="toggleTheme()" title="切换主题">🌓</button>
    
    <div class="container">
        <header>
            <h1>🔍 LLM Detector 检测报告</h1>
            <div class="meta">
                生成时间: %s | 检测目标数: %d | 检测成功率: %.1f%%
            </div>
        </header>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="label">总检测数</div>
                <div class="value">%d</div>
                <div class="sub-value">个目标</div>
            </div>
            <div class="stat-card">
                <div class="label">成功检测</div>
                <div class="value" style="color: var(--success-color);">%d</div>
                <div class="sub-value">%.1f%%</div>
            </div>
            <div class="stat-card">
                <div class="label">平均置信度</div>
                <div class="value" style="color: var(--info-color);">%.1f%%</div>
                <div class="sub-value">综合评估</div>
            </div>
            <div class="stat-card">
                <div class="label">总耗时</div>
                <div class="value" style="color: var(--warning-color);">%.1f</div>
                <div class="sub-value">秒</div>
            </div>
        </div>
        
        <div class="charts-grid">
            <div class="chart-card">
                <h3>📊 组件分布</h3>
                <div class="chart-container">
                    <canvas id="componentChart"></canvas>
                </div>
            </div>
            <div class="chart-card">
                <h3>📈 置信度分布</h3>
                <div class="chart-container">
                    <canvas id="confidenceChart"></canvas>
                </div>
            </div>
        </div>
        
        <div class="results-section">
            <h2>📋 检测结果详情</h2>
            
            <div class="actions">
                <button class="btn" onclick="exportCSV()">
                    <span>📥</span> 导出CSV
                </button>
                <button class="btn btn-secondary" onclick="exportJSON()">
                    <span>📄</span> 导出JSON
                </button>
            </div>
            
            <div class="filter-bar">
                <input type="text" id="searchInput" placeholder="🔍 搜索目标地址..." onkeyup="filterTable()">
                <select id="filterDetected" onchange="filterTable()">
                    <option value="">全部状态</option>
                    <option value="detected">已检测</option>
                    <option value="undetected">未检测</option>
                </select>
                <select id="filterConfidence" onchange="filterTable()">
                    <option value="">全部置信度</option>
                    <option value="high">高 (>80%%)</option>
                    <option value="medium">中 (50-80%%)</option>
                    <option value="low">低 (<50%%)</option>
                </select>
            </div>
            
            <table class="results-table" id="resultsTable">
                <thead>
                    <tr>
                        <th>目标地址</th>
                        <th>检测时间</th>
                        <th>耗时</th>
                        <th>识别模型</th>
                        <th>框架</th>
                        <th>置信度</th>
                        <th>状态</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>
        </div>
    </div>
    
    <script>
        // 图表数据
        const chartData = %s;
        
        // 组件分布饼图
        const componentCtx = document.getElementById('componentChart').getContext('2d');
        new Chart(componentCtx, {
            type: 'doughnut',
            data: {
                labels: chartData.componentLabels,
                datasets: [{
                    data: chartData.componentData,
                    backgroundColor: [
                        '#3b82f6', '#10b981', '#f59e0b', '#ef4444', 
                        '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'right',
                        labels: {
                            color: getComputedStyle(document.body).getPropertyValue('--text-primary')
                        }
                    }
                }
            }
        });
        
        // 置信度分布柱状图
        const confidenceCtx = document.getElementById('confidenceChart').getContext('2d');
        new Chart(confidenceCtx, {
            type: 'bar',
            data: {
                labels: chartData.confidenceLabels,
                datasets: [{
                    label: '目标数量',
                    data: chartData.confidenceData,
                    backgroundColor: '#3b82f6',
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            color: getComputedStyle(document.body).getPropertyValue('--text-secondary')
                        },
                        grid: {
                            color: getComputedStyle(document.body).getPropertyValue('--border-color')
                        }
                    },
                    x: {
                        ticks: {
                            color: getComputedStyle(document.body).getPropertyValue('--text-secondary')
                        },
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });
        
        // 主题切换
        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme') || '%s';
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            html.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            location.reload();
        }
        
        // 加载保存的主题
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme) {
            document.documentElement.setAttribute('data-theme', savedTheme);
        }
        
        // 表格筛选
        function filterTable() {
            const searchValue = document.getElementById('searchInput').value.toLowerCase();
            const detectedFilter = document.getElementById('filterDetected').value;
            const confidenceFilter = document.getElementById('filterConfidence').value;
            const table = document.getElementById('resultsTable');
            const rows = table.getElementsByTagName('tr');
            
            for (let i = 1; i < rows.length; i++) {
                const row = rows[i];
                const cells = row.getElementsByTagName('td');
                if (cells.length === 0) continue;
                
                const target = cells[0].textContent.toLowerCase();
                const model = cells[3].textContent;
                const confidenceText = cells[5].textContent;
                const confidence = parseFloat(confidenceText);
                
                let showRow = target.includes(searchValue);
                
                // 检测状态筛选
                if (detectedFilter) {
                    const isDetected = model && model !== '-';
                    if (detectedFilter === 'detected' && !isDetected) showRow = false;
                    if (detectedFilter === 'undetected' && isDetected) showRow = false;
                }
                
                // 置信度筛选
                if (confidenceFilter && !isNaN(confidence)) {
                    if (confidenceFilter === 'high' && confidence <= 80) showRow = false;
                    if (confidenceFilter === 'medium' && (confidence < 50 || confidence > 80)) showRow = false;
                    if (confidenceFilter === 'low' && confidence >= 50) showRow = false;
                }
                
                row.style.display = showRow ? '' : 'none';
            }
        }
        
        // 导出CSV
        function exportCSV() {
            const table = document.getElementById('resultsTable');
            let csv = [];
            const rows = table.querySelectorAll('tr');
            
            for (let row of rows) {
                if (row.style.display === 'none') continue;
                const cols = row.querySelectorAll('td, th');
                const rowData = [];
                for (let col of cols) {
                    rowData.push('"' + col.textContent.replace(/"/g, '""') + '"');
                }
                csv.push(rowData.join(','));
            }
            
            const csvContent = '\uFEFF' + csv.join('\n');
            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = 'llm-detector-results-' + new Date().toISOString().slice(0,10) + '.csv';
            link.click();
        }
        
        // 导出JSON
        function exportJSON() {
            const jsonData = chartData.rawResults;
            const blob = new Blob([JSON.stringify(jsonData, null, 2)], { type: 'application/json' });
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = 'llm-detector-results-' + new Date().toISOString().slice(0,10) + '.json';
            link.click();
        }
    </script>
</body>
</html>`,
		e.getThemeColors()[0], e.getThemeColors()[1], e.getThemeColors()[2], e.getThemeColors()[3], e.getThemeColors()[4], e.getThemeColors()[5],
		time.Now().Format("2006-01-02 15:04:05"),
		stats.Total,
		stats.DetectionRate*100,
		stats.Total,
		stats.Detected,
		stats.DetectionRate*100,
		stats.AvgConfidence*100,
		stats.TotalDuration.Seconds(),
		e.generateTableRows(results),
		chartData,
		e.theme,
	)
}

// getThemeColors 获取主题颜色
func (e *HTMLExporter) getThemeColors() []string {
	if e.theme == "dark" {
		return []string{
			"#0f172a",    // bg-primary
			"#1e293b",    // bg-secondary
			"#334155",    // bg-card
			"#f8fafc",    // text-primary
			"#94a3b8",    // text-secondary
			"#475569",    // border-color
		}
	}
	return []string{
		"#f8fafc",    // bg-primary
		"#f1f5f9",    // bg-secondary
		"#ffffff",    // bg-card
		"#1e293b",    // text-primary
		"#64748b",    // text-secondary
		"#e2e8f0",    // border-color
	}
}

// generateTableRows 生成表格行HTML
func (e *HTMLExporter) generateTableRows(results []*core.DetectionResult) string {
	var rows strings.Builder

	for _, r := range results {
		status := "未检测"
		statusClass := "badge-danger"
		if r.IsDetected() {
			status = "已检测"
			statusClass = "badge-success"
		}

		modelName := "-"
		if r.ModelGuess != nil && r.ModelGuess.Name != "" {
			modelName = r.ModelGuess.Name
		}

		framework := "-"
		if r.ServiceInfo != nil && r.ServiceInfo.Framework != "" {
			framework = r.ServiceInfo.Framework
		}

		confidence := r.GetConfidence()

		rows.WriteString(fmt.Sprintf(`
                    <tr>
                        <td>%s</td>
                        <td>%s</td>
                        <td>%v</td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>
                            <div class="confidence-bar">
                                <div class="fill" style="width: %.0f%%"></div>
                            </div>
                            <span>%.1f%%</span>
                        </td>
                        <td><span class="badge %s">%s</span></td>
                    </tr>`,
			r.Target,
			r.Timestamp.Format("2006-01-02 15:04:05"),
			r.Duration.Round(time.Millisecond),
			modelName,
			framework,
			confidence*100,
			confidence*100,
			statusClass,
			status,
		))
	}

	return rows.String()
}

// generateChartData 生成图表数据JSON
func (e *HTMLExporter) generateChartData(results []*core.DetectionResult) string {
	// 组件分布统计
	componentCount := make(map[string]int)
	confidenceRanges := map[string]int{
		"0-20%%":   0,
		"20-40%%":  0,
		"40-60%%":  0,
		"60-80%%":  0,
		"80-100%%": 0,
	}

	for _, r := range results {
		// 统计组件
		for _, api := range r.APIResults {
			if api.Available {
				componentCount[api.Type]++
			}
		}

		// 统计置信度分布
		conf := r.GetConfidence() * 100
		switch {
		case conf < 20:
			confidenceRanges["0-20%%"]++
		case conf < 40:
			confidenceRanges["20-40%%"]++
		case conf < 60:
			confidenceRanges["40-60%%"]++
		case conf < 80:
			confidenceRanges["60-80%%"]++
		default:
			confidenceRanges["80-100%%"]++
		}
	}

	// 如果没有组件数据，添加默认值
	if len(componentCount) == 0 {
		componentCount["未识别"] = len(results)
	}

	// 准备图表数据
	var compLabels, compData []string
	for comp, count := range componentCount {
		compLabels = append(compLabels, fmt.Sprintf("\"%s\"", comp))
		compData = append(compData, fmt.Sprintf("%d", count))
	}

	confLabels := []string{"\"0-20%%\"", "\"20-40%%\"", "\"40-60%%\"", "\"60-80%%\"", "\"80-100%%\""}
	confData := []string{
		fmt.Sprintf("%d", confidenceRanges["0-20%%"]),
		fmt.Sprintf("%d", confidenceRanges["20-40%%"]),
		fmt.Sprintf("%d", confidenceRanges["40-60%%"]),
		fmt.Sprintf("%d", confidenceRanges["60-80%%"]),
		fmt.Sprintf("%d", confidenceRanges["80-100%%"]),
	}

	// 准备原始结果数据
	rawResults := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		rawResults = append(rawResults, resultToMap(r))
	}
	rawResultsJSON, _ := json.Marshal(rawResults)

	return fmt.Sprintf(`{
            "componentLabels": [%s],
            "componentData": [%s],
            "confidenceLabels": [%s],
            "confidenceData": [%s],
            "rawResults": %s
        }`,
		strings.Join(compLabels, ","),
		strings.Join(compData, ","),
		strings.Join(confLabels, ","),
		strings.Join(confData, ","),
		string(rawResultsJSON),
	)
}

// resultToMap 将结果转换为map
func resultToMap(r *core.DetectionResult) map[string]interface{} {
	return map[string]interface{}{
		"target":     r.Target,
		"timestamp":  r.Timestamp,
		"duration":   r.Duration.Milliseconds(),
		"mode":       r.Mode,
		"threshold":  r.Threshold,
		"detected":   r.IsDetected(),
		"confidence": r.GetConfidence(),
		"apiResults": r.APIResults,
		"modelGuess": r.ModelGuess,
		"serviceInfo": r.ServiceInfo,
	}
}
