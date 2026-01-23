package account

import (
	"time"

	"gorm.io/gorm"
)

// Account 账号表
type Account struct {
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

func (Account) TableName() string {
	return "accounts"
}

// CreateAccountRequest 创建账号请求
type CreateAccountRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// UpdateAccountRequest 更新账号请求
type UpdateAccountRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=active disabled"`
}

// AccountResponse 单个账号响应
type AccountResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *Account `json:"data,omitempty"`
}

// AccountListResponse 账号列表响应
type AccountListResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []Account `json:"data"`
	Total   int64     `json:"total"`
}
