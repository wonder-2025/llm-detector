package core

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// TargetReader 目标读取器接口
type TargetReader interface {
	ReadTargets() ([]string, error)
}

// FileTargetReader 从文件读取目标
type FileTargetReader struct {
	Path string
}

// NewFileTargetReader 创建文件目标读取器
func NewFileTargetReader(path string) *FileTargetReader {
	return &FileTargetReader{Path: path}
}

// ReadTargets 从文件读取目标列表
func (r *FileTargetReader) ReadTargets() ([]string, error) {
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return readLines(file)
}

// StdinTargetReader 从标准输入读取目标
type StdinTargetReader struct{}

// NewStdinTargetReader 创建标准输入读取器
func NewStdinTargetReader() *StdinTargetReader {
	return &StdinTargetReader{}
}

// ReadTargets 从标准输入读取目标列表
func (r *StdinTargetReader) ReadTargets() ([]string, error) {
	return readLines(os.Stdin)
}

// SliceTargetReader 从字符串切片读取目标
type SliceTargetReader struct {
	Targets []string
}

// NewSliceTargetReader 创建切片目标读取器
func NewSliceTargetReader(targets []string) *SliceTargetReader {
	return &SliceTargetReader{Targets: targets}
}

// ReadTargets 从切片读取目标列表
func (r *SliceTargetReader) ReadTargets() ([]string, error) {
	return r.Targets, nil
}

// readLines 从读取器中读取行
func readLines(r interface{ Read([]byte) (int, error) }) ([]string, error) {
	var targets []string
	seen := make(map[string]bool)

	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 去重
		if seen[line] {
			continue
		}
		seen[line] = true

		targets = append(targets, line)
	}

	return targets, nil
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	engine      *Engine
	workers     int
	rateLimit   int // 每秒请求数限制
	maxRetries  int
	fullScan    bool // 全端口扫描
	onProgress  func(current, total int, target string)
	onResult    func(result *DetectionResult)
	onError     func(target string, err error, attempt int)
}

// BatchOption 批量处理选项
type BatchOption func(*BatchProcessor)

// WithWorkers 设置并发数
func WithWorkers(workers int) BatchOption {
	return func(bp *BatchProcessor) {
		bp.workers = workers
	}
}

// WithRateLimit 设置速率限制
func WithRateLimit(rps int) BatchOption {
	return func(bp *BatchProcessor) {
		bp.rateLimit = rps
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(retries int) BatchOption {
	return func(bp *BatchProcessor) {
		bp.maxRetries = retries
	}
}

// WithProgressCallback 设置进度回调
func WithProgressCallback(cb func(current, total int, target string)) BatchOption {
	return func(bp *BatchProcessor) {
		bp.onProgress = cb
	}
}

// WithResultCallback 设置结果回调
func WithResultCallback(cb func(result *DetectionResult)) BatchOption {
	return func(bp *BatchProcessor) {
		bp.onResult = cb
	}
}

// WithErrorCallback 设置错误回调
func WithErrorCallback(cb func(target string, err error, attempt int)) BatchOption {
	return func(bp *BatchProcessor) {
		bp.onError = cb
	}
}

// WithFullScan 设置全端口扫描
func WithFullScan(fullScan bool) BatchOption {
	return func(bp *BatchProcessor) {
		bp.fullScan = fullScan
	}
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor(engine *Engine, opts ...BatchOption) *BatchProcessor {
	bp := &BatchProcessor{
		engine:     engine,
		workers:    engine.workers,
		rateLimit:  0, // 无限制
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(bp)
	}

	return bp
}

// BatchResult 批量处理结果
type BatchResult struct {
	Results       []*DetectionResult
	Errors        []BatchError
	TotalTargets  int
	SuccessCount  int
	FailCount     int
	SkippedCount  int
	Duration      int64 // 毫秒
}

// BatchError 批量处理错误
type BatchError struct {
	Target  string
	Error   string
	Attempt int
}

// Process 处理目标列表
func (bp *BatchProcessor) Process(ctx context.Context, targets []string) *BatchResult {
	result := &BatchResult{
		TotalTargets: len(targets),
		Results:      make([]*DetectionResult, 0, len(targets)),
		Errors:       make([]BatchError, 0),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建工作池
	semaphore := make(chan struct{}, bp.workers)

	// 速率限制器
	var rateLimiter <-chan struct{}
	if bp.rateLimit > 0 {
		rateLimiter = bp.createRateLimiter(bp.rateLimit)
	}

	for i, targetStr := range targets {
		wg.Add(1)
		semaphore <- struct{}{}

		// 应用速率限制
		if rateLimiter != nil {
			<-rateLimiter
		}

		go func(idx int, t string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 进度回调
			if bp.onProgress != nil {
				bp.onProgress(idx+1, len(targets), t)
			}

			// 解析目标（支持端口发现）
			var targetList []*Target
			parsedTarget, err := ParseTarget(t)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, BatchError{
					Target:  t,
					Error:   fmt.Sprintf("parse error: %v", err),
					Attempt: 0,
				})
				result.FailCount++
				mu.Unlock()
				return
			}

			// 如果是纯IP且启用了全端口扫描，进行端口发现
			if parsedTarget.Type == TargetIP && bp.fullScan {
				timeout := 5 * time.Second
				targetList, err = ResolveTargetWithMode(ctx, t, timeout, true)
				if err != nil {
					// 端口发现失败，使用原始目标
					targetList = []*Target{parsedTarget}
				}
			} else {
				targetList = []*Target{parsedTarget}
			}

			// 对每个发现的目标进行探测
			for _, target := range targetList {
				// 执行探测（带重试）
				detectionResult, err := bp.detectWithRetry(ctx, target)

				mu.Lock()
				if err != nil {
					// 只在最后一个目标失败时记录错误
					if target == targetList[len(targetList)-1] {
						result.Errors = append(result.Errors, BatchError{
							Target:  t,
							Error:   err.Error(),
							Attempt: bp.maxRetries,
						})
						result.FailCount++
					}
				} else {
					result.Results = append(result.Results, detectionResult)
					result.SuccessCount++

					// 结果回调
					if bp.onResult != nil {
						bp.onResult(detectionResult)
					}
				}
				mu.Unlock()
			}
		}(i, targetStr)
	}

	wg.Wait()
	return result
}

// detectWithRetry 带重试的探测
func (bp *BatchProcessor) detectWithRetry(ctx context.Context, target *Target) (*DetectionResult, error) {
	var lastErr error

	for attempt := 0; attempt <= bp.maxRetries; attempt++ {
		if attempt > 0 {
			// 错误回调
			if bp.onError != nil {
				bp.onError(target.String(), lastErr, attempt)
			}
			// 指数退避
			backoff := bp.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		ctx, cancel := context.WithTimeout(ctx, bp.engine.timeout*2)
		result, err := bp.engine.Detect(ctx, target)
		cancel()

		if err == nil {
			return result, nil
		}

		lastErr = err

		// 如果是不可重试的错误，直接返回
		if !isRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateBackoff 计算退避时间
func (bp *BatchProcessor) calculateBackoff(attempt int) time.Duration {
	base := time.Second
	maxBackoff := 30 * time.Second

	backoff := base * time.Duration(1<<uint(attempt))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return backoff
}

// createRateLimiter 创建速率限制器
func (bp *BatchProcessor) createRateLimiter(rps int) <-chan struct{} {
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	ch := make(chan struct{}, 1)

	go func() {
		for range ticker.C {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}()

	return ch
}

// isRetryableError 判断错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"no such host",
		"i/o timeout",
		"temporary failure",
		"too many requests",
		"service unavailable",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// ParseTargetsFromStrings 解析目标字符串列表
func ParseTargetsFromStrings(inputs []string) ([]*Target, []ParseError) {
	var targets []*Target
	var errors []ParseError

	for _, input := range inputs {
		// 处理CIDR
		if strings.Contains(input, "/") {
			cidrTargets, err := parseCIDR(input)
			if err != nil {
				errors = append(errors, ParseError{
					Input: input,
					Error: err.Error(),
				})
				continue
			}
			targets = append(targets, cidrTargets...)
			continue
		}

		// 解析单个目标
		target, err := ParseTarget(input)
		if err != nil {
			errors = append(errors, ParseError{
				Input: input,
				Error: err.Error(),
			})
			continue
		}

		targets = append(targets, target)
	}

	return targets, errors
}

// ParseError 解析错误
type ParseError struct {
	Input string
	Error string
}

// parseCIDR 解析CIDR并展开为IP列表
func parseCIDR(cidr string) ([]*Target, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var targets []*Target
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		// 跳过网络地址和广播地址
		if isNetworkOrBroadcast(ip, ipnet) {
			continue
		}

		targets = append(targets, &Target{
			Type: TargetIP,
			Host: ip.String(),
			Raw:  ip.String(),
		})
	}

	return targets, nil
}

// incrementIP 递增IP地址
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// isNetworkOrBroadcast 判断是否为网络或广播地址
func isNetworkOrBroadcast(ip net.IP, ipnet *net.IPNet) bool {
	mask := ipnet.Mask
	network := ip.Mask(mask)

	// 网络地址
	if ip.Equal(network) {
		return true
	}

	// 广播地址
	broadcast := make(net.IP, len(network))
	copy(broadcast, network)
	for i := range mask {
		broadcast[i] |= ^mask[i]
	}
	if ip.Equal(broadcast) {
		return true
	}

	return false
}

// FilterValidTargets 过滤有效目标
func FilterValidTargets(inputs []string) []string {
	var valid []string
	seen := make(map[string]bool)

	for _, input := range inputs {
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 去重
		if seen[input] {
			continue
		}
		seen[input] = true

		// 验证目标格式
		if isValidTarget(input) {
			valid = append(valid, input)
		}
	}

	return valid
}

// isValidTarget 验证目标格式
func isValidTarget(input string) bool {
	// URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return true
	}

	// CIDR
	if strings.Contains(input, "/") {
		_, _, err := net.ParseCIDR(input)
		return err == nil
	}

	// IP:Port
	if strings.Contains(input, ":") {
		host, _, err := net.SplitHostPort(input)
		if err == nil {
			return net.ParseIP(host) != nil
		}
	}

	// 纯IP
	return net.ParseIP(input) != nil
}
