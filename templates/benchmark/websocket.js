import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { CONFIG } from './config.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export let options = {
    stages: CONFIG.stages,
    thresholds: {
        'ws_ping': ['p(95)<100'],            // WebSocket ping应该在100ms内
        'ws_msgs_sent': ['rate>0'],          // 确保消息发送率大于0
        'ws_msgs_received': ['rate>0']      // 确保消息接收率大于0
    }
};


// TODO 如果需要鉴权，修改这里, 根据不同的场景做不同的测试
export default function () {
    const url = CONFIG.websocket.baseUrl;
    const params = {
        headers: {
            'Sec-WebSocket-Protocol': 'json',
        },
    };

    const response = ws.connect(url, params, function (socket) {
        socket.on('open', () => {
            // console.log('WebSocket connection established');
            
            // 发送订阅消息
            socket.send(JSON.stringify({
                type: 'subscribe',
                channel: 'test'
            }));

            // 设置心跳检测
            const heartbeat = setInterval(() => {
                if (socket.readyState === 1) { // 1 = OPEN
                    socket.send(JSON.stringify({
                        type: 'ping',
                        timestamp: new Date().getTime()
                    }));
                }
            }, 5000); // 每5秒发送一次心跳

            // 30秒后清理心跳并关闭连接
            socket.setTimeout(function () {
                clearInterval(heartbeat);
                socket.close();
            }, 15000);
        });

        socket.on('message', (data) => {
            try {
                const message = JSON.parse(data);
                // console.log('Message received:', message);
                
                check(message, {
                    'message is valid': (m) => m !== null,
                    'message has type': (m) => m.type !== undefined,
                });
            } catch (e) {
                console.error('Error parsing message:', e);
            }
        });

        socket.on('close', () => {
            // console.log('WebSocket connection closed');
        });

        socket.on('error', (e) => {
            // 记录连接失败
            check(null, {
                'WebSocket connection successful': () => false,
                'connection error details': () => {
                    console.error('WebSocket error:', e);
                    return false;
                }
            });
        });
    });

    // 所有的检测都通过才算成功, 1. 状态， 2. 响应时间， 3. 消息
    check(response, {
        'WebSocket connection established': (r) => r && r.status === 101,
        'error is null': (r) => r && r.error === "",
        'message info': (r) => {
            // 如果需要这里可以用来判断真实请求后返回的数据的正确性
            // console.log('Response:', r);
            return true;
        }
    });

    sleep(1);
}

export function handleSummary(data) {
    // console.log('data', data);
    
    // 获取请求统计
    const totalRequests = data.metrics.iterations.values.count;
    const failedChecks = data.metrics.checks.values.fails;
    const successRate = ((totalRequests - failedChecks) / totalRequests * 100).toFixed(2);
    
    // 获取最大并发量
    const maxVUs = data.metrics.vus_max.values.max;
    
    // 获取 WebSocket 会话统计
    const wsSessions = data.metrics.ws_sessions.values;
    const wsConnecting = data.metrics.ws_connecting.values;
    const wsSessionDuration = data.metrics.ws_session_duration.values;
    
    // 获取消息统计
    const messagesSent = data.metrics.ws_msgs_sent.values;
    const messagesReceived = data.metrics.ws_msgs_received.values;
    
    // 获取数据统计
    const dataSent = data.metrics.data_sent.values;
    const dataReceived = data.metrics.data_received.values;
    
    // 在控制台输出统计信息
    console.log(`\n=== WebSocket 测试统计 ===`);
    console.log(`总请求数: ${totalRequests}`);
    console.log(`检查失败数: ${failedChecks}`);
    console.log(`成功率: ${successRate}%`);
    console.log(`最大并发量: ${maxVUs}`);
    console.log(`\n=== WebSocket 会话统计 ===`);
    console.log(`总会话数: ${wsSessions.count}`);
    console.log(`会话建立速率: ${wsSessions.rate.toFixed(2)}/s`);
    console.log(`平均连接时间: ${wsConnecting.avg.toFixed(2)}ms`);
    console.log(`最大连接时间: ${wsConnecting.max.toFixed(2)}ms`);
    console.log(`平均会话持续时间: ${wsSessionDuration.avg.toFixed(2)}ms`);
    console.log(`\n=== 消息统计 ===`);
    console.log(`发送消息数: ${messagesSent.count}`);
    console.log(`发送消息速率: ${messagesSent.rate.toFixed(2)}/s`);
    console.log(`接收消息数: ${messagesReceived.count}`);
    console.log(`接收消息速率: ${messagesReceived.rate.toFixed(2)}/s`);
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
        "WebSocket会话统计": {
            "总会话数": wsSessions.count,
            "会话建立速率": wsSessions.rate.toFixed(2) + "/s",
            "平均连接时间": wsConnecting.avg.toFixed(2) + "ms",
            "最大连接时间": wsConnecting.max.toFixed(2) + "ms",
            "平均会话持续时间": wsSessionDuration.avg.toFixed(2) + "ms"
        },
        "消息统计": {
            "发送消息": {
                "总数": messagesSent.count,
                "速率": messagesSent.rate.toFixed(2) + "/s"
            },
            "接收消息": {
                "总数": messagesReceived.count,
                "速率": messagesReceived.rate.toFixed(2) + "/s"
            }
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
        "reports/websocket-report.html": htmlReport(data, {
            title: "WebSocket 性能测试报告",
            json: true,
            includeMetrics: true,
            includeThresholds: true,
            includeGroups: true,
            includeChecks: true,
            includeTags: true
        }),
        "reports/websocket-stats.json": JSON.stringify(statsJson, null, 2)
    };
} 