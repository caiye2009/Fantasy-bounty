# Bounty CRUD API 文档

## 项目结构

```
back/
├── cmd/
│   └── main.go              # 应用入口
├── config/
│   ├── init.go             # 配置初始化
│   ├── database.go         # 数据库配置
│   └── router.go           # 路由配置
├── internal/
│   └── bounty/
│       ├── model.go        # 数据模型
│       ├── bounty_repo.go  # 数据访问层
│       ├── bounty_service.go # 业务逻辑层
│       └── bounty_handler.go # HTTP 处理层
├── go.mod
└── bounty.db               # SQLite 数据库文件（运行后自动生成）
```

## 启动服务器

```bash
# 运行服务器
go run cmd/main.go

# 或者编译后运行
go build -o bounty-server ./cmd/main.go
./bounty-server
```

服务器将在 `http://localhost:8080` 启动

## API 接口

### 1. 创建赏金

**POST** `/api/v1/bounties`

**请求体:**
```json
{
  "title": "修复登录 Bug",
  "description": "修复用户登录时的验证码错误问题",
  "reward": 500.00,
  "created_by": "user123"
}
```

**响应:**
```json
{
  "code": 201,
  "message": "Bounty created successfully",
  "data": {
    "id": 1,
    "title": "修复登录 Bug",
    "description": "修复用户登录时的验证码错误问题",
    "reward": 500.00,
    "status": "open",
    "created_by": "user123",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. 获取赏金列表

**GET** `/api/v1/bounties?page=1&page_size=10`

**查询参数:**
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 10，最大 100

**响应:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "title": "修复登录 Bug",
      "description": "修复用户登录时的验证码错误问题",
      "reward": 500.00,
      "status": "open",
      "created_by": "user123",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 3. 获取赏金详情

**GET** `/api/v1/bounties/:id`

**响应:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": 1,
    "title": "修复登录 Bug",
    "description": "修复用户登录时的验证码错误问题",
    "reward": 500.00,
    "status": "open",
    "created_by": "user123",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 4. 更新赏金

**PUT** `/api/v1/bounties/:id`

**请求体:**
```json
{
  "title": "修复登录 Bug（紧急）",
  "description": "修复用户登录时的验证码错误问题，优先级高",
  "reward": 800.00,
  "status": "in_progress"
}
```

**响应:**
```json
{
  "code": 200,
  "message": "Bounty updated successfully",
  "data": {
    "id": 1,
    "title": "修复登录 Bug（紧急）",
    "description": "修复用户登录时的验证码错误问题，优先级高",
    "reward": 800.00,
    "status": "in_progress",
    "created_by": "user123",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

### 5. 删除赏金

**DELETE** `/api/v1/bounties/:id`

**响应:**
```json
{
  "code": 200,
  "message": "Bounty deleted successfully"
}
```

### 6. 健康检查

**GET** `/health`

**响应:**
```json
{
  "status": "ok"
}
```

## 赏金状态

- `open`: 开放中
- `in_progress`: 进行中
- `completed`: 已完成
- `closed`: 已关闭

## 错误响应

```json
{
  "code": 400,
  "message": "错误描述"
}
```

**常见错误码:**
- `400`: 请求参数错误
- `404`: 资源不存在
- `500`: 服务器内部错误

## cURL 示例

```bash
# 创建赏金
curl -X POST http://localhost:8080/api/v1/bounties \
  -H "Content-Type: application/json" \
  -d '{
    "title": "修复登录 Bug",
    "description": "修复用户登录时的验证码错误问题",
    "reward": 500.00,
    "created_by": "user123"
  }'

# 获取赏金列表
curl http://localhost:8080/api/v1/bounties?page=1&page_size=10

# 获取赏金详情
curl http://localhost:8080/api/v1/bounties/1

# 更新赏金
curl -X PUT http://localhost:8080/api/v1/bounties/1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'

# 删除赏金
curl -X DELETE http://localhost:8080/api/v1/bounties/1
```

## 数据库

项目使用 SQLite 数据库，数据库文件为 `bounty.db`，会在首次运行时自动创建。

如需切换到 MySQL 或 PostgreSQL，修改 `config/database.go` 中的数据库驱动即可。
