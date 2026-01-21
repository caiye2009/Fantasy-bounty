package search

// SearchRequest 统一搜索请求
type SearchRequest struct {
	Index   string                 `json:"index" binding:"required"` // 索引名，如 "bounty"
	Query   string                 `json:"query"`                    // 关键词搜索
	Filters map[string]interface{} `json:"filters"`                  // 筛选条件
	Sort    *SortOption            `json:"sort"`                     // 排序
	Page    int                    `json:"page"`                     // 页码，默认 1
	Size    int                    `json:"size"`                     // 每页数量，默认 10
}

// SortOption 排序选项
type SortOption struct {
	Field string `json:"field"` // 排序字段
	Order string `json:"order"` // asc 或 desc
}
