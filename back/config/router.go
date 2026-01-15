package config

import (
	"back/internal/auth"
	"back/internal/bounty"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建 Gin 实例
	router := gin.Default()

	// 初始化JWT服务
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtIssuer := getEnv("JWT_ISSUER", "")
	jwtExpiryHours := getEnvInt("JWT_EXPIRY_HOURS", 0)
	jwtService := jwt.NewJWTService(jwtSecret, jwtIssuer, time.Duration(jwtExpiryHours)*time.Hour)

	// 初始化依赖
	authHandler := auth.NewHandler(jwtService)
	bountyRepo := bounty.NewRepository(DB)
	bountyService := bounty.NewService(bountyRepo)
	bountyHandler := bounty.NewHandler(bountyService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由 - 不需要JWT认证
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login) // 登录
		}

		// Bounty 路由 - 需要JWT认证
		bounties := v1.Group("/bounties")
		bounties.Use(middleware.JWTAuth(jwtService))
		{
			bounties.POST("", bountyHandler.CreateBounty)      // 创建赏金
			bounties.GET("", bountyHandler.ListBounties)       // 获取赏金列表
			bounties.GET("/:id", bountyHandler.GetBounty)      // 获取赏金详情
			bounties.PUT("/:id", bountyHandler.UpdateBounty)   // 更新赏金
			bounties.DELETE("/:id", bountyHandler.DeleteBounty) // 删除赏金
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return router
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt 获取整数类型的环境变量，如果不存在或解析失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
