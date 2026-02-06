package supplier

import (
	"time"

	"gorm.io/gorm"
)

// 供应商认证申请状态
const (
	ApplicationStatusPending  = "pending"  // 待审核
	ApplicationStatusApproved = "approved" // 已通过
	ApplicationStatusRejected = "rejected" // 已拒绝
)

// Supplier 供应商表（只存已通过审核的供应商）
type Supplier struct {
	ID                   string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                 string         `json:"name" gorm:"type:varchar(255);not null"`
	BusinessLicenseNo    string         `json:"businessLicenseNo" gorm:"type:varchar(100);not null"`
	BusinessLicenseImage string         `json:"businessLicenseImage" gorm:"type:varchar(500);not null"`
	VerifiedAt           time.Time      `json:"verifiedAt"`
	CreatedAt            time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Supplier) TableName() string {
	return "suppliers"
}

// UserSupplier 用户-供应商绑定表
type UserSupplier struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username   string         `json:"username" gorm:"type:varchar(20);not null;uniqueIndex"`
	SupplierID string         `json:"supplierId" gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (UserSupplier) TableName() string {
	return "user_suppliers"
}

// SupplierApplication 供应商认证申请表
type SupplierApplication struct {
	ID                   string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username             string         `json:"username" gorm:"type:varchar(20);not null;index"`
	Name                 string         `json:"name" gorm:"type:varchar(255);not null"`
	BusinessLicenseNo    string         `json:"businessLicenseNo" gorm:"type:varchar(100);not null"`
	BusinessLicenseImage string         `json:"businessLicenseImage" gorm:"type:varchar(500);not null"`
	Status               string         `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	RejectReason         *string        `json:"rejectReason" gorm:"type:varchar(500)"`
	ReviewedAt           *time.Time     `json:"reviewedAt"`
	CreatedAt            time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SupplierApplication) TableName() string {
	return "supplier_applications"
}

// ========== Request ==========

// ApplySupplierRequest 供应商认证申请请求
type ApplySupplierRequest struct {
	Name              string `json:"name" binding:"required"`
	BusinessLicenseNo string `json:"businessLicenseNo" binding:"required"`
}

// ========== OCR ==========

// OCRResult 营业执照OCR识别结果
type OCRResult struct {
	CompanyName       string `json:"companyName"`
	BusinessLicenseNo string `json:"businessLicenseNo"`
	LegalPerson       string `json:"legalPerson"`
	RegisteredCapital string `json:"registeredCapital"`
	EstablishDate     string `json:"establishDate"`
	BusinessScope     string `json:"businessScope"`
	Address           string `json:"address"`
}

// OCRResponse OCR识别响应
type OCRResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *OCRResult `json:"data,omitempty"`
}

// ========== Response ==========

// SupplierResponse 单个供应商响应
type SupplierResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    *Supplier `json:"data,omitempty"`
}

// SupplierListResponse 供应商列表响应
type SupplierListResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []Supplier `json:"data"`
	Total   int64      `json:"total"`
}

// ApplicationResponse 申请响应
type ApplicationResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *SupplierApplication `json:"data,omitempty"`
}

// MySupplierStatusResponse 我的供应商认证状态响应
type MySupplierStatusResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *MySupplierStatus `json:"data,omitempty"`
}

// MySupplierStatus 我的供应商认证状态
type MySupplierStatus struct {
	HasVerifiedSupplier bool                 `json:"hasVerifiedSupplier"`
	Supplier            *Supplier            `json:"supplier,omitempty"`
	PendingApplication  *SupplierApplication `json:"pendingApplication,omitempty"`
	LatestRejected      *SupplierApplication `json:"latestRejected,omitempty"`
}
