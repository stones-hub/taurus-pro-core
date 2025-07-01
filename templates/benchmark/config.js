export const CONFIG = {
    // HTTP测试配置
    http: {
        baseUrl: 'http://localhost:9080',
        endpoints: {
            test: '/v1/api/?age=30&email=test@demo.com&id=1&phone=13800138000&name=tester'  // 通用测试端点
        }
    },
    
    // gRPC测试配置
    grpc: {
        baseUrl: 'localhost:50051',
        protoDir: '/Users/yelei/data/code/projects/go/Taurus/benchmark/proto/',
        protoFile: 'user.proto',
        // 一元调用配置
        unaryMethod: 'user.UserService/GetUserInfo',
        unaryParams: {
            user_id: 1
        },
        // 一元鉴权
        methodMetadata: {
            'authorization': 'Bearer 123456'
        },
        // 流式调用配置
        streamMethod: 'user.UserService/GetUserList',
        streamRequestCount: 5, // 每次流式调用发送的请求数
        streamParams: {
            user_ids: [1, 2, 3, 4, 5],
            page_size: 10,
            page_num: 1
        },
        // 流式鉴权
        streamMetadata: {
            'authorization': 'Bearer 123456'
        }
    },
    
    // WebSocket测试配置, 如果以后要测试的websocket服务, 是需要做鉴权的，修改websocket.js
    websocket: {
        baseUrl: 'ws://localhost:9080/ws'
    },
    
    // 测试阶段配置
    stages: [
        { duration: '1m', target: 200 },   // 爬坡阶段
        { duration: '2m', target: 300 },   // 稳定阶段
        { duration: '1m', target: 500 },   // 压力阶段
        { duration: '1m', target: 300 },   // 缓慢降低
    ]
}; 