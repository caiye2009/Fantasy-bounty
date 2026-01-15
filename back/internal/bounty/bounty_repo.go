package bounty

import (
	"context"
	"gorm.io/gorm"
)

// Repository 赏金数据访问层接口
type Repository interface {
	Create(ctx context.Context, bounty *Bounty) error
	GetByID(ctx context.Context, id uint) (*Bounty, error)
	Update(ctx context.Context, bounty *Bounty) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]Bounty, int64, error)
}

// repository 赏金数据访问层实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建新的 repository 实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建新赏金
func (r *repository) Create(ctx context.Context, bounty *Bounty) error {
	return r.db.WithContext(ctx).Create(bounty).Error
}

// GetByID 根据 ID 获取赏金
func (r *repository) GetByID(ctx context.Context, id uint) (*Bounty, error) {
	var bounty Bounty
	err := r.db.WithContext(ctx).First(&bounty, id).Error
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

// Update 更新赏金
func (r *repository) Update(ctx context.Context, bounty *Bounty) error {
	return r.db.WithContext(ctx).Save(bounty).Error
}

// Delete 删除赏金
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Bounty{}, id).Error
}

// List 获取赏金列表
func (r *repository) List(ctx context.Context, offset, limit int) ([]Bounty, int64, error) {
	var bounties []Bounty
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&Bounty{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&bounties).Error

	if err != nil {
		return nil, 0, err
	}

	return bounties, total, nil
}
