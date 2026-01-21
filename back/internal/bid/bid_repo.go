package bid

import (
	"context"

	"gorm.io/gorm"
)

// Repository 竞标数据访问层接口
type Repository interface {
	Create(ctx context.Context, bid *Bid) error
	GetByID(ctx context.Context, id string) (*Bid, error)
	Delete(ctx context.Context, id string) error
	ListByBountyID(ctx context.Context, bountyID uint, offset, limit int) ([]Bid, int64, error)
	ListByUserID(ctx context.Context, userID uint, status string, offset, limit int) ([]Bid, int64, error)
}

// repository 竞标数据访问层实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建新的 repository 实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建新竞标
func (r *repository) Create(ctx context.Context, bid *Bid) error {
	return r.db.WithContext(ctx).Create(bid).Error
}

// GetByID 根据 ID 获取竞标（包含规格）
func (r *repository) GetByID(ctx context.Context, id string) (*Bid, error) {
	var bid Bid
	err := r.db.WithContext(ctx).
		Preload("WovenSpec").
		Preload("KnittedSpec").
		First(&bid, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

// Delete 删除竞标
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Bid{}, "id = ?", id).Error
}

// ListByBountyID 根据赏金ID获取竞标列表
func (r *repository) ListByBountyID(ctx context.Context, bountyID uint, offset, limit int) ([]Bid, int64, error) {
	var bids []Bid
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&Bid{}).Where("bounty_id = ?", bountyID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据（包含规格）
	err := r.db.WithContext(ctx).
		Preload("WovenSpec").
		Preload("KnittedSpec").
		Where("bounty_id = ?", bountyID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&bids).Error

	if err != nil {
		return nil, 0, err
	}

	return bids, total, nil
}

// ListByUserID 根据用户ID获取竞标列表（带关联bounty信息和投标规格）
func (r *repository) ListByUserID(ctx context.Context, userID uint, status string, offset, limit int) ([]Bid, int64, error) {
	var bids []Bid
	var total int64

	query := r.db.WithContext(ctx).Model(&Bid{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，并关联bounty表获取额外信息
	err := r.db.WithContext(ctx).
		Table("bids").
		Select("bids.*, bounties.product_name as bounty_product_name, bounties.product_code as bounty_product_code, bounties.bounty_type, bounties.bid_deadline").
		Joins("LEFT JOIN bounties ON bids.bounty_id = bounties.id").
		Where("bids.user_id = ?", userID).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if status != "" {
				return db.Where("bids.status = ?", status)
			}
			return db
		}).
		Offset(offset).
		Limit(limit).
		Order("bids.created_at DESC").
		Find(&bids).Error

	if err != nil {
		return nil, 0, err
	}

	// 预加载投标规格
	bidIDs := make([]string, len(bids))
	for i, bid := range bids {
		bidIDs[i] = bid.ID
	}

	if len(bidIDs) > 0 {
		// 加载梭织规格
		var wovenSpecs []BidWovenSpec
		r.db.WithContext(ctx).Where("bid_id IN ?", bidIDs).Find(&wovenSpecs)
		wovenSpecMap := make(map[string]*BidWovenSpec)
		for i := range wovenSpecs {
			wovenSpecMap[wovenSpecs[i].BidID] = &wovenSpecs[i]
		}

		// 加载针织规格
		var knittedSpecs []BidKnittedSpec
		r.db.WithContext(ctx).Where("bid_id IN ?", bidIDs).Find(&knittedSpecs)
		knittedSpecMap := make(map[string]*BidKnittedSpec)
		for i := range knittedSpecs {
			knittedSpecMap[knittedSpecs[i].BidID] = &knittedSpecs[i]
		}

		// 关联到bid
		for i := range bids {
			bids[i].WovenSpec = wovenSpecMap[bids[i].ID]
			bids[i].KnittedSpec = knittedSpecMap[bids[i].ID]
		}
	}

	return bids, total, nil
}
