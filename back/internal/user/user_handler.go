package user

import (
	"back/pkg/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateUser 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, UserResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		Code:    http.StatusCreated,
		Message: "User created successfully",
		Data:    user,
	})
}

// GetUser 获取用户（只能查看自己的）
func (h *Handler) GetUser(c *gin.Context) {
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, UserResponse{Code: http.StatusUnauthorized, Message: "user not authenticated"})
		return
	}

	id := c.Param("id")

	// 先获取用户信息
	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, UserResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	// 检查是否是自己的账号
	if user.Username != rc.Username {
		c.JSON(http.StatusForbidden, UserResponse{Code: http.StatusForbidden, Message: "只能查看自己的账号"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    user,
	})
}

// UpdateUser 更新用户（只能更新自己的）
func (h *Handler) UpdateUser(c *gin.Context) {
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, UserResponse{Code: http.StatusUnauthorized, Message: "user not authenticated"})
		return
	}

	id := c.Param("id")

	// 先获取用户信息
	existingUser, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, UserResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	// 检查是否是自己的账号
	if existingUser.Username != rc.Username {
		c.JSON(http.StatusForbidden, UserResponse{Code: http.StatusForbidden, Message: "只能更新自己的账号"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, UserResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		Code:    http.StatusOK,
		Message: "User updated successfully",
		Data:    user,
	})
}

// DeleteUser 删除用户（只能删除自己的）
func (h *Handler) DeleteUser(c *gin.Context) {
	rc := middleware.GetRequestContext(c)
	if rc == nil || rc.Username == "" {
		c.JSON(http.StatusUnauthorized, UserResponse{Code: http.StatusUnauthorized, Message: "user not authenticated"})
		return
	}

	id := c.Param("id")

	// 先获取用户信息
	existingUser, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, UserResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	// 检查是否是自己的账号
	if existingUser.Username != rc.Username {
		c.JSON(http.StatusForbidden, UserResponse{Code: http.StatusForbidden, Message: "只能删除自己的账号"})
		return
	}

	err = h.service.DeleteUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		Code:    http.StatusOK,
		Message: "User deleted successfully",
	})
}

// ListUsers 获取用户列表
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserListResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserListResponse{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    users,
		Total:   total,
	})
}
