package auth

import (
	"back/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 认证处理器
type Handler struct {
	jwtService *jwt.JWTService
}

// NewHandler 创建新的 handler 实例
func NewHandler(jwtService *jwt.JWTService) *Handler {
	return &Handler{jwtService: jwtService}
}

// Login 登录
// @Summary 用户登录
// @Description 用户登录获取JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} LoginResponse
// @Failure 401 {object} LoginResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 简单验证：账号密码都是 admin
	if req.Username != "admin" || req.Password != "admin" {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid username or password",
		})
		return
	}

	// 生成 JWT token (固定 userID 为 1)
	token, err := h.jwtService.GenerateToken(1, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Code:    http.StatusOK,
		Message: "Login successful",
		Token:   token,
	})
}
