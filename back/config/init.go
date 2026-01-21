package config

import (
	"fmt"
)

// Init 初始化配置
func Init() error {
	// 初始化数据库
	if err := InitDatabase(); err != nil {
		return err
	}

	// 初始化 Elasticsearch
	if err := InitElasticsearch(); err != nil {
		fmt.Printf("Warning: Elasticsearch init failed: %v\n", err)
		// ES 初始化失败不影响服务启动，只是搜索功能不可用
	}

	// 设置路由
	router := SetupRouter()

	// 启动服务器
	fmt.Println("Server is running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
