package bounty

import (
	"time"
)

// Bounty 赏金模型
type Bounty struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null" binding:"required"`
	Description string    `json:"description" gorm:"type:text"`
	Reward      float64   `json:"reward" gorm:"type:decimal(10,2);not null" binding:"required,gt=0"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'open'" binding:"omitempty,oneof=open in_progress completed closed"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateBountyRequest 创建赏金请求
type CreateBountyRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description string  `json:"description"`
	Reward      float64 `json:"reward" binding:"required,gt=0"`
	CreatedBy   string  `json:"created_by" binding:"required"`
}

// UpdateBountyRequest 更新赏金请求
type UpdateBountyRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description string  `json:"description"`
	Reward      float64 `json:"reward" binding:"omitempty,gt=0"`
	Status      string  `json:"status" binding:"omitempty,oneof=open in_progress completed closed"`
}

// BountyResponse 赏金响应
type BountyResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Bounty `json:"data,omitempty"`
}

// BountyListResponse 赏金列表响应
type BountyListResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []Bounty  `json:"data"`
	Total   int64     `json:"total"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
