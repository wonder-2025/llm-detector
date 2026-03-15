# LLM Detector

🔍 大模型组件识别工具 - 用于安全评估和资产发现

## 功能特性

- **三层探测架构**: 标准API探测 → 主动扫描 → 智能识别
- **103+ 组件指纹**: 覆盖主流LLM框架、模型、工具
- **智能端点发现**: 即使API自定义也能识别
- **多平台支持**: Linux/macOS/Windows

## 快速开始

### 安装

#### 方式1: 一键安装脚本
```bash
curl -sSL https://your-domain.com/install.sh | bash
```

#### 方式2: 下载预编译二进制
```bash
# Linux AMD64
wget https://github.com/yourusername/llm-detector/releases/latest/download/llm-detector-linux-amd64.tar.gz
tar -xzf llm-detector-linux-amd64.tar.gz
sudo cp llm-detector/llm-detector /usr/local/bin/
```

#### 方式3: Docker
```bash
docker pull llm-detector:latest
docker run --rm llm-detector -t target:port
```

#### 方式4: 源码编译
```bash
git clone https://github.com/yourusername/llm-detector.git
cd llm-detector
make build
sudo make install
```

## 使用方法

### 基本用法
```bash
# 探测单个目标
llm-detector -t localhost:11434

# 详细输出
llm-detector -t localhost:11434 -v

# JSON输出
llm-detector -t target:port -o json

# 指定端口扫描
llm-detector -t 192.168.1.100 -p 8000,8080,11434

# HTTPS目标 (自动跳过证书验证)
llm-detector -t https://target.com
```

### 目标格式
```bash
# 纯IP (自动扫描常见端口)
llm-detector -t 192.168.1.100

# IP:端口
llm-detector -t 192.168.1.100:8000

# HTTP URL
llm-detector -t http://target.com:8080/api

# HTTPS URL (支持自签名证书)
llm-detector -t https://target.com
```

## 支持的组件

### 推理框架 (10个)
- Ollama, vLLM, vLLM-Ascend, TGI, SGLang
- LiteLLM, MindIE, GPUStack, FastAPI

### 大模型 (7个)
- GPT-4, Claude 3, Llama 3, Qwen, Gemini, Mistral, DeepSeek

### 向量数据库 (7个)
- Chroma, Pinecone, Milvus, Qdrant, Weaviate, pgvector, **Attu**

### RAG系统 (5个)
- LangChain, LlamaIndex, Haystack, Semantic Kernel, Flowise

### LLM UI (7个)
- Dify, ComfyUI, OpenWebUI, Gradio, ChuanhuChatGPT, NextChat, H2O

### MLOps (4个)
- MLflow, Apache Airflow, ZenML, Kubeflow

### 开发框架 (10个)
- PyTorch, NPU-Torch, LLaMA-Factory, SWIFT, MindFormers
- LangChain, LlamaIndex, Axolotl, Unsloth

### 辅助开发 (5个)
- Jupyter, EvalScope, MindIEBench, OpenCompass, LLMPerf

### 数据库 (6个)
- ClickHouse, Apache HugeGraph, Neo4j, Redis, PostgreSQL, MongoDB

### AI Agent (6个)
- AutoGPT, CrewAI, AutoGen, MetaGPT, BabyAGI, SuperAGI

### 安全防护 (5个)
- NeMo Guardrails, Llama Guard, Azure Content Safety

### 可观测性 (6个)
- LangSmith, Langfuse, Weights & Biases, Helicone

### 编排部署 (7个)
- Kubernetes, Docker, Ray, BentoML, Triton, KServe, Seldon

### 提示词管理 (5个)
- PromptLayer, Pezzo, PromptFlow, Humanloop, Langtail

### 嵌入服务 (5个)
- OpenAI Embedding, Sentence Transformers, HuggingFace TEI

### Notebook (6个)
- JupyterHub, JupyterLab, Google Colab, Deepnote, Kaggle

### 模型注册 (5个)
- HuggingFace Hub, ModelScope, W&B Registry

### 管理界面 (1个)
- **Attu** (Milvus GUI)

**总计: 104+ 组件**

## 构建

```bash
# 当前平台
make build

# 所有平台
make build-all

# Docker镜像
make docker

# 安装到系统
sudo make install
```

## API探测插件

| 插件 | 说明 | 支持HTTPS |
|------|------|-----------|
| OpenAI | OpenAI API 兼容服务 | ✅ |
| Ollama | Ollama 本地模型服务 | ✅ |
| vLLM | vLLM 推理框架 | ✅ |
| TGI | HuggingFace TGI | ✅ |
| LiteLLM | LiteLLM API 网关 | ✅ |
| FastAPI | FastAPI 框架 | ✅ |
| Jupyter | Jupyter Notebook/Hub | ✅ |
| Attu | Milvus 管理界面 | ✅ |
| Generic | 通用端点探测 | ✅ |

## 项目结构

```
llm-detector/
├── cmd/detector/          # 主程序入口
├── pkg/
│   ├── core/              # 核心逻辑
│   │   ├── target.go      # 目标解析
│   │   ├── engine.go      # 探测引擎
│   │   ├── active_probe.go    # 主动探测
│   │   └── smart_probe.go     # 智能探测
│   ├── plugins/           # 探测插件
│   │   └── api/           # API探测插件
│   │       ├── openai.go
│   │       ├── ollama.go
│   │       ├── vllm.go
│   │       ├── jupyter.go
│   │       ├── attu.go
│   │       └── ...
│   └── fingerprints/      # 指纹数据
│       └── data/
│           ├── models/    # 模型指纹
│           ├── frameworks/# 框架指纹
│           └── components/# 组件指纹
├── scripts/
│   └── install.sh         # 安装脚本
├── Makefile
├── Dockerfile
└── README.md
```

## 安全声明

**本工具仅用于安全评估和资产发现，遵循以下原则：**

✅ **只读探测**
- 仅使用 GET 请求获取信息
- 不发送任何修改数据的请求
- 不消耗目标服务的计算资源

✅ **无破坏性操作**
- 不执行暴力破解
- 不尝试绕过认证
- 不造成拒绝服务

✅ **合规使用**
- 请确保您有权限扫描目标
- 遵守当地法律法规
- 仅用于授权的安全测试

## 许可证

MIT License

## 贡献

欢迎提交Issue和PR！
