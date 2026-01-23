# Internal 包模型字段总结

## 项目概述

纺织行业悬赏竞标平台后端，采用分层架构：
- **Handler** — HTTP 请求/响应处理
- **Service** — 业务逻辑
- **Repository** — 数据访问 (GORM)
- **Model** — 数据结构定义

技术栈：Go + Gin + GORM + PostgreSQL + Elasticsearch + JWT

---

## 1. auth 包 (认证)

### 业务逻辑

1. 用户请求发送验证码 → 生成6位随机数字，存内存 (RWMutex 保护，1分钟过期)
2. 用户提交手机号+验证码 → 校验验证码是否匹配且未过期
3. 通过 phone_hash 查询账号是否存在：
   - 不存在 → 自动创建账号 (加密手机号、生成6位用户名、status=active)
   - 存在 → 校验账号状态是否为 active
4. 生成 JWT Token (含 userID + username)
5. 更新 last_login_at

> 注: 验证码存储当前为内存实现，生产环境应替换为 Redis

### SendCodeRequest 发送验证码请求
> **作用**: 客户端请求向指定手机号发送短信验证码时使用的请求体

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Phone | string | phone | 手机号 (必填) |

### VerifyCodeRequest 验证码登录/注册请求
> **作用**: 用户提交手机号和验证码进行登录或自动注册时使用的请求体

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Phone | string | phone | 手机号 (必填) |
| Code | string | code | 验证码 (必填) |

### VerifyCodeResponse 验证码登录/注册响应
> **作用**: 验证码验证成功后的响应体，包含JWT令牌、账号ID及是否为新用户标识

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Code | int | code | 响应码 |
| Message | string | message | 响应消息 |
| Token | string | token | JWT Token |
| AccountID | string | accountId | 账号ID |
| IsNewUser | bool | isNewUser | 是否新用户（首次注册） |
| Username | string | username | 用户名 |

### ErrorResponse 错误响应
> **作用**: 统一的错误响应格式，所有接口出错时返回此结构

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Code | int | code | 错误码 |
| Message | string | message | 错误消息 |

---

## 2. account 包 (账号)

### 业务逻辑

- 手机号双重保护：AES-GCM 加密存储 + SHA256 哈希索引
- 用户名自动生成：6位随机字母数字，重复则重试
- 手机号脱敏：运行时生成 `138****8888` 格式
- 支持软删除 (DeletedAt)
- 分页查询

### Account 账号表
> **作用**: 用户账号实体(数据库表)，存储用户基本信息、状态和最后登录时间。手机号采用加密存储+哈希索引的双重机制保护隐私

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | string (uuid) | id | 主键 |
| Username | string | username | 用户名 (6位随机字符) |
| PhoneHash | string | - | 手机号SHA256哈希 (用于查询索引，不返回前端) |
| PhoneEncrypted | string | - | 手机号AES-GCM加密存储 (不返回前端) |
| Phone | string | phone | 解密后的手机号 (不存数据库，运行时解密) |
| PhoneMasked | string | phoneMasked | 脱敏手机号 138****8888 (不存数据库，运行时生成) |
| Status | string | status | 状态: active/disabled |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| DeletedAt | gorm.DeletedAt | - | 软删除时间 |
| LastLoginAt | *time.Time | lastLoginAt | 最后登录时间 |

### Service 方法

| 方法 | 说明 |
|------|------|
| CreateAccount(phone) | 加密手机号 → 生成用户名 → 存储 |
| GetAccount(id) | 查询 + 解密手机号 + 生成脱敏号 |
| GetAccountByPhone(phone) | 通过 phone_hash 查询 |
| UpdateAccount(id, status) | 更新账号状态 |
| DeleteAccount(id) | 软删除 |
| ListAccounts(page, size) | 分页查询 |
| UpdateLastLogin(id) | 更新最后登录时间 |

---

## 3. company 包 (企业认证)

### 业务逻辑

1. 用户上传营业执照图片 → 调用 OCR 识别 (当前为 Mock)
2. 用户提交认证申请 (企业名、执照号、图片路径)
3. 系统校验：已有认证企业或已有待审核申请 → 拒绝重复申请
4. 审核通过后：创建 Company 记录 + AccountCompany 绑定
5. 投标时通过 IsAccountVerified 检查是否已认证

### 企业认证申请状态常量
| 常量 | 值 | 说明 |
|------|------|------|
| ApplicationStatusPending | pending | 待审核 |
| ApplicationStatusApproved | approved | 已通过 |
| ApplicationStatusRejected | rejected | 已拒绝 |

### Company 企业表
> **作用**: 企业实体(数据库表)，只存储已通过审核的企业信息

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | string (uuid) | id | 主键 |
| Name | string | name | 企业名称 |
| BusinessLicenseNo | string | businessLicenseNo | 营业执照号 |
| BusinessLicenseImage | string | businessLicenseImage | 营业执照图片URL |
| VerifiedAt | time.Time | verifiedAt | 认证通过时间 |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| DeletedAt | gorm.DeletedAt | - | 软删除时间 |

### AccountCompany 账号-企业绑定表
> **作用**: 账号与企业的一对一关联表，一个账号只能绑定一个企业

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | string (uuid) | id | 主键 |
| AccountID | string (uuid) | accountId | 账号ID (唯一索引) |
| CompanyID | string (uuid) | companyId | 企业ID |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| DeletedAt | gorm.DeletedAt | - | 软删除时间 |

### CompanyApplication 企业认证申请表
> **作用**: 企业认证申请实体(数据库表)，存储用户提交的企业认证申请

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | string (uuid) | id | 主键 |
| AccountID | string (uuid) | accountId | 申请人账号ID |
| Name | string | name | 企业名称 |
| BusinessLicenseNo | string | businessLicenseNo | 营业执照号 |
| BusinessLicenseImage | string | businessLicenseImage | 营业执照图片URL |
| Status | string | status | 状态: pending/approved/rejected |
| RejectReason | *string | rejectReason | 拒绝原因 |
| ReviewedAt | *time.Time | reviewedAt | 审核时间 |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| DeletedAt | gorm.DeletedAt | - | 软删除时间 |

### OCRResult OCR识别结果
> **作用**: 营业执照 OCR 识别返回结构 (当前为 Mock 实现)

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| CompanyName | string | companyName | 企业名称 |
| BusinessLicenseNo | string | businessLicenseNo | 营业执照号 |
| LegalPerson | string | legalPerson | 法人 |
| RegisteredCapital | string | registeredCapital | 注册资本 |
| EstablishDate | string | establishDate | 成立日期 |
| BusinessScope | string | businessScope | 经营范围 |
| Address | string | address | 地址 |

### MyCompanyStatus 我的企业认证状态
> **作用**: 用户查询自己企业认证状态的响应数据

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| HasVerifiedCompany | bool | hasVerifiedCompany | 是否已有认证企业 |
| Company | *Company | company | 已认证的企业信息 |
| PendingApplication | *CompanyApplication | pendingApplication | 待审核的申请 |
| LatestRejected | *CompanyApplication | latestRejected | 最近被拒绝的申请 |

### Service 方法

| 方法 | 说明 |
|------|------|
| ApplyCompany(accountID, req, imagePath) | 防重复 → 创建申请 |
| GetMyCompanyStatus(accountID) | 查询认证/待审核/被拒状态 |
| IsAccountVerified(accountID) | 是否已通过企业认证 |
| RecognizeLicense(imagePath) | OCR 识别 (Mock) |
| GetCompany(id) | 获取企业详情 |
| ListCompanies(page, size) | 分页列表 |

---

## 4. bounty 包 (悬赏)

### 业务逻辑

1. 采购方创建悬赏 → 选择类型 (梭织/针织) → 填写规格
2. 日期解析支持 "2006-01-02" 和 RFC3339 两种格式
3. 梭织成分 Composition 自动校验百分比之和≈1.0
4. 创建/更新后异步同步到 Elasticsearch (不阻塞请求)
5. 删除时级联删除规格 + 异步删除 ES 文档
6. 支持全量重建索引 (ReindexAllBounties)

### 状态流转
```
open → in_progress → completed
  ↘         ↗
    closed
```

### Bounty 悬赏（聚合根）
> **作用**: 业务核心聚合根，描述采购方发布的面料需求

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | uint | id | 主键 (自增) |
| BountyType | string | bountyType | 类型: woven(梭织)/knitted(针织) |
| ProductName | string | productName | 产品品名 |
| ProductCode | string | productCode | 产品编码 |
| SampleType | string | sampleType | 需要的样品类型 |
| ExpectedDeliveryDate | time.Time | expectedDeliveryDate | 预计交货日期 |
| BidDeadline | time.Time | bidDeadline | 投标截止时间 |
| Status | string | status | 状态: open/in_progress/completed/closed |
| CreatedBy | string (uuid) | createdBy | 发布人账号ID |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| WovenSpec | *BountyWovenSpec | wovenSpec | 梭织规格 (一对一) |
| KnittedSpec | *BountyKnittedSpec | knittedSpec | 针织规格 (一对一) |

### BountyWovenSpec 悬赏-梭织规格
> **作用**: 梭织类悬赏的详细规格，包含克重、幅宽、经纬密度、成分等参数

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | uint | id | 主键 |
| BountyID | uint | bountyId | 所属悬赏ID |
| FabricWeight | float64 | fabricWeight | 成品克重 (g/m²) |
| FabricWidth | float64 | fabricWidth | 成品幅宽 |
| WarpDensity | float64 | warpDensity | 成品经密 (根/英寸) |
| WeftDensity | float64 | weftDensity | 成品纬密 (根/英寸) |
| Composition | Composition | composition | 面料成分 {"棉": 0.6, "涤纶": 0.4} |
| WarpMaterial | string | warpMaterial | 经向原料 |
| WeftMaterial | string | weftMaterial | 纬向原料 |
| QuantityMeters | float64 | quantityMeters | 需求数量 (米) |

### BountyKnittedSpec 悬赏-针织规格
> **作用**: 针织类悬赏的详细规格，包含克重、幅宽、机型、原料明细等参数

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | uint | id | 主键 |
| BountyID | uint | bountyId | 所属悬赏ID |
| FabricWeight | float64 | fabricWeight | 成品克重 (g/m²) |
| FabricWidth | float64 | fabricWidth | 成品幅宽 |
| MachineType | string | machineType | 机型/针织设备类型 |
| Composition | string | composition | 面料成分 (展示用字符串) |
| Materials | Materials | materials | 原料明细 (结构化数组) |
| QuantityKg | float64 | quantityKg | 需求数量 (kg) |

### 自定义类型

**Composition** — `map[string]float64`
- 用于梭织规格，键值对存储各原料占比
- 实现 `driver.Valuer` / `sql.Scanner` 接口，JSON 格式存储到数据库
- `Validate()` 校验百分比之和≈1.0
- `String()` 格式化为 "Cotton 60% / Polyester 40%"

**Materials** — `[]Material`
- 用于针织规格，结构化原料列表
- Material: `{Name string, Percentage float64}`

### BountyDocument ES文档
> **作用**: Bounty 在 Elasticsearch 中的扁平化文档结构

| 字段 | 说明 |
|------|------|
| id, bountyType, productName, productCode | 基本信息 |
| sampleType, status, createdBy | 状态信息 |
| expectedDeliveryDate, bidDeadline, createdAt, updatedAt | 时间信息 |
| fabricWeight, fabricWidth, warpDensity, weftDensity | 规格参数 (梭织) |
| composition, warpMaterial, weftMaterial, quantityMeters | 材料信息 (梭织) |
| machineType, materials, quantityKg | 规格参数 (针织) |

### Service 方法

| 方法 | 说明 |
|------|------|
| CreateBounty(req, userID) | 解析日期 → 创建 Bounty + Spec → 异步同步ES |
| GetBounty(id) | Preload WovenSpec/KnittedSpec |
| UpdateBounty(id, req) | 更新字段 + 规格 → 异步同步ES |
| DeleteBounty(id) | 删除 Bounty + 规格 → 异步删除ES文档 |
| ListBounties(page, size) | 分页 + Preload |

---

## 5. bid 包 (投标)

### 业务逻辑

1. 供应商投标前检查：必须已通过企业认证 (Company Service)
2. 创建投标：状态初始为 pending，关联到指定 Bounty
3. 查询投标时通过 JOIN 关联 Bounty 表，附带产品名、编码、类型、截止时间
4. 支持按 BountyID 或 AccountID 分页查询，AccountID 查询支持按状态过滤
5. Spec 通过单独查询预加载，再按 BidID map 回关联

### 投标状态常量
| 常量 | 值 | 说明 |
|------|------|------|
| BidStatusPending | pending | 待审核 |
| BidStatusInProgress | in_progress | 进行中 |
| BidStatusPendingAcceptance | pending_acceptance | 待验收 |
| BidStatusCompleted | completed | 已完成 |

### Bid 投标模型
> **作用**: 供应商对悬赏进行报价投标的核心模型

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | string (uuid) | id | 主键 |
| BountyID | uint | bountyId | 悬赏ID |
| AccountID | string (uuid) | accountId | 投标账号ID |
| BidPrice | float64 | bidPrice | 投标价格 |
| Status | string | status | 状态: pending/in_progress/pending_acceptance/completed |
| CreatedAt | time.Time | createdAt | 创建时间 |
| UpdatedAt | time.Time | updatedAt | 更新时间 |
| WovenSpec | *BidWovenSpec | wovenSpec | 梭织规格 |
| KnittedSpec | *BidKnittedSpec | knittedSpec | 针织规格 |
| BountyProductName | string | bountyProductName | 悬赏产品名 (JOIN 查询填充) |
| BountyProductCode | string | bountyProductCode | 悬赏产品编码 (JOIN 查询填充) |
| BountyType | string | bountyType | 悬赏类型 (JOIN 查询填充) |
| BidDeadline | time.Time | bidDeadline | 投标截止时间 (JOIN 查询填充) |

### BidWovenSpec 投标-梭织规格
> **作用**: 梭织类投标的供应商报价参数

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | uint | id | 主键 |
| BidID | string (uuid) | bidId | 投标ID |
| SizeLength | float64 | sizeLength | 尺码 (长度) |
| GreigeFabricType | string | greigeFabricType | 胚布类型: 现货/定织 |
| GreigeDeliveryDate | time.Time | greigeDeliveryDate | 胚布交期 |
| DeliveryMethod | string | deliveryMethod | 交货方式 |

### BidKnittedSpec 投标-针织规格
> **作用**: 针织类投标的供应商报价参数

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| ID | uint | id | 主键 |
| BidID | string (uuid) | bidId | 投标ID |
| SizeWeight | float64 | sizeWeight | 尺码 (重量/皮重) |
| GreigeFabricType | string | greigeFabricType | 胚布类型: 现货/定织 |
| GreigeDeliveryDate | time.Time | greigeDeliveryDate | 胚布交期 |
| DeliveryMethod | string | deliveryMethod | 交货方式 |

### Service 方法

| 方法 | 说明 |
|------|------|
| CreateBid(req, accountID) | 创建投标 (status=pending) |
| GetBid(id) | 获取投标 + Preload Spec |
| DeleteBid(id) | 删除投标 |
| ListBidsByBountyID(bountyID, page, size) | 按悬赏分页查询 |
| ListBidsByAccountID(accountID, status, page, size) | 按账号分页查询，支持状态过滤 |

---

## 6. search 包 (搜索)

### 业务逻辑

- 统一搜索入口，通过 Elasticsearch 实现全文检索
- 默认分页: page=1, size=10, 最大100
- 支持排序字段映射和聚合解析 (动态筛选项)
- 计算总页数

### SearchRequest 统一搜索请求
> **作用**: 客户端发起搜索请求的统一入参

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Index | string | index | 索引名 (必填), 如 "bounty" |
| Query | string | query | 关键词搜索 |
| Filters | map[string]interface{} | filters | 筛选条件 |
| Sort | *SortOption | sort | 排序 |
| Page | int | page | 页码, 默认 1 |
| Size | int | size | 每页数量, 默认 10 |

### SortOption 排序选项

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Field | string | field | 排序字段 |
| Order | string | order | 排序方向: asc/desc |

### SearchResultData 搜索结果数据

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| Hits | []map[string]interface{} | hits | 搜索结果 |
| Total | int64 | total | 总数 |
| Page | int | page | 当前页 |
| Size | int | size | 每页数量 |
| TotalPages | int | totalPages | 总页数 |
| Filters | []elasticsearch.FilterBucket | filters | 动态筛选项 |

---

## 7. pkg/crypto 包 (加密)

### Crypto AES-GCM 加密服务
> **作用**: 提供 AES-GCM 对称加密服务，用于手机号等敏感信息的加密存储

| 字段 | 类型 | 说明 |
|------|------|------|
| key | []byte | 加密密钥 (16/24/32 字节对应 AES-128/192/256) |

### 方法

| 函数/方法 | 签名 | 说明 |
|------|------|------|
| NewCrypto | `NewCrypto(key string) (*Crypto, error)` | 创建加密服务实例 |
| Encrypt | `(c *Crypto) Encrypt(plaintext string) (string, error)` | AES-GCM加密 → base64 (每次生成随机nonce) |
| Decrypt | `(c *Crypto) Decrypt(ciphertext string) (string, error)` | base64解码 → AES-GCM解密 |
| MaskPhone | `MaskPhone(phone string) string` | 手机号脱敏 (138****8888) |
| Hash | `Hash(data string) string` | SHA256哈希 (用于索引查询) |

---

## 8. pkg/jwt 包 (JWT认证)

### Claims JWT声明
> **作用**: JWT Token 的自定义声明结构

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| UserID | string | user_id | 用户ID |
| Username | string | username | 用户名 |
| RegisteredClaims | jwt.RegisteredClaims | - | 标准JWT声明 (issuer, exp等) |

### JWTService JWT服务

| 字段 | 类型 | 说明 |
|------|------|------|
| secretKey | []byte | HS256 签名密钥 |
| issuer | string | 签发者 |
| expiry | time.Duration | 过期时间 |

### 方法

| 方法 | 签名 | 说明 |
|------|------|------|
| NewJWTService | `NewJWTService(secretKey, issuer string, expiry time.Duration) *JWTService` | 创建JWT服务 |
| GenerateToken | `(s *JWTService) GenerateToken(userID, username string) (string, error)` | 生成JWT Token |
| ValidateToken | `(s *JWTService) ValidateToken(tokenString string) (*Claims, error)` | 验证JWT Token |

---

## 9. pkg/middleware 包 (中间件)

### JWTAuth 认证中间件
> **作用**: Gin 路由中间件，保护需要登录的接口

逻辑：
1. 从 `Authorization` Header 提取 `Bearer <token>`
2. 调用 JWTService.ValidateToken 验证签名和过期
3. 验证通过 → 将 `user_id`、`username` 存入 Gin Context
4. 验证失败 → 返回 401

### 辅助函数

| 函数 | 说明 |
|------|------|
| GetUserID(c *gin.Context) string | 从 Context 获取当前用户ID |
| GetUsername(c *gin.Context) string | 从 Context 获取当前用户名 |

---

## 10. pkg/elasticsearch 包 (搜索引擎)

### 配置与连接
- TLS 跳过验证 (开发环境)
- 连接池: 10 空闲连接, 30s 超时
- Basic Auth 认证

### 主要函数

| 函数 | 说明 |
|------|------|
| InitClient(Config) | 初始化全局 ES 客户端 |
| GetClient() | 获取全局客户端 |
| Ping(ctx) | 健康检查 |
| IndexDocument(ctx, index, id, doc) | 索引文档 |
| SearchDocuments(ctx, index, query) | 搜索文档 |
| DeleteDocument(ctx, index, id) | 删除文档 |

---

## API 路由结构

### 公开接口 (无需登录)
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/auth/send-code | 发送验证码 |
| POST | /api/v1/auth/verify-code | 验证码登录/注册 |
| GET | /api/v1/bounties/peek | 预览前10条悬赏 |

### 受保护接口 (需 JWT)
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/search | 统一搜索 |
| GET | /api/v1/bounties | 悬赏列表 |
| GET | /api/v1/bounties/:id | 悬赏详情 |
| POST | /api/v1/bounties | 创建悬赏 |
| PUT | /api/v1/bounties/:id | 更新悬赏 |
| DELETE | /api/v1/bounties/:id | 删除悬赏 |
| GET | /api/v1/bids/bounty/:bountyId | 悬赏下的投标列表 |
| GET | /api/v1/bids/my | 我的投标列表 |
| POST | /api/v1/bids | 创建投标 (需企业认证) |
| DELETE | /api/v1/bids/:id | 删除投标 |
| GET | /api/v1/accounts/:id | 账号详情 |
| GET | /api/v1/companies | 企业列表 |
| GET | /api/v1/companies/:id | 企业详情 |
| POST | /api/v1/companies/apply | 提交企业认证申请 |
| GET | /api/v1/companies/my-status | 我的认证状态 |
| POST | /api/v1/companies/recognize | OCR识别营业执照 |

### 静态资源
| 路径 | 说明 |
|------|------|
| /uploads/ | 上传的营业执照图片 |

---

## 包间依赖关系

```
Auth Handler
    ├── pkg/jwt (Token 生成)
    ├── account Service (账号查询/创建)
    └── pkg/crypto (手机号加密/哈希)

Bid Handler
    ├── bid Service (投标 CRUD)
    └── company Service (企业认证校验)

Bounty Handler
    ├── bounty Service (悬赏 CRUD)
    └── pkg/elasticsearch (异步同步)

Company Handler
    ├── company Service (申请/查询)
    └── 文件上传 (图片保存)

Search Handler
    └── pkg/elasticsearch (全文检索)

pkg/middleware
    └── pkg/jwt (Token 验证)
```

---

## 数据库表 (GORM AutoMigrate)

| 表名 | 主键 | 说明 |
|------|------|------|
| accounts | UUID | 用户账号 |
| bounties | uint (自增) | 悬赏 |
| bounty_woven_specs | uint | 悬赏梭织规格 (1:1) |
| bounty_knitted_specs | uint | 悬赏针织规格 (1:1) |
| bids | UUID | 投标 |
| bid_woven_specs | uint | 投标梭织规格 (1:1) |
| bid_knitted_specs | uint | 投标针织规格 (1:1) |
| companies | UUID | 已认证企业 |
| account_companies | UUID | 账号-企业绑定 (唯一AccountID) |
| company_applications | UUID | 企业认证申请 |

---

## 数据安全措施

- 手机号 AES-GCM 加密存储，不以明文存在数据库
- 手机号 SHA256 哈希索引，支持查询但不可逆推
- 响应中手机号自动脱敏 (138****8888)
- JWT Token 有过期时间控制
- 敏感表支持软删除 (DeletedAt)
- 投标前强制企业认证校验

---

## 通用响应结构

所有接口遵循统一格式:

```go
// 单条数据
{
    "code":    200,
    "message": "success",
    "data":    { ... }
}

// 列表数据
{
    "code":    200,
    "message": "success",
    "data":    [ ... ],
    "total":   100
}

// 错误
{
    "code":    400/401/500,
    "message": "错误描述"
}
```
