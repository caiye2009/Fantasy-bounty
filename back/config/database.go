package config

import (
	"back/internal/account"
	"back/internal/audit"
	"back/internal/bid"
	"back/internal/company"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	var err error

	// 从环境变量读取数据库配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "fantasy_bounty")

	if dbPassword == "" {
		return fmt.Errorf("DB_PASSWORD 环境变量未设置")
	}

	fmt.Printf("Connecting to database: host=%s port=%s user=%s dbname=%s\n", dbHost, dbPort, dbUser, dbName)

	// 构建 PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接 PostgreSQL 数据库
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	fmt.Println("Database connection established, running AutoMigrate...")

	// 自动迁移数据库表
	if err := DB.AutoMigrate(
		&bid.Bid{},
		&bid.BidWovenSpec{},
		&bid.BidKnittedSpec{},
		&account.Account{},
		&company.Company{},
		&company.AccountCompany{},
		&company.CompanyApplication{},
		&audit.AuditLog{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Println("AutoMigrate completed successfully - tables created/updated")
	fmt.Println("Database connected successfully")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
