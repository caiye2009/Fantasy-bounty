package config

import (
	"back/internal/audit"
	"back/internal/auth"
	"back/internal/proxy"
	"back/internal/supplier"
	"back/internal/user"
	"back/pkg/crypto"
	"back/pkg/internal_token"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// SetupRouter 设置路由，返回 router 和清理函数
func SetupRouter() (*gin.Engine, func()) {
	// 创建 Gin 实例
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestContextMiddleware()) // global request context

	// 初始化JWT服务
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET 环境变量未设置")
	}
	jwtIssuer := getEnv("JWT_ISSUER", "fantasy-bounty")
	jwtExpiryHours := getEnvInt("JWT_EXPIRY_HOURS", 24)
	jwtService := jwt.NewJWTService(jwtSecret, jwtIssuer, time.Duration(jwtExpiryHours)*time.Hour)

	// 初始化加密服务
	defaultKey := "1234567890123456" // 16字节示例，24或32字节也行
	cryptoKey := getEnv("CRYPTO_KEY", defaultKey)
	if cryptoKey == "" {
		log.Fatal("CRYPTO_KEY 环境变量未设置，必须是 16/24/32 字节的密钥")
	}
	hashPepper := getEnv("HASH_PEPPER", "")
	if hashPepper == "" {
		log.Fatal("HASH_PEPPER 环境变量未设置")
	}
	cryptoService, err := crypto.NewCrypto(cryptoKey, hashPepper)
	if err != nil {
		log.Fatal("初始化加密服务失败: ", err)
	}

	// 初始化审计服务
	auditRepo := audit.NewRepository(DB)
	auditService := audit.NewService(auditRepo)
	auditService.Start()

	// 初始化用户服务（供 auth 使用）
	userRepo := user.NewRepository(DB)
	userService := user.NewService(userRepo, cryptoService)

	// 初始化供应商服务
	supplierRepo := supplier.NewRepository(DB)
	supplierService := supplier.NewService(supplierRepo)

	// 初始化内部系统 token 管理器
	internalTokenLifetimeHours := getEnvInt("INTERNAL_TOKEN_LIFETIME_HOURS", 10)
	tokenManager := internal_token.NewManager(
		getEnv("INTERNAL_API_URL", ""),
		getEnv("INTERNAL_AUTH_PATH", "/auth/login"),
		getEnv("INTERNAL_USERNAME", ""),
		getEnv("INTERNAL_PASSWORD", ""),
		time.Duration(internalTokenLifetimeHours)*time.Hour,
	)
	tokenManager.Start()

	authHandler := auth.NewHandler(jwtService, userService, tokenManager.GetToken, getEnv("INTERNAL_API_URL", ""))
	supplierHandler := supplier.NewHandler(supplierService, userService)
	internalProxy := proxy.NewInternalProxy(tokenManager, getEnv("INTERNAL_API_URL", ""), jwtService)

	v1 := router.Group("/api/v1")

	// ========== PUBLIC：无需 JWT ==========
	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.Audit(auditService))
	{
		authGroup.POST("/wechat-login", authHandler.WechatLogin) // 微信登录，返回外部JWT
		authGroup.POST("/refresh", authHandler.RefreshToken)     // 刷新外部JWT
	}

	// ========== EXTERNAL JWT：前端携带外部JWT ==========
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(jwtService))
	protected.Use(middleware.Audit(auditService))
	{
		// 代理转发到内部系统（后端用内部token，前端无感知）
		proxyGroup := protected.Group("/proxy")
		{
			proxyGroup.POST("/bind-wechat", internalProxy.BindWeChatHandler())
			proxyGroup.POST("/get-by-wechat", internalProxy.GetByWeChatHandler())
			proxyGroup.POST("/inquiry-query", internalProxy.InquiryQueryHandler())
			proxyGroup.POST("/inquiry-detail", internalProxy.InquiryDetailHandler())
			proxyGroup.POST("/quote-delete", internalProxy.QuoteDeleteHandler())
			proxyGroup.POST("/quote-save", internalProxy.QuoteSaveHandler())
			proxyGroup.POST("/inquiry-quoted", internalProxy.InquiryBySupplierQuotedHandler())
		}

		// 供应商管理
		supplierGroup := protected.Group("/supplier")
		{
			supplierGroup.POST("/profile", supplierHandler.CreateOrUpdateProfile)     // 创建/更新供应商档案
			supplierGroup.GET("/profile", supplierHandler.GetProfile)                 // 获取供应商档案
			supplierGroup.GET("/full-info", supplierHandler.GetFullInfo)             // 获取供应商完整信息
			supplierGroup.POST("/capabilities", supplierHandler.UpdateCapabilities)   // 更新机器能力
		}
	}

	// ========== ADMIN：管理接口，不走外部JWT ==========
	adminGroup := v1.Group("")
	adminGroup.Use(middleware.Audit(auditService))
	{
		adminGroup.POST("/internal/login", authHandler.InternalLogin)          // 内部系统登录代理
		adminGroup.POST("/proxy/refresh-token", internalProxy.ForceRefreshTokenHandler()) // 强制刷新内部token
	}

	// 静态文件服务 - 营业执照图片
	router.Static("/uploads", "./uploads")

	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	cleanup := func() {
		auditService.Stop()
	}

	return router, cleanup
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
