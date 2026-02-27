package internal_token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager 内部系统 Token 管理器
// 负责从内部系统获取 JWT token，缓存并在过期前自动刷新
type Manager struct {
	mu        sync.RWMutex
	token     string
	expiresAt time.Time

	apiURL   string // 内部系统基础 URL
	authPath string // 内部系统登录路径
	username string // 内部系统用户名
	password string // 内部系统密码

	httpClient *http.Client
}

// refreshThreshold 距离过期小于此时间时触发刷新（提前30分钟）
const refreshThreshold = 30 * time.Minute

// NewManager 创建内部 Token 管理器
func NewManager(apiURL, authPath, username, password string) *Manager {
	return &Manager{
		apiURL:   apiURL,
		authPath: authPath,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start 启动时立即刷新一次 token，并开启后台每5分钟检查
func (m *Manager) Start() {
	log.Printf("[TOKEN-MGR] 启动初始化: 执行首次token刷新...")
	m.mu.Lock()
	if err := m.refreshWithRetry(); err != nil {
		log.Printf("[TOKEN-MGR] 启动初始刷新失败: %v", err)
	}
	m.mu.Unlock()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			m.mu.RLock()
			needRefresh := m.token == "" || time.Until(m.expiresAt) <= refreshThreshold
			remaining := time.Until(m.expiresAt).Round(time.Minute)
			m.mu.RUnlock()

			if needRefresh {
				m.mu.Lock()
				// double-check
				if m.token == "" || time.Until(m.expiresAt) <= refreshThreshold {
					log.Printf("[TOKEN-MGR] 后台检测: token剩余有效期 %v，触发刷新...", remaining)
					if err := m.refreshWithRetry(); err != nil {
						log.Printf("[TOKEN-MGR] 后台刷新失败: %v", err)
					}
				}
				m.mu.Unlock()
			}
		}
	}()
}

// ForceRefresh 手动强制刷新 token（不管当前是否即将过期）
func (m *Manager) ForceRefresh() error {
	log.Printf("[TOKEN-MGR] 手动强制刷新token")
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.refreshWithRetry()
}

// ExpiresAt 返回当前 token 的过期时间
func (m *Manager) ExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiresAt
}

// GetToken 获取内部系统 token，必要时自动刷新
func (m *Manager) GetToken() (string, error) {
	// 1. 读锁快速检查
	m.mu.RLock()
	token := m.token
	expiresAt := m.expiresAt
	m.mu.RUnlock()

	if token != "" && time.Until(expiresAt) > refreshThreshold {
		return token, nil
	}

	// 2. 需要刷新，升级为写锁
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check：其他 goroutine 可能已经刷新
	if m.token != "" && time.Until(m.expiresAt) > refreshThreshold {
		return m.token, nil
	}

	oldToken := m.token
	oldExpiresAt := m.expiresAt

	if err := m.refreshWithRetry(); err != nil {
		// 刷新失败，旧 token 还没真正过期则继续使用
		if oldToken != "" && time.Now().Before(oldExpiresAt) {
			log.Printf("[TOKEN-MGR] 刷新失败，继续使用旧token(剩余: %v)", time.Until(oldExpiresAt).Round(time.Minute))
			return oldToken, nil
		}
		return "", fmt.Errorf("获取内部系统 token 失败: %w", err)
	}

	return m.token, nil
}

// refreshWithRetry 带重试的 token 刷新，最多尝试3次
func (m *Manager) refreshWithRetry() error {
	const maxAttempts = 3
	var lastErr error
	for i := 1; i <= maxAttempts; i++ {
		log.Printf("[TOKEN-MGR] 刷新尝试 %d/%d", i, maxAttempts)
		if err := m.refreshToken(); err != nil {
			lastErr = err
			log.Printf("[TOKEN-MGR] 第%d次刷新失败: %v", i, lastErr)
			continue
		}
		return nil
	}
	return fmt.Errorf("token刷新失败(已尝试%d次): %w", maxAttempts, lastErr)
}

// refreshToken 调用内部系统登录接口获取新 token
func (m *Manager) refreshToken() error {
	loginURL := m.apiURL + m.authPath

	log.Printf("[TOKEN-MGR] 开始刷新内部token")
	log.Printf("[TOKEN-MGR]   登录URL: %s", loginURL)
	log.Printf("[TOKEN-MGR]   账号: %s", m.username)

	// 适配内部系统的请求格式: {"user":{"Account":"xxx","PassWord":"xxx"}}
	body, err := json.Marshal(map[string]interface{}{
		"user": map[string]string{
			"Account":  m.username,
			"PassWord": m.password,
		},
	})
	if err != nil {
		return fmt.Errorf("序列化登录请求失败: %w", err)
	}
	log.Printf("[TOKEN-MGR]   请求体: %s", string(body))

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Printf("[TOKEN-MGR] 请求内部登录接口失败: %v", err)
		return fmt.Errorf("调用内部登录接口失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取登录响应失败: %w", err)
	}

	log.Printf("[TOKEN-MGR]   登录响应 HTTP状态码: %d", resp.StatusCode)
	if len(respBody) > 300 {
		log.Printf("[TOKEN-MGR]   登录响应体(前300字符): %s...", string(respBody[:300]))
	} else {
		log.Printf("[TOKEN-MGR]   登录响应体: %s", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("内部登录接口返回非 200 状态码: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		IsSucceed  bool   `json:"isSucceed"`
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode"`
		Data       struct {
			Token string `json:"Token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析登录响应失败: %w, body: %s", err, string(respBody))
	}

	log.Printf("[TOKEN-MGR]   isSucceed: %v, message: %s", result.IsSucceed, result.Message)

	if !result.IsSucceed {
		return fmt.Errorf("登录失败: %s", result.Message)
	}

	if result.Data.Token == "" {
		return fmt.Errorf("登录响应中未包含 token")
	}

	// 解析 JWT 获取过期时间
	expiresAt, err := parseJWTExpiry(result.Data.Token)
	if err != nil {
		log.Printf("[TOKEN-MGR]   无法解析JWT过期时间: %v, 使用默认10h", err)
		expiresAt = time.Now().Add(10 * time.Hour)
	}

	m.token = result.Data.Token
	m.expiresAt = expiresAt

	log.Printf("[TOKEN-MGR] token刷新成功, 过期时间: %v (剩余: %v)",
		expiresAt.Format(time.RFC3339),
		time.Until(expiresAt).Round(time.Minute),
	)
	return nil
}

// parseJWTExpiry 从 JWT token 中解析过期时间（不验证签名）
// 兼容标准 exp 字段和内部系统自定义的 Expire 字段
func parseJWTExpiry(tokenString string) (time.Time, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	mapClaims := jwt.MapClaims{}

	_, _, err := parser.ParseUnverified(tokenString, mapClaims)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析 JWT 失败: %w", err)
	}

	// 优先尝试标准 exp 字段（Unix 时间戳）
	if exp, ok := mapClaims["exp"]; ok {
		if v, ok := exp.(float64); ok {
			return time.Unix(int64(v), 0), nil
		}
	}

	// 兼容内部系统自定义的 Expire 字段（RFC3339 字符串）
	if expireStr, ok := mapClaims["Expire"].(string); ok && expireStr != "" {
		if t, err := time.Parse(time.RFC3339Nano, expireStr); err == nil {
			return t, nil
		}
		if t, err := time.Parse(time.RFC3339, expireStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("JWT 中未包含有效的过期时间字段(exp/Expire)")
}
