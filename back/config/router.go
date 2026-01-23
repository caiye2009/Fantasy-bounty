package config

import (
	"back/internal/account"
	"back/internal/auth"
	"back/internal/bid"
	"back/internal/bounty"
	"back/internal/company"
	"back/internal/search"
	"back/pkg/crypto"
	"back/pkg/jwt"
	"back/pkg/middleware"
	"log"
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

	// 初始化加密服务
	defaultKey := "1234567890123456" // 16字节示例，24或32字节也行
	cryptoKey := getEnv("CRYPTO_KEY", defaultKey)
	if cryptoKey == "" {
		log.Fatal("CRYPTO_KEY 环境变量未设置，必须是 16/24/32 字节的密钥")
	}
	cryptoService, err := crypto.NewCrypto(cryptoKey)
	if err != nil {
		log.Fatal("初始化加密服务失败: ", err)
	}

	// 初始化账号服务（需要先初始化，供 auth 使用）
	accountRepo := account.NewRepository(DB)
	accountService := account.NewService(accountRepo, cryptoService)

	// 初始化依赖
	authHandler := auth.NewHandler(jwtService, accountService)
	bountyRepo := bounty.NewRepository(DB)
	bountyService := bounty.NewService(bountyRepo)
	bountyHandler := bounty.NewHandler(bountyService)

	// 初始化企业服务（需要在 bid 之前初始化）
	companyRepo := company.NewRepository(DB)
	companyService := company.NewService(companyRepo)
	companyHandler := company.NewHandler(companyService)

	// 初始化竞标服务（需要企业服务来校验认证状态）
	bidRepo := bid.NewRepository(DB)
	bidService := bid.NewService(bidRepo)
	bidHandler := bid.NewHandler(bidService, companyService)

	// 初始化搜索服务
	searchService := search.NewService()
	searchHandler := search.NewHandler(searchService)

	// 初始化账号 handler（accountService 已在上面初始化）
	accountHandler := account.NewHandler(accountService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由 - 不需要JWT认证
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/send-code", authHandler.SendCode)     // 发送验证码
			authGroup.POST("/verify-code", authHandler.VerifyCode) // 验证码登录/注册
		}

		// 搜索路由 - 需要JWT认证
		searchGroup := v1.Group("/search")
		searchGroup.Use(middleware.JWTAuth(jwtService))
		{
			searchGroup.POST("", searchHandler.Search) // 统一搜索接口
		}

		// Bounty 公开路由
		v1.GET("/bounties/peek", bountyHandler.PeekBounties)
		v1.POST("/bounties", bountyHandler.CreateBounty) // 临时：不需要token

		// Bounty 保护路由 - 需要JWT认证
		bounties := v1.Group("/bounties")
		bounties.Use(middleware.JWTAuth(jwtService))
		{
			bounties.GET("", bountyHandler.ListBounties)
			bounties.GET("/:id", bountyHandler.GetBounty)
			bounties.PUT("/:id", bountyHandler.UpdateBounty)
			bounties.DELETE("/:id", bountyHandler.DeleteBounty)
		}

		// Bid 路由 - 需要JWT认证
		bids := v1.Group("/bids")
		bids.Use(middleware.JWTAuth(jwtService))
		{
			bids.POST("", bidHandler.CreateBid)       // 创建竞标（需要企业认证）
			bids.GET("", bidHandler.ListBids)         // 获取竞标列表
			bids.GET("/my", bidHandler.ListMyBids)    // 获取我的竞标列表
			bids.DELETE("/:id", bidHandler.DeleteBid) // 删除竞标
		}

		// Account 路由 - 需要JWT认证
		accounts := v1.Group("/accounts")
		accounts.Use(middleware.JWTAuth(jwtService))
		{
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("", accountHandler.ListAccounts)
			accounts.GET("/:id", accountHandler.GetAccount)
			accounts.PUT("/:id", accountHandler.UpdateAccount)
			accounts.DELETE("/:id", accountHandler.DeleteAccount)
		}

		// Company 路由 - 需要JWT认证
		companies := v1.Group("/companies")
		companies.Use(middleware.JWTAuth(jwtService))
		{
			companies.GET("", companyHandler.ListCompanies)
			companies.GET("/:id", companyHandler.GetCompany)
			// 用户操作
			companies.POST("/recognize", companyHandler.RecognizeLicense) // 上传营业执照OCR识别
			companies.POST("/apply", companyHandler.ApplyCompany)         // 提交企业认证申请
			companies.GET("/my", companyHandler.GetMyCompanyStatus)       // 获取我的企业认证状态
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
