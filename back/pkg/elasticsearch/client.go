package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var client *elasticsearch.Client

// InitClient 初始化 ES 客户端
func InitClient(cfg Config) error {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	var err error
	client, err = elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("failed to create ES client: %w", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect ES: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES connection error: %s", res.String())
	}

	return nil
}

// GetClient 获取 ES 客户端
func GetClient() *elasticsearch.Client {
	return client
}

// Ping 检查 ES 连接
func Ping(ctx context.Context) error {
	res, err := client.Ping(
		client.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("ES ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES ping error: %s", res.String())
	}

	return nil
}
