package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"llm-detector/pkg/core"
	"llm-detector/pkg/fingerprints"
	"llm-detector/pkg/plugins"
	"llm-detector/pkg/plugins/api"
)

var (
	targetStr   string
	targetFile  string
	useStdin    bool
	timeout     time.Duration
	jsonOut     bool
	verbose     bool
	workers     int
	strictMode  bool
	looseMode   bool
	threshold   float64
	showVersion bool
	csvOut      string
	htmlOut     string
	jsonlOut    string
	outputPath  string
	useHTML     bool  // --html 作为开关使用
	fullScan    bool  // 全端口扫描
	rateLimit   int
	maxRetries  int
	noProgress  bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "llm-detector",
		Short: "大模型组件识别工具",
		Long: `LLM Detector - 识别目标系统使用的大模型组件

支持识别:
  - 大模型类型 (GPT-4, Claude, Llama, Qwen等)
  - API框架 (OpenAI, Ollama, vLLM等)
  - 部署方式 (Docker, K8s等)

评分系统:
  - 响应头匹配 (30%)
  - 响应体关键词 (40%)
  - JSON结构匹配 (30%)
  - 置信度阈值: 70% (严格模式) / 50% (宽松模式)

使用示例:
  # 探测单个IP地址
  llm-detector -t 192.168.1.100

  # 探测指定端口
  llm-detector -t 192.168.1.100:11434

  # 探测URL
  llm-detector -t http://192.168.1.100:8080

  # 从文件批量扫描
  llm-detector -f targets.txt -v

  # 管道输入
  cat ips.txt | llm-detector --stdin -v

  # 批量导出CSV
  llm-detector -f targets.txt --csv results.csv

  # 生成HTML报告
  llm-detector -f targets.txt --html report.html

  # 使用 -o 指定输出文件
  llm-detector -f targets.txt --html -o report.html

  # 全端口扫描 (扫描1-65535端口)
  llm-detector -t 192.168.1.100 --full-scan

  # 指定并发数和速率限制
  llm-detector -f targets.txt -w 20 --rate 10 --csv results.csv

  # 严格模式 (高置信度要求)
  llm-detector -t 192.168.1.100 --strict

  # 宽松模式 (低置信度要求)
  llm-detector -t 192.168.1.100 --loose

  # 自定义置信度阈值
  llm-detector -t 192.168.1.100 --threshold 0.8

  # 输出JSON格式
  llm-detector -t 192.168.1.100 --json

  # 详细输出
  llm-detector -t 192.168.1.100 --verbose
`,
		RunE: run,
	}

	rootCmd.Flags().StringVarP(&targetStr, "target", "t", "", "目标地址 (IP, IP:Port, URL, CIDR)")
	rootCmd.Flags().StringVarP(&targetFile, "file", "f", "", "从文件读取目标列表")
	rootCmd.Flags().BoolVar(&useStdin, "stdin", false, "从标准输入读取目标")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "探测超时时间")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "输出JSON格式")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 0, "并发数 (0=自动)")
	rootCmd.Flags().BoolVar(&strictMode, "strict", false, "严格模式 - 高置信度阈值 (70%)")
	rootCmd.Flags().BoolVar(&looseMode, "loose", false, "宽松模式 - 低置信度阈值 (50%)")
	rootCmd.Flags().Float64Var(&threshold, "threshold", 0, "自定义置信度阈值 (0.0-1.0)")
	rootCmd.Flags().StringVar(&csvOut, "csv", "", "导出CSV格式到指定文件")
	rootCmd.Flags().StringVar(&htmlOut, "html", "", "导出HTML报告到指定文件路径")
	rootCmd.Flags().BoolVar(&useHTML, "html-fmt", false, "使用HTML格式输出 (配合 -o 使用)")
	rootCmd.Flags().StringVar(&jsonlOut, "jsonl", "", "导出JSON Lines格式到指定文件")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件路径 (配合 --csv/--html/--jsonl 使用，或单独使用根据扩展名推断格式)")
	rootCmd.Flags().BoolVar(&fullScan, "full-scan", false, "全端口扫描模式 (扫描1-65535端口，较慢)")
	rootCmd.Flags().IntVar(&rateLimit, "rate", 0, "速率限制 (每秒请求数, 0=无限制)")
	rootCmd.Flags().IntVar(&maxRetries, "retries", 3, "失败重试次数")
	rootCmd.Flags().BoolVar(&noProgress, "no-progress", false, "禁用进度条")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "显示版本信息")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// 显示版本
	if showVersion {
		printVersion()
		return nil
	}

	// 打印Banner
	printBanner()

	// 读取目标列表
	targets, err := readTargets()
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		return fmt.Errorf("no targets specified. Use -t, -f, or --stdin")
	}

	if verbose {
		color.Cyan("[*] Loaded %d target(s)\n", len(targets))
	}

	// 加载指纹库
	loader := fingerprints.NewLoader()
	fpDir := findFingerprintDir()

	if err := loader.LoadAll(fpDir); err != nil {
		if verbose {
			color.Yellow("[!] Warning: failed to load fingerprints: %v\n", err)
		}
	} else {
		if verbose {
			color.Green("[+] Loaded %d model fingerprints\n", loader.ModelCount())
			color.Green("[+] Loaded %d framework fingerprints\n", loader.FrameworkCount())
		}
	}

	// 初始化插件注册中心
	registry := plugins.NewRegistry()
	api.RegisterAll(registry, timeout)

	if verbose {
		color.Green("[+] Registered %d API plugins\n", len(registry.AllAPIs()))
	}

	// 创建探测引擎
	engine := createEngine(registry, loader, timeout, workers, len(targets))
	setScoringMode(engine)

	if verbose {
		mode := engine.GetMode()
		color.Cyan("[*] Scoring mode: %s (threshold: %.0f%%)\n", mode.String(), engine.GetThreshold()*100)
		if workers > 0 {
			color.Cyan("[*] Workers: %d\n", workers)
		} else {
			color.Cyan("[*] Workers: auto (%d)\n", engine.GetWorkers())
		}
		if rateLimit > 0 {
			color.Cyan("[*] Rate limit: %d req/sec\n", rateLimit)
		}
	}

	// 创建输出写入器
	outputWriter, err := createOutputWriter()
	if err != nil {
		return err
	}
	if outputWriter != nil {
		defer outputWriter.Close()
	}

	// 执行批量扫描
	if len(targets) > 1 || targetFile != "" || useStdin {
		return runBatchScan(engine, targets, outputWriter)
	}

	// 单目标扫描
	return runSingleScan(engine, targets[0], outputWriter)
}

// readTargets 读取目标列表
func readTargets() ([]string, error) {
	var reader core.TargetReader

	switch {
	case targetFile != "":
		reader = core.NewFileTargetReader(targetFile)
	case useStdin:
		reader = core.NewStdinTargetReader()
	case targetStr != "":
		reader = core.NewSliceTargetReader([]string{targetStr})
	default:
		return nil, fmt.Errorf("no target specified")
	}

	targets, err := reader.ReadTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to read targets: %w", err)
	}

	// 过滤有效目标
	validTargets := core.FilterValidTargets(targets)

	return validTargets, nil
}

// runBatchScan 执行批量扫描
func runBatchScan(engine *core.Engine, targets []string, outputWriter core.OutputWriter) error {
	// 创建进度管理器
	var progress *core.BatchProgress
	if !noProgress {
		progress = core.NewBatchProgress(len(targets))
	}

	// 创建统计信息
	stats := core.NewScanStatistics(len(targets))

	// 创建批量处理器
	opts := []core.BatchOption{
		core.WithWorkers(engine.GetWorkers()),
		core.WithMaxRetries(maxRetries),
	}

	if rateLimit > 0 {
		opts = append(opts, core.WithRateLimit(rateLimit))
	}

	// 全端口扫描模式
	if fullScan {
		opts = append(opts, core.WithFullScan(true))
	}

	if !noProgress {
		opts = append(opts, core.WithProgressCallback(func(current, total int, target string) {
			progress.Update(current, target)
			progress.Print()
		}))
	}

	opts = append(opts, core.WithResultCallback(func(result *core.DetectionResult) {
		// 写入输出文件
		if outputWriter != nil {
			outputWriter.Write(result)
		}

		// 更新统计
		progress.IncrementSuccess()

		// 收集组件统计
		for _, api := range result.APIResults {
			if api.Available {
				stats.AddComponent(api.Type)
			}
		}

		// 收集端口统计
		if result.ServiceInfo != nil && result.ServiceInfo.Deployment != "" {
			// 从部署信息中提取端口
			var port int
			fmt.Sscanf(result.ServiceInfo.Deployment, "Port %d", &port)
			if port > 0 {
				stats.AddPort(port)
			}
		}
	}))

	opts = append(opts, core.WithErrorCallback(func(target string, err error, attempt int) {
		if verbose {
			color.Yellow("[!] Retry %d for %s: %v\n", attempt, target, err)
		}
	}))

	processor := core.NewBatchProcessor(engine, opts...)

	// 开始扫描
	startTime := time.Now()
	result := processor.Process(context.Background(), targets)
	duration := time.Since(startTime)

	// 更新统计
	stats.SuccessCount = result.SuccessCount
	stats.FailCount = result.FailCount
	stats.SkippedCount = result.SkippedCount
	stats.Finish()

	// 完成进度显示
	if !noProgress {
		progress.Finish()
	}

	// 打印结果摘要
	fmt.Println()
	color.Cyan("[+] Scan completed in %v\n", duration)
	fmt.Println()

	// 打印统计摘要
	if !noProgress {
		fmt.Print(progress.GetSummary())
	}

	// 打印详细统计
	fmt.Print(stats.String())

	// 如果有错误，打印错误摘要
	if len(result.Errors) > 0 && verbose {
		color.Red("\n[-] Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			if i >= 10 {
				fmt.Printf("    ... and %d more errors\n", len(result.Errors)-10)
				break
			}
			fmt.Printf("    - %s: %s\n", err.Target, err.Error)
		}
	}

	// 设置HTML统计信息
	if htmlWriter, ok := outputWriter.(*core.HTMLWriter); ok {
		htmlWriter.SetStatistics(stats)
	}

	return nil
}

// runSingleScan 执行单目标扫描
func runSingleScan(engine *core.Engine, targetStr string, outputWriter core.OutputWriter) error {
	// 解析目标并自动发现端口（如果是纯IP）
	ctx, cancel := context.WithTimeout(context.Background(), timeout*3)
	defer cancel()

	// 根据是否全端口扫描选择解析方式
	var targets []*core.Target
	var err error
	if fullScan {
		targets, err = core.ResolveTargetWithMode(ctx, targetStr, timeout, true)
	} else {
		targets, err = core.ResolveTarget(ctx, targetStr, timeout)
	}
	if err != nil {
		// 如果ResolveTarget失败（如无开放端口），回退到直接解析
		target, err := core.ParseTarget(targetStr)
		if err != nil {
			return fmt.Errorf("failed to parse target: %w", err)
		}
		targets = []*core.Target{target}
	}

	// 创建统计信息（原始目标数为1）
	stats := core.NewScanStatistics(1)
	defer stats.Finish()

	var hasDetection bool
	var detectedComponents []string

	// 收集所有结果
	var allResults []*core.DetectionResult

	// 扫描所有发现的目标（端口）
	for _, target := range targets {
		if verbose {
			color.Cyan("\n[*] Detecting target: %s\n", target.String())
		}

		result, err := engine.Detect(ctx, target)
		if err != nil {
			if verbose {
				color.Red("[-] Detection failed for %s: %v\n", target.String(), err)
			}
			continue
		}

		// 收集检测到的组件
		if result.IsDetected() {
			hasDetection = true
			for _, api := range result.APIResults {
				if api.Available {
					detectedComponents = append(detectedComponents, api.Type)
					stats.AddComponent(api.Type)
				}
			}
			if result.ServiceInfo != nil && result.ServiceInfo.Framework != "" {
				stats.AddComponent(result.ServiceInfo.Framework)
			}
		}

		// 收集结果
		allResults = append(allResults, result)

		// 写入输出文件
		if outputWriter != nil {
			outputWriter.Write(result)
		}
	}

	// 更新成功率（基于检测到的组件数）
	if len(detectedComponents) > 0 {
		stats.SuccessCount = len(detectedComponents)
	}

	// 设置HTML统计信息
	if htmlWriter, ok := outputWriter.(*core.HTMLWriter); ok {
		htmlWriter.SetStatistics(stats)
	}

	// 输出最终结果
	if hasDetection {
		if jsonOut {
			// JSON输出所有结果
			jsonData, err := json.MarshalIndent(allResults, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %w", err)
			}
			fmt.Println(string(jsonData))
		} else if verbose {
			for _, result := range allResults {
				fmt.Println(result.StringVerbose())
			}
		} else {
			for _, result := range allResults {
				fmt.Println(result.String())
			}
		}
	} else {
		color.Yellow("\n[!] No LLM components detected on %s\n", targetStr)
	}

	return nil
}

// createOutputWriter 创建输出写入器
func createOutputWriter() (core.OutputWriter, error) {
	var writers []core.OutputWriter

	// 如果使用了 --html-fmt 开关，设置htmlOut为outputPath
	if useHTML && outputPath != "" {
		htmlOut = outputPath
	}

	// 如果有 -o 参数，用它作为默认输出路径
	if outputPath != "" {
		if csvOut == "" && htmlOut == "" && jsonlOut == "" {
			// 如果只指定了 -o 但没有指定格式，根据扩展名推断
			ext := filepath.Ext(outputPath)
			switch ext {
			case ".csv":
				csvOut = outputPath
			case ".html":
				htmlOut = outputPath
			case ".jsonl":
				jsonlOut = outputPath
			default:
				// 默认使用HTML
				htmlOut = outputPath
			}
		}
	}

	if csvOut != "" {
		writer, err := core.CreateOutputWriter(core.FormatCSV, csvOut)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
		if verbose {
			color.Cyan("[*] Output to CSV: %s\n", csvOut)
		}
	}

	if htmlOut != "" {
		writer, err := core.CreateOutputWriter(core.FormatHTML, htmlOut)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
		if verbose {
			color.Cyan("[*] Output to HTML: %s\n", htmlOut)
		}
	}

	if jsonlOut != "" {
		writer, err := core.CreateOutputWriter(core.FormatJSONL, jsonlOut)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
		if verbose {
			color.Cyan("[*] Output to JSONL: %s\n", jsonlOut)
		}
	}

	if len(writers) == 0 {
		return nil, nil
	}

	if len(writers) == 1 {
		return writers[0], nil
	}

	return core.NewMultiOutputWriter(writers...), nil
}

func printBanner() {
	banner := `
 _     _ _                  _            _           
| |   | | |                | |          | |          
| |   | | |_   _ _ __   ___| |_ ___  ___| |_ ___ _ __ 
| |   | | | | | | '_ \ / _ \ __/ _ \/ __| __/ _ \ '__|
| |___| | | |_| | | | |  __/ ||  __/\__ \ ||  __/ |   
|_____|_|_|\__,_|_| |_|\___|\__\___||___/\__\___|_|   
                                                        
              LLM Component Detector v2.1.0
              Multi-dimensional Scoring System
              Batch Scanning Enabled
`
	color.Cyan(banner)
}

func printVersion() {
	fmt.Println("LLM Detector v2.1.0")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Multi-dimensional scoring system")
	fmt.Println("  - Response header matching (30%)")
	fmt.Println("  - Body keyword matching (40%)")
	fmt.Println("  - JSON structure matching (30%)")
	fmt.Println("  - Version extraction")
	fmt.Println("  - Strict/Loose scoring modes")
	fmt.Println("  - 185+ enhanced fingerprints")
	fmt.Println("  - Batch scanning with progress bar")
	fmt.Println("  - CSV/HTML/JSONL export")
	fmt.Println("  - Rate limiting and retry")
	fmt.Println("  - File and stdin input")
}

func getExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}

// findFingerprintDir 查找指纹库目录
func findFingerprintDir() string {
	// 尝试多个可能的指纹目录位置
	fpDirs := []string{
		// 发布包结构: 可执行文件同级目录下的 fingerprints/
		filepath.Join(getExecutableDir(), "fingerprints"),
		// 开发环境: pkg/fingerprints/data/
		filepath.Join(getExecutableDir(), "pkg", "fingerprints", "data"),
		filepath.Join(getExecutableDir(), "..", "pkg", "fingerprints", "data"),
		"/root/.openclaw/workspace/llm-detector/pkg/fingerprints/data",
	}

	for _, dir := range fpDirs {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return fpDirs[0]
}

// createEngine 创建探测引擎（支持自适应并发）
func createEngine(registry *plugins.Registry, loader *fingerprints.Loader, timeout time.Duration, workers, targetCount int) *core.Engine {
	// 如果指定了并发数，使用指定值
	if workers > 0 {
		return core.NewEngineWithWorkers(registry, loader, timeout, workers)
	}

	// 自适应并发：根据目标数量动态调整
	adaptiveWorkers := calculateAdaptiveWorkers(targetCount)
	return core.NewEngineWithWorkers(registry, loader, timeout, adaptiveWorkers)
}

// calculateAdaptiveWorkers 计算自适应并发数
func calculateAdaptiveWorkers(targetCount int) int {
	switch {
	case targetCount >= 500:
		return 50
	case targetCount >= 200:
		return 30
	case targetCount >= 100:
		return 20
	case targetCount >= 50:
		return 15
	case targetCount >= 20:
		return 10
	case targetCount >= 10:
		return 8
	default:
		return 5
	}
}

// setScoringMode 设置评分模式
func setScoringMode(engine *core.Engine) {
	// 优先级: --threshold > --strict/--loose > 默认
	if threshold > 0 {
		engine.SetThreshold(threshold)
		return
	}

	if strictMode {
		engine.SetMode(core.ModeStrict)
		return
	}

	if looseMode {
		engine.SetMode(core.ModeLoose)
		return
	}

	// 默认使用严格模式
	engine.SetMode(core.ModeStrict)
}
