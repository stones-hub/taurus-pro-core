#!/bin/bash

# User API 测试脚本
# 测试所有User Controller的函数

BASE_URL="http://localhost:9080"
LOG_FILE="user_api_test.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a $LOG_FILE
}

success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a $LOG_FILE
}

error() {
    echo -e "${RED}❌ $1${NC}" | tee -a $LOG_FILE
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a $LOG_FILE
}

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    log "测试: $description"
    log "请求: $method $BASE_URL$endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method -d "$data" "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        success "成功 - HTTP $http_code"
        log "响应: $body"
    else
        error "失败 - HTTP $http_code"
        log "响应: $body"
    fi
    
    echo "----------------------------------------"
}

# 清空日志文件
> $LOG_FILE

log "开始User API测试..."
log "基础URL: $BASE_URL"
echo ""

# ==================== 基础CRUD操作测试 ====================

log "=== 基础CRUD操作测试 ==="

# 1. 创建用户
test_api "POST" "/user/create" "name=测试用户1&password=test123" "创建用户"
test_api "POST" "/user/create" "name=测试用户2&password=test456" "创建第二个用户"
test_api "POST" "/user/create" "name=张三&password=zhang123" "创建中文用户"


# 2. 获取用户
test_api "GET" "/user/get?id=1" "" "根据ID获取用户"
test_api "GET" "/user/getByName?name=测试用户1" "" "根据用户名获取用户"
test_api "GET" "/user/getByName?name=张三" "" "根据中文用户名获取用户"

# 3. 更新用户
test_api "POST" "/user/updateName" "id=1&name=测试用户1(已更新)" "更新用户名"
test_api "POST" "/user/updatePassword" "id=1&password=newpassword123" "更新用户密码"

# 4. 查询操作
test_api "GET" "/user/all" "" "获取所有用户"
test_api "GET" "/user/page?page=1&pageSize=5" "" "分页查询用户"
test_api "GET" "/user/count" "" "获取用户总数"
test_api "GET" "/user/stats" "" "获取用户统计信息"

# ==================== 高级查询测试 ====================

log "=== 高级查询测试 ==="

# 搜索功能
test_api "GET" "/user/search?keyword=测试" "" "搜索包含'测试'的用户"
test_api "GET" "/user/search?keyword=张" "" "搜索包含'张'的用户"
test_api "GET" "/user/search?keyword=不存在&page=1&pageSize=10" "" "搜索不存在的用户"

# 模糊查询
test_api "GET" "/user/like?name=测试" "" "模糊查询包含'测试'的用户"
test_api "GET" "/user/like?name=张" "" "模糊查询包含'张'的用户"

# 最近用户
test_api "GET" "/user/recent?days=7" "" "获取最近7天创建的用户"
test_api "GET" "/user/recent?days=30" "" "获取最近30天创建的用户"

# ==================== 存在性检查测试 ====================

log "=== 存在性检查测试 ==="

test_api "GET" "/user/exists?id=1" "" "检查用户ID=1是否存在"
test_api "GET" "/user/exists?id=999" "" "检查不存在的用户ID"
test_api "GET" "/user/existsByName?name=测试用户1" "" "检查用户名'测试用户1'是否存在"
test_api "GET" "/user/existsByName?name=不存在的用户" "" "检查不存在的用户名"

# ==================== 批量操作测试 ====================

log "=== 批量操作测试 ==="

test_api "POST" "/user/createBatch" "" "批量创建用户"
test_api "POST" "/user/deleteBatch?ids=1,2,3" "" "批量删除用户"

# ==================== SQL操作测试 ====================

log "=== SQL操作测试 ==="

# 执行SQL
test_api "POST" "/user/executeSQL" "sql=UPDATE users SET username = 'SQL更新用户' WHERE id = 1" "执行更新SQL"
test_api "POST" "/user/executeSQL" "sql=INSERT INTO users (username, password) VALUES ('SQL插入用户', 'sql123')" "执行插入SQL"

# 查询SQL
test_api "GET" "/user/querySQL?sql=SELECT * FROM users LIMIT 5" "" "SQL查询所有用户"
test_api "GET" "/user/querySQL?sql=SELECT COUNT(*) as count FROM users" "" "SQL查询用户总数"
test_api "GET" "/user/querySQL?sql=SELECT username, created_at FROM users WHERE username LIKE '%测试%'" "" "SQL模糊查询"

# ==================== 业务逻辑测试 ====================

log "=== 业务逻辑测试 ==="

# 如果不存在则创建
test_api "POST" "/user/createIfNotExists" "name=新用户&password=new123" "如果不存在则创建新用户"
test_api "POST" "/user/createIfNotExists" "name=测试用户1&password=test123" "尝试创建已存在的用户"

# ==================== 错误处理测试 ====================

log "=== 错误处理测试 ==="

# 参数缺失测试
test_api "POST" "/user/create" "name=测试用户" "创建用户时缺少密码"
test_api "POST" "/user/create" "password=test123" "创建用户时缺少用户名"
test_api "GET" "/user/get" "" "获取用户时缺少ID参数"
test_api "GET" "/user/getByName" "" "根据用户名获取用户时缺少name参数"

# 无效参数测试
test_api "GET" "/user/get?id=invalid" "" "使用无效的ID参数"
test_api "POST" "/user/updateName" "id=invalid&name=新名字" "使用无效的ID更新用户名"

# 不存在的资源测试
test_api "GET" "/user/get?id=999" "" "获取不存在的用户"
test_api "POST" "/user/updateName" "id=999&name=新名字" "更新不存在的用户"
test_api "POST" "/user/delete" "id=999" "删除不存在的用户"

# ==================== 边界条件测试 ====================

log "=== 边界条件测试 ==="

# 分页边界测试
test_api "GET" "/user/page?page=0&pageSize=10" "" "分页查询第0页"
test_api "GET" "/user/page?page=1&pageSize=0" "" "分页查询页面大小为0"
test_api "GET" "/user/page?page=999&pageSize=10" "" "分页查询不存在的页面"

# 空字符串测试
test_api "GET" "/user/search?keyword=" "" "搜索空关键词"
test_api "GET" "/user/like?name=" "" "模糊查询空名称"

# ==================== 最终清理测试 ====================

log "=== 最终清理测试 ==="

# 删除测试用户
test_api "POST" "/user/delete" "id=1" "删除测试用户1"
test_api "POST" "/user/delete" "id=2" "删除测试用户2"

# 最终统计
test_api "GET" "/user/count" "" "最终用户总数"
test_api "GET" "/user/stats" "" "最终统计信息"

# ==================== 测试总结 ====================

log "=== 测试总结 ==="
log "所有User API测试完成！"
log "详细日志请查看: $LOG_FILE"

echo ""
echo -e "${GREEN}🎉 User API测试完成！${NC}"
echo -e "${BLUE}📋 测试日志: $LOG_FILE${NC}"
echo ""
echo -e "${YELLOW}测试覆盖的API端点:${NC}"
echo "✅ /user/create - 创建用户"
echo "✅ /user/get - 根据ID获取用户"
echo "✅ /user/getByName - 根据用户名获取用户"
echo "✅ /user/updateName - 更新用户名"
echo "✅ /user/updatePassword - 更新用户密码"
echo "✅ /user/delete - 删除用户"
echo "✅ /user/all - 获取所有用户"
echo "✅ /user/page - 分页查询用户"
echo "✅ /user/search - 搜索用户"
echo "✅ /user/count - 获取用户总数"
echo "✅ /user/stats - 获取用户统计信息"
echo "✅ /user/like - 模糊查询用户"
echo "✅ /user/recent - 获取最近用户"
echo "✅ /user/exists - 检查用户是否存在"
echo "✅ /user/existsByName - 检查用户名是否存在"
echo "✅ /user/createBatch - 批量创建用户"
echo "✅ /user/deleteBatch - 批量删除用户"
echo "✅ /user/executeSQL - 执行原生SQL"
echo "✅ /user/querySQL - SQL查询用户"
echo "✅ /user/createIfNotExists - 如果不存在则创建用户"
echo ""
echo -e "${GREEN}总计测试了 20 个API端点！${NC}" 