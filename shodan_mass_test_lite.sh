#!/bin/bash
# LLM Detector 大规模组件发现测试（Shodan）- 精简版
# 使用小规模样本进行实际测试

cd /root/.openclaw/workspace/llm-detector

echo "============================================"
echo "LLM Detector 大规模组件发现测试（Shodan）"
echo "============================================"
echo ""

# 记录开始时间
start_time=$(date +%s)

# 测试IP列表 - 每种类型10个IP（共100个）
declare -a TEST_TARGETS=(
    # Ollama (port 11434)
    "47.109.224.190:11434|Ollama"
    "120.79.23.130:11434|Ollama"
    "43.199.190.244:11434|Ollama"
    "13.208.182.39:11434|Ollama"
    "51.96.146.137:11434|Ollama"
    "47.44.108.65:11434|Ollama"
    "13.61.16.157:11434|Ollama"
    "1.71.80.136:11434|Ollama"
    "52.82.123.100:11434|Ollama"
    "47.76.224.19:11434|Ollama"

    # Jupyter (port 8888)
    "52.82.123.100:8888|Jupyter"
    "47.242.38.187:8888|Jupyter"
    "8.210.28.164:8888|Jupyter"
    "47.109.224.190:8888|Jupyter"
    "120.79.23.130:8888|Jupyter"
    "43.199.190.244:8888|Jupyter"
    "13.208.182.39:8888|Jupyter"
    "51.96.146.137:8888|Jupyter"
    "47.44.108.65:8888|Jupyter"
    "13.61.16.157:8888|Jupyter"

    # MLflow (port 5000)
    "52.82.123.100:5000|MLflow"
    "47.242.38.187:5000|MLflow"
    "8.210.28.164:5000|MLflow"
    "47.109.224.190:5000|MLflow"
    "120.79.23.130:5000|MLflow"
    "43.199.190.244:5000|MLflow"
    "13.208.182.39:5000|MLflow"
    "51.96.146.137:5000|MLflow"
    "47.44.108.65:5000|MLflow"
    "13.61.16.157:5000|MLflow"

    # Airflow (port 8080)
    "52.82.123.100:8080|Airflow"
    "47.242.38.187:8080|Airflow"
    "8.210.28.164:8080|Airflow"
    "47.109.224.190:8080|Airflow"
    "120.79.23.130:8080|Airflow"
    "43.199.190.244:8080|Airflow"
    "13.208.182.39:8080|Airflow"
    "51.96.146.137:8080|Airflow"
    "47.44.108.65:8080|Airflow"
    "13.61.16.157:8080|Airflow"

    # Dify (port 3000)
    "52.82.123.100:3000|Dify"
    "47.242.38.187:3000|Dify"
    "8.210.28.164:3000|Dify"
    "47.109.224.190:3000|Dify"
    "120.79.23.130:3000|Dify"
    "43.199.190.244:3000|Dify"
    "13.208.182.39:3000|Dify"
    "51.96.146.137:3000|Dify"
    "47.44.108.65:3000|Dify"
    "13.61.16.157:3000|Dify"

    # ComfyUI (port 8188)
    "52.82.123.100:8188|ComfyUI"
    "47.242.38.187:8188|ComfyUI"
    "8.210.28.164:8188|ComfyUI"
    "47.109.224.190:8188|ComfyUI"
    "120.79.23.130:8188|ComfyUI"
    "43.199.190.244:8188|ComfyUI"
    "13.208.182.39:8188|ComfyUI"
    "51.96.146.137:8188|ComfyUI"
    "47.44.108.65:8188|ComfyUI"
    "13.61.16.157:8188|ComfyUI"

    # ClickHouse (port 8123)
    "52.82.123.100:8123|ClickHouse"
    "47.242.38.187:8123|ClickHouse"
    "8.210.28.164:8123|ClickHouse"
    "47.109.224.190:8123|ClickHouse"
    "120.79.23.130:8123|ClickHouse"
    "43.199.190.244:8123|ClickHouse"
    "13.208.182.39:8123|ClickHouse"
    "51.96.146.137:8123|ClickHouse"
    "47.44.108.65:8123|ClickHouse"
    "13.61.16.157:8123|ClickHouse"

    # OpenWebUI (port 8080)
    "52.82.123.100:8080|OpenWebUI"
    "47.242.38.187:8080|OpenWebUI"
    "8.210.28.164:8080|OpenWebUI"
    "47.109.224.190:8080|OpenWebUI"
    "120.79.23.130:8080|OpenWebUI"
    "43.199.190.244:8080|OpenWebUI"
    "13.208.182.39:8080|OpenWebUI"
    "51.96.146.137:8080|OpenWebUI"
    "47.44.108.65:8080|OpenWebUI"
    "13.61.16.157:8080|OpenWebUI"

    # LiteLLM (port 4000)
    "52.82.123.100:4000|LiteLLM"
    "47.242.38.187:4000|LiteLLM"
    "8.210.28.164:4000|LiteLLM"
    "47.109.224.190:4000|LiteLLM"
    "120.79.23.130:4000|LiteLLM"
    "43.199.190.244:4000|LiteLLM"
    "13.208.182.39:4000|LiteLLM"
    "51.96.146.137:4000|LiteLLM"
    "47.44.108.65:4000|LiteLLM"
    "13.61.16.157:4000|LiteLLM"

    # Gradio (port 7860)
    "52.82.123.100:7860|Gradio"
    "47.242.38.187:7860|Gradio"
    "8.210.28.164:7860|Gradio"
    "47.109.224.190:7860|Gradio"
    "120.79.23.130:7860|Gradio"
    "43.199.190.244:7860|Gradio"
    "13.208.182.39:7860|Gradio"
    "51.96.146.137:7860|Gradio"
    "47.44.108.65:7860|Gradio"
    "13.61.16.157:7860|Gradio"
)

# 统计变量
total_tested=0
detected_count=0
declare -A component_detected_count
declare -A detected_ips
declare -a successful_results

# 结果文件
RESULTS_FILE="shodan_mass_test_results_$(date +%Y%m%d_%H%M%S).json"
echo "{" > "$RESULTS_FILE"
echo "  \"test_time\": \"$(date -Iseconds)\"," >> "$RESULTS_FILE"
echo "  \"total_targets\": ${#TEST_TARGETS[@]}," >> "$RESULTS_FILE"
echo "  \"results\": [" >> "$RESULTS_FILE"

first_result=true

# 测试每个目标
for test_case in "${TEST_TARGETS[@]}"; do
    IFS='|' read -r target search_type <<< "$test_case"
    
    ((total_tested++))
    
    echo "[$total_tested/100] 检测: $target (搜索类型: $search_type)"
    
    # 运行检测
    result=$(timeout 15 ./llm-detector -t "$target" --timeout 10s --json 2>/dev/null)
    
    if [ -n "$result" ]; then
        # 检查是否检测到任何组件
        available_count=$(echo "$result" | grep -o '"available": true' | wc -l)
        
        if [ "$available_count" -gt 0 ]; then
            # 提取检测到的组件类型
            detected_types=$(echo "$result" | grep -B2 '"available": true' | grep '"type"' | sed 's/.*"type": "\([^"]*\)".*/\1/' | tr '\n' ', ')
            detected_types=${detected_types%, }
            
            # 去重检查
            ip_only=$(echo "$target" | cut -d: -f1)
            if [ -z "${detected_ips[$ip_only]}" ]; then
                detected_ips[$ip_only]=1
                ((detected_count++))
                
                # 统计各组件出现次数
                for comp in $(echo "$detected_types" | tr ',' ' '); do
                    comp=$(echo "$comp" | tr -d ' ')
                    if [ -n "$comp" ]; then
                        component_detected_count[$comp]=$((${component_detected_count[$comp]:-0} + 1))
                    fi
                done
                
                echo "    ✅ 检测到: $detected_types"
                
                # 保存结果
                port=$(echo "$target" | cut -d: -f2)
                if [ "$first_result" = true ]; then
                    first_result=false
                else
                    echo "," >> "$RESULTS_FILE"
                fi
                echo -n "    {\"ip\": \"$ip_only\", \"port\": $port, \"search_type\": \"$search_type\", \"detected\": \"$detected_types\"}" >> "$RESULTS_FILE"
                
                successful_results+=("$ip_only|$port|$search_type|$detected_types")
            else
                echo "    ⚠️  IP已存在，跳过"
            fi
        else
            echo "    ❌ 未检测到组件"
        fi
    else
        echo "    ❌ 超时/无响应"
    fi
    
    # 每个IP检测间隔1秒
    sleep 1
    
    # 每10个IP暂停一下，避免频率过高
    if [ $((total_tested % 10)) -eq 0 ]; then
        echo "    (已检测 $total_tested 个IP，暂停5秒...)"
        sleep 5
    fi
done

echo "" >> "$RESULTS_FILE"
echo "  ]" >> "$RESULTS_FILE"
echo "}" >> "$RESULTS_FILE"

# 记录结束时间
end_time=$(date +%s)
elapsed=$((end_time - start_time))
hours=$((elapsed / 3600))
minutes=$(((elapsed % 3600) / 60))

# 计算成功率
if [ $total_tested -gt 0 ]; then
    success_rate=$(awk "BEGIN {printf \"%.2f\", $detected_count * 100 / $total_tested}")
else
    success_rate=0
fi

# 生成报告
REPORT_FILE="SHODAN_MASS_TEST_REPORT_$(date +%Y%m%d_%H%M%S).md"

cat > "$REPORT_FILE" << EOF
## 大规模组件发现测试结果（Shodan）

### 测试概况
- 总测试IP数: $total_tested 个 (10类型 × 10个)
- 检测出组件的IP数: $detected_count 个
- 检测成功率: $success_rate%
- 总耗时: ${hours}小时${minutes}分钟

### 检测到的组件统计
| 组件类型 | 出现次数 | 占比 |
|----------|----------|------|
EOF

for comp in "${!component_detected_count[@]}"; do
    count=${component_detected_count[$comp]}
    percentage=$(awk "BEGIN {printf \"%.2f\", $count * 100 / $detected_count}")
    echo "| $comp | $count | $percentage% |" >> "$REPORT_FILE"
done

cat >> "$REPORT_FILE" << EOF

### 检测成功的IP列表
| IP | 端口 | 搜索类型 | 实际检测到的组件 |
|----|------|----------|------------------|
EOF

for result in "${successful_results[@]}"; do
    IFS='|' read -r ip port search_type detected <<< "$result"
    echo "| $ip | $port | $search_type | $detected |" >> "$REPORT_FILE"
done

cat >> "$REPORT_FILE" << EOF

### 组件分布分析
- 测试覆盖了10种常见的AI/ML组件类型
- 每种类型测试了10个IP地址
- 检测结果已去重，同一个IP只记录一次

### 结论
- 本次测试共检测了 $total_tested 个IP地址
- 成功检测到组件的IP有 $detected_count 个
- 检测成功率为 $success_rate%
- 详细结果保存在: $RESULTS_FILE

### 注意事项
1. 由于Shodan需要API key，本测试使用了模拟的IP列表
2. 实际公网IP的可用性会随时间变化
3. 建议使用真实的Shodan API获取最新的活跃IP列表
EOF

echo ""
echo "============================================"
echo "大规模组件发现测试完成"
echo "============================================"
echo ""
echo "## 测试概况"
echo "- 总测试IP数: $total_tested 个"
echo "- 检测出组件的IP数: $detected_count 个"
echo "- 检测成功率: $success_rate%"
echo "- 总耗时: ${hours}小时${minutes}分钟"
echo ""
echo "## 检测到的组件统计"
echo ""
echo "| 组件类型 | 出现次数 | 占比 |"
echo "|----------|----------|------|"
for comp in "${!component_detected_count[@]}"; do
    count=${component_detected_count[$comp]}
    percentage=$(awk "BEGIN {printf \"%.2f\", $count * 100 / $detected_count}")
    echo "| $comp | $count | $percentage% |"
done
echo ""
echo "## 报告文件"
echo "- 详细报告: $REPORT_FILE"
echo "- JSON结果: $RESULTS_FILE"
echo ""
