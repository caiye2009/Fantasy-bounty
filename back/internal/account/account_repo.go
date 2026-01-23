package account

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByPhoneHash(ctx context.Context, phoneHash string) (*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]Account, int64, error)
	UpdateLastLogin(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, account *Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*Account, error) {
	var account Account
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *repository) GetByPhoneHash(ctx context.Context, phoneHash string) (*Account, error) {
	var account Account
	err := r.db.WithContext(ctx).First(&account, "phone_hash = ?", phoneHash).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *repository) Update(ctx context.Context, account *Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Account{}, "id = ?", id).Error
}

func (r *repository) List(ctx context.Context, offset, limit int) ([]Account, int64, error) {
	var accounts []Account
	var total int64

	if err := r.db.WithContext(ctx).Model(&Account{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&accounts).Error

	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", id).Update("last_login_at", now).Error
}
