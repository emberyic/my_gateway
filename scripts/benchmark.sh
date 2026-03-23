#!/bin/bash

echo "  开始网关性能压测..."
echo "----------------------------------------"

# 1. 基础健康检查
echo "1. 基础健康检查:"
curl -s http://localhost:8080/ | head -2

# 2. 单请求测试
echo -e "\n2. 单请求测试:"
time curl -s -o /dev/null http://localhost:8080/api/test

# 3. 使用wrk进行压力测试（如果安装了wrk）
if command -v wrk &> /dev/null; then
    echo -e "\n3. 压力测试 (wrk):"
    wrk -t4 -c100 -d30s --latency http://localhost:8080/api/test
else
    echo -e "\n3. 安装wrk后进行压力测试:"
    echo "   Ubuntu: sudo apt-get install wrk"
    echo "   Mac: brew install wrk"
fi

# 4. 限流测试
echo -e "\n4. 限流功能测试 (快速发送20个请求):"
for i in {1..20}; do
    status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/test)
    echo -n "请求$i: $status | "
    if (( i % 5 == 0 )); then echo; fi
    sleep 0.1
done

echo -e "\n----------------------------------------"
echo "压测完成。"