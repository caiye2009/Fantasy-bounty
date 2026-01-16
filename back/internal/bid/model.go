package bid

import (
	"time"
)

// Bid 竞标模型
type Bid struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	BountyID  uint      `json:"bounty_id" gorm:"not null;index" binding:"required"`
	Price     float64   `json:"price" gorm:"type:decimal(10,2);not null" binding:"required,gt=0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CreateBidRequest 创建竞标请求
type CreateBidRequest struct {
	BountyID uint    `json:"bounty_id" binding:"required"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

// BidResponse 竞标响应
type BidResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Bid   `json:"data,omitempty"`
}

// BidListResponse 竞标列表响应
type BidListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []Bid  `json:"data"`
	Total   int64  `json:"total"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
