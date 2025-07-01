import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { CONFIG } from './config.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export let options = {
    stages: CONFIG.stages,
};

// 初始化客户端
const client = new grpc.Client();
let isConnected = true;

// 加载proto文件
try {
    // 先importPaths，再protoFiles, 引入要测试的proto文件
    client.load([CONFIG.grpc.protoDir], CONFIG.grpc.protoFile);
    console.log('Proto file loaded successfully');
} catch (e) {
    console.error('Error loading proto file:', e);
    throw e; // 中断执行
}

// 一元调用
export function callUnaryMethod() {
    // 连接gRPC服务，添加重试和超时配置
    try {
        console.log('Attempting to connect to:', CONFIG.grpc.baseUrl);
        client.connect(CONFIG.grpc.baseUrl, {
            plaintext: true,
            timeout: '5s'  // 连接超时时间
        });
        console.log('Connected to gRPC server successfully');

        // 调用gRPC服务
        console.log('Calling method:', CONFIG.grpc.unaryMethod);
        console.log('With params:', JSON.stringify(CONFIG.grpc.unaryParams));
        console.log('With metadata:', JSON.stringify(CONFIG.grpc.methodMetadata));
        
        const userResponse = client.invoke(CONFIG.grpc.unaryMethod, CONFIG.grpc.unaryParams, {
            metadata: CONFIG.grpc.methodMetadata,
            timeout: '10s'  // 调用超时时间
        });
        
        console.log('Raw response:', JSON.stringify(userResponse));
        
        // 所有的检测都通过才算成功, 1. 状态， 2. 响应时间， 3. 消息
        check(userResponse, {
            'get user info status is OK': (r) => {
                console.log('Response status:', r ? r.status : 'no status');
                return r && r.status === grpc.StatusOK;
            },
            'error is null': (r) => {
                if (r && r.error) {
                    console.log('Response error:', r.error);
                }
                return r && r.error === null;
            },
            'message info': (r) => {
                console.log('Response message:', r ? r.message : 'no message');
                return true;
            }
        });

    } catch (e) {
        // 记录连接失败
        console.error('Detailed error:', {
            error: e,
            message: e.message,
            stack: e.stack
        });
        check(null, {
            'gRPC connection successful': () => false,
            'connection error details': () => {
                console.error('gRPC connection error:', {
                    error: e,
                    message: e.message,
                    stack: e.stack
                });
                return false;
            }
        });
    } finally {
        // 确保连接被关闭
        try {
            client.close();
            console.log('Connection closed');
        } catch (e) {
            console.error('Error closing gRPC connection:', e);
        }
    }

    sleep(1);
}

// 服务端流式测试 - GetUserList
export function callStreamMethod() {
    try {
        // 连接服务器
        client.connect(CONFIG.grpc.baseUrl, {
            plaintext: true,
            timeout: '10s'
        });

        // 创建 Stream 实例
        const stream = new grpc.Stream(client, CONFIG.grpc.streamMethod, {
            metadata: CONFIG.grpc.streamMetadata
        });
        let messageCount = 0;

        // 设置数据处理器
        stream.on('data', (data) => {
            console.log('Received user data:', JSON.stringify(data));
            messageCount++;
            check(data, {
                'user data is valid': (d) => d !== null,
            });
        });

        // 设置错误处理器
        stream.on('error', (err) => {
            console.error('Stream error:', err);
            check(null, {
                'stream error check': () => false,
            });
        });

        // 设置结束处理器
        stream.on('end', () => {
            console.log('Stream ended, total messages:', messageCount);
            check(messageCount, {
                'received messages': (count) => count > 0,
            });
        });

        // 发送请求
        stream.write(CONFIG.grpc.streamParams);

        // 结束请求
        stream.end();

        // 等待接收数据
        sleep(2);

    } catch (e) {
        console.error('Server stream error:', e);
    } finally {
        client.close();
    }
}

export function handleSummary(data) {
    // console.log('data', data);
    
    // 获取请求统计
    const totalRequests = data.metrics.iterations.values.count;
    const failedChecks = data.metrics.checks.values.fails;
    const successRate = ((totalRequests - failedChecks) / totalRequests * 100).toFixed(2);
    
    // 获取最大并发量
    const maxVUs = data.metrics.vus_max.values.max;
    
    // 获取响应时间统计
    const avgDuration = data.metrics.grpc_req_duration.values.avg;
    const maxDuration = data.metrics.grpc_req_duration.values.max;
    const p90Duration = data.metrics.grpc_req_duration.values['p(90)'];
    const p95Duration = data.metrics.grpc_req_duration.values['p(95)'];
    
    // 获取数据统计
    const dataSent = data.metrics.data_sent.values;
    const dataReceived = data.metrics.data_received.values;
    
    // 在控制台输出统计信息
    console.log(`\n=== gRPC 测试统计 ===`);
    console.log(`总请求数: ${totalRequests}`);
    console.log(`检查失败数: ${failedChecks}`);
    console.log(`成功率: ${successRate}%`);
    console.log(`最大并发量: ${maxVUs}`);
    console.log(`\n=== 响应时间统计 ===`);
    console.log(`平均响应时间: ${avgDuration.toFixed(2)}ms`);
    console.log(`最大响应时间: ${maxDuration.toFixed(2)}ms`);
    console.log(`90%响应时间: ${p90Duration.toFixed(2)}ms`);
    console.log(`95%响应时间: ${p95Duration.toFixed(2)}ms`);
    console.log(`\n=== 数据传输统计 ===`);
    console.log(`发送数据: ${(dataSent.count/1024).toFixed(2)}KB (${(dataSent.rate/1024).toFixed(2)}KB/s)`);
    console.log(`接收数据: ${(dataReceived.count/1024).toFixed(2)}KB (${(dataReceived.rate/1024).toFixed(2)}KB/s)`);
    console.log(`====================\n`);

    // 构建统计数据的 JSON 对象
    const statsJson = {
        "测试基本信息": {
            "总请求数": totalRequests,
            "检查失败数": failedChecks,
            "成功率": successRate + "%",
            "最大并发用户数": maxVUs
        },
        "响应时间统计": {
            "平均响应时间": avgDuration.toFixed(2) + "ms",
            "最大响应时间": maxDuration.toFixed(2) + "ms",
            "90%响应时间": p90Duration.toFixed(2) + "ms",
            "95%响应时间": p95Duration.toFixed(2) + "ms"
        },
        "数据传输统计": {
            "发送数据": {
                "总量": (dataSent.count/1024).toFixed(2) + "KB",
                "速率": (dataSent.rate/1024).toFixed(2) + "KB/s"
            },
            "接收数据": {
                "总量": (dataReceived.count/1024).toFixed(2) + "KB",
                "速率": (dataReceived.rate/1024).toFixed(2) + "KB/s"
            }
        }
    };

    return {
        "reports/grpc-report.html": htmlReport(data, {
            title: "gRPC 性能测试报告",
            json: true,
            includeMetrics: true,
            includeThresholds: true,
            includeGroups: true,
            includeChecks: true,
            includeTags: true
        }),
        "reports/grpc-stats.json": JSON.stringify(statsJson, null, 2)
    };
} 

export function close() {
    if (isConnected) {
        client.close();
        isConnected = false;
    }
}

export default function() {
    // callUnaryMethod();
    callStreamMethod();
    close();
}



