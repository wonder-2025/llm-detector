#!/bin/bash
# 批量测试LLM Detector工具

cd /root/.openclaw/workspace/llm-detector

# 测试IP列表
declare -A TEST_IPS=(
    # Ollama (port=11434)
    ["47.109.224.190:11434"]="Ollama"
    ["1.71.80.136:11434"]="Ollama"
    ["120.79.23.130:11434"]="Ollama"

    # Jupyter (port=8888)
    ["52.82.123.100:8888"]="Jupyter"
    ["47.242.38.187:8888"]="Jupyter"
    ["8.210.28.164:8888"]="Jupyter"

    # MLflow (port=5000)
    ["47.242.38.187:5000"]="MLflow"
    ["8.210.28.164:5000"]="MLflow"
    ["52.82.123.100:5000"]="MLflow"

    # Airflow (port=8080)
    ["47.242.38.187:8080"]="Airflow"
    ["8.210.28.164:8080"]="Airflow"
    ["52.82.123.100:8080"]="Airflow"

    # Dify (port=3000)
    ["47.242.38.187:3000"]="Dify"
    ["8.210.28.164:3000"]="Dify"
    ["52.82.123.100:3000"]="Dify"

    # ComfyUI (port=8188)
    ["47.242.38.187:8188"]="ComfyUI"
    ["8.210.28.164:8188"]="ComfyUI"
    ["52.82.123.100:8188"]="ComfyUI"

    # ClickHouse (port=8123)
    ["47.242.38.187:8123"]="ClickHouse"
    ["8.210.28.164:8123"]="ClickHouse"
    ["52.82.123.100:8123"]="ClickHouse"

    # OpenWebUI (port=8080) - 与Airflow同端口，不同IP
    ["47.109.224.190:8080"]="OpenWebUI"
    ["1.71.80.136:8080"]="OpenWebUI"
    ["120.79.23.130:8080"]="OpenWebUI"

    # LiteLLM (port=4000)
    ["47.242.38.187:4000"]="LiteLLM"
    ["8.210.28.164:4000"]="LiteLLM"
    ["52.82.123.100:4000"]="LiteLLM"

    # Gradio (port=7860)
    ["47.242.38.187:7860"]="Gradio"
    ["8.210.28.164:7860"]="Gradio"
    ["52.82.123.100:7860"]="Gradio"
)

echo "开始批量测试LLM Detector工具..."
echo "================================"

for ip in "${!TEST_IPS[@]}"; do
    component="${TEST_IPS[$ip]}"
    echo ""
    echo "测试: $ip (组件: $component)"
    echo "--------------------------------"
    timeout 10 ./llm-detector -t "$ip" --timeout 8s --json 2>&1 | grep -E '"type"|"available"|"error"' | head -20
    sleep 1
done

echo ""
echo "================================"
echo "测试完成"
