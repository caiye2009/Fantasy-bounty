# Fantasy Bounty 后端系统架构总结

## 📋 目录
- [系统概览](#系统概览)
- [路由架构](#路由架构)
- [数据模型](#数据模型)
- [认证机制](#认证机制)
- [核心服务](#核心服务)

---

## 系统概览

### 系统定位
本系统为悬赏竞标平台后端服务，服务于以下三个终端：

| 终端 | 用户类型 | 认证方式 | 说明 |
|-----|---------|---------|------|
| **供应商H5** | 外部供应商 | 手机号验证码 | 供应商参与竞标、认证申请 |
| **供应商小程序** | 外部供应商 | 手机号验证码 | 同H5功能 |
| **数据分析平台H5** | 内部员工 | 工号密码（老系统） | 数据查看、管理功能 |

### 技术栈
- **框架**: Gin (Go Web Framework)
- **数据库**: PostgreSQL (通过GORM)
- **认证**: JWT (golang-jwt/jwt)
- **加密**: AES-256-GCM + HMAC-SHA256

---

## 路由架构

### 路由结构总览

```
/api/v1/
├── auth/                      # 认证路由（公开访问）
│   ├── POST /send-code        # 发送验证码（供应商）
│   ├── POST /verify-code      # 验证码登录（供应商）
│
├── supplier/                  # 供应商业务路由（需外部JWT）
│   ├── bids/
│   │   ├── POST ""            # 创建竞标
│   │   ├── GET ""             # 竞标列表
│   │   ├── GET "/my"          # 我的竞标
│   │   └── DELETE "/:id"      # 删除竞标
│   │
│   ├── suppliers/
│   │   ├── GET ""             # 供应商列表
│   │   ├── GET "/:id"         # 供应商详情
│   │   ├── POST "/recognize"  # 营业执照OCR识别
│   │   ├── POST "/apply"      # 提交供应商认证
│   │   └── GET "/my"          # 我的认证状态
│   │
│   └── users/
│       ├── POST ""            # 创建用户
│       ├── GET ""             # 用户列表
│       ├── GET "/:id"         # 用户详情
│       ├── PUT "/:id"         # 更新用户
│       └── DELETE "/:id"      # 删除用户
│
└── internal/                  # 内部系统路由（透传到老系统）
    ├── POST /login            # 内部用户登录（转发到老系统）
    └── ANY /*path             # 其他所有API（透传）

/uploads/*                     # 静态文件服务（营业执照）
/health                        # 健康检查
```

### 路由组说明

#### 1. Auth 路由组 (`/api/v1/auth`)
- **中间件**: 审计
- **认证**: 不需要JWT
- **用途**: 供应商登录注册

#### 2. Supplier 路由组 (`/api/v1/supplier`)
- **中间件**: JWT认证 + 审计
- **认证**: 外部JWT (本系统生成)
- **用途**: 供应商业务操作

#### 3. Internal 路由组 (`/api/v1/internal`)
- **中间件**: 审计
- **认证**: 无JWT认证（直接透传老系统token）
- **用途**:
  - `/login` - 登录请求转发到老系统
  - `/*path` - 其他请求透传到老系统

---

## 数据模型

### 1. User (用户表)

```go
type User struct {
    ID             string         // UUID
    Username       string         // 唯一用户名（自动生成8位base62随机码）
    PhoneHash      string         // 手机号哈希（查询索引）
    PhoneEncrypted string         // 手机号加密存储
    Phone          string         // 解密后的手机号（不存数据库）
    PhoneMasked    string         // 脱敏手机号（不存数据库）
    Status         string         // active / disabled
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt
    LastLoginAt    *time.Time
}
```

**表名**: `users`

**索引**:
- `username` - 唯一索引
- `phone_hash` - 唯一索引

**说明**:
- 手机号使用双重保护：Hash用于查询，加密存储用于展示
- 用户名自动生成：8位base62（0-9a-zA-Z），共62^8≈218万亿种组合
- 数据库唯一索引兜底，冲突时自动重试1次（百万用户下冲突率<0.046%）

---

### 2. Supplier (供应商表)

```go
type Supplier struct {
    ID                   string    // UUID
    Name                 string    // 公司名称
    BusinessLicenseNo    string    // 营业执照号
    BusinessLicenseImage string    // 营业执照图片路径
    VerifiedAt           time.Time // 认证通过时间
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt
}
```

**表名**: `suppliers`

**说明**:
- 仅存储已通过审核的供应商
- 营业执照图片存储在 `uploads/` 目录

---

### 3. UserSupplier (用户-供应商绑定表)

```go
type UserSupplier struct {
    ID         string    // UUID
    Username   string    // 关联用户名
    SupplierID string    // 关联供应商ID
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  gorm.DeletedAt
}
```

**表名**: `user_suppliers`

**索引**:
- `username` - 唯一索引（一个用户只能绑定一个供应商）

---

### 4. SupplierApplication (供应商认证申请表)

```go
type SupplierApplication struct {
    ID                   string     // UUID
    Username             string     // 申请用户
    Name                 string     // 公司名称
    BusinessLicenseNo    string     // 营业执照号
    BusinessLicenseImage string     // 营业执照图片
    Status               string     // pending / approved / rejected
    RejectReason         *string    // 拒绝原因
    ReviewedAt           *time.Time // 审核时间
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt
}
```

**表名**: `supplier_applications`

**索引**:
- `username` - 索引

**状态流转**:
```
pending (待审核) → approved (已通过) / rejected (已拒绝)
```

---

### 5. Bid (竞标表)

```go
type Bid struct {
    ID        string    // UUID
    BountyID  uint      // 悬赏ID
    Username  string    // 竞标用户
    BidPrice  float64   // 竞标价格
    Status    string    // 状态
    CreatedAt time.Time
    UpdatedAt time.Time

    // 关联规格（二选一）
    WovenSpec   *BidWovenSpec
    KnittedSpec *BidKnittedSpec
}
```

**表名**: `bids`

**索引**:
- `bounty_id` - 索引
- `username` - 索引

**状态枚举**:
- `pending` - 审核中
- `in_progress` - 进行中
- `pending_acceptance` - 待验收
- `completed` - 已完成

---

### 6. BidWovenSpec (竞标-梭织规格)

```go
type BidWovenSpec struct {
    ID                 uint
    BidID              string    // 关联竞标ID
    SizeLength         float64   // 尺码（长度）
    GreigeFabricType   string    // 胚布类型（现货/定织）
    GreigeDeliveryDate time.Time // 胚布交期
    DeliveryMethod     string    // 交货方式
}
```

**表名**: `bid_woven_specs`

**索引**:
- `bid_id` - 唯一索引

---

### 7. BidKnittedSpec (竞标-针织规格)

```go
type BidKnittedSpec struct {
    ID                 uint
    BidID              string    // 关联竞标ID
    SizeWeight         float64   // 尺码（重量/皮重）
    GreigeFabricType   string    // 胚布类型（现货/定织）
    GreigeDeliveryDate time.Time // 胚布交期
    DeliveryMethod     string    // 交货方式
}
```

**表名**: `bid_knitted_specs`

**索引**:
- `bid_id` - 唯一索引

---

### 8. AuditLog (审计日志表)

```go
type AuditLog struct {
    ID         string    // UUID
    RequestID  string    // 请求ID
    Username   string    // 操作用户
    Action     string    // 操作动作
    Resource   string    // 资源类型
    ResourceID string    // 资源ID
    Method     string    // HTTP方法
    Path       string    // 请求路径
    StatusCode int       // 响应状态码
    ClientIP   string    // 客户端IP
    UserAgent  string    // User Agent
    Duration   int64     // 耗时（毫秒）
    Detail     string    // 详情（JSONB）
    CreatedAt  time.Time
}
```

**表名**: `audit_logs`

**索引**:
- `request_id` - 索引
- `username` - 索引
- `action` - 索引
- `created_at` - 索引

---

## 认证机制

### 1. 供应商认证流程

```
┌─────────────────┐
│  1. 发送验证码   │  POST /api/v1/auth/send-code
│  { phone }       │  → 生成6位验证码，存储1分钟
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  2. 验证码登录   │  POST /api/v1/auth/verify-code
│  { phone, code } │  → 验证码校验
└────────┬────────┘  → 查询/创建用户
         │           → 生成JWT (username)
         ▼
┌─────────────────┐
│  3. 获得JWT      │  { token: "jwt_token", username: "a7B9cD2e", isNewUser: true }
│  存储到本地      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  4. 访问业务API  │  Header: Authorization: Bearer {token}
│  /supplier/*     │  → JWT中间件验证
└─────────────────┘  → 提取username到context
```

**JWT Payload**:
```json
{
  "username": "a7B9cD2e",
  "iss": "fantasy-bounty",
  "iat": 1234567890,
  "exp": 1234654290
}
```

**有效期**: 24小时（可通过 `JWT_EXPIRY_HOURS` 配置）

---

### 2. 内部用户认证流程

```
┌─────────────────┐
│  1. 输入工号密码 │
│  username        │
│  password        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  2. 登录请求     │  POST /api/v1/internal/login
│  本系统转发      │  → 转发到老系统 POST /auth/login
└────────┬────────┘  ← 老系统返回 { token, ... }
         │           ← 本系统透传返回
         ▼
┌─────────────────┐
│  3. 获得老系统    │  { token, username, ... }
│  JWT token      │  存储到本地
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  4. 访问内部API  │  Header: Authorization: Bearer {老系统token}
│  /internal/*     │  → 本系统直接透传到老系统
└─────────────────┘  → 老系统验证token并返回数据
```

**说明**:
- 本系统不存储内部用户信息
- 本系统不验证老系统token
- 完全依赖老系统的认证授权

---

### 3. Token转换机制（供应商访问内部数据）

某些供应商业务可能需要访问老系统数据，此时使用默认账号自动换取老系统token：

```
┌─────────────────┐
│  供应商请求      │  Header: Authorization: Bearer {外部JWT}
│  /supplier/xxx   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  后端处理        │  → 验证外部JWT
│  需要内部数据    │  → 用默认账号登录老系统（internal_token.Manager）
└────────┬────────┘  → 获取老系统token（自动缓存）
         │           → 调用老系统API
         ▼
┌─────────────────┐
│  返回数据        │  ← 老系统响应
│                 │  ← 本系统处理后返回给供应商
└─────────────────┘
```

**配置**:
```env
INTERNAL_API_URL=https://old-system.com
INTERNAL_AUTH_PATH=/auth/login
INTERNAL_USERNAME=service_account
INTERNAL_PASSWORD=******
```

---

## 核心服务

### 1. JWT服务 (`pkg/jwt`)

**功能**:
- 生成JWT token
- 验证JWT token

**接口**:
```go
type JWTService struct {
    secretKey []byte
    issuer    string
    expiry    time.Duration
}

func (s *JWTService) GenerateToken(username string) (string, error)
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error)
```

---

### 2. 加密服务 (`pkg/crypto`)

**功能**:
- AES-256-GCM 加密/解密（手机号）
- HMAC-SHA256 哈希（手机号索引）
- 手机号脱敏

**接口**:
```go
type Crypto struct {
    key    []byte
    pepper string
}

func (c *Crypto) Encrypt(plaintext string) (string, error)
func (c *Crypto) Decrypt(ciphertext string) (string, error)
func (c *Crypto) Hash(data string) string
func MaskPhone(phone string) string  // 138****8000
```

---

### 3. 内部Token管理器 (`pkg/internal_token`)

**功能**:
- 自动登录老系统获取token
- Token缓存（过期前30分钟自动刷新）
- 并发安全

**接口**:
```go
type Manager struct {
    mu        sync.RWMutex
    token     string
    expiresAt time.Time
    // ...
}

func (m *Manager) GetToken() (string, error)
```

**刷新策略**:
- 距离过期 > 30分钟：返回缓存token
- 距离过期 ≤ 30分钟：自动刷新
- 刷新失败且旧token未过期：继续使用旧token

---

### 4. 内部系统代理 (`pkg/proxy`)

**功能**:
- 反向代理到老系统
- 路径重写（去除 `/api/v1/internal` 前缀）
- 支持两种模式：
  - 透传模式：前端传了Authorization，直接透传
  - 转换模式：前端未传Authorization，用默认账号token

**接口**:
```go
type InternalProxy struct {
    tokenManager *internal_token.Manager
    targetURL    *url.URL
    proxy        *httputil.ReverseProxy
}

func (p *InternalProxy) Handler() gin.HandlerFunc
```

---

### 5. 审计服务 (`internal/audit`)

**功能**:
- 异步记录审计日志
- 批量写入数据库（每5秒或100条）
- 优雅关闭

**接口**:
```go
type Service interface {
    Start()
    Stop()
    Log(log *AuditLog)
}
```

**审计中间件**:
```go
func Audit(auditService audit.Service) gin.HandlerFunc
```

---

### 6. 用户服务 (`internal/user`)

**功能**:
- 用户CRUD
- 手机号加密/解密
- 用户名自动生成

**接口**:
```go
type Service interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    GetUser(ctx context.Context, id string) (*User, error)
    GetUserByUsername(ctx context.Context, username string) (*User, error)
    GetUserByPhone(ctx context.Context, phone string) (*User, error)
    UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error)
    DeleteUser(ctx context.Context, id string) error
    ListUsers(ctx context.Context, page, pageSize int) ([]User, int64, error)
    UpdateLastLogin(ctx context.Context, id string) error
}
```

---

### 7. 供应商服务 (`internal/supplier`)

**功能**:
- 营业执照OCR识别
- 供应商认证申请
- 认证状态查询
- 供应商CRUD

**接口**:
```go
type Service interface {
    RecognizeLicense(imagePath string) (*OCRResult, error)
    ApplySupplier(ctx context.Context, username string, req *ApplySupplierRequest, imagePath string) (*SupplierApplication, error)
    GetMySupplierStatus(ctx context.Context, username string) (*MySupplierStatus, error)
    IsVerifiedSupplier(ctx context.Context, username string) (bool, error)
    // ...
}
```

---

### 8. 竞标服务 (`internal/bid`)

**功能**:
- 创建竞标（需供应商认证）
- 竞标列表查询
- 我的竞标查询
- 删除竞标

**接口**:
```go
type Service interface {
    CreateBid(ctx context.Context, username string, req *CreateBidRequest) (*Bid, error)
    ListBids(ctx context.Context, page, pageSize int, status string) ([]Bid, int64, error)
    ListMyBids(ctx context.Context, username string, page, pageSize int) ([]Bid, int64, error)
    DeleteBid(ctx context.Context, id string, username string) error
    GetBid(ctx context.Context, id string) (*Bid, error)
}
```

---

## 中间件

### 1. RequestContext 中间件

**功能**: 为每个请求创建上下文，存储请求信息供审计使用

**字段**:
```go
type RequestContext struct {
    RequestID  string
    Username   string
    Action     string
    Resource   string
    ResourceID string
    Detail     map[string]any
}
```

---

### 2. JWTAuth 中间件

**功能**:
- 验证JWT token
- 提取username到RequestContext
- 401响应处理

**使用**:
```go
protected.Use(middleware.JWTAuth(jwtService))
```

---

### 3. Audit 中间件

**功能**:
- 记录请求信息（IP、UserAgent、Method、Path等）
- 记录响应信息（StatusCode、Duration）
- 异步写入审计日志

**使用**:
```go
group.Use(middleware.Audit(auditService))
```

---

## 环境变量配置

### 数据库配置
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=fantasy_bounty
```

### JWT配置
```env
JWT_SECRET=your-secret-key-here
JWT_ISSUER=fantasy-bounty
JWT_EXPIRY_HOURS=24
```

### 加密配置
```env
CRYPTO_KEY=1234567890123456    # 16/24/32字节
HASH_PEPPER=your-pepper-here
```

### 老系统配置
```env
INTERNAL_API_URL=https://old-system.com
INTERNAL_AUTH_PATH=/auth/login
INTERNAL_USERNAME=service_account
INTERNAL_PASSWORD=******
```

---

## 安全特性

### 1. 手机号保护
- ✅ 存储：AES-256-GCM加密
- ✅ 查询：HMAC-SHA256哈希索引
- ✅ 展示：脱敏处理（138****8000）

### 2. 验证码安全
- ⚠️ 当前：内存存储（1分钟有效）
- 💡 建议：改用Redis存储

### 3. Token安全
- ✅ JWT有效期控制（24小时）
- ✅ HMAC-SHA256签名
- ✅ 自动刷新机制（老系统token）

### 4. 审计完整
- ✅ 所有API操作记录
- ✅ 包含请求详情、响应状态、耗时
- ✅ 异步写入，不影响性能

### 5. 权限隔离
- ✅ 供应商：仅访问 `/supplier/*`，需外部JWT
- ✅ 内部用户：访问 `/internal/*`，使用老系统token
- ✅ 数据隔离：供应商无法访问内部系统

---

## 版本信息

- **Go版本**: 1.21+
- **Gin版本**: v1.9+
- **GORM版本**: v1.25+
- **PostgreSQL版本**: 14+

---

## 更新日志

### v1.0.0 (当前版本)
- ✅ 供应商手机号验证码登录
- ✅ 供应商认证申请流程
- ✅ 竞标创建与管理
- ✅ 内部系统登录代理
- ✅ 内部系统API透传
- ✅ 完整审计日志
- ✅ 手机号加密存储
- ✅ Token自动刷新机制

---

## 待优化项

1. **验证码存储**: 从内存改为Redis
2. **限流防护**: 添加接口限流中间件
3. **文件上传**: 优化文件大小限制和格式校验
4. **OCR服务**: 接入真实OCR API
5. **监控告警**: 添加Prometheus metrics
6. **日志管理**: 结构化日志输出

---

*文档生成时间: 2026-02-10*
*维护者: Fantasy Team*
