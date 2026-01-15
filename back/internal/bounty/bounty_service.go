package bounty

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// Service 赏金业务逻辑层接口
type Service interface {
	CreateBounty(ctx context.Context, req *CreateBountyRequest) (*Bounty, error)
	GetBounty(ctx context.Context, id uint) (*Bounty, error)
	UpdateBounty(ctx context.Context, id uint, req *UpdateBountyRequest) (*Bounty, error)
	DeleteBounty(ctx context.Context, id uint) error
	ListBounties(ctx context.Context, page, pageSize int) ([]Bounty, int64, error)
}

// service 赏金业务逻辑层实现
type service struct {
	repo Repository
}

// NewService 创建新的 service 实例
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateBounty 创建新赏金
func (s *service) CreateBounty(ctx context.Context, req *CreateBountyRequest) (*Bounty, error) {
	bounty := &Bounty{
		Title:       req.Title,
		Description: req.Description,
		Reward:      req.Reward,
		CreatedBy:   req.CreatedBy,
		Status:      "open",
	}

	if err := s.repo.Create(ctx, bounty); err != nil {
		return nil, err
	}

	return bounty, nil
}

// GetBounty 根据 ID 获取赏金
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

// UpdateBounty 更新赏金
func (s *service) UpdateBounty(ctx context.Context, id uint, req *UpdateBountyRequest) (*Bounty, error) {
	// 先查询是否存在
	bounty, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bounty not found")
		}
		return nil, err
	}

	// 更新字段
	if req.Title != "" {
		bounty.Title = req.Title
	}
	if req.Description != "" {
		bounty.Description = req.Description
	}
	if req.Reward > 0 {
		bounty.Reward = req.Reward
	}
	if req.Status != "" {
		bounty.Status = req.Status
	}

	// 保存更新
	if err := s.repo.Update(ctx, bounty); err != nil {
		return nil, err
	}

	return bounty, nil
}

// DeleteBounty 删除赏金
func (s *service) DeleteBounty(ctx context.Context, id uint) error {
	// 先检查是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bounty not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

// ListBounties 获取赏金列表
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
