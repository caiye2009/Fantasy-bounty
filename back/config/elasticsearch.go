package config

import (
	"back/pkg/elasticsearch"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// InitElasticsearch 初始化 Elasticsearch 连接
func InitElasticsearch() error {
	// 从环境变量读取配置
	addresses := os.Getenv("ES_ADDRESSES")
	if addresses == "" {
		addresses = "http://localhost:9200"
	}

	username := os.Getenv("ES_USERNAME")
	password := os.Getenv("ES_PASSWORD")

	cfg := elasticsearch.Config{
		Addresses: strings.Split(addresses, ","),
		Username:  username,
		Password:  password,
	}

	// 初始化客户端
	if err := elasticsearch.InitClient(cfg); err != nil {
		return fmt.Errorf("failed to init ES client: %w", err)
	}

	// 确保 bounty 索引存在
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := elasticsearch.EnsureIndex(ctx, "bounty", elasticsearch.BountyMapping); err != nil {
		return fmt.Errorf("failed to ensure bounty index: %w", err)
	}

	fmt.Println("Elasticsearch connected and bounty index ready")
	return nil
}
