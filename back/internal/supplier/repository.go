package supplier

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// SupplierProfile
	CreateOrUpdateProfile(ctx context.Context, profile *SupplierProfile) error
	GetProfileByUserID(ctx context.Context, userID string) (*SupplierProfile, error)
	UpdateCapabilities(ctx context.Context, userID string, capabilities map[string]int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateOrUpdateProfile(ctx context.Context, profile *SupplierProfile) error {
	// 检查是否已存在
	existing, err := r.GetProfileByUserID(ctx, profile.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil {
		// 更新现有记录
		existing.CompanyType = profile.CompanyType
		existing.CompanyName = profile.CompanyName
		existing.Capabilities = profile.Capabilities
		existing.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(existing).Error
	}

	// 创建新记录
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *repository) GetProfileByUserID(ctx context.Context, userID string) (*SupplierProfile, error) {
	var profile SupplierProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *repository) UpdateCapabilities(ctx context.Context, userID string, capabilities map[string]int) error {
	// 将map转换为JSON
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return err
	}

	// 先尝试获取现有记录
	existing, err := r.GetProfileByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil {
		// 更新现有记录
		existing.Capabilities = capabilitiesJSON
		existing.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(existing).Error
	}

	// 创建新记录（只有capabilities，没有其他基本信息）
	profile := &SupplierProfile{
		UserID:      userID,
		Capabilities: capabilitiesJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return r.db.WithContext(ctx).Create(profile).Error
}