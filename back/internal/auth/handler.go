package auth

import (
	"back/internal/user"
	"back/pkg/crypto"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 验证码存储（内存，生产环境应使用 Redis）
var (
	codeStore     = make(map[string]codeEntry)
	codeStoreLock sync.RWMutex
)

type codeEntry struct {
	Code      string
	ExpiresAt time.Time
}

// Handler 认证处理器
type Handler struct {
	jwtService  *jwt.JWTService
	userService user.Service
}

// NewHandler 创建新的 handler 实例
func NewHandler(jwtService *jwt.JWTService, userService user.Service) *Handler {
	// 启动后台清理过期验证码（每2分钟执行一次）
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			codeStoreLock.Lock()
			for phone, entry := range codeStore {
				if now.After(entry.ExpiresAt) {
					delete(codeStore, phone)
				}
			}
			codeStoreLock.Unlock()
		}
	}()

	return &Handler{
		jwtService:  jwtService,
		userService: userService,
	}
}

// SendCode 发送验证码
// @Summary 发送验证码
// @Description 向手机号发送登录/注册验证码
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SendCodeRequest true "手机号"
// @Success 200 {object} SendCodeResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/send-code [post]
func (h *Handler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 存储验证码（1分钟有效）
	codeStoreLock.Lock()
	codeStore[req.Phone] = codeEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(1 * time.Minute),
	}
	codeStoreLock.Unlock()

	// 打印验证码到控制台（模拟短信发送）
	fmt.Println("========================================")
	fmt.Printf("手机号: %s\n", req.Phone)
	fmt.Printf("验证码: %s\n", code)
	fmt.Printf("有效期: 1分钟\n")
	fmt.Println("========================================")

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "auth.send_code"
		rc.Resource = "auth"
		rc.Detail = map[string]any{
			"phone_masked": crypto.MaskPhone(req.Phone),
		}
	}

	c.JSON(http.StatusOK, SendCodeResponse{
		Code:    http.StatusOK,
		Message: "验证码已发送",
	})
}

// VerifyCode 验证码登录/注册
// @Summary 验证码登录/注册
// @Description 通过手机号和验证码进行登录，若手机号未注册则自动注册
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyCodeRequest true "手机号和验证码"
// @Success 200 {object} VerifyCodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/verify-code [post]
func (h *Handler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证验证码
	codeStoreLock.RLock()
	entry, exists := codeStore[req.Phone]
	codeStoreLock.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "请先获取验证码",
		})
		return
	}

	if time.Now().After(entry.ExpiresAt) {
		// 清理过期验证码
		codeStoreLock.Lock()
		delete(codeStore, req.Phone)
		codeStoreLock.Unlock()

		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "验证码已过期，请重新获取",
		})
		return
	}

	if req.Code != entry.Code {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "验证码错误",
		})
		return
	}

	// 验证成功，清理验证码
	codeStoreLock.Lock()
	delete(codeStore, req.Phone)
	codeStoreLock.Unlock()

	ctx := context.Background()
	isNewUser := false

	// 查询用户是否存在
	u, err := h.userService.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		// 用户不存在，自动注册
		u, err = h.userService.CreateUser(ctx, &user.CreateUserRequest{
			Phone: req.Phone,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "创建用户失败: " + err.Error(),
			})
			return
		}
		isNewUser = true
	}

	// 检查用户状态
	if u.Status != "active" {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "账号已被禁用",
		})
		return
	}

	// 更新最后登录时间
	_ = h.userService.UpdateLastLogin(ctx, u.ID)

	// 生成 JWT token（只使用 username）
	token, err := h.jwtService.GenerateToken(u.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "生成Token失败: " + err.Error(),
		})
		return
	}

	message := "登录成功"
	if isNewUser {
		message = "注册成功"
	}

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "auth.verify_code"
		rc.Resource = "auth"
		rc.Username = u.Username
		rc.Detail = map[string]any{
			"phone_masked": crypto.MaskPhone(req.Phone),
			"is_new_user":  isNewUser,
		}
	}

	c.JSON(http.StatusOK, VerifyCodeResponse{
		Code:      http.StatusOK,
		Message:   message,
		Token:     token,
		Username:  u.Username,
		IsNewUser: isNewUser,
	})
}
