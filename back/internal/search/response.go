package search

import "back/pkg/elasticsearch"

// SearchResponse 搜索响应
type SearchResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *SearchResultData `json:"data"`
}

// SearchResultData 搜索结果数据
type SearchResultData struct {
	Hits       []map[string]interface{}     `json:"hits"`       // 搜索结果
	Total      int64                        `json:"total"`      // 总数
	Page       int                          `json:"page"`       // 当前页
	Size       int                          `json:"size"`       // 每页数量
	TotalPages int                          `json:"totalPages"` // 总页数
	Filters    []elasticsearch.FilterBucket `json:"filters"`    // 动态筛选项
}
