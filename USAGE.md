# LLM Detector 使用指南

## 输出格式

### 1. 默认文本输出
```bash
llm-detector -t 192.168.1.100
```

### 2. 增强JSON输出
```bash
# 输出到控制台
llm-detector -t 192.168.1.100 --json

# 保存到文件
llm-detector -t 192.168.1.100 --json -o results.json
```

JSON包含以下增强字段：
- `detector_info`: 检测器版本信息
- `config`: 检测配置
- `detection.components`: 组件分类列表
- `statistics`: 统计信息
- `confidence_level`: 置信度级别 (high/medium/low)

### 3. CSV格式导出
```bash
# 单个目标
llm-detector -t 192.168.1.100 --csv -o results.csv

# 批量目标
llm-detector -t 192.168.1.0/24 --csv -o batch_results.csv
```

CSV字段包括：
- 目标地址
- 检测时间
- 检测耗时(ms)
- 评分模式
- 置信度阈值
- 检测到的API
- 识别模型/提供商/类型
- 模型置信度/版本
- 识别框架/版本
- 部署方式

### 4. HTML报告
```bash
# 浅色主题（默认）
llm-detector -t 192.168.1.100 --html -o report.html

# 深色主题
llm-detector -t 192.168.1.100 --html --dark -o report.html

# 批量检测
llm-detector -t 192.168.1.0/24 --html -o batch_report.html
```

HTML报告特性：
- 📊 组件分布饼图（Chart.js）
- 📈 置信度分布柱状图
- 📋 交互式检测结果表格
- 🔍 实时搜索和筛选
- 📥 一键导出CSV/JSON
- 🌓 深色/浅色主题切换
- 📱 响应式设计（支持移动端）

## CLI参数

| 参数 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `--target` | `-t` | 目标地址 | `-t 192.168.1.100` |
| `--json` | | 输出JSON格式 | `--json` |
| `--csv` | | 输出CSV格式 | `--csv` |
| `--html` | | 输出HTML报告 | `--html` |
| `--output` | `-o` | 指定输出文件 | `-o report.html` |
| `--dark` | | 深色主题（HTML） | `--dark` |
| `--strict` | | 严格模式 | `--strict` |
| `--loose` | | 宽松模式 | `--loose` |
| `--threshold` | | 自定义阈值 | `--threshold 0.8` |
| `--workers` | `-w` | 并发数 | `-w 10` |
| `--verbose` | `-v` | 详细输出 | `-v` |
| `--version` | | 显示版本 | `--version` |

## 批量检测示例

```bash
# 扫描整个子网并生成HTML报告
llm-detector -t 10.0.0.0/24 --html -o subnet_report.html

# 扫描多个IP并导出CSV
llm-detector -t 192.168.1.1,192.168.1.2,192.168.1.3 --csv -o results.csv

# 从文件读取目标列表
cat targets.txt | xargs -I {} llm-detector -t {} --json -o results.json
```

## 输出示例

### JSON输出示例
```json
{
  "report_info": {
    "generated_at": "2026-03-04T18:46:56.932241733+08:00",
    "total_count": 1,
    "version": "1.1.0"
  },
  "results": [{
    "target": "192.168.1.100:8080",
    "timestamp": "2026-03-04T18:46:56.659848275+08:00",
    "duration": 272365235,
    "detector_info": {
      "version": "1.1.0",
      "version_name": "Enhanced Output Edition",
      "go_version": "go1.22.2",
      "platform": "linux/amd64"
    },
    "detection": {
      "detected": true,
      "components": [
        {
          "type": "ollama",
          "name": "ollama",
          "category": "api",
          "confidence": 0.95
        }
      ],
      "model_guess": {
        "name": "Llama 2",
        "confidence": 0.85,
        "confidence_level": "high"
      }
    }
  }]
}
```

### CSV输出示例
```csv
目标地址,检测时间,检测耗时(ms),评分模式,置信度阈值,检测到的API,API数量,识别模型,模型提供商,模型类型,模型置信度,模型版本,识别框架,框架置信度,框架版本,部署方式,是否检测到,最高置信度
192.168.1.100:8080,2026-03-04 18:46:59,315,strict,70%,ollama,1,Llama 2,Meta,llama,0.85,2.1,FastAPI,0.90,0.95,Port 8080,true,0.90
```
