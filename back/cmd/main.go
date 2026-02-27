// @title           Fantasy Bounty API
// @version         1.0
// @description     极绎贸易招投标系统对外接口
// @basePath        /api/v1

package main

import (
	"back/config"
	_ "back/docs"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	if err := config.Init(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
