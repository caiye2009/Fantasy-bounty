package company

import (
	"time"

	"gorm.io/gorm"
)

// 企业认证申请状态
const (
	ApplicationStatusPending  = "pending"  // 待审核
	ApplicationStatusApproved = "approved" // 已通过
	ApplicationStatusRejected = "rejected" // 已拒绝
)

// Company 企业表（只存已通过审核的企业）
type Company struct {
	ID                   string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                 string         `json:"name" gorm:"type:varchar(255);not null"`
	BusinessLicenseNo    string         `json:"businessLicenseNo" gorm:"type:varchar(100);not null"`
	BusinessLicenseImage string         `json:"businessLicenseImage" gorm:"type:varchar(500);not null"`
	VerifiedAt           time.Time      `json:"verifiedAt"`
	CreatedAt            time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Company) TableName() string {
	return "companies"
}

// AccountCompany 账号-企业绑定表
type AccountCompany struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AccountID string         `json:"accountId" gorm:"type:uuid;not null;uniqueIndex"`
	CompanyID string         `json:"companyId" gorm:"type:uuid;not null"`
	CreatedAt time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AccountCompany) TableName() string {
	return "account_companies"
}

// CompanyApplication 企业认证申请表
type CompanyApplication struct {
	ID                   string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AccountID            string         `json:"accountId" gorm:"type:uuid;not null;index"`
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

func (CompanyApplication) TableName() string {
	return "company_applications"
}

// ========== Request ==========

// ApplyCompanyRequest 企业认证申请请求
type ApplyCompanyRequest struct {
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

// CompanyResponse 单个企业响应
type CompanyResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *Company `json:"data,omitempty"`
}

// CompanyListResponse 企业列表响应
type CompanyListResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []Company `json:"data"`
	Total   int64     `json:"total"`
}


// ApplicationResponse 申请响应
type ApplicationResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *CompanyApplication `json:"data,omitempty"`
}


// MyCompanyStatusResponse 我的企业认证状态响应
type MyCompanyStatusResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *MyCompanyStatus `json:"data,omitempty"`
}

// MyCompanyStatus 我的企业认证状态
type MyCompanyStatus struct {
	HasVerifiedCompany bool                `json:"hasVerifiedCompany"`
	Company            *Company            `json:"company,omitempty"`
	PendingApplication *CompanyApplication `json:"pendingApplication,omitempty"`
	LatestRejected     *CompanyApplication `json:"latestRejected,omitempty"`
}
