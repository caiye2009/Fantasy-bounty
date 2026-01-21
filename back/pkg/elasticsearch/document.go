package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// IndexDocument 索引文档（创建或更新）
func IndexDocument(ctx context.Context, index string, id string, doc interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	res, err := client.Index(
		index,
		&buf,
		client.Index.WithContext(ctx),
		client.Index.WithDocumentID(id),
		client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index document error: %s", res.String())
	}

	return nil
}

// DeleteDocument 删除文档
func DeleteDocument(ctx context.Context, index string, id string) error {
	res, err := client.Delete(
		index,
		id,
		client.Delete.WithContext(ctx),
		client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	// 404 表示文档不存在，忽略此错误
	if res.StatusCode == 404 {
		return nil
	}

	if res.IsError() {
		return fmt.Errorf("delete document error: %s", res.String())
	}

	return nil
}

// GetDocument 获取文档
func GetDocument(ctx context.Context, index string, id string) (map[string]interface{}, error) {
	res, err := client.Get(
		index,
		id,
		client.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if res.IsError() {
		return nil, fmt.Errorf("get document error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if source, ok := result["_source"].(map[string]interface{}); ok {
		return source, nil
	}

	return nil, nil
}

// BulkIndex 批量索引文档
func BulkIndex(ctx context.Context, index string, docs map[string]interface{}) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for id, doc := range docs {
		// 写入操作元数据
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    id,
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return fmt.Errorf("failed to encode meta: %w", err)
		}

		// 写入文档数据
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return fmt.Errorf("failed to encode document: %w", err)
		}
	}

	res, err := client.Bulk(
		&buf,
		client.Bulk.WithContext(ctx),
		client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.String())
	}

	return nil
}
