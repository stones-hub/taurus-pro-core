# 完整认证系统使用流程总结

## 一、组件概览

### 1. JWT Token（用户身份认证）
- **职责**：标识用户身份，无状态认证
- **存储**：Cookie (HttpOnly=true) + Header（可选）
- **生成时机**：登录成功、设置密码、修改密码
- **验证时机**：每次请求通过 JWT 中间件验证

### 2. CSRF Token（防止跨站请求伪造）
- **职责**：防止 CSRF 攻击，采用 Double Submit Cookie 模式
- **存储**：Cookie (HttpOnly=false) + Header（前端手动设置）
- **生成时机**：登录成功时
- **验证时机**：POST/PUT/PATCH/DELETE 请求通过 CSRF 中间件验证

### 3. State/Nonce（OAuth 第三方登录防护）
- **职责**：防止 OAuth 授权流程中的重放攻击和中间人攻击
- **存储**：Redis（TTL 5分钟）
- **生成时机**：OAuth 初始化接口
- **验证时机**：OAuth 登录回调时

### 4. 密码修改/重置时间戳（Token 失效机制）
- **职责**：记录密码修改时间，使旧 token 失效
- **存储**：Redis（永不过期）
- **更新时机**：密码修改/重置成功后
- **验证时机**：通过密码修改验证切面中间件验证

---

## 二、完整使用流程

### 场景 1：用户登录流程

```
┌─────────────┐
│   用户登录   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: Login()                 │
│  - 验证请求方法 (POST)               │
│  - 解析请求参数                      │
│  - 根据 login_type 分发              │
└──────┬──────────────────────────────┘
       │
       ├─→ 用户名密码登录
       │   └→ AdminAuthService.PasswordLogin()
       │
       ├─→ 手机验证码登录
       │   └→ AdminAuthService.SmsLogin()
       │
       └─→ 第三方登录（OAuth）
           ├─→ 验证 state/nonce（必须）
           └→ AdminAuthService.OauthLoginByCode()
       │
       ▼
┌─────────────────────────────────────┐
│  服务层：验证凭据成功                │
│  - 验证密码/验证码/授权码            │
│  - 创建/查找用户                     │
│  - 生成 JWT Token                    │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: 登录成功处理            │
│  1. SetJWTToken()                    │
│     → 设置到 Cookie (HttpOnly=true)   │
│     → 设置到响应 Header              │
│                                      │
│  2. GenerateCSRFToken()              │
│     → 生成随机字符串                  │
│     → SetCSRFToken()                 │
│       → 设置到 Cookie (HttpOnly=false) │
│       → 设置到响应 Header            │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  前端接收响应                        │
│  - 浏览器自动保存 JWT Cookie         │
│  - 浏览器保存 CSRF Cookie             │
│  - 后续请求自动携带 Cookie            │
└─────────────────────────────────────┘
```

### 场景 2：普通 API 请求流程（GET/POST/PUT/DELETE）

```
┌─────────────┐
│   API 请求   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  中间件链（按顺序执行）               │
│                                      │
│  1. RecoveryMiddleware               │
│     → 错误恢复                       │
│                                      │
│  2. JWTMiddleware                   │
│     → 从 Cookie/Header 提取 token   │
│     → 验证签名和过期时间              │
│     → 解析用户信息到 context         │
│     ⚠️ 不验证密码修改时间戳          │
│                                      │
│  3. PasswordChangeValidatorMiddleware│
│     → 从 context 获取用户信息         │
│     → 重新解析 token 获取 IssuedAt   │
│     → 从 Redis 查询 last_pw_change   │
│     → 验证 token.iat >= last_pw_change│
│     → 失败则拒绝请求                  │
│                                      │
│  4. CSRFMiddleware (仅写操作)         │
│     → POST/PUT/DELETE 请求时：       │
│       - 从 Header 读取 csrf_token    │
│       - 从 Cookie 读取 csrf_token    │
│       - 验证两者相等                  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: 业务逻辑处理            │
│  - 从 context 获取用户ID/用户名      │
│  - 执行业务逻辑                      │
└─────────────────────────────────────┘
```

### 场景 3：OAuth 第三方登录完整流程

```
┌─────────────┐
│ 用户点击第三方登录│
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  前端：调用 /admin/user/oauth-init   │
│  → GET/POST 请求（无需认证）        │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: OAuthInit()            │
│  → CreateOAuthStateNonce()           │
│    - 生成 state（16字节随机）        │
│    - 生成 nonce（16字节随机）        │
│    - 存入 Redis (TTL 5分钟)          │
│  → 返回 {state, nonce} 给前端        │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  前端跳转第三方授权页面              │
│  - 携带 state 参数                  │
│  - 用户授权                         │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  第三方回调（携带 code + state）    │
│  前端调用登录接口，携带：            │
│  - code: 授权码                     │
│  - state: 防 CSRF                   │
│  - nonce: 防重放                    │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: Login()                │
│  → VerifyAndConsumeOAuthStateNonce() │
│    - 从 Redis 查找 state/nonce       │
│    - 验证存在性                      │
│    - 立即删除（一次性使用）          │
│  → OauthLoginByCode()               │
│    - 用 code 换取第三方用户信息      │
│    - 创建/绑定用户                   │
│    - 生成 JWT + CSRF Token           │
└─────────────────────────────────────┘
```

### 场景 4：设置密码流程（首次设置/重置）

```
┌─────────────┐
│ 用户设置密码 │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  请求：POST /admin/user/set-password │
│  Header: jwt_token (Cookie自动)      │
│  Header: csrf_token (前端手动设置)   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  中间件验证                          │
│  1. JWT 中间件：验证 token 签名      │
│  2. CSRF 中间件：验证 CSRF token     │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: SetPassword()          │
│  - 从 JWT 获取用户ID                 │
│  - 验证新密码强度                    │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Service: SetPassword()             │
│  1. 验证密码强度                     │
│  2. 生成 salt 和哈希密码              │
│  3. 事务中更新数据库                 │
│     - 如果登录方式不存在，创建        │
│     - 如果登录方式存在，更新密码      │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  事务成功后：                        │
│  1. InvalidateUserSessions(userID)    │
│     → 更新 Redis:                    │
│        user:last_pw_change:{userID}  │
│        = 当前时间戳（秒级）           │
│                                      │
│  2. GenerateJWTToken()                │
│     → 生成新 JWT Token               │
│     → 新 token.iat >= last_pw_change │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: 返回响应                │
│  1. SetJWTToken(newToken)            │
│     → 更新 Cookie 和 Header          │
│  2. 返回成功响应（包含新 token）     │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  后续请求                           │
│  - 旧 token 会自动失效               │
│    (PasswordChangeValidatorMiddleware)│
│  - 新 token 正常工作                 │
└─────────────────────────────────────┘
```

### 场景 5：修改密码流程

```
┌─────────────┐
│ 用户修改密码 │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  请求：POST /admin/user/change-password│
│  Body: {old_password, new_password} │
│  Header: jwt_token (Cookie自动)      │
│  Header: csrf_token (前端手动设置)   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: ChangePassword()        │
│  - 从 JWT 获取用户ID                 │
│  - 验证旧密码和新密码强度            │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Service: ChangePassword()           │
│  1. 验证新密码强度                   │
│  2. 验证旧密码（防止未授权修改）     │
│  3. 生成新的 salt 和哈希密码          │
│  4. 事务中更新数据库密码             │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  事务成功后：                        │
│  1. InvalidateUserSessions(userID)    │
│     → 更新 Redis 时间戳               │
│                                      │
│  2. GenerateJWTToken()                │
│     → 生成新 JWT Token               │
│     → 新 token.iat >= last_pw_change │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: 返回响应                │
│  1. SetJWTToken(newToken)            │
│  2. 返回成功响应（包含新 token）     │
└─────────────────────────────────────┘
```

### 场景 6：用户登出流程

```
┌─────────────┐
│   用户登出   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  请求：POST /admin/user/logout       │
│  Header: jwt_token (Cookie自动)      │
│  Header: csrf_token (前端手动设置)   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  中间件验证                          │
│  1. JWT 中间件：验证 token           │
│  2. CSRF 中间件：验证 CSRF token     │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Controller: Logout()               │
│  1. 从 JWT 获取用户ID                 │
│  2. 记录登出日志                     │
│  3. ClearJWTToken()                  │
│     → 清除 JWT Cookie 和 Header      │
│  4. ClearCSRFToken()                 │
│     → 清除 CSRF Cookie 和 Header     │
│  5. 返回成功响应                     │
└─────────────────────────────────────┘
```

---

## 三、中间件执行顺序（重要）

### 推荐顺序

```go
Middleware: []router.MiddlewareFunc{
    // 1. 错误恢复（最外层）
    middleware.RecoveryMiddleware(...),
    
    // 2. JWT 验证（身份认证）
    tmid.JWTMiddleware(),
    
    // 3. 密码修改时间戳验证（可选，按需启用）
    tmid.PasswordChangeValidatorMiddleware(),
    
    // 4. CSRF 验证（写操作保护）
    tmid.CSRFMiddleware(),
}
```

### 执行逻辑

1. **RecoveryMiddleware**：捕获 panic，防止服务崩溃
2. **JWTMiddleware**：
   - 验证 token 签名和过期时间
   - 提取用户信息到 context
3. **PasswordChangeValidatorMiddleware**（可选）：
   - 验证 token 是否在密码修改之后签发
   - 使密码修改后的旧 token 失效
4. **CSRFMiddleware**：
   - 仅对 POST/PUT/PATCH/DELETE 验证
   - 验证 Header 和 Cookie 中的 CSRF token 是否匹配

---

## 四、关键时序关系

### 密码修改后的 Token 失效机制

```
时间轴：
T0: 用户登录，签发 token1 (iat = T0)
T1: 用户修改密码
    → InvalidateUserSessions() 
    → Redis: last_pw_change = T1
    → 签发 token2 (iat = T1, T1 >= T1 ✅)
T2: 用户使用 token1 请求
    → PasswordChangeValidatorMiddleware
    → token1.iat (T0) < last_pw_change (T1) ❌
    → 拒绝请求："JWT令牌已失效，请重新登录"
T3: 用户使用 token2 请求
    → token2.iat (T1) >= last_pw_change (T1) ✅
    → 验证通过
```

### State/Nonce 的时效性

```
时间轴：
T0: 调用 OAuthInit
    → 生成 state/nonce
    → Redis: oauth:state:{state} (TTL 5分钟)
    → Redis: oauth:nonce:{nonce} (TTL 5分钟)
T1: 跳转第三方授权（通常 < 1分钟）
T2: 第三方回调，验证 state/nonce
    → 验证成功，立即删除（一次性使用）
T3: 再次使用相同的 state/nonce
    → 验证失败（已删除）
```

---

## 五、前端配合要点

### 1. JWT Token 使用

```javascript
// 前端不需要手动处理 JWT Token
// 浏览器自动通过 Cookie 发送
// HttpOnly=true，JavaScript 无法访问（安全）
```

### 2. CSRF Token 使用

```javascript
// 1. 从 Cookie 读取 CSRF Token（登录后自动设置）
function getCSRFToken() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'csrf_token') {
            return value;
        }
    }
    return '';
}

// 2. 每次 POST/PUT/DELETE 请求时放入 Header
fetch('/api/xxx', {
    method: 'POST',
    headers: {
        'csrf_token': getCSRFToken(),  // 从 Cookie 读取并放入 Header
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
});
```

### 3. OAuth 登录流程

```javascript
// 1. 跳转第三方前，先获取 state/nonce
const response = await fetch('/admin/user/oauth-init');
const { state, nonce } = await response.json();

// 2. 跳转第三方授权页面（携带 state）
window.location.href = `https://oauth-provider.com/authorize?state=${state}&...`;

// 3. 第三方回调后，携带 code + state + nonce 调用登录接口
const loginResponse = await fetch('/admin/user/login', {
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

### 4. 处理 Token 失效

```javascript
// 当收到 "JWT令牌已失效" 错误时
if (error.message.includes('JWT令牌已失效')) {
    // 1. 清除本地存储
    // 2. 清除 Cookie（如果需要）
    // 3. 跳转登录页面
    window.location.href = '/login';
}
```

---

## 六、安全性检查清单

### JWT Token 安全 ✅
- ✅ HttpOnly Cookie（防止 XSS）
- ✅ SameSite=Lax（防止 CSRF）
- ✅ 密码修改后旧 token 失效（通过切面中间件）
- ⚠️ 生产环境应启用 Secure（HTTPS only）

### CSRF Token 安全 ✅
- ✅ Double Submit Cookie 模式
- ✅ 仅保护写操作
- ✅ HttpOnly=false（前端需要读取）
- ✅ SameSite=Lax

### OAuth State/Nonce 安全 ✅
- ✅ State 防 CSRF
- ✅ Nonce 防重放
- ✅ TTL 自动过期（5分钟）
- ✅ 一次性使用（验证后立即删除）

### 密码安全 ✅
- ✅ 密码强度验证
- ✅ 密码哈希存储（带 salt）
- ✅ 密码修改后更新时间戳
- ✅ 签发新 token 前更新时间戳（确保新 token.iat >= last_pw_change）

---

## 七、配置要求

### 必需配置

```yaml
# config/autoload/http/http.yaml
http:
  jwt:
    enabled: true              # 启用 JWT
    secret: "your-secret-key"  # JWT 密钥（必须配置）
    issuer: "your-app-name"    # JWT 签发者
    expire_hours: 24           # Token 过期时间（小时）
    secure: false              # 生产环境建议设为 true
```

### Redis 配置

需要 Redis 用于存储：
- OAuth state/nonce（TTL 5分钟）
- 密码修改时间戳（永不过期）

---

## 八、常见问题

### Q1: 密码修改后，为什么旧 token 还能用？

**A**: 需要在路由配置中添加 `PasswordChangeValidatorMiddleware` 切面中间件。

### Q2: CSRF Token 什么时候刷新？

**A**: CSRF Token 只在登录时生成，与 JWT 同时过期。不需要单独刷新。

### Q3: State/Nonce 过期了怎么办？

**A**: TTL 5分钟通常足够完成 OAuth 流程。如果过期，前端需要重新调用 `OAuthInit` 获取新的 state/nonce。

### Q4: Redis 故障时怎么办？

**A**: 
- State/Nonce 验证：OAuth 登录会失败（可接受）
- 密码修改时间戳验证：自动跳过验证（避免所有请求被拒绝）

---

## 九、最佳实践

1. ✅ **中间件顺序**：Recovery → JWT → PasswordChangeValidator → CSRF
2. ✅ **按需启用**：PasswordChangeValidator 可在敏感路由中启用
3. ✅ **错误处理**：前端应妥善处理 token 失效错误
4. ✅ **安全配置**：生产环境启用 Secure 和 HTTPS
5. ✅ **监控日志**：记录密码修改、登录失败等关键操作

