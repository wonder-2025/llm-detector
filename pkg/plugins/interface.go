package plugins

import (
	"context"
)

// Target 探测目标 (从core复制，避免循环依赖)
type Target interface {
	String() string
	BaseURL() string
}

// APIResult API探测结果 (从core复制，避免循环依赖)
type APIResult struct {
	Type       string
	Endpoint   string
	Available  bool
	StatusCode int
	Headers    map[string]string
	Body       string
	Error      string
}

// ModelGuess 模型猜测结果 (从core复制，避免循环依赖)
type ModelGuess struct {
	Name        string
	Provider    string
	Type        string
	Confidence  float64
	Features    []string
	Alternative []AlternativeModel
}

// AlternativeModel 备选模型
type AlternativeModel struct {
	Name       string
	Confidence float64
}

// ServiceInfo 服务信息 (从core复制，避免循环依赖)
type ServiceInfo struct {
	Framework  string
	Version    string
	Deployment string
	Headers    map[string]string
}

// Plugin 插件接口
type Plugin interface {
	Name() string
	Version() string
	Detect(ctx context.Context, target Target) (*APIResult, error)
}

// ModelFingerprinter 模型指纹插件接口
type ModelFingerprinter interface {
	Plugin
	Fingerprint(ctx context.Context, target Target, apiResult *APIResult) (*ModelGuess, error)
}

// ServiceFingerprinter 服务指纹插件接口
type ServiceFingerprinter interface {
	Plugin
	FingerprintService(ctx context.Context, target Target, apiResults []*APIResult) (*ServiceInfo, error)
}

// Registry 插件注册中心
type Registry struct {
	apiPlugins     map[string]Plugin
	modelPlugins   map[string]ModelFingerprinter
	servicePlugins map[string]ServiceFingerprinter
}

// NewRegistry 创建插件注册中心
func NewRegistry() *Registry {
	return &Registry{
		apiPlugins:     make(map[string]Plugin),
		modelPlugins:   make(map[string]ModelFingerprinter),
		servicePlugins: make(map[string]ServiceFingerprinter),
	}
}

// RegisterAPI 注册API探测插件
func (r *Registry) RegisterAPI(p Plugin) {
	r.apiPlugins[p.Name()] = p
}

// RegisterModel 注册模型指纹插件
func (r *Registry) RegisterModel(p ModelFingerprinter) {
	r.modelPlugins[p.Name()] = p
}

// RegisterService 注册服务指纹插件
func (r *Registry) RegisterService(p ServiceFingerprinter) {
	r.servicePlugins[p.Name()] = p
}

// GetAPI 获取API探测插件
func (r *Registry) GetAPI(name string) (Plugin, bool) {
	p, ok := r.apiPlugins[name]
	return p, ok
}

// GetModel 获取模型指纹插件
func (r *Registry) GetModel(name string) (ModelFingerprinter, bool) {
	p, ok := r.modelPlugins[name]
	return p, ok
}

// GetService 获取服务指纹插件
func (r *Registry) GetService(name string) (ServiceFingerprinter, bool) {
	p, ok := r.servicePlugins[name]
	return p, ok
}

// AllAPIs 获取所有API探测插件
func (r *Registry) AllAPIs() []Plugin {
	var plugins []Plugin
	for _, p := range r.apiPlugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// AllModels 获取所有模型指纹插件
func (r *Registry) AllModels() []ModelFingerprinter {
	var plugins []ModelFingerprinter
	for _, p := range r.modelPlugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// AllServices 获取所有服务指纹插件
func (r *Registry) AllServices() []ServiceFingerprinter {
	var plugins []ServiceFingerprinter
	for _, p := range r.servicePlugins {
		plugins = append(plugins, p)
	}
	return plugins
}
