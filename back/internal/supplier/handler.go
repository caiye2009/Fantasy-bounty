package supplier

import (
	"back/internal/user"
	"net/http"

	"back/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// Handler 供应商处理器
type Handler struct {
	service     Service
	userService user.Service
}

// NewHandler 创建新的处理器实例
func NewHandler(service Service, userService user.Service) *Handler {
	return &Handler{
		service:     service,
		userService: userService,
	}
}

// CreateOrUpdateProfile 创建/更新供应商档案
// @Summary 创建/更新供应商档案
// @Description 创建或更新供应商的企业信息和机器能力
// @Tags supplier
// @Accept json
// @Produce json
// @Param request body CreateSupplierRequest true "供应商信息"
// @Success 200 {object} SupplierProfileResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/supplier/profile [post]
func (h *Handler) CreateOrUpdateProfile(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 从上下文中获取用户ID
	userID := h.getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	profile, err := h.service.CreateOrUpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建/更新供应商档案失败: " + err.Error(),
		})
		return
	}

	// 设置审计信息
	h.setAuditInfo(c, "supplier.create_or_update", profile.ID)

	c.JSON(http.StatusOK, SupplierProfileResponse{
		Code:    http.StatusOK,
		Message: "供应商档案保存成功",
		Data:    profile,
	})
}

// GetProfile 获取供应商档案
// @Summary 获取供应商档案
// @Description 获取当前用户的供应商档案信息
// @Tags supplier
// @Produce json
// @Success 200 {object} SupplierProfileResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/supplier/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	profile, err := h.service.GetSupplierProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "供应商档案不存在",
		})
		return
	}

	// 设置审计信息
	h.setAuditInfo(c, "supplier.get_profile", profile.ID)

	c.JSON(http.StatusOK, SupplierProfileResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    profile,
	})
}

// UpdateCapabilities 更新机器能力
// @Summary 更新机器能力
// @Description 更新供应商的机器设备数量（增量更新）
// @Tags supplier
// @Accept json
// @Produce json
// @Param request body UpdateCapabilitiesRequest true "机器能力信息"
// @Success 200 {object} CapabilitiesResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/supplier/capabilities [post]
func (h *Handler) UpdateCapabilities(c *gin.Context) {
	var req UpdateCapabilitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	userID := h.getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	if err := h.service.UpdateCapabilities(c.Request.Context(), userID, req.Capabilities); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新机器能力失败: " + err.Error(),
		})
		return
	}

	// 设置审计信息
	h.setAuditInfo(c, "supplier.update_capabilities", "")

	c.JSON(http.StatusOK, CapabilitiesResponse{
		Code:    http.StatusOK,
		Message: "机器能力更新成功",
		Data:    req.Capabilities,
	})
}

// GetFullInfo 获取供应商完整信息
// @Summary 获取供应商完整信息
// @Description 获取供应商的企业信息、资质和生产能力
// @Tags supplier
// @Produce json
// @Success 200 {object} SupplierFullInfoResponse
// @Router /api/v1/supplier/full-info [get]
func (h *Handler) GetFullInfo(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户未登录",
		})
		return
	}

	fullInfo, err := h.service.GetSupplierFullInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取供应商信息失败: " + err.Error(),
		})
		return
	}

	// 设置审计信息
	var resourceID string
	if fullInfo.Info != nil {
		resourceID = fullInfo.Info.ID
	}
	h.setAuditInfo(c, "supplier.get_full_info", resourceID)

	message := "获取成功"
	if fullInfo.Info == nil {
		message = "供应商档案为空，请先创建档案"
	}

	c.JSON(http.StatusOK, SupplierFullInfoResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    fullInfo,
	})
}

// getUserIDFromContext 从上下文中获取用户ID
func (h *Handler) getUserIDFromContext(c *gin.Context) string {
	// 从JWT中间件中获取username
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		return ""
	}

	// 通过username查询用户信息
	user, err := h.userService.GetUserByUsername(c.Request.Context(), rc.Username)
	if err != nil {
		return ""
	}

	return user.ID
}

// setAuditInfo 设置审计信息
func (h *Handler) setAuditInfo(c *gin.Context, action, resource string) {
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = action
		rc.Resource = "supplier"
		if resource != "" {
			rc.ResourceID = resource
		}
	}
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}