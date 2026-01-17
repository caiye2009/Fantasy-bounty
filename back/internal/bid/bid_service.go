package bid

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 竞标业务逻辑层接口
type Service interface {
	CreateBid(ctx context.Context, userID uint, req *CreateBidRequest) (*Bid, error)
	GetBid(ctx context.Context, id string) (*Bid, error)
	DeleteBid(ctx context.Context, id string) error
	ListBidsByBountyID(ctx context.Context, bountyID uint, page, pageSize int) ([]Bid, int64, error)
	ListBidsByUserID(ctx context.Context, userID uint, status string, page, pageSize int) ([]Bid, int64, error)
}

// service 竞标业务逻辑层实现
type service struct {
	repo Repository
}

// NewService 创建新的 service 实例
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateBid 创建新竞标
func (s *service) CreateBid(ctx context.Context, userID uint, req *CreateBidRequest) (*Bid, error) {
	bid := &Bid{
		ID:       uuid.New().String(),
		BountyID: req.BountyID,
		UserID:   userID,
		BidPrice: req.BidPrice,
		Status:   "pending",
	}

	// 处理梭织规格
	if req.WovenSpec != nil {
		greigeDate, _ := time.Parse("2006-01-02", req.WovenSpec.GreigeDeliveryDate)
		bid.WovenSpec = &BidWovenSpec{
			SizeLength:         req.WovenSpec.SizeLength,
			GreigeFabricType:   req.WovenSpec.GreigeFabricType,
			GreigeDeliveryDate: greigeDate,
		}
	}

	// 处理针织规格
	if req.KnittedSpec != nil {
		greigeDate, _ := time.Parse("2006-01-02", req.KnittedSpec.GreigeDeliveryDate)
		bid.KnittedSpec = &BidKnittedSpec{
			SizeWeight:         req.KnittedSpec.SizeWeight,
			GreigeFabricType:   req.KnittedSpec.GreigeFabricType,
			GreigeDeliveryDate: greigeDate,
		}
	}

	if err := s.repo.Create(ctx, bid); err != nil {
		return nil, err
	}

	return bid, nil
}

// GetBid 根据 ID 获取竞标
func (s *service) GetBid(ctx context.Context, id string) (*Bid, error) {
	bid, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bid not found")
		}
		return nil, err
	}
	return bid, nil
}

// DeleteBid 删除竞标
func (s *service) DeleteBid(ctx context.Context, id string) error {
	// 先检查是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bid not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

// ListBidsByBountyID 根据赏金ID获取竞标列表
func (s *service) ListBidsByBountyID(ctx context.Context, bountyID uint, page, pageSize int) ([]Bid, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.ListByBountyID(ctx, bountyID, offset, pageSize)
}

// ListBidsByUserID 根据用户ID获取竞标列表
func (s *service) ListBidsByUserID(ctx context.Context, userID uint, status string, page, pageSize int) ([]Bid, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.ListByUserID(ctx, userID, status, offset, pageSize)
}
