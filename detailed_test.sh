#!/bin/bash
# 详细测试LLM Detector工具并记录结果

cd /root/.openclaw/workspace/llm-detector

OUTPUT_FILE="test_results_$(date +%Y%m%d_%H%M%S).json"
echo "[]" > "$OUTPUT_FILE"

# 测试IP列表 - 从FOFA获取的真实IP
declare -a TEST_TARGETS=(
    "47.109.224.190:11434|Ollama"
    "1.71.80.136:11434|Ollama"
    "120.79.23.130:11434|Ollama"
    "36.111.157.108:11434|Ollama"
    "121.41.228.131:11434|Ollama"
    "47.101.88.61:11434|Ollama"
    "120.78.122.213:11434|Ollama"
)

echo "开始详细测试LLM Detector工具..."
echo "================================"
echo ""

results="["
first=true

for target_info in "${TEST_TARGETS[@]}"; do
    IFS='|' read -r target expected_type <<< "$target_info"

    echo "测试: $target (期望: $expected_type)"

    # 运行检测
    result=$(timeout 15 ./llm-detector -t "$target" --timeout 10s --json 2>/dev/null)

    if [ -n "$result" ]; then
        # 检查是否检测到任何组件
        detected=$(echo "$result" | grep -o '"available": true' | wc -l)

        if [ "$detected" -gt 0 ]; then
            # 提取检测到的组件类型
            detected_type=$(echo "$result" | grep -B2 '"available": true' | grep '"type"' | head -1 | sed 's/.*"type": "\([^"]*\)".*/\1/')
            status="✅ 检测到: $detected_type"
        else
            status="❌ 未检测到"
        fi

        echo "  结果: $status"

        # 添加到JSON结果
        if [ "$first" = true ]; then
            first=false
        else
            results+=","
        fi

        # 转义JSON字符串
        result_escaped=$(echo "$result" | sed 's/"/\\"/g' | tr '\n' ' ')
        results+="{\"target\":\"$target\",\"expected\":\"$expected_type\",\"result\":\"$result_escaped\",\"status\":\"$status\"}"
    else
        echo "  结果: ❌ 超时或无响应"
        if [ "$first" = true ]; then
            first=false
        else
            results+=","
        fi
        results+="{\"target\":\"$target\",\"expected\":\"$expected_type\",\"result\":\"\",\"status\":\"❌ 超时\"}"
    fi

    echo ""
    sleep 2
done

results+="]"
echo "$results" > "$OUTPUT_FILE"

echo "================================"
echo "测试完成，结果保存到: $OUTPUT_FILE"
