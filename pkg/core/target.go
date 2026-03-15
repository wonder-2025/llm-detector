package core

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TargetType 目标类型
type TargetType int

const (
	TargetIP TargetType = iota
	TargetIPPort
	TargetURL
)

// Target 探测目标
type Target struct {
	Type     TargetType
	Host     string
	Port     int
	Path     string
	Scheme   string
	Raw      string
	Endpoints []string // 发现的API端点
}

// String 返回目标字符串表示
func (t *Target) String() string {
	switch t.Type {
	case TargetIP:
		return t.Host
	case TargetIPPort:
		return fmt.Sprintf("%s:%d", t.Host, t.Port)
	case TargetURL:
		return t.Raw
	}
	return t.Raw
}

// BaseURL 返回基础URL
func (t *Target) BaseURL() string {
	if t.Type == TargetURL {
		// URL类型时，如果端口非标准，需要包含端口
		if t.Port != 80 && t.Port != 443 {
			return fmt.Sprintf("%s://%s:%d%s", t.Scheme, t.Host, t.Port, t.Path)
		}
		return fmt.Sprintf("%s://%s%s", t.Scheme, t.Host, t.Path)
	}
	// IP:Port类型，使用Scheme
	scheme := t.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if t.Port == 443 {
		return fmt.Sprintf("https://%s", t.Host)
	}
	if t.Port == 80 {
		return fmt.Sprintf("http://%s", t.Host)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, t.Host, t.Port)
}

// ParseTarget 解析目标字符串
func ParseTarget(input string) (*Target, error) {
	input = strings.TrimSpace(input)

	// 尝试解析为URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return parseURL(input)
	}

	// 尝试解析为 IP:Port
	if strings.Contains(input, ":") {
		return parseIPPort(input)
	}

	// 纯IP
	return parseIP(input)
}

func parseURL(input string) (*Target, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		if u.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	return &Target{
		Type:   TargetURL,
		Host:   host,
		Port:   port,
		Path:   u.Path,
		Scheme: u.Scheme,
		Raw:    input,
	}, nil
}

func parseIPPort(input string) (*Target, error) {
	host, portStr, err := net.SplitHostPort(input)
	if err != nil {
		return nil, fmt.Errorf("invalid IP:Port format: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	return &Target{
		Type: TargetIPPort,
		Host: host,
		Port: port,
		Raw:  input,
	}, nil
}

func parseIP(input string) (*Target, error) {
	ip := net.ParseIP(input)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", input)
	}

	return &Target{
		Type: TargetIP,
		Host: input,
		Raw:  input,
	}, nil
}

// PortScanner 端口扫描器
type PortScanner struct {
	Timeout   time.Duration
	FullScan  bool // 是否全端口扫描
}

// NewPortScanner 创建端口扫描器
func NewPortScanner(timeout time.Duration) *PortScanner {
	return &PortScanner{
		Timeout:  timeout,
		FullScan: false,
	}
}

// NewFullPortScanner 创建全端口扫描器
func NewFullPortScanner(timeout time.Duration) *PortScanner {
	return &PortScanner{
		Timeout:  timeout,
		FullScan: true,
	}
}

// CommonLLMPorts 常见大模型服务端口
var CommonLLMPorts = []int{
	// AI/LLM 服务
	11434, // Ollama
	8080,  // vLLM, TGI, 通用API
	8000,  // FastAPI, 通用API
	3000,  // Node.js API
	5000,  // Flask
	5001,  // Flask development
	8888,  // Jupyter Notebook/Lab
	8889,  // Jupyter 备用
	8501,  // Streamlit
	7860,  // Gradio
	
	// HTTPS/Web
	443,   // HTTPS standard
	80,    // HTTP standard
	8443,  // HTTPS alternate
	9443,  // HTTPS alternate
	8081,  // HTTP alternate
	8082,  // HTTP alternate
	
	// 数据库/存储
	9090,  // Prometheus/Grafana
	3306,  // MySQL
	5432,  // PostgreSQL
	6379,  // Redis
	27017, // MongoDB
	9200,  // Elasticsearch
	5601,  // Kibana
	
	// 消息队列/缓存
	5672,  // RabbitMQ
	15672, // RabbitMQ Management
	9092,  // Kafka
	2181,  // Zookeeper
	
	// 容器/K8s
	10250, // Kubelet
	6443,  // Kubernetes API
	
	// 其他常用
	22,    // SSH
	3389,  // RDP
	5900,  // VNC
}

// FullScanPorts 全端口扫描范围 (1-65535)
var FullScanPorts []int

func init() {
	// 初始化全端口列表
	FullScanPorts = make([]int, 65535)
	for i := 1; i <= 65535; i++ {
		FullScanPorts[i-1] = i
	}
}

// ScanPorts 扫描端口
func (s *PortScanner) ScanPorts(ctx context.Context, host string) ([]int, error) {
	var portsToScan []int
	if s.FullScan {
		portsToScan = FullScanPorts
	} else {
		portsToScan = CommonLLMPorts
	}

	var openPorts []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 全端口扫描使用更大的并发限制
	concurrency := 50
	if s.FullScan {
		concurrency = 1000
	}
	semaphore := make(chan struct{}, concurrency)

	for _, port := range portsToScan {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(p int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if s.isPortOpen(ctx, host, p) {
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts, nil
}

func (s *PortScanner) isPortOpen(ctx context.Context, host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	conn, err := net.DialTimeout("tcp", address, s.Timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ResolveTarget 解析并发现目标
func ResolveTarget(ctx context.Context, input string, timeout time.Duration) ([]*Target, error) {
	return ResolveTargetWithMode(ctx, input, timeout, false)
}

// ResolveTargetWithMode 解析并发现目标（支持全端口扫描）
func ResolveTargetWithMode(ctx context.Context, input string, timeout time.Duration, fullScan bool) ([]*Target, error) {
	target, err := ParseTarget(input)
	if err != nil {
		return nil, err
	}

	// 如果是纯IP，扫描端口
	if target.Type == TargetIP {
		var scanner *PortScanner
		if fullScan {
			scanner = NewFullPortScanner(timeout)
		} else {
			scanner = NewPortScanner(timeout)
		}
		ports, err := scanner.ScanPorts(ctx, target.Host)
		if err != nil {
			return nil, err
		}

		if len(ports) == 0 {
			return nil, fmt.Errorf("no open ports found on %s", target.Host)
		}

		var targets []*Target
		for _, port := range ports {
			scheme := "http"
			// HTTPS端口识别
			switch port {
			case 443, 8443, 9443, 6443: // 标准HTTPS和K8s API
				scheme = "https"
			}
			targets = append(targets, &Target{
				Type:   TargetIPPort,
				Host:   target.Host,
				Port:   port,
				Scheme: scheme,
				Raw:    fmt.Sprintf("%s:%d", target.Host, port),
			})
		}
		return targets, nil
	}

	return []*Target{target}, nil
}
