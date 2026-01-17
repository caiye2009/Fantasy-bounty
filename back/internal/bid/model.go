package bid

import (
	"time"
)

// Bid 投标模型
type Bid struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	BountyID  uint      `json:"bountyId" gorm:"not null;index"`
	UserID    uint      `json:"userId" gorm:"not null;index"`
	BidPrice  float64   `json:"bidPrice" gorm:"type:decimal(10,2);not null"` // 投标价格
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, accepted, rejected, completed
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联规格 - 根据悬赏类型只会有一个有值
	WovenSpec   *BidWovenSpec   `json:"wovenSpec,omitempty" gorm:"foreignKey:BidID"`
	KnittedSpec *BidKnittedSpec `json:"knittedSpec,omitempty" gorm:"foreignKey:BidID"`

	// 关联字段 - 用于返回时展示悬赏信息
	BountyProductName string    `json:"bountyProductName" gorm:"-"`
	BountyType        string    `json:"bountyType" gorm:"-"`
	BidDeadline       time.Time `json:"bidDeadline" gorm:"-"`
}

// BidWovenSpec 投标-梭织规格
type BidWovenSpec struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	BidID             string    `json:"bidId" gorm:"type:uuid;not null;uniqueIndex"`
	SizeLength        float64   `json:"sizeLength" gorm:"type:decimal(10,2)"`       // 尺码（长度）
	GreigeFabricType  string    `json:"greigeFabricType" gorm:"type:varchar(100)"`  // 胚布类型
	GreigeDeliveryDate time.Time `json:"greigeDeliveryDate" gorm:"type:date"`       // 胚布交期
}

// BidKnittedSpec 投标-针织规格
type BidKnittedSpec struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	BidID             string    `json:"bidId" gorm:"type:uuid;not null;uniqueIndex"`
	SizeWeight        float64   `json:"sizeWeight" gorm:"type:decimal(10,2)"`       // 尺码（重量/皮重）
	GreigeFabricType  string    `json:"greigeFabricType" gorm:"type:varchar(100)"`  // 胚布类型
	GreigeDeliveryDate time.Time `json:"greigeDeliveryDate" gorm:"type:date"`       // 胚布交期
}

// CreateBidRequest 创建投标请求
type CreateBidRequest struct {
	BountyID    uint                      `json:"bountyId" binding:"required"`
	BidPrice    float64                   `json:"bidPrice" binding:"required,gt=0"`
	WovenSpec   *CreateBidWovenSpecRequest   `json:"wovenSpec"`
	KnittedSpec *CreateBidKnittedSpecRequest `json:"knittedSpec"`
}

// CreateBidWovenSpecRequest 创建投标-梭织规格请求
type CreateBidWovenSpecRequest struct {
	SizeLength        float64 `json:"sizeLength"`
	GreigeFabricType  string  `json:"greigeFabricType"`
	GreigeDeliveryDate string  `json:"greigeDeliveryDate"`
}

// CreateBidKnittedSpecRequest 创建投标-针织规格请求
type CreateBidKnittedSpecRequest struct {
	SizeWeight        float64 `json:"sizeWeight"`
	GreigeFabricType  string  `json:"greigeFabricType"`
	GreigeDeliveryDate string  `json:"greigeDeliveryDate"`
}

// BidResponse 投标响应
type BidResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Bid   `json:"data,omitempty"`
}

// BidListResponse 投标列表响应
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
