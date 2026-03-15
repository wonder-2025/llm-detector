package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// ProgressBar 进度条
type ProgressBar struct {
	total         int
	current       int
	startTime     time.Time
	width         int
	showPercent   bool
	showETA       bool
	showCount     bool
	currentTarget string
	mu            sync.RWMutex
}

// ProgressOption 进度条选项
type ProgressOption func(*ProgressBar)

// WithWidth 设置进度条宽度
func WithWidth(width int) ProgressOption {
	return func(pb *ProgressBar) {
		pb.width = width
	}
}

// WithShowPercent 显示百分比
func WithShowPercent(show bool) ProgressOption {
	return func(pb *ProgressBar) {
		pb.showPercent = show
	}
}

// WithShowETA 显示预估时间
func WithShowETA(show bool) ProgressOption {
	return func(pb *ProgressBar) {
		pb.showETA = show
	}
}

// WithShowCount 显示计数
func WithShowCount(show bool) ProgressOption {
	return func(pb *ProgressBar) {
		pb.showCount = show
	}
}

// NewProgressBar 创建进度条
func NewProgressBar(total int, opts ...ProgressOption) *ProgressBar {
	pb := &ProgressBar{
		total:       total,
		startTime:   time.Now(),
		width:       40,
		showPercent: true,
		showETA:     true,
		showCount:   true,
	}

	for _, opt := range opts {
		opt(pb)
	}

	return pb
}

// Update 更新进度
func (pb *ProgressBar) Update(current int) {
	pb.mu.Lock()
	pb.current = current
	pb.mu.Unlock()
}

// Increment 递增进度
func (pb *ProgressBar) Increment() {
	pb.mu.Lock()
	pb.current++
	pb.mu.Unlock()
}

// SetCurrentTarget 设置当前目标（用于显示）
func (pb *ProgressBar) SetCurrentTarget(target string) {
	pb.mu.Lock()
	pb.currentTarget = target
	pb.mu.Unlock()
}

// currentTarget 当前正在处理的目标
func (pb *ProgressBar) GetCurrentTarget() string {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.currentTarget
}



// String 返回进度条字符串
func (pb *ProgressBar) String() string {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	if pb.total == 0 {
		return ""
	}

	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

	// 构建进度条
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	var parts []string

	// 计数
	if pb.showCount {
		parts = append(parts, fmt.Sprintf("%d/%d", pb.current, pb.total))
	}

	// 进度条
	parts = append(parts, fmt.Sprintf("[%s]", bar))

	// 百分比
	if pb.showPercent {
		parts = append(parts, fmt.Sprintf("%.1f%%", percent*100))
	}

	// 预估时间
	if pb.showETA && pb.current > 0 {
		eta := pb.calculateETA()
		parts = append(parts, fmt.Sprintf("ETA: %s", eta))
	}

	return strings.Join(parts, " ")
}

// ColoredString 返回带颜色的进度条
func (pb *ProgressBar) ColoredString() string {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	if pb.total == 0 {
		return ""
	}

	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

	// 根据进度选择颜色
	var barColor func(string, ...interface{}) string
	switch {
	case percent >= 1.0:
		barColor = color.GreenString
	case percent >= 0.7:
		barColor = color.CyanString
	case percent >= 0.4:
		barColor = color.YellowString
	default:
		barColor = color.RedString
	}

	// 构建进度条
	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", pb.width-filled)
	bar := barColor("%s", filledBar) + color.WhiteString("%s", emptyBar)

	var parts []string

	// 计数
	if pb.showCount {
		parts = append(parts, color.WhiteString("%d/%d", pb.current, pb.total))
	}

	// 进度条
	parts = append(parts, fmt.Sprintf("[%s]", bar))

	// 百分比
	if pb.showPercent {
		percentStr := fmt.Sprintf("%.1f%%", percent*100)
		switch {
		case percent >= 1.0:
			parts = append(parts, color.GreenString(percentStr))
		case percent >= 0.7:
			parts = append(parts, color.CyanString(percentStr))
		case percent >= 0.4:
			parts = append(parts, color.YellowString(percentStr))
		default:
			parts = append(parts, color.RedString(percentStr))
		}
	}

	// 预估时间
	if pb.showETA && pb.current > 0 {
		eta := pb.calculateETA()
		parts = append(parts, color.WhiteString("ETA: %s", eta))
	}

	return strings.Join(parts, " ")
}

// calculateETA 计算预估剩余时间
func (pb *ProgressBar) calculateETA() string {
	if pb.current == 0 || pb.current >= pb.total {
		return "--:--"
	}

	elapsed := time.Since(pb.startTime)
	rate := float64(pb.current) / elapsed.Seconds()
	remaining := float64(pb.total-pb.current) / rate

	return formatDuration(time.Duration(remaining) * time.Second)
}

// GetStats 获取统计信息
func (pb *ProgressBar) GetStats() ProgressStats {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	elapsed := time.Since(pb.startTime)
	var rate float64
	if elapsed.Seconds() > 0 {
		rate = float64(pb.current) / elapsed.Seconds()
	}

	return ProgressStats{
		Total:      pb.total,
		Current:    pb.current,
		Percent:    float64(pb.current) / float64(pb.total) * 100,
		Elapsed:    elapsed,
		Rate:       rate,
		ETA:        pb.calculateETA(),
	}
}

// ProgressStats 进度统计
type ProgressStats struct {
	Total   int
	Current int
	Percent float64
	Elapsed time.Duration
	Rate    float64 // items per second
	ETA     string
}

// formatDuration 格式化持续时间
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// BatchProgress 批量扫描进度管理器
type BatchProgress struct {
	bar         *ProgressBar
	startTime   time.Time
	successCount int
	failCount    int
	skippedCount int
	mu          sync.RWMutex
}

// NewBatchProgress 创建批量扫描进度管理器
func NewBatchProgress(total int) *BatchProgress {
	return &BatchProgress{
		bar:       NewProgressBar(total),
		startTime: time.Now(),
	}
}

// Update 更新进度
func (bp *BatchProgress) Update(current int, target string) {
	bp.bar.Update(current)
	// 这里可以存储当前目标用于显示
}

// IncrementSuccess 递增成功计数
func (bp *BatchProgress) IncrementSuccess() {
	bp.mu.Lock()
	bp.successCount++
	bp.mu.Unlock()
	bp.bar.Increment()
}

// IncrementFail 递增失败计数
func (bp *BatchProgress) IncrementFail() {
	bp.mu.Lock()
	bp.failCount++
	bp.mu.Unlock()
	bp.bar.Increment()
}

// IncrementSkipped 递增跳过计数
func (bp *BatchProgress) IncrementSkipped() {
	bp.mu.Lock()
	bp.skippedCount++
	bp.mu.Unlock()
	bp.bar.Increment()
}

// Print 打印当前进度
func (bp *BatchProgress) Print() {
	fmt.Printf("\r%s", bp.bar.ColoredString())
}

// Println 打印进度并换行
func (bp *BatchProgress) Println() {
	fmt.Printf("\r%s\n", bp.bar.ColoredString())
}

// Finish 完成进度显示
func (bp *BatchProgress) Finish() {
	bp.bar.Update(bp.bar.total)
	fmt.Printf("\r%s\n", bp.bar.ColoredString())
}

// GetSummary 获取扫描摘要
func (bp *BatchProgress) GetSummary() string {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	elapsed := time.Since(bp.startTime)
	total := bp.successCount + bp.failCount + bp.skippedCount

	var parts []string
	parts = append(parts, color.CyanString("\n[+] Scan Summary:\n"))
	parts = append(parts, fmt.Sprintf("    Total Targets: %d\n", total))
	parts = append(parts, fmt.Sprintf("    Successful:    %s\n", color.GreenString("%d", bp.successCount)))
	parts = append(parts, fmt.Sprintf("    Failed:        %s\n", color.RedString("%d", bp.failCount)))
	parts = append(parts, fmt.Sprintf("    Skipped:       %s\n", color.YellowString("%d", bp.skippedCount)))
	parts = append(parts, fmt.Sprintf("    Duration:      %s\n", formatDuration(elapsed)))

	if elapsed.Seconds() > 0 {
		rate := float64(total) / elapsed.Seconds()
		parts = append(parts, fmt.Sprintf("    Rate:          %.2f targets/sec\n", rate))
	}

	if total > 0 {
		successRate := float64(bp.successCount) / float64(total) * 100
		parts = append(parts, fmt.Sprintf("    Success Rate:  %.1f%%\n", successRate))
	}

	return strings.Join(parts, "")
}

// ScanStatistics 扫描统计
type ScanStatistics struct {
	StartTime       time.Time
	EndTime         time.Time
	TotalTargets    int
	SuccessCount    int
	FailCount       int
	SkippedCount    int
	ComponentDist   map[string]int // 组件分布
	PortDist        map[int]int    // 端口分布
}

// NewScanStatistics 创建扫描统计
func NewScanStatistics(total int) *ScanStatistics {
	return &ScanStatistics{
		StartTime:     time.Now(),
		TotalTargets:  total,
		ComponentDist: make(map[string]int),
		PortDist:      make(map[int]int),
	}
}

// Finish 完成统计
func (ss *ScanStatistics) Finish() {
	ss.EndTime = time.Now()
}

// AddComponent 添加组件统计
func (ss *ScanStatistics) AddComponent(component string) {
	ss.ComponentDist[component]++
}

// AddPort 添加端口统计
func (ss *ScanStatistics) AddPort(port int) {
	ss.PortDist[port]++
}

// Duration 获取扫描持续时间
func (ss *ScanStatistics) Duration() time.Duration {
	if ss.EndTime.IsZero() {
		return time.Since(ss.StartTime)
	}
	return ss.EndTime.Sub(ss.StartTime)
}

// SuccessRate 获取成功率
func (ss *ScanStatistics) SuccessRate() float64 {
	if ss.TotalTargets == 0 {
		return 0
	}
	return float64(ss.SuccessCount) / float64(ss.TotalTargets) * 100
}

// String 返回统计字符串
func (ss *ScanStatistics) String() string {
	var parts []string

	parts = append(parts, color.CyanString("\n[+] Scan Statistics:\n"))
	parts = append(parts, fmt.Sprintf("    Total Duration:  %s\n", formatDuration(ss.Duration())))
	parts = append(parts, fmt.Sprintf("    Total Targets:   %d\n", ss.TotalTargets))
	parts = append(parts, fmt.Sprintf("    Successful:      %s\n", color.GreenString("%d", ss.SuccessCount)))
	parts = append(parts, fmt.Sprintf("    Failed:          %s\n", color.RedString("%d", ss.FailCount)))
	parts = append(parts, fmt.Sprintf("    Skipped:         %s\n", color.YellowString("%d", ss.SkippedCount)))
	parts = append(parts, fmt.Sprintf("    Success Rate:    %.1f%%\n", ss.SuccessRate()))

	// 组件分布
	if len(ss.ComponentDist) > 0 {
		parts = append(parts, color.CyanString("\n[+] Component Distribution:\n"))
		for component, count := range ss.ComponentDist {
			percent := float64(count) / float64(ss.SuccessCount) * 100
			parts = append(parts, fmt.Sprintf("    %s: %d (%.1f%%)\n", component, count, percent))
		}
	}

	// 端口分布
	if len(ss.PortDist) > 0 {
		parts = append(parts, color.CyanString("\n[+] Port Distribution:\n"))
		for port, count := range ss.PortDist {
			percent := float64(count) / float64(ss.SuccessCount) * 100
			parts = append(parts, fmt.Sprintf("    Port %d: %d (%.1f%%)\n", port, count, percent))
		}
	}

	return strings.Join(parts, "")
}

// Spinner 旋转等待指示器
type Spinner struct {
	frames []string
	index  int
	mu     sync.Mutex
}

// NewSpinner 创建旋转指示器
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Next 获取下一帧
func (s *Spinner) Next() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)
	return frame
}

// Spin 打印旋转指示器
func (s *Spinner) Spin(message string) {
	fmt.Printf("\r%s %s", color.CyanString(s.Next()), message)
}

// Stop 停止旋转指示器
func (s *Spinner) Stop() {
	fmt.Println()
}
