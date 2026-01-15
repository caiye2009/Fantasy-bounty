package main

import (
	"back/config"
	"log"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
