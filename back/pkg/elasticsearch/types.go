package elasticsearch

// Config ES 连接配置
type Config struct {
	Addresses []string
	Username  string
	Password  string
}

// SearchParams 搜索参数
type SearchParams struct {
	Index   string
	Query   string
	Filters map[string]interface{}
	Sort    map[string]string
	From    int
	Size    int
}

// SearchResult 搜索结果
type SearchResult struct {
	Hits         []map[string]interface{}
	Total        int64
	Aggregations map[string]interface{}
}

// FilterBucket 筛选项桶
type FilterBucket struct {
	Field   string       `json:"field"`
	Label   string       `json:"label"`
	Type    string       `json:"type"` // terms | range | date_range
	Buckets []BucketItem `json:"buckets"`
}

// BucketItem 桶项
type BucketItem struct {
	Key      interface{} `json:"key"`
	Label    string      `json:"label"`
	DocCount int64       `json:"docCount"`
}

// AggregationConfig 聚合配置
type AggregationConfig struct {
	Field      string
	Label      string
	Type       string            // terms | range | date_range
	Size       int
	Ranges     []RangeConfig     // 用于 range 类型
	DateRanges []DateRangeConfig // 用于 date_range 类型
}

// RangeConfig 范围配置
type RangeConfig struct {
	Key  string
	From *float64
	To   *float64
}

// DateRangeConfig 日期范围配置
type DateRangeConfig struct {
	Key  string
	From string // 支持 "now-7d", "now-30d" 等相对时间
	To   string
}
