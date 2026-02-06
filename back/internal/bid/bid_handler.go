package bid

import (
	"back/internal/supplier"
	"back/pkg/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 竞标处理器
type Handler struct {
	service         Service
	supplierService supplier.Service
}

// NewHandler 创建新的 handler 实例
func NewHandler(service Service, supplierService supplier.Service) *Handler {
	return &Handler{
		service:         service,
		supplierService: supplierService,
	}
}

// CreateBid 创建竞标
// @Summary 创建竞标
// @Description 为指定赏金创建竞标
// @Tags bid
// @Accept json
// @Produce json
// @Param bid body CreateBidRequest true "竞标信息"
// @Success 201 {object} BidResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/bids [post]
func (h *Handler) CreateBid(c *gin.Context) {
	// 从 RequestContext 中获取用户名
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "user not authenticated",
		})
		return
	}
	username := rc.Username

	// 检查用户是否已认证供应商
	verified, err := h.supplierService.IsUserVerified(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "检查供应商认证状态失败",
		})
		return
	}
	if !verified {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "请先完成供应商认证后再进行竞标",
		})
		return
	}

	var req CreateBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	bid, err := h.service.CreateBid(c.Request.Context(), username, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// 设置审计信息
	rc.Action = "bid.create"
	rc.Resource = "bid"
	rc.ResourceID = bid.ID
	rc.Detail = map[string]any{
		"bounty_id": req.BountyID,
	}

	c.JSON(http.StatusCreated, BidResponse{
		Code:    http.StatusCreated,
		Message: "Bid created successfully",
		Data:    bid,
	})
}

// ListBids 获取竞标列表
// @Summary 获取竞标列表
// @Description 根据赏金ID获取竞标列表
// @Tags bid
// @Produce json
// @Param bounty_id query int true "赏金ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} BidListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/bids [get]
func (h *Handler) ListBids(c *gin.Context) {
	bountyIDStr := c.Query("bounty_id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "bounty_id is required",
		})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid bounty_id",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	bids, total, err := h.service.ListBidsByBountyID(c.Request.Context(), uint(bountyID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BidListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    bids,
		Total:   total,
	})
}

// ListMyBids 获取我的竞标列表
// @Summary 获取我的竞标列表
// @Description 获取当前用户的竞标列表
// @Tags bid
// @Produce json
// @Param status query string false "状态筛选: draft, pending, in_progress, pending_acceptance, completed"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} BidListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/bids/my [get]
func (h *Handler) ListMyBids(c *gin.Context) {
	// 从 RequestContext 中获取用户名
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "user not authenticated",
		})
		return
	}
	username := rc.Username

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	bids, total, err := h.service.ListBidsByUsername(c.Request.Context(), username, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BidListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    bids,
		Total:   total,
	})
}

// DeleteBid 删除竞标
// @Summary 删除竞标
// @Description 删除指定竞标
// @Tags bid
// @Produce json
// @Param id path string true "竞标ID (UUID)"
// @Success 200 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/bids/{id} [delete]
func (h *Handler) DeleteBid(c *gin.Context) {
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "user not authenticated",
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid bid ID",
		})
		return
	}

	// 先查询竞标，校验所有权
	existingBid, err := h.service.GetBid(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "bid not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}
	if existingBid.Username != rc.Username {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "只能删除自己的竞标",
		})
		return
	}

	err = h.service.DeleteBid(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "bid.delete"
		rc.Resource = "bid"
		rc.ResourceID = id
	}

	c.JSON(http.StatusOK, ErrorResponse{
		Code:    http.StatusOK,
		Message: "Bid deleted successfully",
	})
}
