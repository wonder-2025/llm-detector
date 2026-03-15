#!/bin/bash

# LLM Detector 输出格式测试脚本

echo "=========================================="
echo "LLM Detector 输出格式测试"
echo "=========================================="

DETECTOR="./llm-detector"
TEST_TARGET="127.0.0.1"

echo ""
echo "1. 测试默认文本输出..."
echo "------------------------------------------"
$DETECTOR -t $TEST_TARGET -v 2>&1 | head -30

echo ""
echo "2. 测试增强JSON输出..."
echo "------------------------------------------"
$DETECTOR -t $TEST_TARGET --json 2>&1 | head -50

echo ""
echo "3. 测试CSV输出..."
echo "------------------------------------------"
$DETECTOR -t $TEST_TARGET --csv -o test_results.csv 2>&1
echo "CSV内容:"
cat test_results.csv

echo ""
echo "4. 测试HTML报告输出..."
echo "------------------------------------------"
$DETECTOR -t $TEST_TARGET --html -o test_report.html 2>&1
echo "HTML报告已生成: test_report.html"
ls -lh test_report.html

echo ""
echo "5. 测试深色主题HTML..."
echo "------------------------------------------"
$DETECTOR -t $TEST_TARGET --html --dark -o test_report_dark.html 2>&1
echo "深色HTML报告已生成: test_report_dark.html"

echo ""
echo "6. 测试版本信息..."
echo "------------------------------------------"
$DETECTOR --version

echo ""
echo "=========================================="
echo "测试完成!"
echo "=========================================="
echo ""
echo "生成的文件:"
ls -lh test_results.csv test_report.html test_report_dark.html 2>/dev/null
