package config

import (
	"back/internal/audit"
	"back/internal/auth"
	"back/internal/bid"
	"back/internal/supplier"
	"back/internal/user"
	"back/pkg/crypto"
	"back/pkg/internal_token"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"back/pkg/proxy"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

	// 初始化用户服务（需要先初始化，供 auth 使用）
	userRepo := user.NewRepository(DB)
	userService := user.NewService(userRepo, cryptoService)

	// 初始化依赖
	authHandler := auth.NewHandler(jwtService, userService)

	// 初始化供应商服务（需要在 bid 之前初始化）
	supplierRepo := supplier.NewRepository(DB)
	supplierService := supplier.NewService(supplierRepo)
	supplierHandler := supplier.NewHandler(supplierService)

	// 初始化竞标服务（需要供应商服务来校验认证状态）
	bidRepo := bid.NewRepository(DB)
	bidService := bid.NewService(bidRepo)
	bidHandler := bid.NewHandler(bidService, supplierService)

	// 初始化用户 handler（userService 已在上面初始化）
	userHandler := user.NewHandler(userService)

	// 初始化内部系统 token 管理器
	tokenManager := internal_token.NewManager(
		getEnv("INTERNAL_API_URL", ""),
		getEnv("INTERNAL_AUTH_PATH", "/auth/login"),
		getEnv("INTERNAL_USERNAME", ""),
		getEnv("INTERNAL_PASSWORD", ""),
	)

	// 初始化内部系统反向代理
	internalProxy := proxy.NewInternalProxy(tokenManager, getEnv("INTERNAL_API_URL", ""), jwtService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// ========== 供应商认证路由 ==========
		// 不需要JWT认证，但有审计
		authGroup := v1.Group("/auth")
		authGroup.Use(middleware.Audit(auditService))
		{
			authGroup.POST("/send-code", authHandler.SendCode)     // 发送验证码
			authGroup.POST("/verify-code", authHandler.VerifyCode) // 验证码登录/注册
			authGroup.POST("/refresh", authHandler.RefreshToken)   // 刷新 token
		}

		// ========== 供应商业务路由 ==========
		// 需要外部JWT认证 + 审计
		supplierGroup := v1.Group("/supplier")
		supplierGroup.Use(middleware.JWTAuth(jwtService))
		supplierGroup.Use(middleware.Audit(auditService))
		{
			// Bid 路由
			bids := supplierGroup.Group("/bids")
			{
				bids.POST("", bidHandler.CreateBid)       // 创建竞标（需要供应商认证）
				bids.GET("", bidHandler.ListBids)         // 获取竞标列表
				bids.GET("/my", bidHandler.ListMyBids)    // 获取我的竞标列表
				bids.DELETE("/:id", bidHandler.DeleteBid) // 删除竞标
			}

			// Supplier 路由
			suppliers := supplierGroup.Group("/suppliers")
			{
				suppliers.GET("", supplierHandler.ListSuppliers)
				suppliers.GET("/:id", supplierHandler.GetSupplier)
				suppliers.POST("/recognize", supplierHandler.RecognizeLicense) // 上传营业执照OCR识别
				suppliers.POST("/apply", supplierHandler.ApplySupplier)        // 提交供应商认证申请
				suppliers.GET("/my", supplierHandler.GetMySupplierStatus)      // 获取我的供应商认证状态
			}

			// User 路由
			users := supplierGroup.Group("/users")
			{
				users.POST("", userHandler.CreateUser)
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}
		}

		// ========== 内部系统路由 ==========
		internalGroup := v1.Group("/internal")
		internalGroup.Use(middleware.Audit(auditService))
		{
			// 内部用户登录（不需要JWT）
			internalGroup.POST("/login", authHandler.InternalLogin)

			// ========== 外部用户访问的路径（供应商） ==========
			// 需要验证外部JWT，通过后转换成内部token访问老系统
			// 使用专门的 BountiesHandler（通用 Handler 依赖 *path 参数，这里不适用）
			internalGroup.GET("/bounties", middleware.JWTAuth(jwtService), internalProxy.BountiesHandler())
			internalGroup.GET("/bounties/:id", middleware.JWTAuth(jwtService), internalProxy.BountiesHandler())

			// TODO: 在这里添加其他外部用户可以访问的路径
			// internalGroup.GET("/search", middleware.JWTAuth(jwtService), internalProxy.Handler())

			// ========== 内部用户访问的路径（员工） ==========
			// 直接透传Authorization header到老系统，不做JWT验证
			// 注意：由于 Gin 不允许通配符和具体路径混用，
			// 需要为内部用户明确添加需要的路由，或者在这里使用 NoRoute
			// 暂时注释掉通配符路由，避免冲突
			// internalGroup.Any("/*path", internalProxy.Handler())
		}
	}

	// 静态文件服务 - 营业执照图片
	router.Static("/uploads", "./uploads")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
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
