package client

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateOrUpdateProfile(ctx context.Context, profile *ClientProfile) error
	GetProfileByUserID(ctx context.Context, userID string) (*ClientProfile, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetProfileByUserID(ctx context.Context, userID string) (*ClientProfile, error) {
	var p ClientProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) CreateOrUpdateProfile(ctx context.Context, profile *ClientProfile) error {
	existing, err := r.GetProfileByUserID(ctx, profile.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		existing.CompanyName = profile.CompanyName
		existing.ContactName = profile.ContactName
		existing.Address = profile.Address
		existing.Industry = profile.Industry
		existing.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(existing).Error
	}
	return r.db.WithContext(ctx).Create(profile).Error
}

