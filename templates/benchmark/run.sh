#!/bin/bash

# 清理报告目录
rm -rf reports

# 创建报告目录
mkdir -p reports

# 运行HTTP测试
echo "Running HTTP tests..."
# k6 run http.js

# 运行gRPC测试
echo "Running gRPC tests..."
k6 run grpc.js

# 运行WebSocket测试
echo "Running WebSocket tests..."
# k6 run websocket.js

echo "All tests completed. Reports are available in the benchmark/reports directory." 