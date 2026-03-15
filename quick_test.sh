#!/bin/bash
# 快速测试 - 测试少量IP获取样本数据

cd /root/.openclaw/workspace/llm-detector

echo "快速测试 - 获取样本数据"
echo "========================"

# 测试一些已知的活跃IP
declare -a TEST_IPS=(
    "13.208.182.39:11434"
    "httpbin.org:80"
    "httpbin.org:443"
    "47.242.38.187:80"
    "8.210.28.164:80"
    "120.79.23.130:80"
    "43.199.190.244:80"
    "51.96.146.137:80"
    "47.44.108.65:80"
    "13.61.16.157:80"
)

for target in "${TEST_IPS[@]}"; do
    echo ""
    echo "测试: $target"
    timeout 10 ./llm-detector -t "$target" --timeout 8s --json 2>/dev/null | head -50
    sleep 2
done
