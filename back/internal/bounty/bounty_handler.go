package bounty

import (
	"back/pkg/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 悬赏处理器
type Handler struct {
	service Service
}

// NewHandler 创建新的 handler 实例
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateBounty 创建悬赏
// @Summary 创建悬赏
// @Description 创建新的悬赏任务
// @Tags bounty
// @Accept json
// @Produce json
// @Param bounty body CreateBountyRequest true "悬赏信息"
// @Success 201 {object} BountyResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/bounties [post]
func (h *Handler) CreateBounty(c *gin.Context) {
	// 临时：使用默认用户ID，跳过JWT验证
	userID, exists := middleware.GetUserID(c)
	if !exists {
		userID = 1 // 默认用户ID
	}

	var req CreateBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bounty, err := h.service.CreateBounty(c.Request.Context(), userID, &req)
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

// GetBounty 获取悬赏详情
// @Summary 获取悬赏详情
// @Description 根据 ID 获取悬赏详情
// @Tags bounty
// @Produce json
// @Param id path int true "悬赏 ID"
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

// UpdateBounty 更新悬赏
// @Summary 更新悬赏
// @Description 更新悬赏信息
// @Tags bounty
// @Accept json
// @Produce json
// @Param id path int true "悬赏 ID"
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

// DeleteBounty 删除悬赏
// @Summary 删除悬赏
// @Description 删除指定悬赏
// @Tags bounty
// @Produce json
// @Param id path int true "悬赏 ID"
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

// ListBounties 获取悬赏列表
// @Summary 获取悬赏列表
// @Description 分页获取悬赏列表
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

// PeekBounties 预览悬赏列表（公开接口）
// @Summary 预览悬赏列表
// @Description 公开接口，仅返回第一页10条数据，不支持筛选
// @Tags bounty
// @Produce json
// @Success 200 {object} BountyListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/bounties/peek [get]
func (h *Handler) PeekBounties(c *gin.Context) {
	// 固定只返回第一页，10条数据
	bounties, total, err := h.service.ListBounties(c.Request.Context(), 1, 10)
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
