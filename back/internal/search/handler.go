package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 搜索处理器
type Handler struct {
	service Service
}

// NewHandler 创建搜索处理器
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Search 统一搜索接口
// POST /api/v1/search
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// 验证 index
	validIndexes := map[string]bool{
		"bounty": true,
		// 未来可以扩展其他索引
	}
	if !validIndexes[req.Index] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid index: " + req.Index,
		})
		return
	}

	result, err := h.service.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Search failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SearchResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}
