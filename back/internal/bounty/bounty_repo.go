package bounty

import (
	"context"
	"gorm.io/gorm"
)

// Repository 悬赏数据访问层接口
type Repository interface {
	Create(ctx context.Context, bounty *Bounty) error
	CreateWovenSpec(ctx context.Context, spec *BountyWovenSpec) error
	CreateKnittedSpec(ctx context.Context, spec *BountyKnittedSpec) error
	GetByID(ctx context.Context, id uint) (*Bounty, error)
	Update(ctx context.Context, bounty *Bounty) error
	UpdateWovenSpec(ctx context.Context, spec *BountyWovenSpec) error
	UpdateKnittedSpec(ctx context.Context, spec *BountyKnittedSpec) error
	Delete(ctx context.Context, id uint) error
	DeleteWovenSpec(ctx context.Context, bountyID uint) error
	DeleteKnittedSpec(ctx context.Context, bountyID uint) error
	List(ctx context.Context, offset, limit int) ([]Bounty, int64, error)
}

// repository 悬赏数据访问层实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建新的 repository 实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建新悬赏
func (r *repository) Create(ctx context.Context, bounty *Bounty) error {
	return r.db.WithContext(ctx).Create(bounty).Error
}

// CreateWovenSpec 创建梭织规格
func (r *repository) CreateWovenSpec(ctx context.Context, spec *BountyWovenSpec) error {
	return r.db.WithContext(ctx).Create(spec).Error
}

// CreateKnittedSpec 创建针织规格
func (r *repository) CreateKnittedSpec(ctx context.Context, spec *BountyKnittedSpec) error {
	return r.db.WithContext(ctx).Create(spec).Error
}

// GetByID 根据 ID 获取悬赏（包含规格）
func (r *repository) GetByID(ctx context.Context, id uint) (*Bounty, error) {
	var bounty Bounty
	err := r.db.WithContext(ctx).
		Preload("WovenSpec").
		Preload("KnittedSpec").
		First(&bounty, id).Error
	if err != nil {
		return nil, err
	}
	return &bounty, nil
}

// Update 更新悬赏
func (r *repository) Update(ctx context.Context, bounty *Bounty) error {
	return r.db.WithContext(ctx).Save(bounty).Error
}

// UpdateWovenSpec 更新梭织规格
func (r *repository) UpdateWovenSpec(ctx context.Context, spec *BountyWovenSpec) error {
	return r.db.WithContext(ctx).Save(spec).Error
}

// UpdateKnittedSpec 更新针织规格
func (r *repository) UpdateKnittedSpec(ctx context.Context, spec *BountyKnittedSpec) error {
	return r.db.WithContext(ctx).Save(spec).Error
}

// Delete 删除悬赏
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Bounty{}, id).Error
}

// DeleteWovenSpec 删除梭织规格
func (r *repository) DeleteWovenSpec(ctx context.Context, bountyID uint) error {
	return r.db.WithContext(ctx).Where("bounty_id = ?", bountyID).Delete(&BountyWovenSpec{}).Error
}

// DeleteKnittedSpec 删除针织规格
func (r *repository) DeleteKnittedSpec(ctx context.Context, bountyID uint) error {
	return r.db.WithContext(ctx).Where("bounty_id = ?", bountyID).Delete(&BountyKnittedSpec{}).Error
}

// List 获取悬赏列表（包含规格）
func (r *repository) List(ctx context.Context, offset, limit int) ([]Bounty, int64, error) {
	var bounties []Bounty
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&Bounty{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据（包含规格）
	err := r.db.WithContext(ctx).
		Preload("WovenSpec").
		Preload("KnittedSpec").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&bounties).Error

	if err != nil {
		return nil, 0, err
	}

	return bounties, total, nil
}
