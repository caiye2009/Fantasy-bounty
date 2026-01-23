package bounty

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Bounty 悬赏（聚合根）
type Bounty struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	BountyType string    `json:"bountyType" gorm:"type:varchar(50);not null"` 
	// 悬赏类型：woven（梭织） / knitted（针织）

	ProductName string `json:"productName" gorm:"type:varchar(255);not null"`
	// 产品品名

	ProductCode string `json:"productCode" gorm:"type:varchar(255);not null"`
	// 产品编码 / 产品编号

	SampleType string `json:"sampleType" gorm:"type:varchar(100)"`
	// 需要的样品类型

	ExpectedDeliveryDate time.Time `json:"expectedDeliveryDate" gorm:"type:date"`
	// 预计交货日期

	BidDeadline time.Time `json:"bidDeadline" gorm:"type:timestamp"`
	// 投标截止时间

	Status string `json:"status" gorm:"type:varchar(50);default:'open'"`
	// 悬赏状态：open / in_progress / completed / closed

	CreatedBy string `json:"createdBy" gorm:"type:uuid;not null;index"`
	// 发布人账号 ID (Account UUID)

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 悬赏规格（根据 BountyType 只会有一个有值）
	WovenSpec   *BountyWovenSpec   `json:"wovenSpec,omitempty" gorm:"foreignKey:BountyID"`
	KnittedSpec *BountyKnittedSpec `json:"knittedSpec,omitempty" gorm:"foreignKey:BountyID"`
}

// BountyWovenSpec 悬赏-梭织规格
type BountyWovenSpec struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	BountyID uint `json:"bountyId" gorm:"not null;uniqueIndex"`
	// 所属悬赏 ID（一对一）

	FabricWeight float64 `json:"fabricWeight" gorm:"type:decimal(10,2)"`
	// 成品克重（g/m²）

	FabricWidth float64 `json:"fabricWidth" gorm:"type:decimal(10,2)"`
	// 成品幅宽

	WarpDensity float64 `json:"warpDensity" gorm:"type:decimal(10,2)"`
	// 成品经密（经 / 英寸）

	WeftDensity float64 `json:"weftDensity" gorm:"type:decimal(10,2)"`
	// 成品纬密（纬 / 英寸）

	Composition Composition `json:"composition" gorm:"type:json"`
	// 面料成分，如 {"棉": 0.6, "涤纶": 0.4}

	WarpMaterial string `json:"warpMaterial" gorm:"type:varchar(255)"`
	// 经向原料

	WeftMaterial string `json:"weftMaterial" gorm:"type:varchar(255)"`
	// 纬向原料

	QuantityMeters float64 `json:"quantityMeters" gorm:"type:decimal(10,2)"`
	// 需求数量（米）
}

// BountyKnittedSpec 悬赏-针织规格
type BountyKnittedSpec struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	BountyID uint `json:"bountyId" gorm:"not null;uniqueIndex"`
	// 所属悬赏 ID（一对一）

	FabricWeight float64 `json:"fabricWeight" gorm:"type:decimal(10,2)"`
	// 成品克重（g/m²）

	FabricWidth float64 `json:"fabricWidth" gorm:"type:decimal(10,2)"`
	// 成品幅宽

	MachineType string `json:"machineType" gorm:"type:varchar(100)"`
	// 机型 / 针织设备类型

	Composition string `json:"composition" gorm:"type:varchar(255)"`
	// 面料成分（展示用，如：Cotton 60% / Polyester 40%）

	Materials Materials `json:"materials" gorm:"type:json"`
	// 原料明细（结构化，多原料组成）

	QuantityKg float64 `json:"quantityKg" gorm:"type:decimal(10,2)"`
	// 需求数量（kg）
}

// Material 单一原料
type Material struct {
	Name       string  `json:"name"`
	// 原料名称（如：Cotton、Polyester、Spandex）

	Percentage float64 `json:"percentage"`
	// 占比（百分比，如 60 表示 60%）
}

// Materials 原料列表
type Materials []Material

// Value 实现 driver.Valuer 接口，用于写入数据库
func (m Materials) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (m *Materials) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Materials: invalid type")
	}
	return json.Unmarshal(bytes, m)
}

// Composition 面料成分（map[原料名称]占比）
// 例如：{"棉": 0.6, "涤纶": 0.4}
type Composition map[string]float64

// Value 实现 driver.Valuer 接口，用于写入数据库
func (c Composition) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (c *Composition) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Composition: invalid type")
	}
	return json.Unmarshal(bytes, c)
}

// Validate 验证成分总和是否为 1
func (c Composition) Validate() error {
	if c == nil || len(c) == 0 {
		return nil // 允许为空
	}
	var total float64
	for _, v := range c {
		if v < 0 || v > 1 {
			return errors.New("composition percentage must be between 0 and 1")
		}
		total += v
	}
	// 允许小误差 (0.999 ~ 1.001)
	if total < 0.999 || total > 1.001 {
		return errors.New("composition percentages must sum to 1")
	}
	return nil
}

// String 格式化为展示字符串，如 "棉 60% / 涤纶 40%"
func (c Composition) String() string {
	if c == nil || len(c) == 0 {
		return ""
	}
	var parts []string
	for name, pct := range c {
		parts = append(parts, name+" "+formatPercent(pct))
	}
	return strings.Join(parts, " / ")
}

// formatPercent 将小数转换为百分比字符串
func formatPercent(v float64) string {
	pct := v * 100
	if pct == float64(int(pct)) {
		return fmt.Sprintf("%.0f%%", pct)
	}
	return fmt.Sprintf("%.1f%%", pct)
}

// CreateBountyRequest 创建悬赏请求
type CreateBountyRequest struct {
	BountyType           string                          `json:"bountyType" binding:"required,oneof=woven knitted"`
	ProductName          string                          `json:"productName" binding:"required,min=1,max=255"`
	ProductCode          string                          `json:"productCode"`
	SampleType           string                          `json:"sampleType"`
	ExpectedDeliveryDate string                          `json:"expectedDeliveryDate"`
	BidDeadline          string                          `json:"bidDeadline" binding:"required"`
	WovenSpec            *CreateBountyWovenSpecRequest   `json:"wovenSpec"`
	KnittedSpec          *CreateBountyKnittedSpecRequest `json:"knittedSpec"`
}

// CreateBountyWovenSpecRequest 创建悬赏-梭织规格请求
type CreateBountyWovenSpecRequest struct {
	FabricWeight   float64     `json:"fabricWeight"`
	FabricWidth    float64     `json:"fabricWidth"`
	WarpDensity    float64     `json:"warpDensity"`
	WeftDensity    float64     `json:"weftDensity"`
	Composition    Composition `json:"composition"` // {"棉": 0.6, "涤纶": 0.4}
	WarpMaterial   string      `json:"warpMaterial"`
	WeftMaterial   string      `json:"weftMaterial"`
	QuantityMeters float64     `json:"quantityMeters"`
}

// CreateBountyKnittedSpecRequest 创建悬赏-针织规格请求
type CreateBountyKnittedSpecRequest struct {
	FabricWeight float64   `json:"fabricWeight"`
	FabricWidth  float64   `json:"fabricWidth"`
	MachineType  string    `json:"machineType"`
	Composition  string    `json:"composition"`
	Materials    Materials `json:"materials"`
	QuantityKg   float64   `json:"quantityKg"`
}

// UpdateBountyRequest 更新悬赏请求
type UpdateBountyRequest struct {
	ProductName          string                          `json:"productName" binding:"omitempty,min=1,max=255"`
	ProductCode          string                          `json:"productCode"`
	SampleType           string                          `json:"sampleType"`
	ExpectedDeliveryDate string                          `json:"expectedDeliveryDate"`
	BidDeadline          string                          `json:"bidDeadline"`
	Status               string                          `json:"status" binding:"omitempty,oneof=open in_progress completed closed"`
	WovenSpec            *CreateBountyWovenSpecRequest   `json:"wovenSpec"`
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
