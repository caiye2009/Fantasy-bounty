package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// BountyMapping bounty 索引的 mapping 定义
var BountyMapping = map[string]interface{}{
	"settings": map[string]interface{}{
		"number_of_shards":   1,
		"number_of_replicas": 0,
	},
	"mappings": map[string]interface{}{
		"properties": map[string]interface{}{
			"id":          map[string]interface{}{"type": "keyword"},
			"bounty_type": map[string]interface{}{"type": "keyword"},
			"product_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},

			// ✅ 新增字段
			"product_code": map[string]interface{}{"type": "keyword"},
			"materials": map[string]interface{}{
				"type": "nested",
				"properties": map[string]interface{}{
					"name":       map[string]interface{}{"type": "keyword"},
					"percentage": map[string]interface{}{"type": "float"},
				},
			},

			"sample_type": map[string]interface{}{"type": "keyword"},
			"status":      map[string]interface{}{"type": "keyword"},
			"created_by":  map[string]interface{}{"type": "keyword"},
			"created_at":  map[string]interface{}{"type": "date"},
			"updated_at":  map[string]interface{}{"type": "date"},
			"expected_delivery_date": map[string]interface{}{"type": "date"},
			"bid_deadline":           map[string]interface{}{"type": "date"},

			// 规格字段（扁平化）
			"composition": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "keyword"},
				},
			},
			"fabric_weight": map[string]interface{}{"type": "float"},
			"fabric_width":  map[string]interface{}{"type": "float"},

			// 梭织特有
			"warp_density":    map[string]interface{}{"type": "float"},
			"weft_density":    map[string]interface{}{"type": "float"},
			"warp_material":   map[string]interface{}{"type": "keyword"},
			"weft_material":   map[string]interface{}{"type": "keyword"},
			"quantity_meters": map[string]interface{}{"type": "float"},

			// 针织特有
			"factory_article_no": map[string]interface{}{"type": "keyword"},
			"machine_type":       map[string]interface{}{"type": "keyword"},
			"quantity_kg":        map[string]interface{}{"type": "float"},

			// 全文搜索字段
			"search_text": map[string]interface{}{"type": "text"},
		},
	},
}

// CreateIndex 创建索引
func CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	res, err := client.Indices.Create(
		index,
		client.Indices.Create.WithContext(ctx),
		client.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}

	return nil
}

// IndexExists 检查索引是否存在
func IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := client.Indices.Exists(
		[]string{index},
		client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check index exists: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// EnsureIndex 确保索引存在，不存在则创建
func EnsureIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	exists, err := IndexExists(ctx, index)
	if err != nil {
		return err
	}

	if !exists {
		return CreateIndex(ctx, index, mapping)
	}

	return nil
}

// DeleteIndex 删除索引
func DeleteIndex(ctx context.Context, index string) error {
	res, err := client.Indices.Delete(
		[]string{index},
		client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.String())
	}

	return nil
}