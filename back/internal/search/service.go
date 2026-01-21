package search

import (
	"back/pkg/elasticsearch"
	"context"
	"math"
)

// Service 搜索服务接口
type Service interface {
	Search(ctx context.Context, req *SearchRequest) (*SearchResultData, error)
}

type service struct{}

// NewService 创建搜索服务
func NewService() Service {
	return &service{}
}

// Search 执行搜索
func (s *service) Search(ctx context.Context, req *SearchRequest) (*SearchResultData, error) {
	// 参数默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 构建搜索参数
	params := elasticsearch.SearchParams{
		Index:   req.Index,
		Query:   req.Query,
		Filters: req.Filters,
		From:    (req.Page - 1) * req.Size,
		Size:    req.Size,
	}

	// 处理排序
	if req.Sort != nil && req.Sort.Field != "" {
		order := req.Sort.Order
		if order == "" {
			order = "desc"
		}
		params.Sort = map[string]string{req.Sort.Field: order}
	}

	// 执行搜索
	result, err := elasticsearch.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(result.Total) / float64(req.Size)))

	// 解析聚合结果为筛选项
	filters := elasticsearch.ParseAggregationsToFilters(result.Aggregations)

	return &SearchResultData{
		Hits:       result.Hits,
		Total:      result.Total,
		Page:       req.Page,
		Size:       req.Size,
		TotalPages: totalPages,
		Filters:    filters,
	}, nil
}
