#!/bin/bash

# 测试JWT和ratelimit中间件的脚本
# 作者: yelei
# 日期: 2025-08-18

# 设置基础URL
BASE_URL="http://localhost:9080"
AUTH_URL="$BASE_URL/auth"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务是否运行
check_service() {
    print_info "检查服务状态..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        print_success "服务正在运行"
        return 0
    else
        print_error "服务未运行，请先启动服务"
        return 1
    fi
}

# 测试用户注册
test_register() {
    print_info "测试用户注册..."
    
    # 生成随机用户名
    USERNAME="testuser_$(date +%s)"
    PASSWORD="testpass123"
    
    RESPONSE=$(curl -s -X POST "$AUTH_URL/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
    if echo "$RESPONSE" | grep -q "token"; then
        print_success "用户注册成功: $USERNAME"
        # 提取token
        TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "$TOKEN" > /tmp/test_token.txt
        echo "$USERNAME" > /tmp/test_username.txt
        return 0
    else
        print_error "用户注册失败: $RESPONSE"
        return 1
    fi
}

# 测试用户登录
test_login() {
    print_info "测试用户登录..."
    
    if [ ! -f /tmp/test_username.txt ] || [ ! -f /tmp/test_token.txt ]; then
        print_warning "跳过登录测试，需要先注册用户"
        return 1
    fi
    
    USERNAME=$(cat /tmp/test_username.txt)
    PASSWORD="testpass123"
    
    RESPONSE=$(curl -s -X POST "$AUTH_URL/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
    if echo "$RESPONSE" | grep -q "token"; then
        print_success "用户登录成功: $USERNAME"
        # 更新token
        TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "$TOKEN" > /tmp/test_token.txt
        return 0
    else
        print_error "用户登录失败: $RESPONSE"
        return 1
    fi
}

# 测试限流中间件
test_rate_limit() {
    print_info "测试限流中间件..."
    
    print_info "快速发送20个请求测试限流..."
    
    for i in {1..20}; do
        RESPONSE=$(curl -s -X GET "$AUTH_URL/test/ratelimit")
        if echo "$RESPONSE" | grep -q "请求过于频繁"; then
            print_success "限流中间件工作正常，第$i个请求被限流"
            return 0
        fi
        echo -n "."
        sleep 0.05  # 减少间隔时间
    done
    
    print_warning "限流中间件可能未生效，所有请求都通过了"
    return 1
}

# 测试JWT中间件（无token）
test_jwt_no_token() {
    print_info "测试JWT中间件 - 无token访问..."
    
    RESPONSE=$(curl -s -X GET "$AUTH_URL/test/jwt")
    
    if echo "$RESPONSE" | grep -q "缺少JWT令牌\|未授权访问"; then
        print_success "JWT中间件工作正常，无token访问被拒绝"
        return 0
    else
        print_error "JWT中间件可能未生效，无token访问被允许"
        return 1
    fi
}

# 测试JWT中间件（有token）
test_jwt_with_token() {
    print_info "测试JWT中间件 - 有token访问..."
    
    if [ ! -f /tmp/test_token.txt ]; then
        print_warning "跳过JWT token测试，需要先获取token"
        return 1
    fi
    
    TOKEN=$(cat /tmp/test_token.txt)
    
    RESPONSE=$(curl -s -X GET "$AUTH_URL/test/jwt" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$RESPONSE" | grep -q "JWT中间件测试成功"; then
        print_success "JWT中间件工作正常，有token访问成功"
        return 0
    else
        print_error "JWT中间件可能有问题，有token访问失败: $RESPONSE"
        return 1
    fi
}

# 测试受保护的端点
test_protected_endpoint() {
    print_info "测试受保护的端点..."
    
    if [ ! -f /tmp/test_token.txt ]; then
        print_warning "跳过受保护端点测试，需要先获取token"
        return 1
    fi
    
    TOKEN=$(cat /tmp/test_token.txt)
    
    RESPONSE=$(curl -s -X GET "$AUTH_URL/test/protected" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$RESPONSE" | grep -q "受保护端点访问成功"; then
        print_success "受保护端点访问成功"
        return 0
    else
        print_error "受保护端点访问失败: $RESPONSE"
        return 1
    fi
}

# 测试用户资料获取
test_profile() {
    print_info "测试获取用户资料..."
    
    if [ ! -f /tmp/test_token.txt ]; then
        print_warning "跳过用户资料测试，需要先获取token"
        return 1
    fi
    
    TOKEN=$(cat /tmp/test_token.txt)
    
    RESPONSE=$(curl -s -X GET "$AUTH_URL/profile" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$RESPONSE" | grep -q "id\|name"; then
        print_success "用户资料获取成功"
        return 0
    else
        print_error "用户资料获取失败: $RESPONSE"
        return 1
    fi
}

# 测试令牌刷新
test_token_refresh() {
    print_info "测试令牌刷新..."
    
    if [ ! -f /tmp/test_token.txt ]; then
        print_warning "跳过令牌刷新测试，需要先获取token"
        return 1
    fi
    
    TOKEN=$(cat /tmp/test_token.txt)
    
    RESPONSE=$(curl -s -X POST "$AUTH_URL/refresh" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$RESPONSE" | grep -q "token"; then
        print_success "令牌刷新成功"
        # 更新token
        NEW_TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "$NEW_TOKEN" > /tmp/test_token.txt
        return 0
    else
        print_error "令牌刷新失败: $RESPONSE"
        return 1
    fi
}

# 测试用户登出
test_logout() {
    print_info "测试用户登出..."
    
    RESPONSE=$(curl -s -X POST "$AUTH_URL/logout")
    
    if echo "$RESPONSE" | grep -q "登出成功"; then
        print_success "用户登出成功"
        return 0
    else
        print_error "用户登出失败: $RESPONSE"
        return 1
    fi
}

# 清理测试数据
cleanup() {
    print_info "清理测试数据..."
    rm -f /tmp/test_token.txt /tmp/test_username.txt
    print_success "测试数据清理完成"
}

# 主测试流程
main() {
    print_info "开始测试JWT和ratelimit中间件..."
    echo
    
    # 检查服务状态
    if ! check_service; then
        exit 1
    fi
    
    # 测试计数器
    TOTAL_TESTS=0
    PASSED_TESTS=0
    
    # 测试用户注册
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_register; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试用户登录
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_login; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试限流中间件
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_rate_limit; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试JWT中间件（无token）
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_jwt_no_token; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试JWT中间件（有token）
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_jwt_with_token; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试受保护的端点
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_protected_endpoint; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试用户资料获取
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_profile; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试令牌刷新
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_token_refresh; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 测试用户登出
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if test_logout; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo
    
    # 显示测试结果
    print_info "测试完成！"
    print_info "总测试数: $TOTAL_TESTS"
    print_info "通过测试: $PASSED_TESTS"
    print_info "失败测试: $((TOTAL_TESTS - PASSED_TESTS))"
    
    if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
        print_success "所有测试都通过了！🎉"
    else
        print_warning "部分测试失败，请检查相关功能"
    fi
    
    # 清理测试数据
    cleanup
}

# 运行主函数
main "$@"
