package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// DefaultAggregations 默认聚合配置
var DefaultAggregations = []AggregationConfig{
	{Field: "bounty_type", Label: "悬赏类型", Type: "terms", Size: 10},
	{Field: "sample_type", Label: "样品类型", Type: "terms", Size: 20},
	{Field: "composition.keyword", Label: "成分", Type: "terms", Size: 20},
	{Field: "machine_type", Label: "机型", Type: "terms", Size: 10},
	{Field: "warp_material", Label: "经向原料", Type: "terms", Size: 10},
	{Field: "weft_material", Label: "纬向原料", Type: "terms", Size: 10},
	{
		Field: "fabric_weight",
		Label: "克重范围",
		Type:  "range",
		Ranges: []RangeConfig{
			{Key: "0-100", To: ptrFloat64(100)},
			{Key: "100-200", From: ptrFloat64(100), To: ptrFloat64(200)},
			{Key: "200-300", From: ptrFloat64(200), To: ptrFloat64(300)},
			{Key: "300+", From: ptrFloat64(300)},
		},
	},
	// 发布时间范围
	{
		Field: "created_at",
		Label: "发布时间",
		Type:  "date_range",
		DateRanges: []DateRangeConfig{
			{Key: "today", From: "now/d", To: "now"},
			{Key: "3days", From: "now-3d/d", To: "now"},
			{Key: "7days", From: "now-7d/d", To: "now"},
			{Key: "30days", From: "now-30d/d", To: "now"},
		},
	},
	// 截止接单时间范围
	{
		Field: "bid_deadline",
		Label: "截止接单",
		Type:  "date_range",
		DateRanges: []DateRangeConfig{
			{Key: "today", From: "now/d", To: "now+1d/d"},
			{Key: "3days", From: "now/d", To: "now+3d/d"},
			{Key: "7days", From: "now/d", To: "now+7d/d"},
			{Key: "30days", From: "now/d", To: "now+30d/d"},
		},
	},
	// 梭织需求量范围
	{
		Field: "quantity_meters",
		Label: "需求量(米)",
		Type:  "range",
		Ranges: []RangeConfig{
			{Key: "0-1000", To: ptrFloat64(1000)},
			{Key: "1000-5000", From: ptrFloat64(1000), To: ptrFloat64(5000)},
			{Key: "5000-10000", From: ptrFloat64(5000), To: ptrFloat64(10000)},
			{Key: "10000+", From: ptrFloat64(10000)},
		},
	},
	// 针织需求量范围
	{
		Field: "quantity_kg",
		Label: "需求量(kg)",
		Type:  "range",
		Ranges: []RangeConfig{
			{Key: "0-100", To: ptrFloat64(100)},
			{Key: "100-500", From: ptrFloat64(100), To: ptrFloat64(500)},
			{Key: "500-1000", From: ptrFloat64(500), To: ptrFloat64(1000)},
			{Key: "1000+", From: ptrFloat64(1000)},
		},
	},
}

func ptrFloat64(v float64) *float64 {
	return &v
}

// Search 执行搜索
func Search(ctx context.Context, params SearchParams) (*SearchResult, error) {
	query := buildQuery(params)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(params.Index),
		client.Search.WithBody(&buf),
		client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseSearchResult(result), nil
}

// 日期范围 key 到实际范围的映射
var dateRangeKeyMappings = map[string]map[string]map[string]string{
	"created_at": {
		"today":  {"gte": "now/d", "lte": "now"},
		"3days":  {"gte": "now-3d/d", "lte": "now"},
		"7days":  {"gte": "now-7d/d", "lte": "now"},
		"30days": {"gte": "now-30d/d", "lte": "now"},
	},
	"bid_deadline": {
		"today":  {"gte": "now/d", "lt": "now+1d/d"},
		"3days":  {"gte": "now/d", "lt": "now+3d/d"},
		"7days":  {"gte": "now/d", "lt": "now+7d/d"},
		"30days": {"gte": "now/d", "lt": "now+30d/d"},
	},
}

// 数值范围 key 到实际范围的映射
var numericRangeKeyMappings = map[string]map[string]map[string]interface{}{
	"quantity_meters": {
		"0-1000":     {"lt": float64(1000)},
		"1000-5000":  {"gte": float64(1000), "lt": float64(5000)},
		"5000-10000": {"gte": float64(5000), "lt": float64(10000)},
		"10000+":     {"gte": float64(10000)},
	},
	"quantity_kg": {
		"0-100":    {"lt": float64(100)},
		"100-500":  {"gte": float64(100), "lt": float64(500)},
		"500-1000": {"gte": float64(500), "lt": float64(1000)},
		"1000+":    {"gte": float64(1000)},
	},
	"fabric_weight": {
		"0-100":   {"lt": float64(100)},
		"100-200": {"gte": float64(100), "lt": float64(200)},
		"200-300": {"gte": float64(200), "lt": float64(300)},
		"300+":    {"gte": float64(300)},
	},
}

// buildQuery 构建查询
func buildQuery(params SearchParams) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	// 关键词搜索
	if params.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  params.Query,
				"fields": []string{"product_name^3", "composition^2", "search_text", "sample_type"},
				"type":   "best_fields",
			},
		})
	}

	// 筛选条件
	for field, value := range params.Filters {
		// 检查是否是日期范围 key
		if dateRanges, ok := dateRangeKeyMappings[field]; ok {
			if key, isString := value.(string); isString {
				if rangeVal, exists := dateRanges[key]; exists {
					filter = append(filter, map[string]interface{}{
						"range": map[string]interface{}{field: rangeVal},
					})
					continue
				}
			}
		}

		// 检查是否是数值范围 key
		if numericRanges, ok := numericRangeKeyMappings[field]; ok {
			if key, isString := value.(string); isString {
				if rangeVal, exists := numericRanges[key]; exists {
					filter = append(filter, map[string]interface{}{
						"range": map[string]interface{}{field: rangeVal},
					})
					continue
				}
			}
		}

		switch v := value.(type) {
		case []interface{}:
			if len(v) > 0 {
				filter = append(filter, map[string]interface{}{
					"terms": map[string]interface{}{field: v},
				})
			}
		case []string:
			if len(v) > 0 {
				filter = append(filter, map[string]interface{}{
					"terms": map[string]interface{}{field: v},
				})
			}
		case map[string]interface{}:
			// range 查询
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{field: v},
			})
		default:
			// term 查询
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{field: v},
			})
		}
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filter,
			},
		},
		"from": params.From,
		"size": params.Size,
		"aggs": buildAggregations(DefaultAggregations),
	}

	// 排序
	if len(params.Sort) > 0 {
		var sort []map[string]interface{}
		for field, order := range params.Sort {
			sort = append(sort, map[string]interface{}{
				field: map[string]interface{}{"order": order},
			})
		}
		query["sort"] = sort
	} else {
		// 默认按创建时间降序
		query["sort"] = []map[string]interface{}{
			{"created_at": map[string]interface{}{"order": "desc"}},
		}
	}

	return query
}

// buildAggregations 构建聚合查询
func buildAggregations(configs []AggregationConfig) map[string]interface{} {
	aggs := make(map[string]interface{})

	for _, cfg := range configs {
		switch cfg.Type {
		case "terms":
			size := cfg.Size
			if size == 0 {
				size = 10
			}
			aggs[cfg.Field] = map[string]interface{}{
				"terms": map[string]interface{}{
					"field": cfg.Field,
					"size":  size,
				},
			}
		case "range":
			ranges := make([]map[string]interface{}, 0, len(cfg.Ranges))
			for _, r := range cfg.Ranges {
				rangeItem := map[string]interface{}{"key": r.Key}
				if r.From != nil {
					rangeItem["from"] = *r.From
				}
				if r.To != nil {
					rangeItem["to"] = *r.To
				}
				ranges = append(ranges, rangeItem)
			}
			aggs[cfg.Field+"_ranges"] = map[string]interface{}{
				"range": map[string]interface{}{
					"field":  cfg.Field,
					"ranges": ranges,
				},
			}
		case "date_range":
			ranges := make([]map[string]interface{}, 0, len(cfg.DateRanges))
			for _, r := range cfg.DateRanges {
				rangeItem := map[string]interface{}{"key": r.Key}
				if r.From != "" {
					rangeItem["from"] = r.From
				}
				if r.To != "" {
					rangeItem["to"] = r.To
				}
				ranges = append(ranges, rangeItem)
			}
			aggs[cfg.Field+"_ranges"] = map[string]interface{}{
				"date_range": map[string]interface{}{
					"field":  cfg.Field,
					"ranges": ranges,
				},
			}
		}
	}

	return aggs
}

// parseSearchResult 解析搜索结果
func parseSearchResult(result map[string]interface{}) *SearchResult {
	sr := &SearchResult{
		Hits:         make([]map[string]interface{}, 0),
		Aggregations: make(map[string]interface{}),
	}

	// 解析 total
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				sr.Total = int64(value)
			}
		}

		// 解析 hits
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						sr.Hits = append(sr.Hits, source)
					}
				}
			}
		}
	}

	// 解析 aggregations
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		sr.Aggregations = aggs
	}

	return sr
}

// ParseAggregationsToFilters 将聚合结果转换为筛选项
func ParseAggregationsToFilters(aggs map[string]interface{}) []FilterBucket {
	filters := make([]FilterBucket, 0)

	// 字段标签映射
	fieldLabels := map[string]string{
		"bounty_type":           "悬赏类型",
		"sample_type":           "样品类型",
		"composition.keyword":   "成分",
		"machine_type":          "机型",
		"warp_material":         "经向原料",
		"weft_material":         "纬向原料",
		"fabric_weight_ranges":  "克重范围",
		"created_at_ranges":     "发布时间",
		"bid_deadline_ranges":   "截止接单",
		"quantity_meters_ranges": "需求量(米)",
		"quantity_kg_ranges":    "需求量(kg)",
	}

	// 值标签映射
	valueLabels := map[string]map[string]string{
		"bounty_type": {"woven": "梭织", "knitted": "针织"},
	}

	// 范围标签映射（按字段区分）
	rangeLabels := map[string]map[string]string{
		"fabric_weight_ranges": {
			"0-100":   "100g以下",
			"100-200": "100-200g",
			"200-300": "200-300g",
			"300+":    "300g以上",
		},
		"created_at_ranges": {
			"today":  "今天",
			"3days":  "近3天",
			"7days":  "近7天",
			"30days": "近30天",
		},
		"bid_deadline_ranges": {
			"today":  "今天截止",
			"3days":  "3天内",
			"7days":  "7天内",
			"30days": "30天内",
		},
		"quantity_meters_ranges": {
			"0-1000":     "1000米以下",
			"1000-5000":  "1000-5000米",
			"5000-10000": "5000-10000米",
			"10000+":     "10000米以上",
		},
		"quantity_kg_ranges": {
			"0-100":    "100kg以下",
			"100-500":  "100-500kg",
			"500-1000": "500-1000kg",
			"1000+":    "1000kg以上",
		},
	}

	for field, data := range aggs {
		aggData, ok := data.(map[string]interface{})
		if !ok {
			continue
		}

		filter := FilterBucket{
			Field: field,
			Label: fieldLabels[field],
		}

		// 处理 buckets
		if buckets, ok := aggData["buckets"].([]interface{}); ok {
			// 判断是否是 range 类型（字段名以 _ranges 结尾的都是范围类型）
			if strings.HasSuffix(field, "_ranges") {
				filter.Type = "range"
			} else {
				filter.Type = "terms"
			}

			for _, b := range buckets {
				bucket, ok := b.(map[string]interface{})
				if !ok {
					continue
				}

				key := bucket["key"]
				docCount := int64(0)
				if dc, ok := bucket["doc_count"].(float64); ok {
					docCount = int64(dc)
				}

				// 跳过空桶
				if docCount == 0 {
					continue
				}

				// 获取标签
				label := fmt.Sprintf("%v", key)
				if labels, ok := valueLabels[field]; ok {
					if l, ok := labels[fmt.Sprintf("%v", key)]; ok {
						label = l
					}
				}
				// range 类型的标签（按字段查找）
				if filter.Type == "range" {
					if fieldLabels, ok := rangeLabels[field]; ok {
						if l, ok := fieldLabels[fmt.Sprintf("%v", key)]; ok {
							label = l
						}
					}
				}

				filter.Buckets = append(filter.Buckets, BucketItem{
					Key:      key,
					Label:    label,
					DocCount: docCount,
				})
			}
		}

		// 只添加有数据的筛选项
		if len(filter.Buckets) > 0 {
			filters = append(filters, filter)
		}
	}

	return filters
}
