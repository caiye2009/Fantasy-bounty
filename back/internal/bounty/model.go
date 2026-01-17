package bounty

import (
	"time"
)

// Bounty 悬赏模型
type Bounty struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	BountyType           string    `json:"bountyType" gorm:"type:varchar(50);not null"` // woven(梭织) or knitted(针织)
	ProductName          string    `json:"productName" gorm:"type:varchar(255);not null"`
	SampleType           string    `json:"sampleType" gorm:"type:varchar(100)"`
	ExpectedDeliveryDate time.Time `json:"expectedDeliveryDate" gorm:"type:date"`
	BidDeadline          time.Time `json:"bidDeadline" gorm:"type:timestamp"`
	Status               string    `json:"status" gorm:"type:varchar(50);default:'open'"` // open, in_progress, completed, closed
	CreatedBy            uint      `json:"createdBy" gorm:"not null;index"`
	CreatedAt            time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联规格 - 根据 BountyType 只会有一个有值
	WovenSpec   *BountyWovenSpec   `json:"wovenSpec,omitempty" gorm:"foreignKey:BountyID"`
	KnittedSpec *BountyKnittedSpec `json:"knittedSpec,omitempty" gorm:"foreignKey:BountyID"`
}

// BountyWovenSpec 悬赏-梭织规格
type BountyWovenSpec struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	BountyID       uint    `json:"bountyId" gorm:"not null;uniqueIndex"`
	FabricWeight   float64 `json:"fabricWeight" gorm:"type:decimal(10,2)"`   // 成品克重
	FabricWidth    float64 `json:"fabricWidth" gorm:"type:decimal(10,2)"`    // 成品幅宽
	WarpDensity    float64 `json:"warpDensity" gorm:"type:decimal(10,2)"`    // 成品经密
	WeftDensity    float64 `json:"weftDensity" gorm:"type:decimal(10,2)"`    // 成品纬密
	Composition    string  `json:"composition" gorm:"type:varchar(255)"`     // 成分
	WarpMaterial   string  `json:"warpMaterial" gorm:"type:varchar(255)"`    // 经向原料
	WeftMaterial   string  `json:"weftMaterial" gorm:"type:varchar(255)"`    // 纬向原料
	QuantityMeters float64 `json:"quantityMeters" gorm:"type:decimal(10,2)"` // 需求数量（米）
}

// BountyKnittedSpec 悬赏-针织规格
type BountyKnittedSpec struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	BountyID         uint    `json:"bountyId" gorm:"not null;uniqueIndex"`
	FactoryArticleNo string  `json:"factoryArticleNo" gorm:"type:varchar(100)"` // 工厂货号
	FabricWeight     float64 `json:"fabricWeight" gorm:"type:decimal(10,2)"`    // 成品克重
	FabricWidth      float64 `json:"fabricWidth" gorm:"type:decimal(10,2)"`     // 成品幅宽
	MachineType      string  `json:"machineType" gorm:"type:varchar(100)"`      // 机型
	Composition      string  `json:"composition" gorm:"type:varchar(255)"`      // 成分
	QuantityKg       float64 `json:"quantityKg" gorm:"type:decimal(10,2)"`      // 需求数量（kg）
}

// CreateBountyRequest 创建悬赏请求
type CreateBountyRequest struct {
	BountyType           string                        `json:"bountyType" binding:"required,oneof=woven knitted"`
	ProductName          string                        `json:"productName" binding:"required,min=1,max=255"`
	SampleType           string                        `json:"sampleType"`
	ExpectedDeliveryDate string                        `json:"expectedDeliveryDate"`
	BidDeadline          string                        `json:"bidDeadline" binding:"required"`
	WovenSpec            *CreateBountyWovenSpecRequest `json:"wovenSpec"`
	KnittedSpec          *CreateBountyKnittedSpecRequest `json:"knittedSpec"`
}

// CreateBountyWovenSpecRequest 创建悬赏-梭织规格请求
type CreateBountyWovenSpecRequest struct {
	FabricWeight   float64 `json:"fabricWeight"`
	FabricWidth    float64 `json:"fabricWidth"`
	WarpDensity    float64 `json:"warpDensity"`
	WeftDensity    float64 `json:"weftDensity"`
	Composition    string  `json:"composition"`
	WarpMaterial   string  `json:"warpMaterial"`
	WeftMaterial   string  `json:"weftMaterial"`
	QuantityMeters float64 `json:"quantityMeters"`
}

// CreateBountyKnittedSpecRequest 创建悬赏-针织规格请求
type CreateBountyKnittedSpecRequest struct {
	FactoryArticleNo string  `json:"factoryArticleNo"`
	FabricWeight     float64 `json:"fabricWeight"`
	FabricWidth      float64 `json:"fabricWidth"`
	MachineType      string  `json:"machineType"`
	Composition      string  `json:"composition"`
	QuantityKg       float64 `json:"quantityKg"`
}

// UpdateBountyRequest 更新悬赏请求
type UpdateBountyRequest struct {
	ProductName          string                        `json:"productName" binding:"omitempty,min=1,max=255"`
	SampleType           string                        `json:"sampleType"`
	ExpectedDeliveryDate string                        `json:"expectedDeliveryDate"`
	BidDeadline          string                        `json:"bidDeadline"`
	Status               string                        `json:"status" binding:"omitempty,oneof=open in_progress completed closed"`
	WovenSpec            *CreateBountyWovenSpecRequest `json:"wovenSpec"`
	KnittedSpec          *CreateBountyKnittedSpecRequest `json:"knittedSpec"`
}

// BountyResponse 悬赏响应
type BountyResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Bounty `json:"data,omitempty"`
}

// BountyListResponse 悬赏列表响应
type BountyListResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []Bounty `json:"data"`
	Total   int64    `json:"total"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
