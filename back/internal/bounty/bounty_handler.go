package bounty

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 赏金处理器
type Handler struct {
	service Service
}

// NewHandler 创建新的 handler 实例
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateBounty 创建赏金
// @Summary 创建赏金
// @Description 创建新的赏金任务
// @Tags bounty
// @Accept json
// @Produce json
// @Param bounty body CreateBountyRequest true "赏金信息"
// @Success 201 {object} BountyResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/bounties [post]
func (h *Handler) CreateBounty(c *gin.Context) {
	var req CreateBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bounty, err := h.service.CreateBounty(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, BountyResponse{
		Code:    http.StatusCreated,
		Message: "Bounty created successfully",
		Data:    bounty,
	})
}

// GetBounty 获取赏金详情
// @Summary 获取赏金详情
// @Description 根据 ID 获取赏金详情
// @Tags bounty
// @Produce json
// @Param id path int true "赏金 ID"
// @Success 200 {object} BountyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/bounties/{id} [get]
func (h *Handler) GetBounty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid bounty ID",
		})
		return
	}

	bounty, err := h.service.GetBounty(c.Request.Context(), uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "bounty not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BountyResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    bounty,
	})
}

// UpdateBounty 更新赏金
// @Summary 更新赏金
// @Description 更新赏金信息
// @Tags bounty
// @Accept json
// @Produce json
// @Param id path int true "赏金 ID"
// @Param bounty body UpdateBountyRequest true "更新信息"
// @Success 200 {object} BountyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/bounties/{id} [put]
func (h *Handler) UpdateBounty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid bounty ID",
		})
		return
	}

	var req UpdateBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bounty, err := h.service.UpdateBounty(c.Request.Context(), uint(id), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "bounty not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BountyResponse{
		Code:    http.StatusOK,
		Message: "Bounty updated successfully",
		Data:    bounty,
	})
}

// DeleteBounty 删除赏金
// @Summary 删除赏金
// @Description 删除指定赏金
// @Tags bounty
// @Produce json
// @Param id path int true "赏金 ID"
// @Success 200 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/bounties/{id} [delete]
func (h *Handler) DeleteBounty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid bounty ID",
		})
		return
	}

	err = h.service.DeleteBounty(c.Request.Context(), uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "bounty not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ErrorResponse{
		Code:    http.StatusOK,
		Message: "Bounty deleted successfully",
	})
}

// ListBounties 获取赏金列表
// @Summary 获取赏金列表
// @Description 分页获取赏金列表
// @Tags bounty
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} BountyListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/bounties [get]
func (h *Handler) ListBounties(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	bounties, total, err := h.service.ListBounties(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BountyListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    bounties,
		Total:   total,
	})
}
