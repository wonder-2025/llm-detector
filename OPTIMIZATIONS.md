# LLM Detector 优化记录

## 已完成的优化

### 1. 网络优化
- **TLS支持**: 所有插件支持HTTPS，自动跳过证书验证
- **重定向支持**: Jupyter/Attu插件支持HTTP重定向
- **统一HTTP客户端**: `pkg/plugins/api/client.go` 统一配置

### 2. 端口解析优化
- **修复Bug**: URL类型目标不再丢失非标准端口
- **支持格式**: IP, IP:Port, http://, https://

### 3. 探测插件优化
- **9个API插件**: OpenAI, Ollama, vLLM, TGI, FastAPI, LiteLLM, Jupyter, Attu, Generic
- **智能识别**: 基于响应头、HTML内容、API特征

### 4. 性能优化
- **并发探测**: 多个插件并行执行
- **超时控制**: 默认10秒，可配置
- **快速失败**: 连接失败立即返回

## 使用建议

```bash
# 基础使用
./llm-detector -t target:port

# 详细输出
./llm-detector -t target:port -v

# 增加超时
./llm-detector -t target:port --timeout 30

# HTTPS目标
./llm-detector -t https://target.com
```

## 故障排查

| 问题 | 解决 |
|------|------|
| 全部Error | 检查网络/防火墙/代理 |
| 超时 | 增加--timeout参数 |
| 无权限 | chmod +x llm-detector |
