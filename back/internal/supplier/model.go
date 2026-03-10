package supplier

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SupplierProfile 供应商档案
type SupplierProfile struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string         `json:"userId" gorm:"type:uuid;not null;uniqueIndex"` // 关联用户，唯一索引
	CompanyType  string         `json:"companyType" gorm:"type:varchar(50)"`          // 企业类型
	CompanyName  string         `json:"companyName" gorm:"type:varchar(255)"`         // 企业名称
	Capabilities datatypes.JSON `json:"capabilities" gorm:"type:jsonb"`                // 机器能力 {a:1, b:2}
	CreatedAt    time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SupplierProfile) TableName() string {
	return "supplier_profiles"
}

// ========== Request ==========

// CreateSupplierRequest 创建/更新供应商请求
type CreateSupplierRequest struct {
	UserID       string         `json:"-"` // 从上下文中获取，不从前端传递
	CompanyType  string         `json:"companyType" binding:"required"`
	CompanyName  string         `json:"companyName" binding:"required"`
	Capabilities map[string]int `json:"capabilities"` // 机器能力 {a:1, b:2}
}

// UpdateCapabilitiesRequest 更新机器能力请求
type UpdateCapabilitiesRequest struct {
	Capabilities map[string]int `json:"capabilities" binding:"required"`
}

// ========== Response ==========

// SupplierProfileResponse 供应商档案响应
type SupplierProfileResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *SupplierProfile `json:"data,omitempty"`
}

// CapabilitiesResponse 机器能力响应
type CapabilitiesResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    map[string]int    `json:"data,omitempty"`
}

// SupplierFullInfoResponse 供应商完整信息响应
type SupplierFullInfoResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *SupplierFullInfo  `json:"data,omitempty"`
}

// SupplierFullInfo 供应商完整信息
type SupplierFullInfo struct {
	Info          *SupplierInfo      `json:"info"`           // 企业信息
	Qualification *SupplierQualification `json:"qualification"` // 资质信息
	Capabilities  map[string]int    `json:"capabilities"`  // 生产能力
}

// SupplierInfo 企业基本信息
type SupplierInfo struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	CompanyType string `json:"companyType"`
	CompanyName string `json:"companyName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SupplierQualification 供应商资质信息
type SupplierQualification struct {
	// 可以在这里添加资质相关字段，目前暂时为空
	// 例如：营业执照、生产许可证等
}
