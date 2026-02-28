package auth

import (
	"back/internal/user"
	"back/pkg/crypto"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
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
	jwtService       *jwt.JWTService
	userService      user.Service
	getInternalToken func() (string, error) // 获取内部系统 token
	internalAPIURL   string                 // 内部系统基础 URL
}

// NewHandler 创建新的 handler 实例
func NewHandler(
	jwtService *jwt.JWTService,
	userService user.Service,
	getInternalToken func() (string, error),
	internalAPIURL string,
) *Handler {
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
		jwtService:       jwtService,
		userService:      userService,
		getInternalToken: getInternalToken,
		internalAPIURL:   internalAPIURL,
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

// RefreshToken 刷新用户 JWT
// 接受过期不超过 7 天的旧 token，签发新 token
func (h *Handler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "missing authorization header",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "invalid authorization header format",
		})
		return
	}

	tokenString := parts[1]

	// 先尝试正常验证（token 还没过期，直接返回）
	if claims, err := h.jwtService.ValidateToken(tokenString); err == nil {
		c.JSON(http.StatusOK, RefreshTokenResponse{
			Code:    http.StatusOK,
			Message: "token 仍然有效",
			Token:   tokenString,
		})
		_ = claims
		return
	}

	// token 过期了，解析忽略过期检查
	claims, err := h.jwtService.ParseTokenIgnoreExpiry(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "token 无效，请重新登录",
		})
		return
	}

	// 检查过期时间，超过 7 天不允许刷新
	if claims.ExpiresAt != nil {
		expiredDuration := time.Since(claims.ExpiresAt.Time)
		if expiredDuration > 7*24*time.Hour {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "token 已过期超过7天，请重新登录",
			})
			return
		}
	}

	// 签发新 token
	newToken, err := h.jwtService.GenerateToken(claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "生成新Token失败",
		})
		return
	}

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "auth.refresh_token"
		rc.Resource = "auth"
		rc.Username = claims.Username
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		Code:    http.StatusOK,
		Message: "token 刷新成功",
		Token:   newToken,
	})
}

// InternalLogin 内部系统登录代理
// @Summary 内部系统登录
// @Description 将登录请求转发到内部系统，返回内部系统的token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body InternalLoginRequest true "工号和密码"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
// @Router /api/v1/internal/login [post]
func (h *Handler) InternalLogin(c *gin.Context) {
	var req InternalLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取内部系统配置
	internalAPIURL := os.Getenv("INTERNAL_API_URL")
	internalAuthPath := os.Getenv("INTERNAL_AUTH_PATH")
	if internalAPIURL == "" || internalAuthPath == "" {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "内部系统配置错误",
		})
		return
	}

	// 构造转发请求
	loginURL := internalAPIURL + internalAuthPath
	body, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "请求序列化失败",
		})
		return
	}

	// 发送请求到内部系统
	httpReq, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建请求失败",
		})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Code:    http.StatusBadGateway,
			Message: "内部系统连接失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "读取响应失败",
		})
		return
	}

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "auth.internal_login"
		rc.Resource = "auth"
		rc.Detail = map[string]any{
			"username":    req.Username,
			"status_code": resp.StatusCode,
		}
	}

	// 透传内部系统的响应（包括状态码和body）
	c.Data(resp.StatusCode, "application/json", respBody)
}

// WechatLogin 微信小程序登录
// @Summary 微信小程序登录
// @Description 用前端临时 code 换取 openid，再查询内部系统中该 openid 绑定的供应商信息
// @Tags auth
// @Accept json
// @Produce json
// @Param request body WechatLoginRequest true "微信登录参数"
// @Success 200 {object} WechatLoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/wechat-login [post]
func (h *Handler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// Step 1: 用 code 向微信服务器换取 openid
	appid := os.Getenv("WECHAT_APPID")
	secret := os.Getenv("WECHAT_SECRET")
	if appid == "" || secret == "" {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "微信配置未设置(WECHAT_APPID/WECHAT_SECRET)",
		})
		return
	}

	wxURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appid, secret, req.Code,
	)

	wxResp, err := http.Get(wxURL) //nolint:gosec
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Code:    http.StatusBadGateway,
			Message: "微信服务器请求失败",
		})
		return
	}
	defer wxResp.Body.Close()

	var session wechatSessionResult
	if err := json.NewDecoder(wxResp.Body).Decode(&session); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "解析微信响应失败",
		})
		return
	}

	if session.ErrCode != 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: fmt.Sprintf("微信授权失败: %s", session.ErrMsg),
		})
		return
	}
	fmt.Printf("[WECHAT-LOGIN] 获取 openid 成功: %s\n", session.OpenID)

	// Step 2: 用 openid 查询内部系统绑定的供应商信息
	supplierRaw, err := h.querySupplierByOpenID(session.OpenID)
	if err != nil {
		fmt.Printf("[WECHAT-LOGIN] 查询供应商失败: %v\n", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Code:    http.StatusBadGateway,
			Message: "查询供应商信息失败: " + err.Error(),
		})
		return
	}
	fmt.Printf("[WECHAT-LOGIN] 内部系统原始响应: %s\n", string(supplierRaw))

	// 解析内部系统响应，只取 data 字段
	var internalResp struct {
		IsSucceed bool              `json:"isSucceed"`
		Data      []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(supplierRaw, &internalResp); err != nil {
		fmt.Printf("[WECHAT-LOGIN] 解析内部响应失败: %v\n", err)
	}
	isBound := internalResp.IsSucceed && len(internalResp.Data) > 0
	fmt.Printf("[WECHAT-LOGIN] isBound=%v isSucceed=%v dataLen=%d\n", isBound, internalResp.IsSucceed, len(internalResp.Data))

	// 将 data 序列化为 JSON（未绑定时为空数组 []）
	supplierData, _ := json.Marshal(internalResp.Data)

	// Step 3: 生成 JWT（以 openid 作为标识）
	token, err := h.jwtService.GenerateToken(session.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "生成 token 失败",
		})
		return
	}

	message := "登录成功"
	if !isBound {
		message = "登录成功，该微信暂未绑定供应商"
	}

	// 设置审计信息
	if rc := middleware.GetRequestContext(c); rc != nil {
		rc.Action = "auth.wechat_login"
		rc.Resource = "auth"
		rc.Username = session.OpenID
		rc.Detail = map[string]any{
			"openid":   session.OpenID,
			"is_bound": isBound,
		}
	}

	c.JSON(http.StatusOK, WechatLoginResponse{
		Code:    http.StatusOK,
		Message: message,
		Data: WechatLoginData{
			Token:        token,
			OpenID:       session.OpenID,
			IsBound:      isBound,
			SupplierInfo: json.RawMessage(supplierData),
		},
	})
}

// querySupplierByOpenID 调用内部系统 BC_Customer_GetByWeChat，返回原始 JSON
func (h *Handler) querySupplierByOpenID(openid string) ([]byte, error) {
	internalToken, err := h.getInternalToken()
	if err != nil {
		return nil, fmt.Errorf("获取内部 token 失败: %w", err)
	}

	requestBody := map[string]interface{}{
		"code": "BC_Customer_GetByWeChat",
		"pars": map[string]interface{}{
			"Openid": openid,
			"Type":   "1", // 1=小程序
		},
		"outPars": map[string]interface{}{},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	apiURL := strings.TrimRight(h.internalAPIURL, "/") + "/api/Public/GetProcedureDataSet"
	fmt.Printf("[WECHAT-LOGIN] 查询供应商: URL=%s body=%s\n", apiURL, string(bodyBytes))

	httpReq, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", internalToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求内部系统失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Printf("[WECHAT-LOGIN] 内部系统 HTTP %d, 响应: %s\n", resp.StatusCode, string(body))
	return body, err
}
