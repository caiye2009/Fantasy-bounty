package user

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username       string         `json:"username" gorm:"type:varchar(20);uniqueIndex"`            // 唯一用户名（自动生成6位）
	PhoneHash      string         `json:"-" gorm:"type:varchar(64);not null;uniqueIndex"`          // 手机号哈希（用于查询索引）
	PhoneEncrypted string         `json:"-" gorm:"type:varchar(255);not null"`                     // 手机号加密存储（用于解密显示）
	Phone          string         `json:"phone" gorm:"-"`                                          // 解密后的手机号（不存数据库）
	PhoneMasked    string         `json:"phoneMasked" gorm:"-"`                                    // 脱敏手机号（不存数据库）
	Status         string         `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // active, disabled
	CreatedAt      time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	LastLoginAt    *time.Time     `json:"lastLoginAt"`
}

func (User) TableName() string {
	return "users"
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=active disabled"`
}

// UserResponse 单个用户响应
type UserResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *User  `json:"data,omitempty"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []User `json:"data"`
	Total   int64  `json:"total"`
}
