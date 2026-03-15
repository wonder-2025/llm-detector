# LLM Detector 批量扫描功能指南

## 概述

LLM Detector v2.1.0 引入了强大的批量扫描功能，支持从文件、管道等多种方式输入目标，并提供实时进度显示、多种格式导出等特性。

## 新功能特性

### 1. 文件输入支持 (`-f, --file`)

从文件读取目标列表，支持多种格式：
- IP 地址: `192.168.1.100`
- IP:Port: `192.168.1.100:11434`
- URL: `http://example.com:8080`
- CIDR: `192.168.1.0/24`

```bash
# 从文件读取目标
./llm-detector -f targets.txt -v

# 目标文件示例 (targets.txt)
# 注释行以 # 开头
192.168.1.100
192.168.1.101:11434
http://api.example.com:8080
10.0.0.0/24
```

### 2. 管道输入支持 (`--stdin`)

从标准输入读取目标，可配合其他工具使用：

```bash
# 配合 nmap 使用
cat ips.txt | ./llm-detector --stdin -v

# 配合子网扫描
echo "192.168.1.0/24" | ./llm-detector --stdin

# 配合其他命令
seq -f "192.168.1.%g" 1 254 | ./llm-detector --stdin -w 50
```

### 3. 进度显示

实时显示扫描进度：
- 进度条（彩色显示）
- 已完成/总数
- 预估剩余时间 (ETA)
- 当前检测目标

```bash
# 显示进度条（默认）
./llm-detector -f targets.txt

# 禁用进度条
./llm-detector -f targets.txt --no-progress
```

### 4. 批量输出格式

#### CSV 导出 (`--csv`)
```bash
./llm-detector -f targets.txt --csv results.csv
```
CSV 包含字段：
- target: 目标地址
- timestamp: 检测时间
- duration_ms: 耗时
- detected: 是否检测到
- framework: 检测到的框架
- framework_confidence: 框架置信度
- model: 检测到的模型
- model_confidence: 模型置信度
- provider: 提供商
- model_type: 模型类型
- version: 版本
- components: 检测到的组件列表
- error: 错误信息

#### HTML 报告 (`--html`)
```bash
./llm-detector -f targets.txt --html report.html
```
HTML 报告包含：
- 扫描统计概览
- 组件分布图表
- 端口分布统计
- 详细结果表格
- 置信度可视化

#### JSON Lines (`--jsonl`)
```bash
./llm-detector -f targets.txt --jsonl results.jsonl
```
每行一个 JSON 对象，便于流式处理。

#### 多格式同时导出
```bash
./llm-detector -f targets.txt --csv results.csv --html report.html --jsonl results.jsonl
```

### 5. 并发控制

#### 指定并发数 (`-w, --workers`)
```bash
# 指定20个并发 worker
./llm-detector -f targets.txt -w 20
```

#### 自适应并发（默认）
根据目标数量自动调整并发数：
- ≥500 目标: 50 workers
- ≥200 目标: 30 workers
- ≥100 目标: 20 workers
- ≥50 目标: 15 workers
- ≥20 目标: 10 workers
- ≥10 目标: 8 workers
- <10 目标: 5 workers

#### 速率限制 (`--rate`)
```bash
# 限制每秒10个请求
./llm-detector -f targets.txt --rate 10
```

### 6. 错误处理

#### 失败重试 (`--retries`)
```bash
# 设置重试3次（默认）
./llm-detector -f targets.txt --retries 3

# 禁用重试
./llm-detector -f targets.txt --retries 0
```

#### 超时跳过
自动跳过超时目标，继续扫描其他目标。

#### 指数退避
重试时使用指数退避策略，避免对目标服务器造成压力。

### 7. 统计信息

扫描完成后显示：
- 扫描总耗时
- 成功率统计
- 组件分布
- 端口分布
- 处理速率 (targets/sec)

## 使用示例

### 基础用法

```bash
# 扫描单个目标
./llm-detector -t 192.168.1.100

# 扫描 CIDR
./llm-detector -t 192.168.1.0/24

# 扫描 URL
./llm-detector -t http://example.com:8080
```

### 批量扫描

```bash
# 从文件扫描
./llm-detector -f targets.txt -v

# 管道输入
cat targets.txt | ./llm-detector --stdin -v

# 生成 CSV 和 HTML 报告
./llm-detector -f targets.txt --csv results.csv --html report.html
```

### 高级用法

```bash
# 高并发 + 速率限制
./llm-detector -f targets.txt -w 50 --rate 20

# 严格模式 + 详细输出
./llm-detector -f targets.txt --strict -v

# 自定义阈值
./llm-detector -f targets.txt --threshold 0.8

# 短超时 + 无重试（快速扫描）
./llm-detector -f targets.txt --timeout 3s --retries 0
```

### 配合其他工具

```bash
# 配合 nmap
nmap -p 11434,8080,8000 192.168.1.0/24 -oG - | grep "open" | cut -d' ' -f2 | ./llm-detector --stdin

# 配合 masscan
masscan 192.168.1.0/24 -p11434,8080 --rate 1000 -oL - | tail -n +2 | cut -d' ' -f4 | ./llm-detector --stdin -w 100

# 配合 shodan
shodan download --limit 100 llm.json.gz 'port:11434'
shodan parse --fields ip_str llm.json.gz | ./llm-detector --stdin
```

## 命令行选项

```
Flags:
  -t, --target string      目标地址 (IP, IP:Port, URL, CIDR)
  -f, --file string        从文件读取目标列表
      --stdin              从标准输入读取目标
      --timeout duration   探测超时时间 (默认 10s)
      --json               输出JSON格式
  -v, --verbose            详细输出
  -w, --workers int        并发数 (0=自动)
      --strict             严格模式 - 高置信度阈值 (70%)
      --loose              宽松模式 - 低置信度阈值 (50%)
      --threshold float    自定义置信度阈值 (0.0-1.0)
      --csv string         导出CSV格式到指定文件
      --html string        导出HTML报告到指定文件
      --jsonl string       导出JSON Lines格式到指定文件
      --rate int           速率限制 (每秒请求数, 0=无限制)
      --retries int        失败重试次数 (默认 3)
      --no-progress        禁用进度条
      --version            显示版本信息
  -h, --help               帮助信息
```

## 性能优化建议

1. **大规模扫描** (>1000 目标)
   - 使用 `-w 50` 或更高并发
   - 添加 `--rate` 限制避免被防火墙拦截
   - 使用 `--timeout 5s` 减少等待时间

2. **内网扫描**
   - 可以使用更高并发 `-w 100`
   - 可以降低超时 `--timeout 3s`

3. **互联网扫描**
   - 使用 `--rate 10` 限制速率
   - 保持默认重试次数
   - 使用较长的超时 `--timeout 15s`

## 输出文件示例

### CSV 示例
```csv
target,timestamp,duration_ms,detected,framework,framework_confidence,model,model_confidence,provider,model_type,version,components,error
192.168.1.100:11434,2024-01-15T10:30:00Z,1250,true,Ollama,0.95,Llama2,0.88,Meta,llm,2.1,ollama;api,
```

### JSON Lines 示例
```json
{"target":"192.168.1.100:11434","timestamp":"2024-01-15T10:30:00Z","duration":1250000000,"api_results":[...],"model_guess":{...},"service_info":{...}}
```

## 注意事项

1. **权限**: 扫描某些端口可能需要管理员权限
2. **法律**: 仅扫描您有权限的目标
3. **网络**: 大规模扫描可能触发防火墙规则
4. **资源**: 高并发扫描会消耗较多内存和CPU

## 故障排除

### 进度条不显示
- 检查是否使用了 `--no-progress`
- 检查终端是否支持 ANSI 转义序列

### 导出文件为空
- 检查是否有写入权限
- 检查目录是否存在
- 检查是否有扫描结果

### 扫描速度过慢
- 增加并发数 `-w`
- 减少超时时间 `--timeout`
- 禁用重试 `--retries 0`

### 内存占用过高
- 降低并发数
- 使用 `--rate` 限制速率
- 分批扫描大目标列表
