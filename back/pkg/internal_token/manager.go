package internal_token

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// refreshThreshold 距离过期小于此时间时触发刷新
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

	// 3. 调用内部系统登录接口
	oldToken := m.token
	oldExpiresAt := m.expiresAt

	err := m.refreshToken()
	if err != nil {
		// 刷新失败，旧 token 还没过期则继续使用
		if oldToken != "" && time.Now().Before(oldExpiresAt) {
			return oldToken, nil
		}
		return "", fmt.Errorf("获取内部系统 token 失败: %w", err)
	}

	return m.token, nil
}

// refreshToken 调用内部系统登录接口获取新 token
func (m *Manager) refreshToken() error {
	loginURL := m.apiURL + m.authPath

	body, err := json.Marshal(map[string]string{
		"username": m.username,
		"password": m.password,
	})
	if err != nil {
		return fmt.Errorf("序列化登录请求失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("调用内部登录接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("内部登录接口返回非 200 状态码: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}

	if result.Token == "" {
		return fmt.Errorf("登录响应中未包含 token")
	}

	// 解析 JWT 获取过期时间
	expiresAt, err := parseJWTExpiry(result.Token)
	if err != nil {
		// 无法解析过期时间，使用默认 10h
		expiresAt = time.Now().Add(10 * time.Hour)
	}

	m.token = result.Token
	m.expiresAt = expiresAt

	return nil
}

// parseJWTExpiry 从 JWT token 中解析过期时间（不验证签名）
func parseJWTExpiry(tokenString string) (time.Time, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claims := &jwt.RegisteredClaims{}

	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析 JWT 失败: %w", err)
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("JWT 中未包含过期时间")
	}

	return claims.ExpiresAt.Time, nil
}
