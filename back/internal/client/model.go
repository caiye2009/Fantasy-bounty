package client

import (
	"time"

	"gorm.io/gorm"
)

// ClientProfile 客户档案（id=users.id）
type ClientProfile struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey"`                 // equals users.id
	UserID      string         `json:"userId" gorm:"type:uuid;not null;uniqueIndex"`   // equals ID, kept for clarity
	CompanyName string         `json:"companyName" gorm:"type:varchar(255)"`
	ContactName string         `json:"contactName" gorm:"type:varchar(100)"`
	Address     string         `json:"address" gorm:"type:varchar(255)"`
	Industry    string         `json:"industry" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ClientProfile) TableName() string {
	return "client_profiles"
}

type CreateClientProfileRequest struct {
	CompanyName string `json:"companyName" binding:"required"`
	ContactName string `json:"contactName" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Industry    string `json:"industry" binding:"required"`
}

type ClientProfileResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *ClientProfile `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

