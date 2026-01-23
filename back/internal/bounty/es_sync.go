package bounty

import (
	"back/pkg/elasticsearch"
	"context"
	"fmt"
	"log"
	"strings"
)

// BountyDocument ES 文档结构
type BountyDocument struct {
	ID                   uint    `json:"id"`
	BountyType           string  `json:"bounty_type"`
	ProductName          string  `json:"product_name"`
	ProductCode          string  `json:"product_code,omitempty"`
	SampleType           string  `json:"sample_type,omitempty"`
	Status               string  `json:"status"`
	CreatedBy            string  `json:"created_by"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
	ExpectedDeliveryDate string  `json:"expected_delivery_date,omitempty"`
	BidDeadline          string  `json:"bid_deadline"`

	// 扁平化规格字段
	Composition  string  `json:"composition,omitempty"`
	FabricWeight float64 `json:"fabric_weight,omitempty"`
	FabricWidth  float64 `json:"fabric_width,omitempty"`

	// 梭织特有
	WarpDensity    float64 `json:"warp_density,omitempty"`
	WeftDensity    float64 `json:"weft_density,omitempty"`
	WarpMaterial   string  `json:"warp_material,omitempty"`
	WeftMaterial   string  `json:"weft_material,omitempty"`
	QuantityMeters float64 `json:"quantity_meters,omitempty"`

	// 针织特有
	MachineType string    `json:"machine_type,omitempty"`
	Materials   Materials `json:"materials,omitempty"`
	QuantityKg  float64   `json:"quantity_kg,omitempty"`

	// 全文搜索字段
	SearchText string `json:"search_text"`
}

// ToESDocument 将 Bounty 转换为 ES 文档
func (b *Bounty) ToESDocument() *BountyDocument {
	doc := &BountyDocument{
		ID:          b.ID,
		BountyType:  b.BountyType,
		ProductName: b.ProductName,
		ProductCode: b.ProductCode,
		SampleType:  b.SampleType,
		Status:      b.Status,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		BidDeadline: b.BidDeadline.Format("2006-01-02T15:04:05Z"),
	}

	if !b.ExpectedDeliveryDate.IsZero() {
		doc.ExpectedDeliveryDate = b.ExpectedDeliveryDate.Format("2006-01-02")
	}

	// 搜索文本拼接
	searchParts := []string{b.ProductName}
	if b.SampleType != "" {
		searchParts = append(searchParts, b.SampleType)
	}

	// 梭织规格
	if b.WovenSpec != nil {
		doc.Composition = b.WovenSpec.Composition.String() // 转换为展示字符串
		doc.FabricWeight = b.WovenSpec.FabricWeight
		doc.FabricWidth = b.WovenSpec.FabricWidth
		doc.WarpDensity = b.WovenSpec.WarpDensity
		doc.WeftDensity = b.WovenSpec.WeftDensity
		doc.WarpMaterial = b.WovenSpec.WarpMaterial
		doc.WeftMaterial = b.WovenSpec.WeftMaterial
		doc.QuantityMeters = b.WovenSpec.QuantityMeters

		// 添加成分名称到搜索文本
		for name := range b.WovenSpec.Composition {
			searchParts = append(searchParts, name)
		}
		if b.WovenSpec.WarpMaterial != "" {
			searchParts = append(searchParts, b.WovenSpec.WarpMaterial)
		}
		if b.WovenSpec.WeftMaterial != "" {
			searchParts = append(searchParts, b.WovenSpec.WeftMaterial)
		}
	}

	// 针织规格
	if b.KnittedSpec != nil {
		doc.Composition = b.KnittedSpec.Composition
		doc.FabricWeight = b.KnittedSpec.FabricWeight
		doc.FabricWidth = b.KnittedSpec.FabricWidth
		doc.MachineType = b.KnittedSpec.MachineType
		doc.Materials = b.KnittedSpec.Materials
		doc.QuantityKg = b.KnittedSpec.QuantityKg

		if b.KnittedSpec.Composition != "" {
			searchParts = append(searchParts, b.KnittedSpec.Composition)
		}
		if b.KnittedSpec.MachineType != "" {
			searchParts = append(searchParts, b.KnittedSpec.MachineType)
		}
		// 添加原料名称到搜索文本
		for _, m := range b.KnittedSpec.Materials {
			if m.Name != "" {
				searchParts = append(searchParts, m.Name)
			}
		}
	}

	doc.SearchText = strings.Join(searchParts, " ")
	return doc
}

// SyncToES 同步到 ES
func SyncToES(ctx context.Context, bounty *Bounty) error {
	doc := bounty.ToESDocument()
	return elasticsearch.IndexDocument(ctx, "bounty", fmt.Sprintf("%d", bounty.ID), doc)
}

// DeleteFromES 从 ES 删除
func DeleteFromES(ctx context.Context, id uint) error {
	return elasticsearch.DeleteDocument(ctx, "bounty", fmt.Sprintf("%d", id))
}

// SyncToESAsync 异步同步到 ES（不阻塞主流程）
func SyncToESAsync(bounty *Bounty) {
	go func() {
		if err := SyncToES(context.Background(), bounty); err != nil {
			log.Printf("Failed to sync bounty %d to ES: %v", bounty.ID, err)
		}
	}()
}

// DeleteFromESAsync 异步从 ES 删除（不阻塞主流程）
func DeleteFromESAsync(id uint) {
	go func() {
		if err := DeleteFromES(context.Background(), id); err != nil {
			log.Printf("Failed to delete bounty %d from ES: %v", id, err)
		}
	}()
}

// ReindexAllBounties 重建所有 bounty 的 ES 索引
// 返回成功数量和失败数量
func ReindexAllBounties(bounties []Bounty) (successCount int, failCount int) {
	ctx := context.Background()

	log.Printf("Starting reindex of %d bounties...", len(bounties))

	for i := range bounties {
		if err := SyncToES(ctx, &bounties[i]); err != nil {
			log.Printf("[FAIL] Bounty %d (%s): %v", bounties[i].ID, bounties[i].ProductName, err)
			failCount++
		} else {
			log.Printf("[OK] Bounty %d: %s", bounties[i].ID, bounties[i].ProductName)
			successCount++
		}
	}

	log.Printf("Reindex completed: %d success, %d failed", successCount, failCount)
	return successCount, failCount
}
