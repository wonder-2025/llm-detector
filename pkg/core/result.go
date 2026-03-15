package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// DetectionResult 探测结果
type DetectionResult struct {
	Target      string                 `json:"target"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Mode        string                 `json:"mode,omitempty"`        // 评分模式
	Threshold   float64                `json:"threshold,omitempty"`   // 置信度阈值
	APIResults  []APIResult            `json:"api_results"`
	ModelGuess  *ModelGuess            `json:"model_guess"`
	ServiceInfo *ServiceInfo           `json:"service_info"`
	Raw         map[string]interface{} `json:"raw,omitempty"`
}

// APIResult API探测结果
type APIResult struct {
	Type       string            `json:"type"`
	Endpoint   string            `json:"endpoint"`
	Available  bool              `json:"available"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body,omitempty"`
	Error      string            `json:"error,omitempty"`
	Confidence float64           `json:"confidence,omitempty"`
}

// ModelGuess 模型猜测结果
type ModelGuess struct {
	Name           string             `json:"name"`
	Provider       string             `json:"provider"`
	Type           string             `json:"type"`
	Confidence     float64            `json:"confidence"`
	Version        string             `json:"version,omitempty"`        // 检测到的版本
	Features       []string           `json:"features"`
	Alternative    []AlternativeModel `json:"alternative,omitempty"`
	ScoringDetails *ScoringDetails    `json:"scoring_details,omitempty"` // 评分详情
}

// AlternativeModel 备选模型
type AlternativeModel struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Framework     string            `json:"framework,omitempty"`
	Version       string            `json:"version,omitempty"`
	Deployment    string            `json:"deployment,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Confidence    float64           `json:"confidence,omitempty"`
	ScoringDetails *ScoringDetails  `json:"scoring_details,omitempty"`
}

// String 格式化输出结果
func (r *DetectionResult) String() string {
	var output string
	
	output += fmt.Sprintf("\n[+] Target: %s\n", r.Target)
	output += fmt.Sprintf("[+] Detection Time: %s\n", r.Timestamp.Format("2006-01-02 15:04:05"))
	output += fmt.Sprintf("[+] Duration: %v\n", r.Duration)
	
	// 显示评分模式
	if r.Mode != "" {
		output += fmt.Sprintf("[+] Mode: %s\n", r.Mode)
	}
	if r.Threshold > 0 {
		output += fmt.Sprintf("[+] Threshold: %.0f%%\n", r.Threshold*100)
	}
	output += "\n"

	// API检测结果
	if len(r.APIResults) > 0 {
		output += "[*] API Detection:\n"
		for _, api := range r.APIResults {
			if api.Available {
				if api.Confidence > 0 {
					output += fmt.Sprintf("    [+] %s: %s (HTTP %d, Confidence: %.0f%%)\n", api.Type, api.Endpoint, api.StatusCode, api.Confidence*100)
				} else {
					output += fmt.Sprintf("    [+] %s: %s (HTTP %d)\n", api.Type, api.Endpoint, api.StatusCode)
				}
			} else {
				output += fmt.Sprintf("    [-] %s: %s (Error: %s)\n", api.Type, api.Endpoint, api.Error)
			}
		}
		output += "\n"
	}

	// 模型识别结果
	if r.ModelGuess != nil {
		output += "[*] Model Fingerprinting:\n"
		output += fmt.Sprintf("    [+] Detected: %s (Confidence: %.1f%%)\n", r.ModelGuess.Name, r.ModelGuess.Confidence*100)
		output += fmt.Sprintf("    [+] Provider: %s\n", r.ModelGuess.Provider)
		output += fmt.Sprintf("    [+] Type: %s\n", r.ModelGuess.Type)
		
		if r.ModelGuess.Version != "" {
			output += fmt.Sprintf("    [+] Version: %s\n", r.ModelGuess.Version)
		}
		
		if r.ModelGuess.ScoringDetails != nil {
			output += "    [+] Scoring Breakdown:\n"
			d := r.ModelGuess.ScoringDetails
			output += fmt.Sprintf("        - Headers:   %.0f%% (weight: %.0f%%)\n", d.HeaderScore*100, d.HeaderWeight*100)
			output += fmt.Sprintf("        - Body:      %.0f%% (weight: %.0f%%)\n", d.BodyScore*100, d.BodyWeight*100)
			output += fmt.Sprintf("        - JSON:      %.0f%% (weight: %.0f%%)\n", d.JSONScore*100, d.JSONWeight*100)
		}
		
		if len(r.ModelGuess.Features) > 0 {
			output += "    [+] Features:\n"
			for _, f := range r.ModelGuess.Features {
				output += fmt.Sprintf("        - %s\n", f)
			}
		}
		if len(r.ModelGuess.Alternative) > 0 {
			output += "    [*] Alternative models:\n"
			for _, alt := range r.ModelGuess.Alternative {
				output += fmt.Sprintf("        - %s (%.1f%%)\n", alt.Name, alt.Confidence*100)
			}
		}
		output += "\n"
	}

	// 服务信息
	if r.ServiceInfo != nil {
		output += "[*] Service Information:\n"
		if r.ServiceInfo.Framework != "" {
			output += fmt.Sprintf("    [+] Framework: %s", r.ServiceInfo.Framework)
			if r.ServiceInfo.Confidence > 0 {
				output += fmt.Sprintf(" (Confidence: %.1f%%)", r.ServiceInfo.Confidence*100)
			}
			output += "\n"
		}
		if r.ServiceInfo.Version != "" {
			output += fmt.Sprintf("    [+] Version: %s\n", r.ServiceInfo.Version)
		}
		if r.ServiceInfo.Deployment != "" {
			output += fmt.Sprintf("    [+] Deployment: %s\n", r.ServiceInfo.Deployment)
		}
		
		if r.ServiceInfo.ScoringDetails != nil {
			output += "    [+] Scoring Breakdown:\n"
			d := r.ServiceInfo.ScoringDetails
			output += fmt.Sprintf("        - Headers:   %.0f%% (weight: %.0f%%)\n", d.HeaderScore*100, d.HeaderWeight*100)
			output += fmt.Sprintf("        - Body:      %.0f%% (weight: %.0f%%)\n", d.BodyScore*100, d.BodyWeight*100)
			output += fmt.Sprintf("        - JSON:      %.0f%% (weight: %.0f%%)\n", d.JSONScore*100, d.JSONWeight*100)
		}
		output += "\n"
	}

	// 汇总 - 显示识别到的组件
	output += "[+] Results Summary:\n"
	
	// 收集识别到的组件
	var detectedComponents []string
	for _, api := range r.APIResults {
		if api.Available {
			detectedComponents = append(detectedComponents, api.Type)
		}
	}
	
	if len(detectedComponents) > 0 {
		output += "    Detected Components:\n"
		for _, comp := range detectedComponents {
			output += fmt.Sprintf("      - %s\n", comp)
		}
	} else {
		output += "    Detected Components: None\n"
	}
	
	if r.ServiceInfo != nil && r.ServiceInfo.Framework != "" {
		output += fmt.Sprintf("    Framework: %s\n", r.ServiceInfo.Framework)
	}
	if r.ModelGuess != nil && r.ModelGuess.Name != "" {
		output += fmt.Sprintf("    Model: %s\n", r.ModelGuess.Name)
	}

	return output
}

// StringVerbose 详细格式化输出结果
func (r *DetectionResult) StringVerbose() string {
	output := r.String()
	
	// 添加原始数据（如果存在）
	if r.Raw != nil && len(r.Raw) > 0 {
		output += "\n[*] Raw Data:\n"
		rawJSON, err := json.MarshalIndent(r.Raw, "    ", "  ")
		if err == nil {
			output += "    " + string(rawJSON) + "\n"
		}
	}
	
	return output
}

// JSON 返回JSON格式结果
func (r *DetectionResult) JSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CompactJSON 返回紧凑JSON格式
func (r *DetectionResult) CompactJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsDetected 是否检测到任何内容
func (r *DetectionResult) IsDetected() bool {
	for _, api := range r.APIResults {
		if api.Available {
			return true
		}
	}
	// 检查模型识别
	if r.ModelGuess != nil {
		return true
	}
	// 检查服务框架
	if r.ServiceInfo != nil && r.ServiceInfo.Framework != "" {
		return true
	}
	// 检查API组件检测
	for _, api := range r.APIResults {
		if api.Available {
			return true
		}
	}
	return false
}

// GetConfidence 获取最高置信度
func (r *DetectionResult) GetConfidence() float64 {
	maxConfidence := 0.0
	
	if r.ModelGuess != nil && r.ModelGuess.Confidence > maxConfidence {
		maxConfidence = r.ModelGuess.Confidence
	}
	
	if r.ServiceInfo != nil && r.ServiceInfo.Confidence > maxConfidence {
		maxConfidence = r.ServiceInfo.Confidence
	}
	
	return maxConfidence
}
