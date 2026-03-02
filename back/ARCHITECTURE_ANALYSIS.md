# 后端架构分析

## 当前实际运行的功能（路由已注册）

```
/api/v1/auth/
  POST /wechat-login          → auth.Handler.WechatLogin
  POST /refresh               → auth.Handler.RefreshToken

/api/v1/proxy/                (需要外部JWT)
  POST /bind-wechat           → proxy.InternalProxy
  POST /get-by-wechat         → proxy.InternalProxy
  POST /inquiry-query         → proxy.InternalProxy
  POST /inquiry-detail        → proxy.InternalProxy
  POST /quote-delete          → proxy.InternalProxy
  POST /quote-save            → proxy.InternalProxy
  POST /inquiry-quoted        → proxy.InternalProxy  (Pur_InquiryBySupplierQuoted)

/api/v1/internal/             (需要外部JWT)
  GET  /bounties              → proxy.InternalProxy
  GET  /bounties/:id          → proxy.InternalProxy

/api/v1/internal/login        → auth.Handler.InternalLogin (admin，无JWT)
/api/v1/proxy/refresh-token   → proxy.ForceRefreshTokenHandler (admin，无JWT)
```

---

## 问题清单（待处理）

### 1. `pkg/proxy` 位置错误（核心）
`pkg/` 应放可复用工具（crypto/jwt/middleware），但 proxy 是纯业务 handler。
**修改：** `pkg/proxy/` → `internal/proxy/`

### 2. 死代码（路由未注册）

| 包 | 未使用内容 |
|---|---|
| `internal/bid` | 整个包（handler/service/repo/model） |
| `internal/supplier` | 整个包 |
| `internal/user` | handler 完整，路由无注册；user.Service 仅给未注册的方法用 |
| `internal/auth` | `SendCode`、`VerifyCode` 两个方法未注册 |
| `pkg/proxy` | `Handler()` 通用代理方法未使用 |

**待确认：**
- bid / supplier / user 三个模块是否全删？
- SendCode / VerifyCode 是废弃还是后续还会用？

### 3. 代码重复
`handleBountyList` 和 `handleBountyDetail` 各自手写了一遍 http 请求，没有复用 `forwardToInternal`。

### 4. `BountiesHandler` 重复验证 JWT
auth 中间件已经验证过，`BountiesHandler` 内又解析了一遍，多余。

---

## 整体流程

```
前端
  │ Bearer <外部JWT>
  ▼
后端 (this repo)
  │ JWTAuth 中间件验证外部JWT
  │ 静默获取/刷新 内部Token (internal_token.Manager)
  │ 组装请求 → POST /api/Public/GetProcedureDataSet
  ▼
内部系统 (INTERNAL_API_URL)
  │ 返回业务数据
  ▼
后端透传响应
  ▼
前端
```

---

## 目标架构（整理后）

```
back/
├── cmd/main.go
├── config/
│   ├── init.go
│   └── router.go
├── internal/
│   ├── audit/          ← 保留
│   ├── auth/           ← 保留，清理未注册方法
│   └── proxy/          ← 从 pkg/proxy 移过来
│       ├── handler.go
│       └── types.go
└── pkg/
    ├── crypto/
    ├── internal_token/
    ├── jwt/
    └── middleware/
```
