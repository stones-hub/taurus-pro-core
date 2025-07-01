# API性能测试套件

这个测试套件使用k6对HTTP、gRPC和WebSocket协议进行性能测试。

## 前置条件

1. 安装k6
```bash
# MacOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

2. 确保测试服务已启动并监听相应端口：
   - HTTP服务: localhost:8080
   - gRPC服务: localhost:9000
   - WebSocket服务: localhost:8080/ws

## 配置

所有测试配置都在 `config.js` 文件中，你可以根据需要修改：
- 虚拟用户数 (vus)
- 测试持续时间 (duration)
- 测试阶段 (stages)
- 性能指标阈值 (thresholds)
- 服务地址和端口

## 运行测试

1. 运行所有测试
```bash
chmod +x run.sh
./run.sh
```

2. 运行单个协议测试
```bash
# HTTP测试
k6 run http.js

# gRPC测试
k6 run grpc.js

# WebSocket测试
k6 run websocket.js
```

## 测试报告

测试完成后，可以在 `reports` 目录下找到HTML格式的测试报告：
- HTTP测试报告: `reports/http-report.html`
- gRPC测试报告: `reports/grpc-report.html`
- WebSocket测试报告: `reports/websocket-report.html`

## 测试指标说明

1. HTTP/gRPC测试指标：
   - 响应时间 (Response Time)
   - 请求率 (Request Rate)
   - 错误率 (Error Rate)
   - 95th/99th 百分位延迟

2. WebSocket测试指标：
   - 连接建立时间
   - 消息传输延迟
   - 连接成功率
   - 消息处理成功率

## 注意事项

1. 测试前请确保所有服务都已正确启动
2. 根据实际情况调整配置文件中的参数
3. 测试报告会自动保存在reports目录下
4. 建议在测试环境中运行，避免影响生产环境 