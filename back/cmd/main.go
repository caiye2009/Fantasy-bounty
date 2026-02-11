package main

import (
	"back/config"
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
