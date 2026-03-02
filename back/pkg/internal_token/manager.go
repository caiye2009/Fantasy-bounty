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
)

// Manager 内部系统 Token 管理器
// 负责从内部系统获取 token，并按固定周期（tokenLifetime - 30min）自动刷新
type Manager struct {
	mu           sync.RWMutex
	token        string
	expiresAt    time.Time
	refreshTimer *time.Timer

	apiURL        string        // 内部系统基础 URL
	authPath      string        // 内部系统登录路径
	username      string        // 内部系统用户名
	password      string        // 内部系统密码
	tokenLifetime time.Duration // 内部 token 有效期（用于计算刷新周期）

	httpClient *http.Client
}

// NewManager 创建内部 Token 管理器
// tokenLifetime: 内部系统 token 有效期，刷新将在到期前 30 分钟触发
func NewManager(apiURL, authPath, username, password string, tokenLifetime time.Duration) *Manager {
	return &Manager{
		apiURL:        apiURL,
		authPath:      authPath,
		username:      username,
		password:      password,
		tokenLifetime: tokenLifetime,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// refreshInterval 刷新周期 = tokenLifetime - 30min
func (m *Manager) refreshInterval() time.Duration {
	d := m.tokenLifetime - 30*time.Minute
	if d <= 0 {
		d = m.tokenLifetime
	}
	return d
}

// Start 启动时立即刷新一次，之后按固定周期自动刷新
func (m *Manager) Start() {
	log.Printf("[TOKEN-MGR] 启动初始化: token有效期=%v, 刷新周期=%v",
		m.tokenLifetime, m.refreshInterval())

	m.mu.Lock()
	if err := m.refreshWithRetry(); err != nil {
		log.Printf("[TOKEN-MGR] 启动初始刷新失败: %v", err)
	}
	m.scheduleNextRefresh()
	m.mu.Unlock()
}

// scheduleNextRefresh 按固定刷新周期调度下次刷新（调用时须持有写锁）
func (m *Manager) scheduleNextRefresh() {
	if m.refreshTimer != nil {
		m.refreshTimer.Stop()
	}
	interval := m.refreshInterval()
	log.Printf("[TOKEN-MGR] 下次刷新: %v 后", interval.Round(time.Minute))
	m.refreshTimer = time.AfterFunc(interval, m.timedRefresh)
}

// timedRefresh 定时触发的刷新
func (m *Manager) timedRefresh() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[TOKEN-MGR] 定时触发: 执行内部token刷新...")
	if err := m.refreshWithRetry(); err != nil {
		log.Printf("[TOKEN-MGR] 定时刷新失败: %v，5分钟后重试", err)
		if m.refreshTimer != nil {
			m.refreshTimer.Stop()
		}
		m.refreshTimer = time.AfterFunc(5*time.Minute, m.timedRefresh)
		return
	}
	m.scheduleNextRefresh()
}

// ForceRefresh 手动强制刷新 token，并重置定时周期
func (m *Manager) ForceRefresh() error {
	log.Printf("[TOKEN-MGR] 手动强制刷新token")
	m.mu.Lock()
	defer m.mu.Unlock()
	err := m.refreshWithRetry()
	if err == nil {
		m.scheduleNextRefresh()
	}
	return err
}

// ExpiresAt 返回当前 token 的过期时间
func (m *Manager) ExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiresAt
}

// GetToken 返回当前内部 token；仅当 token 已过期时才做兜底刷新
func (m *Manager) GetToken() (string, error) {
	m.mu.RLock()
	token := m.token
	expiresAt := m.expiresAt
	m.mu.RUnlock()

	if token != "" && time.Now().Before(expiresAt) {
		return token, nil
	}

	// 兜底：定时器未能及时触发，立即刷新
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token != "" && time.Now().Before(m.expiresAt) {
		return m.token, nil
	}

	log.Printf("[TOKEN-MGR] GetToken: token已过期，立即刷新...")
	if err := m.refreshWithRetry(); err != nil {
		return "", fmt.Errorf("获取内部系统 token 失败: %w", err)
	}
	m.scheduleNextRefresh()
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

	now := time.Now()
	m.token = result.Data.Token
	m.expiresAt = now.Add(m.tokenLifetime)

	log.Printf("[TOKEN-MGR] token刷新成功, 下次刷新: %v 后 (预计到期: %s)",
		m.refreshInterval().Round(time.Minute),
		m.expiresAt.Format("2006-01-02 15:04:05"),
	)
	return nil
}
