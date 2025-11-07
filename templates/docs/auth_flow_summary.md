# 认证系统使用流程总结

## 📋 组件一览表

| 组件 | 生成时机 | 设置位置 | 验证时机 | 失效时机 |
|------|---------|---------|---------|---------|
| **JWT Token** | 登录成功<br>设置密码<br>修改密码 | Cookie (HttpOnly=true)<br>Header | 每次请求 | 过期(24h)<br>密码修改后（需切面） |
| **CSRF Token** | 登录成功 | Cookie (HttpOnly=false)<br>Header | POST/PUT/DELETE | 过期(24h)<br>登出时 |
| **State/Nonce** | OAuth 初始化 | Redis (TTL 5分钟) | OAuth 登录时 | 验证后删除<br>TTL 过期 |
| **密码修改时间戳** | 密码修改/重置 | Redis (永不过期) | 切面中间件 | 不失效 |

---

## 🔄 核心流程

### 1️⃣ 用户登录流程

```
用户登录请求
    ↓
验证凭据（密码/短信/第三方）
    ↓
生成 JWT Token
    ↓
生成 CSRF Token
    ↓
同时设置到响应：
  - jwt_token Cookie (HttpOnly=true, Secure=配置)
  - jwt_token Header
  - csrf_token Cookie (HttpOnly=false, Secure=配置)
  - csrf_token Header
    ↓
前端保存 Cookie，后续请求自动携带
```

**关键代码位置**：
- 生成：`app/controller/admin/api/user_controller.go:125, 128-134`
- 设置：`pkg/middleware/jwt_middleware.go:137-156`
- 设置：`pkg/middleware/csrf_middleware.go:90-116`

---

### 2️⃣ 普通 API 请求流程

```
请求到达
    ↓
[中间件链]
    ↓
1. RecoveryMiddleware
    - 错误恢复
    ↓
2. JWTMiddleware
    - 从 Cookie/Header 提取 token
    - 验证签名和过期时间
    - 解析用户信息 → context
    ↓
3. PasswordChangeValidatorMiddleware（可选）
    - 从 context 获取用户ID
    - 重新解析 token 获取 IssuedAt
    - 从 Redis 查询 last_pw_change
    - 验证 token.iat >= last_pw_change
    - ❌ 失败：拒绝请求
    ↓
4. CSRFMiddleware（仅写操作）
    - POST/PUT/DELETE 时验证
    - 从 Header 读取 csrf_token
    - 从 Cookie 读取 csrf_token
    - 验证两者相等
    - ❌ 失败：拒绝请求
    ↓
Controller 业务逻辑
```

**中间件配置**：
```go
Middleware: []router.MiddlewareFunc{
    middleware.RecoveryMiddleware(...),
    tmid.JWTMiddleware(),
    tmid.PasswordChangeValidatorMiddleware(),  // 可选
    tmid.CSRFMiddleware(),
}
```

---

### 3️⃣ OAuth 第三方登录流程

```
步骤 1：前端调用 OAuthInit
    POST /admin/user/oauth-init
    ↓
服务端生成 state 和 nonce
    ↓
存入 Redis（TTL 5分钟）：
  - oauth:state:{state} = "1"
  - oauth:nonce:{nonce} = "1"
    ↓
返回 {state, nonce} 给前端
    ↓
步骤 2：前端跳转第三方授权
    https://oauth-provider.com/authorize?state={state}&...
    ↓
用户授权
    ↓
步骤 3：第三方回调
    回调 URL?code={code}&state={state}
    ↓
前端调用登录接口
    POST /admin/user/login
    Body: {
        login_type: "wechat_work",
        code: "{code}",
        state: "{state}",
        nonce: "{nonce}"
    }
    ↓
服务端验证 state/nonce
    VerifyAndConsumeOAuthStateNonce()
    - 从 Redis 查找
    - 验证存在性
    - 立即删除（一次性使用）
    ↓
用 code 换取第三方用户信息
    ↓
生成 JWT + CSRF Token（同登录流程）
```

**关键代码位置**：
- 生成：`app/controller/admin/api/user_controller.go:366`
- 验证：`app/controller/admin/api/user_controller.go:109`
- 实现：`app/helper/store/auth_redis_store.go:28-53, 55-71`

---

### 4️⃣ 设置密码流程（首次/重置）

```
POST /admin/user/set-password
Header: jwt_token (Cookie自动)
Header: csrf_token (前端手动)
Body: {password: "new_password"}
    ↓
[中间件验证]
1. JWT 验证 ✅
2. CSRF 验证 ✅
    ↓
Controller: SetPassword()
- 从 JWT 获取用户ID
- 验证密码强度
    ↓
Service: SetPassword()
1. 验证密码强度
2. 生成 salt 和哈希密码
3. 事务中更新数据库
   - 不存在登录方式：创建
   - 存在登录方式：更新密码
    ↓
事务成功后：
1. InvalidateUserSessions(userID)
   → Redis: last_pw_change = now() ⚠️ 先更新
2. GenerateJWTToken()
   → 新 token.iat = now() >= last_pw_change ✅
    ↓
Controller: 设置新 token
- SetJWTToken(newToken)
- 返回响应（包含新 token）
    ↓
后续请求：
- 旧 token ❌ 失效（PasswordChangeValidatorMiddleware）
- 新 token ✅ 正常使用
```

**关键代码位置**：
- 控制器：`app/controller/admin/api/user_controller.go:166-201`
- 服务层：`app/service/admin_auth_service.go:538-616`
- 时间戳更新：`app/helper/store/auth_redis_store.go:73-79`

---

### 5️⃣ 修改密码流程

```
POST /admin/user/change-password
Header: jwt_token (Cookie自动)
Header: csrf_token (前端手动)
Body: {
    old_password: "old",
    new_password: "new"
}
    ↓
[中间件验证]
1. JWT 验证 ✅
2. CSRF 验证 ✅
    ↓
Controller: ChangePassword()
- 从 JWT 获取用户ID
- 验证旧密码和新密码强度
    ↓
Service: ChangePassword()
1. 验证新密码强度
2. 验证旧密码（防止未授权修改）⚠️
3. 生成新的 salt 和哈希密码
4. 事务中更新数据库密码
    ↓
事务成功后：
1. InvalidateUserSessions(userID)
   → Redis: last_pw_change = now() ⚠️ 先更新
2. GenerateJWTToken()
   → 新 token.iat = now() >= last_pw_change ✅
    ↓
Controller: 设置新 token
- SetJWTToken(newToken)
- 返回响应（包含新 token）
    ↓
后续请求：
- 旧 token ❌ 失效
- 新 token ✅ 正常使用
```

**关键代码位置**：
- 控制器：`app/controller/admin/api/user_controller.go:208-241`
- 服务层：`app/service/admin_auth_service.go:618-684`

---

### 6️⃣ 用户登出流程

```
POST /admin/user/logout
Header: jwt_token (Cookie自动)
Header: csrf_token (前端手动)
    ↓
[中间件验证]
1. JWT 验证 ✅
2. CSRF 验证 ✅
    ↓
Controller: Logout()
1. 从 JWT 获取用户ID
2. 记录登出日志
3. ClearJWTToken()
   → 清除 JWT Cookie 和 Header
4. ClearCSRFToken()
   → 清除 CSRF Cookie 和 Header
5. 返回成功响应
```

**关键代码位置**：
- `app/controller/admin/api/user_controller.go:377-408`
- `pkg/middleware/jwt_middleware.go:185-200`
- `pkg/middleware/csrf_middleware.go:116-131`

---

## 🔐 安全机制总结

### JWT Token
- ✅ **生成**：登录成功、设置密码、修改密码
- ✅ **验证**：签名、过期时间（JWT 中间件）
- ✅ **失效**：过期、密码修改后（切面中间件）
- ✅ **存储**：HttpOnly Cookie（防 XSS）

### CSRF Token
- ✅ **生成**：登录成功
- ✅ **验证**：Double Submit Cookie 模式
- ✅ **失效**：过期、登出
- ✅ **存储**：HttpOnly=false Cookie（前端需读取）

### State/Nonce
- ✅ **生成**：OAuth 初始化
- ✅ **验证**：OAuth 登录时
- ✅ **失效**：验证后删除、TTL 过期
- ✅ **存储**：Redis（TTL 5分钟）

### 密码修改时间戳
- ✅ **更新**：密码修改/重置成功后
- ✅ **验证**：切面中间件（可选）
- ✅ **存储**：Redis（永不过期）

---

## ⚙️ 中间件配置示例

### 完整配置（推荐）

```go
taurus.Container.Http.AddRouterGroup(router.RouteGroup{
    Prefix: "/admin/user",
    Middleware: []router.MiddlewareFunc{
        middleware.RecoveryMiddleware(func(err any, stack string) {
            fmt.Printf("Error: %v\nStack: %s\n", err, stack)
        }),
        tmid.JWTMiddleware(),                        // JWT 身份认证
        tmid.PasswordChangeValidatorMiddleware(),    // 密码修改时间戳验证（切面）
        tmid.CSRFMiddleware(),                       // CSRF 保护
    },
    Routes: [...] 
})
```

### 最小配置（不使用密码修改验证）

```go
taurus.Container.Http.AddRouterGroup(router.RouteGroup{
    Prefix: "/admin/user",
    Middleware: []router.MiddlewareFunc{
        tmid.JWTMiddleware(),
        tmid.CSRFMiddleware(),
    },
    Routes: [...] 
})
```

---

## 📝 前端使用示例

### 发送 POST 请求（需 CSRF）

```javascript
// 1. 从 Cookie 读取 CSRF Token
function getCSRFToken() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'csrf_token') return value;
    }
    return '';
}

// 2. 发送请求（自动携带 JWT Cookie，手动设置 CSRF Header）
fetch('/admin/user/set-password', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'csrf_token': getCSRFToken()  // 必须手动设置
    },
    credentials: 'include',  // 确保携带 Cookie
    body: JSON.stringify({
        password: 'new_password'
    })
});
```

### OAuth 登录流程

```javascript
// 1. 获取 state/nonce
const { state, nonce } = await fetch('/admin/user/oauth-init')
    .then(r => r.json());

// 2. 跳转第三方授权
window.location.href = `https://oauth-provider.com/authorize?state=${state}`;

// 3. 回调后登录
await fetch('/admin/user/login', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'csrf_token': getCSRFToken()
    },
    body: JSON.stringify({
        login_type: 'wechat_work',
        code: oauthCode,
        state: state,
        nonce: nonce
    })
});
```

---

## ✅ 完整流程检查清单

### 登录流程 ✅
- [x] 验证凭据成功
- [x] 生成并设置 JWT Token
- [x] 生成并设置 CSRF Token
- [x] OAuth 登录时验证 state/nonce

### 请求流程 ✅
- [x] JWT 中间件验证 token 签名和过期时间
- [x] 密码修改验证切面（可选）验证时间戳
- [x] CSRF 中间件验证写操作

### 密码修改流程 ✅
- [x] 验证旧密码（仅修改密码）
- [x] 事务中更新数据库
- [x] 先更新 Redis 时间戳
- [x] 再签发新 JWT Token
- [x] 设置新 token 到 Cookie 和 Header

### 登出流程 ✅
- [x] 清除 JWT Token
- [x] 清除 CSRF Token
- [x] 记录登出日志

---

## 🎯 总结

所有组件已正确实现并配合工作：
1. ✅ **JWT**：身份认证，无状态
2. ✅ **CSRF**：写操作保护，Double Submit Cookie
3. ✅ **State/Nonce**：OAuth 安全防护，一次性使用
4. ✅ **密码修改时间戳**：使旧 token 失效（通过切面中间件）

**关键点**：
- JWT 和 CSRF 在登录时同时生成
- 密码修改后通过切面中间件使旧 token 失效
- 中间件顺序很重要：JWT → PasswordChangeValidator → CSRF

