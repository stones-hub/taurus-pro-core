import http from 'k6/http';
import { check, sleep } from 'k6';
import { CONFIG } from './config.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export let options = {
    stages: CONFIG.stages,
    thresholds: {
        'http_req_duration': ['p(95)<500'],  // 95%的请求应该在500ms内完成
        'http_req_failed': ['rate<0.01'],    // 错误率应该小于1%
    }
};

// TODO 如果需要鉴权，修改这里, 根据不同的场景做不同的测试
export default function () {
    const response = http.get(`${CONFIG.http.baseUrl}${CONFIG.http.endpoints.test}`);
    check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 200ms': (r) => r.timings.duration < 200,
    });

    // 模拟用户思考时间
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
    
    // 获取响应时间统计
    const avgDuration = data.metrics.http_req_duration.values.avg;
    const maxDuration = data.metrics.http_req_duration.values.max;
    const p90Duration = data.metrics.http_req_duration.values['p(90)'];
    const p95Duration = data.metrics.http_req_duration.values['p(95)'];
    
    // 获取数据统计
    const dataSent = data.metrics.data_sent.values;
    const dataReceived = data.metrics.data_received.values;
    
    // 在控制台输出统计信息
    console.log(`\n=== HTTP 测试统计 ===`);
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
        "reports/http-report.html": htmlReport(data, {
            title: "HTTP 性能测试报告",
            json: true,
            includeMetrics: true,
            includeThresholds: true,
            includeGroups: true,
            includeChecks: true,
            includeTags: true
        }),
        "reports/http-stats.json": JSON.stringify(statsJson, null, 2)
    };
} 