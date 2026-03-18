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

type Manager struct {
	mu           sync.RWMutex
	token        string
	expiresAt    time.Time
	refreshTimer *time.Timer

	apiURL        string
	authPath      string
	username      string
	password      string
	tokenLifetime time.Duration

	httpClient *http.Client
}

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

func (m *Manager) refreshInterval() time.Duration {
	d := m.tokenLifetime - 30*time.Minute
	if d <= 0 {
		d = m.tokenLifetime
	}
	return d
}

// Start 启动 token 刷新定时器
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 启动时尝试刷新，但不重复打印失败日志
	_ = m.refreshWithRetry()

	m.scheduleNextRefresh()
}

func (m *Manager) scheduleNextRefresh() {
	if m.refreshTimer != nil {
		m.refreshTimer.Stop()
	}

	interval := m.refreshInterval()
	log.Printf("[TOKEN-MGR] 下次刷新: %v 后", interval.Round(time.Minute))

	m.refreshTimer = time.AfterFunc(interval, m.timedRefresh)
}

func (m *Manager) timedRefresh() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.refreshWithRetry(); err != nil {
		log.Printf("[TOKEN-MGR] token刷新失败: %v，5分钟后重试", err)

		if m.refreshTimer != nil {
			m.refreshTimer.Stop()
		}

		m.refreshTimer = time.AfterFunc(5*time.Minute, m.timedRefresh)
		return
	}

	m.scheduleNextRefresh()
}

func (m *Manager) ForceRefresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.refreshWithRetry()
	if err == nil {
		m.scheduleNextRefresh()
	}

	return err
}

func (m *Manager) ExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiresAt
}

func (m *Manager) GetToken() (string, error) {
	m.mu.RLock()
	token := m.token
	expiresAt := m.expiresAt
	m.mu.RUnlock()

	if token != "" && time.Now().Before(expiresAt) {
		return token, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token != "" && time.Now().Before(m.expiresAt) {
		return m.token, nil
	}

	if err := m.refreshWithRetry(); err != nil {
		return "", fmt.Errorf("获取内部系统 token 失败: %w", err)
	}

	m.scheduleNextRefresh()
	return m.token, nil
}

func (m *Manager) refreshWithRetry() error {
	const maxAttempts = 3
	var lastErr error

	for i := 1; i <= maxAttempts; i++ {
		if err := m.refreshToken(); err != nil {
			lastErr = err
			log.Printf("[TOKEN-MGR] 第%d次刷新失败: %v", i, lastErr)
			continue
		}
		return nil
	}

	return fmt.Errorf("token刷新失败(已尝试%d次): %w", maxAttempts, lastErr)
}

func (m *Manager) refreshToken() error {
	loginURL := m.apiURL + m.authPath

	body, _ := json.Marshal(map[string]interface{}{
		"user": map[string]string{
			"Account":  m.username,
			"PassWord": m.password,
		},
	})

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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取登录响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("内部登录接口返回非200状态码: %d", resp.StatusCode)
	}

	var result struct {
		IsSucceed bool   `json:"isSucceed"`
		Message   string `json:"message"`
		Data      struct {
			Token string `json:"Token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}

	if !result.IsSucceed {
		return fmt.Errorf("登录失败: %s", result.Message)
	}

	if result.Data.Token == "" {
		return fmt.Errorf("登录响应未返回token")
	}

	now := time.Now()

	m.token = result.Data.Token
	m.expiresAt = now.Add(m.tokenLifetime)

	log.Printf("[TOKEN-MGR] token刷新成功")

	return nil
}