package client

import (
	"net/http"

	"back/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateOrUpdateProfile 客户档案创建/更新
// @Summary 创建或更新客户档案
// @Description 创建或更新当前登录 client 的公司信息
// @Tags client
// @Accept json
// @Produce json
// @Param request body CreateClientProfileRequest true "客户档案信息"
// @Success 200 {object} ClientProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/client/profile [post]
func (h *Handler) CreateOrUpdateProfile(c *gin.Context) {
	var req CreateClientProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.UserID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	p, err := h.service.CreateOrUpdateProfile(c.Request.Context(), rc.UserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建/更新客户档案失败: " + err.Error(),
		})
		return
	}

	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "client.create_or_update"
		rc.Resource = "client"
		rc.ResourceID = p.ID
	}

	c.JSON(http.StatusOK, ClientProfileResponse{
		Code:    http.StatusOK,
		Message: "客户档案保存成功",
		Data:    p,
	})
}

// GetProfile 获取客户档案
// @Summary 获取客户档案
// @Description 获取当前登录 client 的档案信息
// @Tags client
// @Produce json
// @Success 200 {object} ClientProfileResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/client/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.UserID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	p, err := h.service.GetProfile(c.Request.Context(), rc.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "客户档案不存在",
		})
		return
	}

	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "client.get_profile"
		rc.Resource = "client"
		rc.ResourceID = p.ID
	}

	c.JSON(http.StatusOK, ClientProfileResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    p,
	})
}

