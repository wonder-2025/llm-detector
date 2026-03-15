package fingerprints

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnhancedHeaderPattern 增强的响应头模式
type EnhancedHeaderPattern struct {
	Name        string  `yaml:"name"`
	Pattern     string  `yaml:"pattern,omitempty"`
	Value       string  `yaml:"value,omitempty"`
	Required    bool    `yaml:"required"`
	Weight      float64 `yaml:"weight,omitempty"`      // 匹配权重
	Extract     string  `yaml:"extract,omitempty"`     // 提取组索引
	Description string  `yaml:"description,omitempty"`
}

// EnhancedBodyPattern 增强的响应体模式
type EnhancedBodyPattern struct {
	Field       string   `yaml:"field,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
	Value       string   `yaml:"value,omitempty"`
	JSONPath    string   `yaml:"json_path,omitempty"`   // JSON路径提取
	Required    bool     `yaml:"required"`
	Weight      float64  `yaml:"weight,omitempty"`
	Keywords    []string `yaml:"keywords,omitempty"`    // 关键词列表
	Description string   `yaml:"description,omitempty"`
}

// ScoringConfig 评分配置
type ScoringConfig struct {
	HeaderMatch   float64 `yaml:"header_match"`    // 响应头匹配权重
	BodyKeywords  float64 `yaml:"body_keywords"`   // 响应体关键词权重
	JSONStructure float64 `yaml:"json_structure"`  // JSON结构匹配权重
	Threshold     float64 `yaml:"threshold"`       // 匹配阈值
}

// DefaultScoringConfig 返回默认评分配置
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		HeaderMatch:   0.30,
		BodyKeywords:  0.40,
		JSONStructure: 0.30,
		Threshold:     0.70,
	}
}

// ModelFingerprint 模型指纹定义 (增强版)
type ModelFingerprint struct {
	Name           string                 `yaml:"name"`
	Provider       string                 `yaml:"provider"`
	Type           string                 `yaml:"type"`
	Description    string                 `yaml:"description"`
	Version        string                 `yaml:"version,omitempty"`        // 指纹版本
	Response       ResponseFeatures       `yaml:"response"`
	Behavior       BehaviorFeatures       `yaml:"behavior"`
	Tests          map[string]float64     `yaml:"tests"`
	Fingerprints   []TestFingerprint      `yaml:"fingerprints"`
	Variants       []ModelVariant         `yaml:"variants"`
	Scoring        ScoringConfig          `yaml:"scoring,omitempty"`        // 评分配置
}

// ResponseFeatures 响应特征
type ResponseFeatures struct {
	Headers      []EnhancedHeaderPattern `yaml:"headers"`
	BodyPatterns []EnhancedBodyPattern   `yaml:"body_patterns"`
}

// BehaviorFeatures 行为特征
type BehaviorFeatures struct {
	MaxTokens               int    `yaml:"max_tokens"`
	SupportsFunctionCalling bool   `yaml:"supports_function_calling"`
	SupportsVision          bool   `yaml:"supports_vision"`
	SupportsJSONMode        bool   `yaml:"supports_json_mode"`
	ResponseTimeMs          string `yaml:"response_time_ms"`
	StreamingSupported      bool   `yaml:"streaming_supported"`
}

// TestFingerprint 测试指纹
type TestFingerprint struct {
	Name              string   `yaml:"test"`
	Prompt            string   `yaml:"prompt"`
	ExpectedKeywords  []string `yaml:"expected_keywords,omitempty"`
	ExpectedPatterns  []string `yaml:"expected_patterns,omitempty"`
	ForbiddenKeywords []string `yaml:"forbidden_keywords,omitempty"`
	ForbiddenPatterns []string `yaml:"forbidden_patterns,omitempty"`
	Weight            float64  `yaml:"weight"`
}

// ModelVariant 模型变体
type ModelVariant struct {
	Name     string   `yaml:"name"`
	Pattern  string   `yaml:"pattern"`
	Features []string `yaml:"features"`
}

// EnhancedEndpoint 增强的端点定义
type EnhancedEndpoint struct {
	Path        string   `yaml:"path"`
	Method      string   `yaml:"method"`
	Description string   `yaml:"description"`
	Weight      float64  `yaml:"weight,omitempty"`      // 端点匹配权重
	Patterns    []string `yaml:"patterns,omitempty"`    // 额外匹配模式
}

// EnhancedVersionInfo 增强的版本信息
type EnhancedVersionInfo struct {
	Pattern     string            `yaml:"pattern"`
	Features    []string          `yaml:"features"`
	ExtractPath string            `yaml:"extract_path,omitempty"` // 提取路径
	Metadata    map[string]string `yaml:"metadata,omitempty"`     // 额外元数据
}

// FrameworkFingerprint 框架指纹定义 (增强版)
type FrameworkFingerprint struct {
	Name           string                `yaml:"name"`
	Type           string                `yaml:"type"`
	Description    string                `yaml:"description"`
	Version        string                `yaml:"version,omitempty"`        // 指纹版本
	Endpoints      []EnhancedEndpoint    `yaml:"endpoints"`
	Headers        []EnhancedHeaderPattern `yaml:"headers"`
	BodyPatterns   []EnhancedBodyPattern   `yaml:"body_patterns"`
	ErrorPatterns  []ErrorPattern          `yaml:"error_patterns"`
	Versions       []EnhancedVersionInfo   `yaml:"versions"`
	Deployment     DeploymentInfo          `yaml:"deployment"`
	Scoring        ScoringConfig           `yaml:"scoring,omitempty"`        // 评分配置
}

// Endpoint API端点 (兼容旧版本)
type Endpoint struct {
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	Description string `yaml:"description"`
}

// ErrorPattern 错误模式
type ErrorPattern struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

// VersionInfo 版本信息 (兼容旧版本)
type VersionInfo struct {
	Pattern  string   `yaml:"pattern"`
	Features []string `yaml:"features"`
}

// DeploymentInfo 部署信息
type DeploymentInfo struct {
	DefaultPort string `yaml:"default_port"`
	DockerImage string `yaml:"docker_image,omitempty"`
	GPURequired bool   `yaml:"gpu_required,omitempty"`
	ProcessName string `yaml:"process_name,omitempty"`
}

// Loader 指纹加载器
type Loader struct {
	models     map[string]*ModelFingerprint
	frameworks map[string]*FrameworkFingerprint
}

// NewLoader 创建指纹加载器
func NewLoader() *Loader {
	return &Loader{
		models:     make(map[string]*ModelFingerprint),
		frameworks: make(map[string]*FrameworkFingerprint),
	}
}

// LoadAll 加载所有指纹
func (l *Loader) LoadAll(baseDir string) error {
	// 加载模型指纹 (models目录和子目录)
	if err := l.loadModels(filepath.Join(baseDir, "models")); err != nil {
		return fmt.Errorf("failed to load models: %w", err)
	}

	// 加载中国厂商模型指纹
	if err := l.loadModels(filepath.Join(baseDir, "china")); err != nil {
		// 非致命错误，继续
	}

	// 加载云厂商模型指纹
	if err := l.loadModels(filepath.Join(baseDir, "cloud")); err != nil {
		// 非致命错误，继续
	}

	// 加载框架指纹
	if err := l.loadFrameworks(filepath.Join(baseDir, "frameworks")); err != nil {
		return fmt.Errorf("failed to load frameworks: %w", err)
	}

	// 加载行业指纹
	if err := l.loadFrameworks(filepath.Join(baseDir, "industry")); err != nil {
		// 非致命错误，继续
	}

	// 加载组件指纹
	if err := l.loadFrameworks(filepath.Join(baseDir, "components")); err != nil {
		// 非致命错误，继续
	}

	// 加载部署指纹
	if err := l.loadFrameworks(filepath.Join(baseDir, "deploy")); err != nil {
		// 非致命错误，继续
	}

	// 加载开发工具指纹
	if err := l.loadFrameworks(filepath.Join(baseDir, "devtools")); err != nil {
		// 非致命错误，继续
	}

	return nil
}

func (l *Loader) loadModels(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var fp ModelFingerprint
		if err := yaml.Unmarshal(data, &fp); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// 设置默认评分配置
		if fp.Scoring.Threshold == 0 {
			fp.Scoring = DefaultScoringConfig()
		}

		l.models[fp.Name] = &fp
		return nil
	})
}

func (l *Loader) loadFrameworks(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var fp FrameworkFingerprint
		if err := yaml.Unmarshal(data, &fp); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// 设置默认评分配置
		if fp.Scoring.Threshold == 0 {
			fp.Scoring = DefaultScoringConfig()
		}

		l.frameworks[fp.Name] = &fp
		return nil
	})
}

// GetModel 获取模型指纹
func (l *Loader) GetModel(name string) (*ModelFingerprint, bool) {
	fp, ok := l.models[name]
	return fp, ok
}

// GetFramework 获取框架指纹
func (l *Loader) GetFramework(name string) (*FrameworkFingerprint, bool) {
	fp, ok := l.frameworks[name]
	return fp, ok
}

// AllModels 获取所有模型指纹
func (l *Loader) AllModels() []*ModelFingerprint {
	var models []*ModelFingerprint
	for _, fp := range l.models {
		models = append(models, fp)
	}
	return models
}

// AllFrameworks 获取所有框架指纹
func (l *Loader) AllFrameworks() []*FrameworkFingerprint {
	var frameworks []*FrameworkFingerprint
	for _, fp := range l.frameworks {
		frameworks = append(frameworks, fp)
	}
	return frameworks
}

// ModelCount 返回模型数量
func (l *Loader) ModelCount() int {
	return len(l.models)
}

// FrameworkCount 返回框架数量
func (l *Loader) FrameworkCount() int {
	return len(l.frameworks)
}

// GetModelsByProvider 按提供商获取模型
func (l *Loader) GetModelsByProvider(provider string) []*ModelFingerprint {
	var result []*ModelFingerprint
	for _, fp := range l.models {
		if strings.EqualFold(fp.Provider, provider) {
			result = append(result, fp)
		}
	}
	return result
}

// GetFrameworksByType 按类型获取框架
func (l *Loader) GetFrameworksByType(frameworkType string) []*FrameworkFingerprint {
	var result []*FrameworkFingerprint
	for _, fp := range l.frameworks {
		if strings.EqualFold(fp.Type, frameworkType) {
			result = append(result, fp)
		}
	}
	return result
}

// ValidateFingerprints 验证所有指纹的有效性
func (l *Loader) ValidateFingerprints() []error {
	var errors []error

	// 验证模型指纹
	for name, fp := range l.models {
		if fp.Name == "" {
			errors = append(errors, fmt.Errorf("model fingerprint missing name"))
		}
		if fp.Provider == "" {
			errors = append(errors, fmt.Errorf("model %s: missing provider", name))
		}
		// 验证评分配置
		if err := validateScoringConfig(fp.Scoring); err != nil {
			errors = append(errors, fmt.Errorf("model %s: %w", name, err))
		}
	}

	// 验证框架指纹
	for name, fp := range l.frameworks {
		if fp.Name == "" {
			errors = append(errors, fmt.Errorf("framework fingerprint missing name"))
		}
		if fp.Type == "" {
			errors = append(errors, fmt.Errorf("framework %s: missing type", name))
		}
		// 验证评分配置
		if err := validateScoringConfig(fp.Scoring); err != nil {
			errors = append(errors, fmt.Errorf("framework %s: %w", name, err))
		}
	}

	return errors
}

// validateScoringConfig 验证评分配置
func validateScoringConfig(config ScoringConfig) error {
	total := config.HeaderMatch + config.BodyKeywords + config.JSONStructure
	if config.HeaderMatch > 0 || config.BodyKeywords > 0 || config.JSONStructure > 0 {
		if total < 0.99 || total > 1.01 {
			return fmt.Errorf("scoring weights must sum to 1.0, got %.2f", total)
		}
	}
	if config.Threshold < 0 || config.Threshold > 1 {
		return fmt.Errorf("threshold must be between 0 and 1, got %.2f", config.Threshold)
	}
	return nil
}
