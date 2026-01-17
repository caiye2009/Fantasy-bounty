package bounty

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Service 悬赏业务逻辑层接口
type Service interface {
	CreateBounty(ctx context.Context, userID uint, req *CreateBountyRequest) (*Bounty, error)
	GetBounty(ctx context.Context, id uint) (*Bounty, error)
	UpdateBounty(ctx context.Context, id uint, req *UpdateBountyRequest) (*Bounty, error)
	DeleteBounty(ctx context.Context, id uint) error
	ListBounties(ctx context.Context, page, pageSize int) ([]Bounty, int64, error)
}

// service 悬赏业务逻辑层实现
type service struct {
	repo Repository
}

// NewService 创建新的 service 实例
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateBounty 创建新悬赏
func (s *service) CreateBounty(ctx context.Context, userID uint, req *CreateBountyRequest) (*Bounty, error) {
	// 解析投标截止日期
	bidDeadline, err := time.Parse("2006-01-02", req.BidDeadline)
	if err != nil {
		bidDeadline, err = time.Parse(time.RFC3339, req.BidDeadline)
		if err != nil {
			return nil, errors.New("invalid bidDeadline format")
		}
	}

	// 解析预计交货日期
	var expectedDeliveryDate time.Time
	if req.ExpectedDeliveryDate != "" {
		expectedDeliveryDate, err = time.Parse("2006-01-02", req.ExpectedDeliveryDate)
		if err != nil {
			expectedDeliveryDate, err = time.Parse(time.RFC3339, req.ExpectedDeliveryDate)
			if err != nil {
				return nil, errors.New("invalid expectedDeliveryDate format")
			}
		}
	}

	bounty := &Bounty{
		BountyType:           req.BountyType,
		ProductName:          req.ProductName,
		SampleType:           req.SampleType,
		ExpectedDeliveryDate: expectedDeliveryDate,
		BidDeadline:          bidDeadline,
		CreatedBy:            userID,
		Status:               "open",
	}

	if err := s.repo.Create(ctx, bounty); err != nil {
		return nil, err
	}

	// 根据悬赏类型创建对应规格
	if req.BountyType == "woven" && req.WovenSpec != nil {
		wovenSpec := &BountyWovenSpec{
			BountyID:       bounty.ID,
			FabricWeight:   req.WovenSpec.FabricWeight,
			FabricWidth:    req.WovenSpec.FabricWidth,
			WarpDensity:    req.WovenSpec.WarpDensity,
			WeftDensity:    req.WovenSpec.WeftDensity,
			Composition:    req.WovenSpec.Composition,
			WarpMaterial:   req.WovenSpec.WarpMaterial,
			WeftMaterial:   req.WovenSpec.WeftMaterial,
			QuantityMeters: req.WovenSpec.QuantityMeters,
		}
		if err := s.repo.CreateWovenSpec(ctx, wovenSpec); err != nil {
			return nil, err
		}
		bounty.WovenSpec = wovenSpec
	} else if req.BountyType == "knitted" && req.KnittedSpec != nil {
		knittedSpec := &BountyKnittedSpec{
			BountyID:         bounty.ID,
			FactoryArticleNo: req.KnittedSpec.FactoryArticleNo,
			FabricWeight:     req.KnittedSpec.FabricWeight,
			FabricWidth:      req.KnittedSpec.FabricWidth,
			MachineType:      req.KnittedSpec.MachineType,
			Composition:      req.KnittedSpec.Composition,
			QuantityKg:       req.KnittedSpec.QuantityKg,
		}
		if err := s.repo.CreateKnittedSpec(ctx, knittedSpec); err != nil {
			return nil, err
		}
		bounty.KnittedSpec = knittedSpec
	}

	return bounty, nil
}

// GetBounty 根据 ID 获取悬赏
func (s *service) GetBounty(ctx context.Context, id uint) (*Bounty, error) {
	bounty, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bounty not found")
		}
		return nil, err
	}
	return bounty, nil
}

// UpdateBounty 更新悬赏
func (s *service) UpdateBounty(ctx context.Context, id uint, req *UpdateBountyRequest) (*Bounty, error) {
	bounty, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bounty not found")
		}
		return nil, err
	}

	// 更新基本字段
	if req.ProductName != "" {
		bounty.ProductName = req.ProductName
	}
	if req.SampleType != "" {
		bounty.SampleType = req.SampleType
	}
	if req.ExpectedDeliveryDate != "" {
		expectedDeliveryDate, err := time.Parse("2006-01-02", req.ExpectedDeliveryDate)
		if err != nil {
			expectedDeliveryDate, err = time.Parse(time.RFC3339, req.ExpectedDeliveryDate)
			if err != nil {
				return nil, errors.New("invalid expectedDeliveryDate format")
			}
		}
		bounty.ExpectedDeliveryDate = expectedDeliveryDate
	}
	if req.BidDeadline != "" {
		bidDeadline, err := time.Parse("2006-01-02", req.BidDeadline)
		if err != nil {
			bidDeadline, err = time.Parse(time.RFC3339, req.BidDeadline)
			if err != nil {
				return nil, errors.New("invalid bidDeadline format")
			}
		}
		bounty.BidDeadline = bidDeadline
	}
	if req.Status != "" {
		bounty.Status = req.Status
	}

	if err := s.repo.Update(ctx, bounty); err != nil {
		return nil, err
	}

	// 更新规格
	if bounty.BountyType == "woven" && req.WovenSpec != nil {
		if bounty.WovenSpec == nil {
			wovenSpec := &BountyWovenSpec{
				BountyID:       bounty.ID,
				FabricWeight:   req.WovenSpec.FabricWeight,
				FabricWidth:    req.WovenSpec.FabricWidth,
				WarpDensity:    req.WovenSpec.WarpDensity,
				WeftDensity:    req.WovenSpec.WeftDensity,
				Composition:    req.WovenSpec.Composition,
				WarpMaterial:   req.WovenSpec.WarpMaterial,
				WeftMaterial:   req.WovenSpec.WeftMaterial,
				QuantityMeters: req.WovenSpec.QuantityMeters,
			}
			if err := s.repo.CreateWovenSpec(ctx, wovenSpec); err != nil {
				return nil, err
			}
			bounty.WovenSpec = wovenSpec
		} else {
			bounty.WovenSpec.FabricWeight = req.WovenSpec.FabricWeight
			bounty.WovenSpec.FabricWidth = req.WovenSpec.FabricWidth
			bounty.WovenSpec.WarpDensity = req.WovenSpec.WarpDensity
			bounty.WovenSpec.WeftDensity = req.WovenSpec.WeftDensity
			bounty.WovenSpec.Composition = req.WovenSpec.Composition
			bounty.WovenSpec.WarpMaterial = req.WovenSpec.WarpMaterial
			bounty.WovenSpec.WeftMaterial = req.WovenSpec.WeftMaterial
			bounty.WovenSpec.QuantityMeters = req.WovenSpec.QuantityMeters
			if err := s.repo.UpdateWovenSpec(ctx, bounty.WovenSpec); err != nil {
				return nil, err
			}
		}
	} else if bounty.BountyType == "knitted" && req.KnittedSpec != nil {
		if bounty.KnittedSpec == nil {
			knittedSpec := &BountyKnittedSpec{
				BountyID:         bounty.ID,
				FactoryArticleNo: req.KnittedSpec.FactoryArticleNo,
				FabricWeight:     req.KnittedSpec.FabricWeight,
				FabricWidth:      req.KnittedSpec.FabricWidth,
				MachineType:      req.KnittedSpec.MachineType,
				Composition:      req.KnittedSpec.Composition,
				QuantityKg:       req.KnittedSpec.QuantityKg,
			}
			if err := s.repo.CreateKnittedSpec(ctx, knittedSpec); err != nil {
				return nil, err
			}
			bounty.KnittedSpec = knittedSpec
		} else {
			bounty.KnittedSpec.FactoryArticleNo = req.KnittedSpec.FactoryArticleNo
			bounty.KnittedSpec.FabricWeight = req.KnittedSpec.FabricWeight
			bounty.KnittedSpec.FabricWidth = req.KnittedSpec.FabricWidth
			bounty.KnittedSpec.MachineType = req.KnittedSpec.MachineType
			bounty.KnittedSpec.Composition = req.KnittedSpec.Composition
			bounty.KnittedSpec.QuantityKg = req.KnittedSpec.QuantityKg
			if err := s.repo.UpdateKnittedSpec(ctx, bounty.KnittedSpec); err != nil {
				return nil, err
			}
		}
	}

	return bounty, nil
}

// DeleteBounty 删除悬赏
func (s *service) DeleteBounty(ctx context.Context, id uint) error {
	bounty, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bounty not found")
		}
		return err
	}

	// 删除关联的规格
	if bounty.BountyType == "woven" {
		if err := s.repo.DeleteWovenSpec(ctx, id); err != nil {
			return err
		}
	} else if bounty.BountyType == "knitted" {
		if err := s.repo.DeleteKnittedSpec(ctx, id); err != nil {
			return err
		}
	}

	return s.repo.Delete(ctx, id)
}

// ListBounties 获取悬赏列表
func (s *service) ListBounties(ctx context.Context, page, pageSize int) ([]Bounty, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}
