package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// 原始结构定义 (用于读取旧格式)
type OldFrameworkFingerprint struct {
	Name          string              `yaml:"name"`
	Type          string              `yaml:"type"`
	Description   string              `yaml:"description"`
	Endpoints     []OldEndpoint       `yaml:"endpoints"`
	Headers       []OldHeaderPattern  `yaml:"headers"`
	BodyPatterns  []OldBodyPattern    `yaml:"body_patterns"`
	ErrorPatterns []OldErrorPattern   `yaml:"error_patterns"`
	Versions      []OldVersionInfo    `yaml:"versions"`
	Deployment    OldDeploymentInfo   `yaml:"deployment"`
	// 可能存在的额外字段
	Indicators    []string            `yaml:"indicators,omitempty"`
	Ports         []int               `yaml:"ports,omitempty"`
}

type OldEndpoint struct {
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	Description string `yaml:"description"`
}

type OldHeaderPattern struct {
	Name     string `yaml:"name"`
	Pattern  string `yaml:"pattern,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Required bool   `yaml:"required"`
}

type OldBodyPattern struct {
	Field    string `yaml:"field"`
	Pattern  string `yaml:"pattern,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Required bool   `yaml:"required"`
}

type OldErrorPattern struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

type OldVersionInfo struct {
	Pattern  string   `yaml:"pattern"`
	Features []string `yaml:"features"`
}

type OldDeploymentInfo struct {
	DefaultPort  string `yaml:"default_port"`
	DockerImage  string `yaml:"docker_image,omitempty"`
	GPURequired  bool   `yaml:"gpu_required,omitempty"`
	ProcessName  string `yaml:"process_name,omitempty"`
	ServiceName  string `yaml:"service_name,omitempty"`
	RegionBased  bool   `yaml:"region_based,omitempty"`
}

// 增强结构定义 (用于写入新格式)
type EnhancedFrameworkFingerprint struct {
	Name           string                   `yaml:"name"`
	Type           string                   `yaml:"type"`
	Description    string                   `yaml:"description"`
	Version        string                   `yaml:"version,omitempty"`
	Endpoints      []EnhancedEndpoint       `yaml:"endpoints"`
	Headers        []EnhancedHeaderPattern  `yaml:"headers"`
	BodyPatterns   []EnhancedBodyPattern    `yaml:"body_patterns"`
	ErrorPatterns  []EnhancedErrorPattern   `yaml:"error_patterns"`
	Versions       []EnhancedVersionInfo    `yaml:"versions"`
	Deployment     EnhancedDeploymentInfo   `yaml:"deployment"`
	Scoring        ScoringConfig            `yaml:"scoring"`
}

type EnhancedEndpoint struct {
	Path        string   `yaml:"path"`
	Method      string   `yaml:"method"`
	Description string   `yaml:"description"`
	Weight      float64  `yaml:"weight,omitempty"`
	Patterns    []string `yaml:"patterns,omitempty"`
}

type EnhancedHeaderPattern struct {
	Name        string  `yaml:"name"`
	Pattern     string  `yaml:"pattern,omitempty"`
	Value       string  `yaml:"value,omitempty"`
	Required    bool    `yaml:"required"`
	Weight      float64 `yaml:"weight,omitempty"`
	Extract     string  `yaml:"extract,omitempty"`
	Description string  `yaml:"description,omitempty"`
}

type EnhancedBodyPattern struct {
	Field       string   `yaml:"field,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
	Value       string   `yaml:"value,omitempty"`
	JSONPath    string   `yaml:"json_path,omitempty"`
	Required    bool     `yaml:"required"`
	Weight      float64  `yaml:"weight,omitempty"`
	Keywords    []string `yaml:"keywords,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type EnhancedErrorPattern struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
	Weight  float64 `yaml:"weight,omitempty"`
}

type EnhancedVersionInfo struct {
	Pattern     string            `yaml:"pattern"`
	Features    []string          `yaml:"features"`
	ExtractPath string            `yaml:"extract_path,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

type EnhancedDeploymentInfo struct {
	DefaultPort  string `yaml:"default_port"`
	DockerImage  string `yaml:"docker_image,omitempty"`
	GPURequired  bool   `yaml:"gpu_required,omitempty"`
	ProcessName  string `yaml:"process_name,omitempty"`
	ServiceName  string `yaml:"service_name,omitempty"`
	RegionBased  bool   `yaml:"region_based,omitempty"`
}

type ScoringConfig struct {
	HeaderMatch   float64 `yaml:"header_match"`
	BodyKeywords  float64 `yaml:"body_keywords"`
	JSONStructure float64 `yaml:"json_structure"`
	Threshold     float64 `yaml:"threshold"`
}

// 模型指纹结构
type OldModelFingerprint struct {
	Name         string              `yaml:"name"`
	Provider     string              `yaml:"provider"`
	Type         string              `yaml:"type"`
	Description  string              `yaml:"description"`
	Response     OldModelResponse    `yaml:"response"`
	Behavior     OldModelBehavior    `yaml:"behavior"`
	Tests        map[string]float64  `yaml:"tests"`
	Fingerprints []OldTestFingerprint `yaml:"fingerprints"`
	Variants     []OldModelVariant    `yaml:"variants"`
}

type OldModelResponse struct {
	Headers      []OldHeaderPattern `yaml:"headers"`
	BodyPatterns []OldBodyPattern   `yaml:"body_patterns"`
}

type OldModelBehavior struct {
	MaxTokens               int    `yaml:"max_tokens"`
	SupportsFunctionCalling bool   `yaml:"supports_function_calling"`
	SupportsVision          bool   `yaml:"supports_vision"`
	SupportsJSONMode        bool   `yaml:"supports_json_mode"`
	ResponseTimeMs          string `yaml:"response_time_ms"`
	StreamingSupported      bool   `yaml:"streaming_supported"`
}

type OldTestFingerprint struct {
	Name              string   `yaml:"test"`
	Prompt            string   `yaml:"prompt"`
	ExpectedKeywords  []string `yaml:"expected_keywords,omitempty"`
	ExpectedPatterns  []string `yaml:"expected_patterns,omitempty"`
	ForbiddenKeywords []string `yaml:"forbidden_keywords,omitempty"`
	ForbiddenPatterns []string `yaml:"forbidden_patterns,omitempty"`
	Weight            float64  `yaml:"weight"`
}

type OldModelVariant struct {
	Name     string   `yaml:"name"`
	Pattern  string   `yaml:"pattern"`
	Features []string `yaml:"features"`
}

type EnhancedModelFingerprint struct {
	Name         string                   `yaml:"name"`
	Provider     string                   `yaml:"provider"`
	Type         string                   `yaml:"type"`
	Description  string                   `yaml:"description"`
	Version      string                   `yaml:"version,omitempty"`
	Response     EnhancedModelResponse    `yaml:"response"`
	Behavior     OldModelBehavior         `yaml:"behavior"`
	Tests        map[string]float64       `yaml:"tests"`
	Fingerprints []EnhancedTestFingerprint `yaml:"fingerprints"`
	Variants     []OldModelVariant         `yaml:"variants"`
	Scoring      ScoringConfig             `yaml:"scoring"`
}

type EnhancedModelResponse struct {
	Headers      []EnhancedHeaderPattern `yaml:"headers"`
	BodyPatterns []EnhancedBodyPattern   `yaml:"body_patterns"`
}

type EnhancedTestFingerprint struct {
	Name              string   `yaml:"test"`
	Prompt            string   `yaml:"prompt"`
	ExpectedKeywords  []string `yaml:"expected_keywords,omitempty"`
	ExpectedPatterns  []string `yaml:"expected_patterns,omitempty"`
	ForbiddenKeywords []string `yaml:"forbidden_keywords,omitempty"`
	ForbiddenPatterns []string `yaml:"forbidden_patterns,omitempty"`
	Weight            float64  `yaml:"weight"`
	JSONPath          string   `yaml:"json_path,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: upgrade_fingerprints <base_dir>")
		os.Exit(1)
	}

	baseDir := os.Args[1]
	
	upgradedCount := 0
	errors := []string{}

	// 升级框架指纹
	frameworksDir := filepath.Join(baseDir, "frameworks")
	if _, err := os.Stat(frameworksDir); err == nil {
		count, errs := upgradeFrameworks(frameworksDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级模型指纹
	modelsDir := filepath.Join(baseDir, "models")
	if _, err := os.Stat(modelsDir); err == nil {
		count, errs := upgradeModels(modelsDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级行业指纹
	industryDir := filepath.Join(baseDir, "industry")
	if _, err := os.Stat(industryDir); err == nil {
		count, errs := upgradeFrameworksInDir(industryDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级中国厂商指纹
	chinaDir := filepath.Join(baseDir, "china")
	if _, err := os.Stat(chinaDir); err == nil {
		count, errs := upgradeModelsInDir(chinaDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级云厂商指纹
	cloudDir := filepath.Join(baseDir, "cloud")
	if _, err := os.Stat(cloudDir); err == nil {
		count, errs := upgradeModelsInDir(cloudDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级组件指纹
	componentsDir := filepath.Join(baseDir, "components")
	if _, err := os.Stat(componentsDir); err == nil {
		count, errs := upgradeFrameworksInDir(componentsDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级部署指纹
	deployDir := filepath.Join(baseDir, "deploy")
	if _, err := os.Stat(deployDir); err == nil {
		count, errs := upgradeFrameworksInDir(deployDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	// 升级开发工具指纹
	devtoolsDir := filepath.Join(baseDir, "devtools")
	if _, err := os.Stat(devtoolsDir); err == nil {
		count, errs := upgradeFrameworksInDir(devtoolsDir)
		upgradedCount += count
		errors = append(errors, errs...)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("升级完成!\n")
	fmt.Printf("升级文件数: %d\n", upgradedCount)
	fmt.Printf("错误数: %d\n", len(errors))
	
	if len(errors) > 0 {
		fmt.Printf("\n错误详情:\n")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}
}

func upgradeFrameworks(dir string) (int, []string) {
	return upgradeFrameworksInDir(dir)
}

func upgradeFrameworksInDir(dir string) (int, []string) {
	upgradedCount := 0
	var errors []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("读取失败 %s: %v", path, err))
			return nil
		}

		var oldFp OldFrameworkFingerprint
		if err := yaml.Unmarshal(data, &oldFp); err != nil {
			errors = append(errors, fmt.Sprintf("解析失败 %s: %v", path, err))
			return nil
		}

		// 转换为增强格式
		enhancedFp := convertFrameworkFingerprint(oldFp)

		// 写入文件
		output, err := yaml.Marshal(enhancedFp)
		if err != nil {
			errors = append(errors, fmt.Sprintf("序列化失败 %s: %v", path, err))
			return nil
		}

		if err := os.WriteFile(path, output, 0644); err != nil {
			errors = append(errors, fmt.Sprintf("写入失败 %s: %v", path, err))
			return nil
		}

		upgradedCount++
		fmt.Printf("[✓] 升级: %s\n", path)
		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("遍历目录失败 %s: %v", dir, err))
	}

	return upgradedCount, errors
}

func convertFrameworkFingerprint(old OldFrameworkFingerprint) EnhancedFrameworkFingerprint {
	// 转换端点
	var endpoints []EnhancedEndpoint
	for _, ep := range old.Endpoints {
		endpoints = append(endpoints, EnhancedEndpoint{
			Path:        ep.Path,
			Method:      ep.Method,
			Description: ep.Description,
			Weight:      0.2, // 默认权重
		})
	}

	// 转换响应头
	var headers []EnhancedHeaderPattern
	for _, h := range old.Headers {
		headers = append(headers, EnhancedHeaderPattern{
			Name:        h.Name,
			Pattern:     h.Pattern,
			Value:       h.Value,
			Required:    h.Required,
			Weight:      0.25,
			Description: "",
		})
	}

	// 转换响应体模式
	var bodyPatterns []EnhancedBodyPattern
	for _, bp := range old.BodyPatterns {
		ebp := EnhancedBodyPattern{
			Field:       bp.Field,
			Pattern:     bp.Pattern,
			Value:       bp.Value,
			Required:    bp.Required,
			Weight:      0.25,
			Description: "",
		}
		// 如果有type字段，转换为JSONPath
		if bp.Type != "" {
			ebp.JSONPath = bp.Field
		}
		bodyPatterns = append(bodyPatterns, ebp)
	}

	// 转换错误模式
	var errorPatterns []EnhancedErrorPattern
	for _, ep := range old.ErrorPatterns {
		errorPatterns = append(errorPatterns, EnhancedErrorPattern{
			Pattern: ep.Pattern,
			Type:    ep.Type,
			Weight:  0.15,
		})
	}

	// 转换版本信息
	var versions []EnhancedVersionInfo
	for _, v := range old.Versions {
		versions = append(versions, EnhancedVersionInfo{
			Pattern:     v.Pattern,
			Features:    v.Features,
			ExtractPath: "",
			Metadata:    make(map[string]string),
		})
	}

	return EnhancedFrameworkFingerprint{
		Name:          old.Name,
		Type:          old.Type,
		Description:   old.Description,
		Version:       "1.0",
		Endpoints:     endpoints,
		Headers:       headers,
		BodyPatterns:  bodyPatterns,
		ErrorPatterns: errorPatterns,
		Versions:      versions,
		Deployment: EnhancedDeploymentInfo{
			DefaultPort: old.Deployment.DefaultPort,
			DockerImage: old.Deployment.DockerImage,
			GPURequired: old.Deployment.GPURequired,
			ProcessName: old.Deployment.ProcessName,
			ServiceName: old.Deployment.ServiceName,
			RegionBased: old.Deployment.RegionBased,
		},
		Scoring: ScoringConfig{
			HeaderMatch:   0.30,
			BodyKeywords:  0.40,
			JSONStructure: 0.30,
			Threshold:     0.70,
		},
	}
}

func upgradeModels(dir string) (int, []string) {
	return upgradeModelsInDir(dir)
}

func upgradeModelsInDir(dir string) (int, []string) {
	upgradedCount := 0
	var errors []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("读取失败 %s: %v", path, err))
			return nil
		}

		var oldFp OldModelFingerprint
		if err := yaml.Unmarshal(data, &oldFp); err != nil {
			errors = append(errors, fmt.Sprintf("解析失败 %s: %v", path, err))
			return nil
		}

		// 转换为增强格式
		enhancedFp := convertModelFingerprint(oldFp)

		// 写入文件
		output, err := yaml.Marshal(enhancedFp)
		if err != nil {
			errors = append(errors, fmt.Sprintf("序列化失败 %s: %v", path, err))
			return nil
		}

		if err := os.WriteFile(path, output, 0644); err != nil {
			errors = append(errors, fmt.Sprintf("写入失败 %s: %v", path, err))
			return nil
		}

		upgradedCount++
		fmt.Printf("[✓] 升级: %s\n", path)
		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("遍历目录失败 %s: %v", dir, err))
	}

	return upgradedCount, errors
}

func convertModelFingerprint(old OldModelFingerprint) EnhancedModelFingerprint {
	// 转换响应头
	var headers []EnhancedHeaderPattern
	for _, h := range old.Response.Headers {
		headers = append(headers, EnhancedHeaderPattern{
			Name:     h.Name,
			Pattern:  h.Pattern,
			Value:    h.Value,
			Required: h.Required,
			Weight:   0.25,
		})
	}

	// 转换响应体模式
	var bodyPatterns []EnhancedBodyPattern
	for _, bp := range old.Response.BodyPatterns {
		bodyPatterns = append(bodyPatterns, EnhancedBodyPattern{
			Field:    bp.Field,
			Pattern:  bp.Pattern,
			Value:    bp.Value,
			Required: bp.Required,
			Weight:   0.25,
		})
	}

	// 转换测试指纹
	var fingerprints []EnhancedTestFingerprint
	for _, tf := range old.Fingerprints {
		fingerprints = append(fingerprints, EnhancedTestFingerprint{
			Name:              tf.Name,
			Prompt:            tf.Prompt,
			ExpectedKeywords:  tf.ExpectedKeywords,
			ExpectedPatterns:  tf.ExpectedPatterns,
			ForbiddenKeywords: tf.ForbiddenKeywords,
			ForbiddenPatterns: tf.ForbiddenPatterns,
			Weight:            tf.Weight,
		})
	}

	return EnhancedModelFingerprint{
		Name:        old.Name,
		Provider:    old.Provider,
		Type:        old.Type,
		Description: old.Description,
		Version:     "1.0",
		Response: EnhancedModelResponse{
			Headers:      headers,
			BodyPatterns: bodyPatterns,
		},
		Behavior:     old.Behavior,
		Tests:        old.Tests,
		Fingerprints: fingerprints,
		Variants:     old.Variants,
		Scoring: ScoringConfig{
			HeaderMatch:   0.30,
			BodyKeywords:  0.40,
			JSONStructure: 0.30,
			Threshold:     0.70,
		},
	}
}
