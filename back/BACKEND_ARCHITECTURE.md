# Fantasy-Bounty 后端架构文档

## 目录

1. [项目结构](#1-项目结构)
2. [请求处理流程](#2-请求处理流程)
3. [核心模块详解](#3-核心模块详解)
4. [公共工具包 (pkg)](#4-公共工具包-pkg)
5. [数据库设计](#5-数据库设计)
6. [认证与授权](#6-认证与授权)
7. [API 路由一览](#7-api-路由一览)
8. [设计模式](#8-设计模式)
9. [环境变量配置](#9-环境变量配置)

---

## 1. 项目结构

```
back/
├── cmd/
│   └── main.go                    # 应用程序入口
├── config/
│   ├── init.go                    # 服务器初始化和启动
│   ├── router.go                  # 路由配置
│   └── database.go                # 数据库配置和迁移
├── internal/                      # 内部业务模块
│   ├── auth/                      # 认证模块
│   ├── user/                      # 用户模块
│   ├── bid/                       # 竞标模块
│   ├── supplier/                  # 供应商认证模块
│   └── audit/                     # 审计日志模块
├── pkg/                           # 公共工具包
│   ├── middleware/                # 中间件
│   ├── jwt/                       # JWT服务
│   ├── crypto/                    # 加密服务
│   ├── internal_token/            # 内部系统Token管理
│   └── proxy/                     # 反向代理
├── uploads/                       # 静态文件目录
│   └── licenses/                  # 营业执照图片
├── go.mod
├── Dockerfile
└── Makefile
```

### 技术栈

| 组件 | 技术选型 |
|------|----------|
| Web 框架 | Gin (v1.11.0) |
| ORM | GORM (v1.31.1) |
| 数据库 | PostgreSQL |
| 认证 | JWT (HS256) |
| 加密 | AES-GCM |

---

## 2. 请求处理流程

### 2.1 启动流程

```
cmd/main.go
    │
    └── config.Init()
            │
            ├── InitDatabase()      // 连接 PostgreSQL，自动迁移表
            ├── SetupRouter()       // 配置路由和中间件
            ├── server.ListenAndServe()  // 启动 HTTP 服务 (:8080)
            └── 优雅关闭            // 处理 SIGINT/SIGTERM，5秒超时
```

### 2.2 请求处理链

```
HTTP 请求
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│  全局中间件                                                   │
│  ┌─────────────────┐                                        │
│  │ gin.Recovery()  │  捕获 panic，返回 500                   │
│  └────────┬────────┘                                        │
│           ▼                                                 │
│  ┌─────────────────────────────┐                            │
│  │ RequestContextMiddleware() │  生成 RequestID、记录开始时间 │
│  └────────┬────────────────────┘                            │
└───────────┼─────────────────────────────────────────────────┘
            │
            ▼
    ┌───────────────┐
    │  路由匹配      │
    └───────┬───────┘
            │
    ┌───────┴───────┐
    │               │
    ▼               ▼
公开路由         受保护路由
(/auth/*)        (/bids, /users, /suppliers, /internal)
    │               │
    │               ▼
    │       ┌─────────────┐
    │       │ JWTAuth()   │  验证 JWT Token，填充 Username
    │       └──────┬──────┘
    │              ▼
    │       ┌─────────────┐
    │       │ Audit()     │  记录审计日志（异步）
    │       └──────┬──────┘
    │              │
    └──────┬───────┘
           ▼
    ┌─────────────────┐
    │  Handler 处理器  │  业务逻辑处理
    └────────┬────────┘
             ▼
    ┌─────────────────┐
    │  Service 服务层  │  业务规则、数据聚合
    └────────┬────────┘
             ▼
    ┌─────────────────┐
    │ Repository 仓库  │  数据库操作
    └────────┬────────┘
             ▼
        HTTP 响应
```

### 2.3 请求示例：创建竞标

```
POST /api/v1/bids
Authorization: Bearer <jwt_token>

1. RequestContextMiddleware
   └── 生成 RequestID、记录 ClientIP、UserAgent、StartTime

2. JWTAuth 中间件
   ├── 从 Header 提取 Token
   ├── 验证签名和过期时间
   └── 填充 Username 到 RequestContext

3. Audit 中间件（请求前）
   └── 等待后续处理完成

4. bidHandler.CreateBid()
   ├── 获取 Username
   ├── 检查供应商认证状态 → supplierService.IsUserVerified()
   ├── 解析请求参数
   ├── 调用 bidService.CreateBid()
   │   ├── 生成 UUID
   │   ├── 创建规格（WovenSpec/KnittedSpec）
   │   └── 写入数据库
   └── 返回 201 响应

5. Audit 中间件（请求后）
   └── 异步记录审计日志到数据库
```

---

## 3. 核心模块详解

### 3.1 auth 模块 - 认证

**职责**：手机号验证码登录/注册

#### Handler 函数

| 函数 | 路由 | 说明 |
|------|------|------|
| `SendCode` | POST /auth/send-code | 发送 6 位验证码（1分钟有效） |
| `VerifyCode` | POST /auth/verify-code | 验证码校验，自动注册/登录 |

#### 数据模型

```go
// 请求
type SendCodeRequest struct {
    Phone string `json:"phone" binding:"required"`
}

type VerifyCodeRequest struct {
    Phone string `json:"phone" binding:"required"`
    Code  string `json:"code" binding:"required"`
}

// 响应
type VerifyCodeResponse struct {
    Code      int    `json:"code"`
    Message   string `json:"message"`    // "登录成功" 或 "注册成功"
    Token     string `json:"token"`       // JWT Token（只包含 username）
    Username  string `json:"username"`
    IsNewUser bool   `json:"isNewUser"`
}
```

#### 验证码存储

- 内存存储（sync.Map）
- 有效期：1 分钟
- 后台协程定期清理过期验证码

---

### 3.2 user 模块 - 用户管理

**职责**：用户的 CRUD 操作

#### Handler 函数

| 函数 | 路由 | 说明 |
|------|------|------|
| `CreateUser` | POST /users | 创建用户 |
| `GetUser` | GET /users/:id | 获取用户详情（仅可查看自己） |
| `UpdateUser` | PUT /users/:id | 更新用户状态（仅可更新自己） |
| `DeleteUser` | DELETE /users/:id | 删除用户（仅可删除自己） |
| `ListUsers` | GET /users | 获取用户列表（分页） |

#### Service 函数

```go
type Service interface {
    CreateUser(ctx, req) (*User, error)
        // 生成唯一用户名，加密手机号

    GetUser(ctx, id) (*User, error)
        // 按ID获取，自动解密手机号

    GetUserByUsername(ctx, username) (*User, error)
        // 按用户名获取

    GetUserByPhone(ctx, phone) (*User, error)
        // 按手机号哈希查询

    UpdateUser(ctx, id, req) (*User, error)
        // 更新状态

    DeleteUser(ctx, id) error
        // 软删除

    ListUsers(ctx, page, pageSize) ([]User, int64, error)
        // 分页查询，自动解密手机号

    UpdateLastLogin(ctx, id) error
        // 更新最后登录时间
}
```

#### 数据模型

```go
type User struct {
    ID             string         `gorm:"primaryKey;type:uuid"`
    Username       string         `gorm:"uniqueIndex"`           // 自动生成："用户"+5位随机字符
    PhoneHash      string         `gorm:"uniqueIndex"`           // SHA256 哈希（用于查询）
    PhoneEncrypted string         `gorm:"column:phone_encrypted"` // AES-GCM 加密
    Phone          string         `gorm:"-"`                      // 虚拟字段：解密后
    PhoneMasked    string         `gorm:"-"`                      // 虚拟字段：脱敏
    Status         string         `gorm:"default:active"`         // active | disabled
    LastLoginAt    *time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt `gorm:"index"`                  // 软删除
}
```

---

### 3.3 bid 模块 - 竞标管理

**职责**：竞标的创建、查询、删除

#### Handler 函数

| 函数 | 路由 | 说明 |
|------|------|------|
| `CreateBid` | POST /bids | 创建竞标（需供应商认证） |
| `ListBids` | GET /bids?bounty_id=xxx | 按赏金ID获取竞标列表 |
| `ListMyBids` | GET /bids/my | 获取我的竞标列表 |
| `DeleteBid` | DELETE /bids/:id | 删除竞标（仅可删除自己的） |

#### Service 函数

```go
type Service interface {
    CreateBid(ctx, username, req) (*Bid, error)
        // 创建竞标，设置初始状态为 pending

    GetBid(ctx, id) (*Bid, error)
        // 获取竞标（含关联规格）

    DeleteBid(ctx, id) error

    ListBidsByBountyID(ctx, bountyID, page, pageSize) ([]Bid, int64, error)
        // 按赏金ID查询

    ListBidsByUsername(ctx, username, status, page, pageSize) ([]Bid, int64, error)
        // 按用户名查询（支持状态筛选）
}
```

#### 数据模型

```go
// 竞标状态常量
const (
    BidStatusPending           = "pending"            // 审核中
    BidStatusInProgress        = "in_progress"        // 进行中
    BidStatusPendingAcceptance = "pending_acceptance" // 待验收
    BidStatusCompleted         = "completed"          // 已完成
)

type Bid struct {
    ID          string          `gorm:"primaryKey;type:uuid"`
    BountyID    uint            `gorm:"index"`                    // 关联赏金ID
    Username    string          `gorm:"index;type:varchar(20)"`   // 关联用户名
    BidPrice    float64                                           // 投标价格
    Status      string          `gorm:"default:pending"`
    WovenSpec   *BidWovenSpec   `gorm:"foreignKey:BidID"`        // 梭织规格（一对一）
    KnittedSpec *BidKnittedSpec `gorm:"foreignKey:BidID"`        // 针织规格（一对一）
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type BidWovenSpec struct {
    ID                 uint      `gorm:"primaryKey"`
    BidID              string    `gorm:"uniqueIndex;type:uuid"`
    SizeLength         float64                                  // 尺码（长度）
    GreigeFabricType   string                                   // 胚布类型：现货 | 定织
    GreigeDeliveryDate time.Time                                // 胚布交期
    DeliveryMethod     string                                   // 交货方式
}

type BidKnittedSpec struct {
    ID                 uint      `gorm:"primaryKey"`
    BidID              string    `gorm:"uniqueIndex;type:uuid"`
    SizeWeight         float64                                  // 尺码（重量）
    GreigeFabricType   string
    GreigeDeliveryDate time.Time
    DeliveryMethod     string
}
```

---

### 3.4 supplier 模块 - 供应商认证

**职责**：供应商认证申请、OCR识别、认证状态查询

#### Handler 函数

| 函数 | 路由 | 说明 |
|------|------|------|
| `GetSupplier` | GET /suppliers/:id | 获取供应商详情 |
| `ListSuppliers` | GET /suppliers | 获取供应商列表（分页） |
| `RecognizeLicense` | POST /suppliers/recognize | 上传营业执照 OCR 识别 |
| `ApplySupplier` | POST /suppliers/apply | 提交供应商认证申请 |
| `GetMySupplierStatus` | GET /suppliers/my | 获取我的供应商认证状态 |

#### Service 函数

```go
type Service interface {
    // 供应商查询
    GetSupplier(ctx, id) (*Supplier, error)
    ListSuppliers(ctx, page, pageSize) ([]Supplier, int64, error)

    // 认证申请
    ApplySupplier(ctx, username, req, licenseImage) (*SupplierApplication, error)
        // 提交申请，检查重复申请、已认证状态

    GetMySupplierStatus(ctx, username) (*MySupplierStatus, error)
        // 返回：已认证供应商、待审核申请、最近被拒申请

    // OCR 识别
    RecognizeLicense(ctx, imagePath) (*OCRResult, error)
        // 识别营业执照（当前为模拟）

    // 供应商认证校验（供其他模块调用）
    IsUserVerified(ctx, username) (bool, error)
        // 检查用户是否已完成供应商认证
}
```

#### 数据模型

```go
// 申请状态常量
const (
    ApplicationStatusPending  = "pending"   // 待审核
    ApplicationStatusApproved = "approved"  // 已通过
    ApplicationStatusRejected = "rejected"  // 已拒绝
)

type Supplier struct {
    ID                   string `gorm:"primaryKey;type:uuid"`
    Name                 string                               // 供应商名称
    BusinessLicenseNo    string                               // 营业执照号
    BusinessLicenseImage string                               // 营业执照图片路径
    VerifiedAt           time.Time                            // 认证通过时间
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt `gorm:"index"`
}

type UserSupplier struct {
    ID         string `gorm:"primaryKey;type:uuid"`
    Username   string `gorm:"uniqueIndex;type:varchar(20)"` // 一个用户只能绑定一个供应商
    SupplierID string `gorm:"type:uuid"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type SupplierApplication struct {
    ID                   string  `gorm:"primaryKey;type:uuid"`
    Username             string  `gorm:"index;type:varchar(20)"`
    Name                 string                               // 申请的供应商名称
    BusinessLicenseNo    string
    BusinessLicenseImage string
    Status               string  `gorm:"default:pending"`     // pending | approved | rejected
    RejectReason         *string                              // 拒绝原因
    ReviewedAt           *time.Time                           // 审核时间
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt `gorm:"index"`
}

type OCRResult struct {
    CompanyName       string `json:"companyName"`
    BusinessLicenseNo string `json:"businessLicenseNo"`
    LegalPerson       string `json:"legalPerson"`
    RegisteredCapital string `json:"registeredCapital"`
    EstablishDate     string `json:"establishDate"`
    BusinessScope     string `json:"businessScope"`
    Address           string `json:"address"`
}

type MySupplierStatus struct {
    HasVerifiedSupplier bool                 `json:"hasVerifiedSupplier"`
    Supplier            *Supplier            `json:"supplier,omitempty"`
    PendingApplication  *SupplierApplication `json:"pendingApplication,omitempty"`
    LatestRejected      *SupplierApplication `json:"latestRejected,omitempty"`
}
```

#### 供应商认证流程

```
1. 上传营业执照
   POST /suppliers/recognize
   ├── 上传图片（multipart/form-data，field: license）
   ├── 验证文件类型（jpg/jpeg/png/pdf）
   ├── 保存到 uploads/licenses/<uuid>.<ext>
   ├── 调用 OCR 识别
   └── 返回识别结果 + imagePath

2. 提交认证申请
   POST /suppliers/apply
   ├── 请求体：{ name, businessLicenseNo, imagePath }
   ├── 校验：不允许重复申请、不允许已认证再申请
   └── 创建 SupplierApplication (status=pending)

3. 后台审核（待实现）
   ├── 批准 → 创建 Supplier + UserSupplier 绑定
   └── 拒绝 → 设置 rejectReason

4. 查询认证状态
   GET /suppliers/my
   └── 返回 MySupplierStatus
```

---

### 3.5 audit 模块 - 审计日志

**职责**：记录所有受保护路由的操作日志

#### Service 函数

```go
type Service interface {
    Log(entry *AuditLog)
        // 异步记录审计日志（非阻塞）

    Start()
        // 启动后台日志写入协程

    Stop()
        // 关闭通道并等待写入完成
}
```

#### 数据模型

```go
type AuditLog struct {
    ID         string    `gorm:"primaryKey;type:uuid"`
    RequestID  string    `gorm:"index"`                  // 请求唯一ID
    Username   string    `gorm:"index"`                  // 用户名
    Action     string    `gorm:"index"`                  // 操作类型（如 bid.create）
    Resource   string                                    // 资源类型（如 bid）
    ResourceID string                                    // 资源ID
    Method     string                                    // HTTP 方法
    Path       string                                    // 请求路径
    StatusCode int                                       // HTTP 状态码
    ClientIP   string
    UserAgent  string
    Duration   int64                                     // 处理时间（毫秒）
    Detail     string                                    // JSON 详细信息
    CreatedAt  time.Time `gorm:"index"`
}
```

#### 异步写入机制

```
请求 → Audit 中间件 → auditService.Log(entry)
                           │
                           ▼
                    ┌─────────────┐
                    │ 缓冲通道     │  容量：256
                    │ (非阻塞)     │
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ 后台协程     │  从通道读取并写入数据库
                    └─────────────┘
```

---

## 4. 公共工具包 (pkg)

### 4.1 jwt 包 - JWT 服务

```go
type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

type JWTService struct {
    secretKey []byte
    issuer    string
    expiry    time.Duration
}

// 函数
NewJWTService(secretKey, issuer string, expiry time.Duration) *JWTService

GenerateToken(username string) (string, error)
    // 生成 JWT Token（签名算法：HS256）
    // 只包含 username，不包含 user_id

ValidateToken(tokenString string) (*Claims, error)
    // 验证并解析 Token
```

### 4.2 crypto 包 - 加密服务

```go
type Crypto struct {
    key []byte  // AES 密钥（16/24/32 字节）
}

// 函数
NewCrypto(key string) (*Crypto, error)

Encrypt(plaintext string) (string, error)
    // AES-GCM 加密
    // 流程：生成随机 nonce → 加密 → nonce 附加在密文前 → Base64 编码

Decrypt(ciphertext string) (string, error)
    // AES-GCM 解密

MaskPhone(phone string) string
    // 手机号脱敏：138****5678

Hash(data string) string
    // SHA256 哈希
```

### 4.3 middleware 包 - 中间件

#### RequestContextMiddleware

```go
func RequestContextMiddleware() gin.HandlerFunc
    // 为每个请求初始化 RequestContext：
    // - RequestID (UUID)
    // - ClientIP
    // - UserAgent
    // - StartTime
```

#### JWTAuth

```go
func JWTAuth(jwtService *jwt.JWTService) gin.HandlerFunc
    // JWT 认证中间件：
    // 1. 从 Authorization Header 提取 Token（Bearer <token>）
    // 2. 验证 Token 有效性
    // 3. 将 Username 填入 RequestContext
    // 4. 无效则返回 401，abort 请求
```

#### Audit

```go
func Audit(auditService audit.Service) gin.HandlerFunc
    // 审计中间件：
    // 1. 等待后续处理完成
    // 2. 构建 AuditLog（自动生成 Action）
    // 3. 异步写入
```

#### RequestContext

```go
type RequestContext struct {
    RequestID  string
    ClientIP   string
    UserAgent  string
    StartTime  time.Time
    Username   string         // JWT 中间件填充
    Action     string         // Handler 可选设置
    Resource   string         // Handler 可选设置
    ResourceID string         // Handler 可选设置
    Detail     map[string]any // Handler 可选设置
}

SetRequestContext(c *gin.Context, rc *RequestContext)
GetRequestContext(c *gin.Context) *RequestContext
```

### 4.4 internal_token 包 - 内部系统 Token 管理

```go
type Manager struct {
    mu        sync.RWMutex
    token     string
    expiresAt time.Time
    apiURL    string
    authPath  string
    username  string
    password  string
    httpClient *http.Client
}

// 函数
NewManager(apiURL, authPath, username, password string) *Manager

GetToken() (string, error)
    // 获取内部系统 Token
    // 特性：
    // - 自动缓存
    // - 距离过期 <30 分钟自动刷新
    // - Double-check 防止并发重复刷新
    // - 刷新失败但旧 Token 未过期时降级使用旧 Token
```

### 4.5 proxy 包 - 反向代理

```go
type InternalProxy struct {
    tokenManager *internal_token.Manager
    targetURL    *url.URL
    proxy        *httputil.ReverseProxy
}

// 函数
NewInternalProxy(tokenManager *Manager, targetURL string) *InternalProxy

Handler() gin.HandlerFunc
    // 反向代理处理器：
    // 1. 获取内部系统 Token
    // 2. 替换 Authorization Header
    // 3. 删除 Cookie 等干扰 Header
    // 4. 转发请求到内部系统
```

---

## 5. 数据库设计

### 5.1 ER 关系图

```
┌─────────────────┐
│     users       │
├─────────────────┤
│ ID (PK, UUID)   │
│ Username (UK)   │──────────────┐
│ PhoneHash       │              │
│ PhoneEncrypted  │              │
│ Status          │              │
│ LastLoginAt     │              │
│ CreatedAt       │              │
│ UpdatedAt       │              │
│ DeletedAt       │              │
└─────────────────┘              │
         │                       │
         │ 1:N (by username)     │ 1:1 (by username)
         ▼                       ▼
┌─────────────────┐     ┌───────────────────┐
│      bids       │     │  user_suppliers   │
├─────────────────┤     ├───────────────────┤
│ ID (PK, UUID)   │     │ ID (PK, UUID)     │
│ BountyID        │     │ Username (UK)     │───┐
│ Username (FK)   │     │ SupplierID (FK)   │   │
│ BidPrice        │     │ CreatedAt         │   │
│ Status          │     │ UpdatedAt         │   │
│ CreatedAt       │     │ DeletedAt         │   │
│ UpdatedAt       │     └───────────────────┘   │
└────────┬────────┘              │              │
         │                       │              │
    1:1  │ 1:1                   │ N:1          │
         ▼                       ▼              │
┌────────────────────┐   ┌─────────────────┐   │
│  bid_woven_specs   │   │    suppliers    │◄──┘
├────────────────────┤   ├─────────────────┤
│ ID (PK)            │   │ ID (PK, UUID)   │
│ BidID (UK, UUID)   │   │ Name            │
│ SizeLength         │   │ BusinessLicenseNo│
│ GreigeFabricType   │   │ BusinessLicenseImage│
│ GreigeDeliveryDate │   │ VerifiedAt      │
│ DeliveryMethod     │   │ CreatedAt       │
└────────────────────┘   │ UpdatedAt       │
                         │ DeletedAt       │
┌────────────────────┐   └─────────────────┘
│ bid_knitted_specs  │
├────────────────────┤   ┌───────────────────────┐
│ ID (PK)            │   │ supplier_applications │
│ BidID (UK, UUID)   │   ├───────────────────────┤
│ SizeWeight         │   │ ID (PK, UUID)         │
│ GreigeFabricType   │   │ Username (INDEX)      │
│ GreigeDeliveryDate │   │ Name                  │
│ DeliveryMethod     │   │ BusinessLicenseNo     │
└────────────────────┘   │ BusinessLicenseImage  │
                         │ Status                │
┌─────────────────┐      │ RejectReason          │
│   audit_logs    │      │ ReviewedAt            │
├─────────────────┤      │ CreatedAt             │
│ ID (PK, UUID)   │      │ UpdatedAt             │
│ RequestID (IDX) │      │ DeletedAt             │
│ Username (IDX)  │      └───────────────────────┘
│ Action (IDX)    │
│ Resource        │
│ ResourceID      │
│ Method          │
│ Path            │
│ StatusCode      │
│ ClientIP        │
│ UserAgent       │
│ Duration        │
│ Detail (JSONB)  │
│ CreatedAt (IDX) │
└─────────────────┘
```

### 5.2 索引设计

| 表 | 索引 | 类型 | 用途 |
|---|---|---|---|
| users | Username | UNIQUE | 用户名唯一 |
| users | PhoneHash | UNIQUE | 手机号快速查询 |
| users | DeletedAt | INDEX | 软删除查询 |
| bids | BountyID | INDEX | 按赏金查询竞标 |
| bids | Username | INDEX | 按用户查询竞标 |
| bid_woven_specs | BidID | UNIQUE | 一对一关联 |
| bid_knitted_specs | BidID | UNIQUE | 一对一关联 |
| user_suppliers | Username | UNIQUE | 一个用户只能绑定一个供应商 |
| supplier_applications | Username | INDEX | 查询用户申请 |
| audit_logs | RequestID | INDEX | 请求追踪 |
| audit_logs | Username | INDEX | 用户行为分析 |
| audit_logs | Action | INDEX | 操作类型统计 |
| audit_logs | CreatedAt | INDEX | 时间范围查询 |

---

## 6. 认证与授权

### 6.1 认证流程

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 发送验证码                                                 │
│    POST /api/v1/auth/send-code                               │
│    ├── 生成 6 位随机验证码                                    │
│    ├── 存储到内存（有效期 1 分钟）                            │
│    └── 打印到控制台（模拟短信）                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. 验证码登录/注册                                            │
│    POST /api/v1/auth/verify-code                             │
│    ├── 验证验证码是否存在且未过期                             │
│    ├── 查询用户是否存在（通过 PhoneHash）                     │
│    │   ├── 不存在 → 自动注册（创建 User）                     │
│    │   └── 存在 → 登录                                       │
│    ├── 检查用户状态（active/disabled）                        │
│    ├── 更新 LastLoginAt                                      │
│    └── 生成 JWT Token（只包含 username）                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ JWT Token 结构                                               │
│ {                                                            │
│   "username": "<username>",                                  │
│   "iss": "bounty-backend",                                   │
│   "iat": <issued_at>,                                        │
│   "exp": <expiry_time>                                       │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 授权规则

| 资源 | 操作 | 权限要求 |
|---|---|---|
| /bids | POST (创建) | 登录 + 供应商认证 |
| /bids | GET (列表) | 登录 |
| /bids/my | GET | 登录（只能看自己的） |
| /bids/:id | DELETE | 登录 + 只能删除自己的 |
| /users/:id | GET/PUT/DELETE | 登录 + 只能操作自己的 |
| /suppliers | GET | 登录 |
| /suppliers/apply | POST | 登录 + 无待审核申请 + 未认证供应商 |
| /internal/* | ANY | 登录 |

### 6.3 手机号安全存储

```
手机号处理流程：
┌─────────────────┐
│ 原始手机号       │
│ 13812345678     │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌────────┐  ┌────────────────┐
│ SHA256 │  │ AES-GCM 加密    │
│ 哈希   │  │ + Base64 编码  │
└────┬───┘  └───────┬────────┘
     │              │
     ▼              ▼
┌────────────┐  ┌────────────────────┐
│ PhoneHash  │  │ PhoneEncrypted     │
│ (用于查询)  │  │ (用于存储和解密)    │
└────────────┘  └────────────────────┘
```

---

## 7. API 路由一览

### 7.1 公开路由（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /health | 健康检查 |
| POST | /api/v1/auth/send-code | 发送验证码 |
| POST | /api/v1/auth/verify-code | 验证码登录/注册 |
| GET | /uploads/* | 静态文件（营业执照图片） |

### 7.2 受保护路由（需要 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| **竞标** | | |
| POST | /api/v1/bids | 创建竞标（需供应商认证） |
| GET | /api/v1/bids | 获取竞标列表（?bounty_id=xxx） |
| GET | /api/v1/bids/my | 获取我的竞标列表 |
| DELETE | /api/v1/bids/:id | 删除竞标 |
| **用户** | | |
| POST | /api/v1/users | 创建用户 |
| GET | /api/v1/users | 获取用户列表 |
| GET | /api/v1/users/:id | 获取用户详情 |
| PUT | /api/v1/users/:id | 更新用户 |
| DELETE | /api/v1/users/:id | 删除用户 |
| **供应商** | | |
| GET | /api/v1/suppliers | 获取供应商列表 |
| GET | /api/v1/suppliers/:id | 获取供应商详情 |
| POST | /api/v1/suppliers/recognize | OCR 识别营业执照 |
| POST | /api/v1/suppliers/apply | 提交供应商认证申请 |
| GET | /api/v1/suppliers/my | 获取我的供应商认证状态 |
| **内部代理** | | |
| ANY | /api/v1/internal/* | 转发到内部系统 |

---

## 8. 设计模式

### 8.1 三层架构

```
┌─────────────────────────────────────────┐
│           Handler 层（处理器）            │
│  - HTTP 请求解析和响应                   │
│  - 参数校验                              │
│  - 权限检查                              │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│           Service 层（服务）             │
│  - 业务逻辑实现                          │
│  - 业务规则校验                          │
│  - 数据聚合和转换                        │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│         Repository 层（仓库）            │
│  - 数据库 CRUD 操作                      │
│  - SQL 查询封装                          │
└─────────────────────────────────────────┘
```

### 8.2 依赖注入

```go
// config/router.go
cryptoService := crypto.NewCrypto(os.Getenv("CRYPTO_KEY"))
jwtService := jwt.NewJWTService(secret, issuer, expiry)

userRepo := user.NewRepository(db)
userService := user.NewService(userRepo, cryptoService)
userHandler := user.NewHandler(userService)

authHandler := auth.NewHandler(jwtService, userService)
```

### 8.3 接口隔离

```go
// Service 和 Repository 都定义为 interface，便于测试和解耦
type Service interface {
    CreateBid(ctx context.Context, ...) (*Bid, error)
    // ...
}

type Repository interface {
    Create(ctx context.Context, bid *Bid) error
    // ...
}
```

### 8.4 资源隔离

```go
// 通过 RequestContext 的 Username 实现用户资源隔离
func (h *Handler) GetUser(c *gin.Context) {
    rc := middleware.GetRequestContext(c)
    // 获取用户信息后检查 username 是否匹配
    if user.Username != rc.Username {
        c.JSON(403, gin.H{"message": "Permission denied"})
        return
    }
    // ...
}
```

---

## 9. 环境变量配置

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=bounty_db

# 服务器配置
GIN_MODE=debug          # debug | release | test
PORT=8080

# JWT 配置
JWT_SECRET=your_jwt_secret_key   # 必填，建议 32 字符以上
JWT_ISSUER=bounty-backend        # 默认值
JWT_EXPIRY_HOURS=24              # 默认 24 小时

# 加密配置
CRYPTO_KEY=1234567890123456      # 16/24/32 字节（AES-128/192/256）

# 内部系统代理配置
INTERNAL_API_URL=http://internal-api:8080
INTERNAL_AUTH_PATH=/auth/login
INTERNAL_USERNAME=internal_user
INTERNAL_PASSWORD=internal_password
```

---

## 附录：错误码

| HTTP 状态码 | 场景 |
|-------------|------|
| 400 | 请求参数错误、业务校验失败 |
| 401 | 未认证（缺少或无效的 Token） |
| 403 | 权限不足（未供应商认证、操作他人资源） |
| 404 | 资源不存在 |
| 409 | 资源冲突（重复申请等） |
| 500 | 服务器内部错误 |
