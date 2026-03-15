#!/bin/bash

# LLM Detector 批量扫描功能测试脚本

echo "========================================="
echo "LLM Detector 批量扫描功能测试"
echo "========================================="
echo

# 创建测试目标文件
cat > /tmp/test_targets.txt << 'EOF'
# 测试目标列表
192.168.1.1
192.168.1.100:11434
http://localhost:8080
10.0.0.0/30
EOF

echo "[+] 测试目标文件已创建: /tmp/test_targets.txt"
cat /tmp/test_targets.txt
echo

# 测试1: 文件输入
echo "========================================="
echo "测试1: 文件输入 (-f)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt -v --timeout 5s 2>&1 | head -50
echo

# 测试2: 管道输入
echo "========================================="
echo "测试2: 管道输入 (--stdin)"
echo "========================================="
echo -e "192.168.1.1\n192.168.1.2" | ./llm-detector --stdin -v --timeout 3s 2>&1 | head -30
echo

# 测试3: CSV导出
echo "========================================="
echo "测试3: CSV导出 (--csv)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt --csv /tmp/test_results.csv --timeout 5s 2>&1 | head -20
echo "[+] CSV文件内容:"
cat /tmp/test_results.csv 2>/dev/null || echo "文件未生成"
echo

# 测试4: HTML报告
echo "========================================="
echo "测试4: HTML报告 (--html)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt --html /tmp/test_report.html --timeout 5s 2>&1 | head -20
echo "[+] HTML文件:"
ls -la /tmp/test_report.html 2>/dev/null || echo "文件未生成"
echo

# 测试5: JSON Lines导出
echo "========================================="
echo "测试5: JSON Lines导出 (--jsonl)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt --jsonl /tmp/test_results.jsonl --timeout 5s 2>&1 | head -20
echo "[+] JSONL文件内容:"
cat /tmp/test_results.jsonl 2>/dev/null || echo "文件未生成"
echo

# 测试6: 指定并发数
echo "========================================="
echo "测试6: 指定并发数 (-w)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt -w 10 -v --timeout 5s 2>&1 | head -30
echo

# 测试7: 速率限制
echo "========================================="
echo "测试7: 速率限制 (--rate)"
echo "========================================="
./llm-detector -f /tmp/test_targets.txt --rate 5 -v --timeout 5s 2>&1 | head -30
echo

# 测试8: 显示帮助
echo "========================================="
echo "测试8: 帮助信息"
echo "========================================="
./llm-detector --help
echo

# 测试9: 版本信息
echo "========================================="
echo "测试9: 版本信息"
echo "========================================="
./llm-detector --version
echo

# 清理
rm -f /tmp/test_targets.txt /tmp/test_results.csv /tmp/test_report.html /tmp/test_results.jsonl

echo "========================================="
echo "测试完成!"
echo "========================================="
