# LLM Detector 输出格式增强

## 概述

LLM Detector v2.1.0 引入了多种输出格式支持，包括增强JSON、CSV和交互式HTML报告。

## 新增功能

### 1. 增强JSON格式 (`--json`)

相比原始JSON，增强版本包含：

- **版本信息**: 检测器版本、Go版本、平台信息
- **检测配置**: 评分模式、置信度阈值、超时设置
- **组件分类**: API、模型、框架、部署方式分类
- **置信度级别**: high/medium/low 三级分类
- **统计信息**: API总数、成功数、失败数、平均置信度

```bash
llm-detector -t 192.168.1.100 --json
llm-detector -t 192.168.1.100 --json -o results.json
```

### 2. CSV格式导出 (`--csv`)

Excel兼容的CSV格式，包含18个字段：

| 字段 | 说明 |
|------|------|
| 目标地址 | 检测目标 |
| 检测时间 | 格式化时间戳 |
| 检测耗时 | 毫秒单位 |
| 评分模式 | strict/loose |
| 置信度阈值 | 百分比显示 |
| 检测到的API | 分号分隔列表 |
| API数量 | 可用API计数 |
| 识别模型 | 检测到的模型名称 |
| 模型提供商 | 模型提供商 |
| 模型类型 | 模型类型 |
| 模型置信度 | 0-1数值 |
| 模型版本 | 检测到的版本 |
| 识别框架 | 服务框架 |
| 框架置信度 | 0-1数值 |
| 框架版本 | 框架版本 |
| 部署方式 | 部署信息 |
| 是否检测到 | true/false |
| 最高置信度 | 最高置信度值 |

```bash
llm-detector -t 192.168.1.100 --csv -o results.csv
```

### 3. HTML报告 (`--html`)

交互式HTML报告，包含Chart.js图表：

#### 特性
- 📊 **组件分布饼图**: 可视化展示检测到的组件类型
- 📈 **置信度分布柱状图**: 展示置信度区间分布
- 📋 **检测结果表格**: 完整检测数据展示
- 🔍 **交互式筛选**: 按状态、置信度筛选
- 🔎 **实时搜索**: 目标地址搜索
- 📥 **一键导出**: 导出CSV和JSON
- 🌓 **主题切换**: 深色/浅色模式
- 📱 **响应式设计**: 支持移动端

#### 使用方式
```bash
# 浅色主题（默认）
llm-detector -t 192.168.1.100 --html -o report.html

# 深色主题
llm-detector -t 192.168.1.100 --html --dark -o report.html

# 批量检测
llm-detector -t 10.0.0.0/24 --html -o batch_report.html
```

## CLI参数

### 新增参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--json` | 输出增强JSON | `--json` |
| `--csv` | 输出CSV格式 | `--csv` |
| `--html` | 输出HTML报告 | `--html` |
| `-o, --output` | 指定输出文件 | `-o report.html` |
| `--dark` | HTML深色主题 | `--dark` |

### 完整示例

```bash
# 基础检测
llm-detector -t 192.168.1.100

# 增强JSON输出
llm-detector -t 192.168.1.100 --json -o results.json

# CSV导出
llm-detector -t 192.168.1.100 --csv -o results.csv

# HTML报告（浅色）
llm-detector -t 192.168.1.100 --html -o report.html

# HTML报告（深色）
llm-detector -t 192.168.1.100 --html --dark -o report.html

# 批量检测
llm-detector -t 192.168.1.0/24 --html -o batch_report.html
```

## 技术实现

### 文件结构
```
pkg/output/
├── csv.go            # CSV导出模块
├── html.go           # HTML报告生成器
└── enhanced_json.go  # 增强JSON导出
```

### Chart.js集成
- 使用CDN加载Chart.js（在线模式）
- 支持离线使用（已缓存）
- 饼图：组件类型分布
- 柱状图：置信度区间分布

### 主题系统
- CSS变量控制主题色
- localStorage保存用户偏好
- 支持深色/浅色切换

## 性能

- 二进制大小: ~11MB（符合16MB限制）
- HTML报告大小: ~24KB（含图表）
- CSV导出: 内存友好，流式写入
- JSON生成: 支持美化/压缩

## 兼容性

- Excel: CSV文件已添加UTF-8 BOM，支持中文
- 浏览器: 支持现代浏览器（Chrome/Firefox/Safari/Edge）
- 移动端: 响应式设计，支持手机/平板
