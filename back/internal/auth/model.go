package auth

import "encoding/json"

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// VerifyCodeRequest 验证码登录/注册请求
type VerifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// VerifyCodeResponse 验证码登录/注册响应
type VerifyCodeResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Token     string `json:"token,omitempty"`
	Username  string `json:"username,omitempty"`
	IsNewUser bool   `json:"isNewUser"` // 是否新用户（首次注册）
}

// InternalLoginRequest 内部系统登录请求
type InternalLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenResponse token 刷新响应
type RefreshTokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WechatLoginRequest 微信小程序登录请求
type WechatLoginRequest struct {
	Code     string      `json:"code"     binding:"required"`
	UserInfo interface{} `json:"userInfo"`
}

// wechatSessionResult 微信服务器 jscode2session 响应（内部使用）
type wechatSessionResult struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatLoginData 微信登录响应 data 字段
type WechatLoginData struct {
	Token        string          `json:"token"`
	OpenID       string          `json:"openId"`
	IsBound      bool            `json:"isBound"`
	SupplierInfo json.RawMessage `json:"supplierInfo"` // 内部系统 BC_Customer_GetByWeChat 返回的 data 数组，未绑定时为 []
}

// WechatLoginResponse 微信登录响应
type WechatLoginResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    WechatLoginData `json:"data"`
}
