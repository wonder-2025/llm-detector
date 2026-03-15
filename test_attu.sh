#!/bin/bash

# Attu站点测试脚本
# 从Shodan获取的站点列表

SITES=(
    "84.247.151.177"
    "54.174.246.181"
    "185.196.21.81"
    "34.59.99.95"
    "114.67.114.191"
)

echo "=== Attu Detection Test ==="
echo "Testing ${#SITES[@]} sites from Shodan"
echo ""

for site in "${SITES[@]}"; do
    echo "[*] Testing: $site"
    timeout 10 ./llm-detector -t "http://$site" 2>&1 | grep -E "(attu|Detected|Error)" || echo "  [-] No result"
    echo ""
done

echo "=== Test Complete ==="
