#!/bin/bash
# 全面测试LLM Detector工具 - 测试所有10种组件类型

cd /root/.openclaw/workspace/llm-detector

echo "============================================"
echo "LLM Detector 工具全面测试"
echo "============================================"
echo ""

# 定义测试用例
# 格式: "IP:端口|期望组件类型|描述"

declare -a TEST_CASES=(
    # Ollama (port=11434) - 从FOFA获取的IP
    "47.109.224.190:11434|Ollama|Ollama测试1"
    "1.71.80.136:11434|Ollama|Ollama测试2"
    "120.79.23.130:11434|Ollama|Ollama测试3"

    # Jupyter (port=8888)
    "52.82.123.100:8888|Jupyter|Jupyter测试1"
    "47.242.38.187:8888|Jupyter|Jupyter测试2"
    "8.210.28.164:8888|Jupyter|Jupyter测试3"

    # MLflow (port=5000)
    "52.82.123.100:5000|MLflow|MLflow测试1"
    "47.242.38.187:5000|MLflow|MLflow测试2"
    "8.210.28.164:5000|MLflow|MLflow测试3"

    # Airflow (port=8080)
    "52.82.123.100:8080|Airflow|Airflow测试1"
    "47.242.38.187:8080|Airflow|Airflow测试2"
    "8.210.28.164:8080|Airflow|Airflow测试3"

    # Dify (port=3000)
    "52.82.123.100:3000|Dify|Dify测试1"
    "47.242.38.187:3000|Dify|Dify测试2"
    "8.210.28.164:3000|Dify|Dify测试3"

    # ComfyUI (port=8188)
    "52.82.123.100:8188|ComfyUI|ComfyUI测试1"
    "47.242.38.187:8188|ComfyUI|ComfyUI测试2"
    "8.210.28.164:8188|ComfyUI|ComfyUI测试3"

    # ClickHouse (port=8123)
    "52.82.123.100:8123|ClickHouse|ClickHouse测试1"
    "47.242.38.187:8123|ClickHouse|ClickHouse测试2"
    "8.210.28.164:8123|ClickHouse|ClickHouse测试3"

    # OpenWebUI (port=8080) - 使用不同IP区分Airflow
    "47.109.224.190:8080|OpenWebUI|OpenWebUI测试1"
    "1.71.80.136:8080|OpenWebUI|OpenWebUI测试2"
    "120.79.23.130:8080|OpenWebUI|OpenWebUI测试3"

    # LiteLLM (port=4000)
    "52.82.123.100:4000|LiteLLM|LiteLLM测试1"
    "47.242.38.187:4000|LiteLLM|LiteLLM测试2"
    "8.210.28.164:4000|LiteLLM|LiteLLM测试3"

    # Gradio (port=7860)
    "52.82.123.100:7860|Gradio|Gradio测试1"
    "47.242.38.187:7860|Gradio|Gradio测试2"
    "8.210.28.164:7860|Gradio|Gradio测试3"
)

# 统计变量
total_tests=0
success_tests=0
failed_tests=0
timeout_tests=0

# 按组件类型统计
declare -A component_total
declare -A component_success
declare -A component_failed

# 详细结果
results_log=""

for test_case in "${TEST_CASES[@]}"; do
    IFS='|' read -r target expected_type description <<< "$test_case"

    # 初始化组件统计
    if [ -z "${component_total[$expected_type]}" ]; then
        component_total[$expected_type]=0
        component_success[$expected_type]=0
        component_failed[$expected_type]=0
    fi

    ((total_tests++))
    ((component_total[$expected_type]++))

    echo "[$total_tests/30] 测试: $description"
    echo "    目标: $target"
    echo "    期望组件: $expected_type"

    # 运行检测
    result=$(timeout 12 ./llm-detector -t "$target" --timeout 10s --json 2>/dev/null)

    if [ -z "$result" ]; then
        echo "    结果: ❌ 超时/无响应"
        ((timeout_tests++))
        ((component_failed[$expected_type]++))
        results_log+="| $target | $expected_type | ❌ 超时 |\n"
    else
        # 检查是否检测到期望的组件
        detected_count=$(echo "$result" | grep -o '"available": true' | wc -l)

        if [ "$detected_count" -gt 0 ]; then
            # 提取检测到的组件类型
            detected_types=$(echo "$result" | grep -B2 '"available": true' | grep '"type"' | sed 's/.*"type": "\([^"]*\)".*/\1/' | tr '\n' ', ')
            detected_types=${detected_types%, }

            # 检查是否检测到期望的组件
            if echo "$result" | grep -B2 '"available": true' | grep -q "\"type\": \"$expected_type\""; then
                echo "    结果: ✅ 成功检测到 $expected_type"
                ((success_tests++))
                ((component_success[$expected_type]++))
                results_log+="| $target | $expected_type | ✅ 成功 |\n"
            else
                echo "    结果: ⚠️  检测到其他组件: $detected_types"
                ((failed_tests++))
                ((component_failed[$expected_type]++))
                results_log+="| $target | $expected_type | ⚠️ 检测到其他 |\n"
            fi
        else
            echo "    结果: ❌ 未检测到任何组件"
            ((failed_tests++))
            ((component_failed[$expected_type]++))
            results_log+="| $target | $expected_type | ❌ 未检测到 |\n"
        fi
    fi

    echo ""
    sleep 1
done

# 计算准确率
if [ $total_tests -gt 0 ]; then
    accuracy=$((success_tests * 100 / total_tests))
else
    accuracy=0
fi

echo "============================================"
echo "测试结果汇总"
echo "============================================"
echo ""
echo "## 总体统计"
echo "- 总测试数: $total_tests"
echo "- 检测成功: $success_tests"
echo "- 检测失败: $failed_tests"
echo "- 超时: $timeout_tests"
echo "- 准确率: $accuracy%"
echo ""
echo "## 各组件类型统计"
echo ""
for component in "${!component_total[@]}"; do
    total=${component_total[$component]}
    success=${component_success[$component]:-0}
    failed=${component_failed[$component]:-0}
    if [ $total -gt 0 ]; then
        comp_accuracy=$((success * 100 / total))
    else
        comp_accuracy=0
    fi
    echo "### $component"
    echo "- 测试数: $total"
    echo "- 成功: $success"
    echo "- 失败: $failed"
    echo "- 准确率: $comp_accuracy%"
    echo ""
done

echo "## 详细结果"
echo ""
echo "| IP:端口 | 期望组件 | 结果 |"
echo "|---------|----------|------|"
printf "$results_log"
echo ""
echo "============================================"
