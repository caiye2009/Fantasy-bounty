package auth

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
	AccountID string `json:"accountId,omitempty"`
	IsNewUser bool   `json:"isNewUser"` // 是否新用户（首次注册）
	Username  string `json:"username,omitempty"` // 用户名
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
