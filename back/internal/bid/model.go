package bid

import (
	"time"
)

// 投标状态常量
const (
	BidStatusPending           = "pending"            // 审核中
	BidStatusInProgress        = "in_progress"        // 进行中
	BidStatusPendingAcceptance = "pending_acceptance" // 待验收
	BidStatusCompleted         = "completed"          // 已完成
)

// ValidBidStatuses 有效的投标状态列表
var ValidBidStatuses = []string{
	BidStatusPending,
	BidStatusInProgress,
	BidStatusPendingAcceptance,
	BidStatusCompleted,
}

// IsValidBidStatus 检查状态是否有效
func IsValidBidStatus(status string) bool {
	for _, s := range ValidBidStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// Bid 投标模型
type Bid struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	BountyID  uint      `json:"bountyId" gorm:"not null;index"`
	Username  string    `json:"username" gorm:"type:varchar(20);not null;index"` // 关联用户名
	BidPrice  float64   `json:"bidPrice" gorm:"type:decimal(10,2);not null"`     // 投标价格
	Status    string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, in_progress, pending_acceptance, completed
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联规格 - 根据悬赏类型只会有一个有值
	WovenSpec   *BidWovenSpec   `json:"wovenSpec,omitempty" gorm:"foreignKey:BidID"`
	KnittedSpec *BidKnittedSpec `json:"knittedSpec,omitempty" gorm:"foreignKey:BidID"`
}

// BidWovenSpec 投标-梭织规格
type BidWovenSpec struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	BidID              string    `json:"bidId" gorm:"type:uuid;not null;uniqueIndex"`
	SizeLength         float64   `json:"sizeLength" gorm:"type:decimal(10,2)"`       // 尺码（长度）
	GreigeFabricType   string    `json:"greigeFabricType" gorm:"type:varchar(100)"`  // 胚布类型（现货/定织）
	GreigeDeliveryDate time.Time `json:"greigeDeliveryDate" gorm:"type:date"`        // 胚布交期
	DeliveryMethod     string    `json:"deliveryMethod" gorm:"type:varchar(100)"`    // 交货方式
}

// BidKnittedSpec 投标-针织规格
type BidKnittedSpec struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	BidID              string    `json:"bidId" gorm:"type:uuid;not null;uniqueIndex"`
	SizeWeight         float64   `json:"sizeWeight" gorm:"type:decimal(10,2)"`       // 尺码（重量/皮重）
	GreigeFabricType   string    `json:"greigeFabricType" gorm:"type:varchar(100)"`  // 胚布类型（现货/定织）
	GreigeDeliveryDate time.Time `json:"greigeDeliveryDate" gorm:"type:date"`        // 胚布交期
	DeliveryMethod     string    `json:"deliveryMethod" gorm:"type:varchar(100)"`    // 交货方式
}

// CreateBidRequest 创建投标请求
type CreateBidRequest struct {
	BountyID    uint                         `json:"bountyId" binding:"required"`
	BidPrice    float64                      `json:"bidPrice" binding:"required,gt=0"`
	WovenSpec   *CreateBidWovenSpecRequest   `json:"wovenSpec"`
	KnittedSpec *CreateBidKnittedSpecRequest `json:"knittedSpec"`
}

// CreateBidWovenSpecRequest 创建投标-梭织规格请求
type CreateBidWovenSpecRequest struct {
	SizeLength         float64 `json:"sizeLength"`
	GreigeFabricType   string  `json:"greigeFabricType"`   // 现货 / 定织
	GreigeDeliveryDate string  `json:"greigeDeliveryDate"`
	DeliveryMethod     string  `json:"deliveryMethod"`     // 竞标确认后 / 签订合同后 / 收到预付款后
}

// CreateBidKnittedSpecRequest 创建投标-针织规格请求
type CreateBidKnittedSpecRequest struct {
	SizeWeight         float64 `json:"sizeWeight"`
	GreigeFabricType   string  `json:"greigeFabricType"`   // 现货 / 定织
	GreigeDeliveryDate string  `json:"greigeDeliveryDate"`
	DeliveryMethod     string  `json:"deliveryMethod"`     // 竞标确认后 / 签订合同后 / 收到预付款后
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
