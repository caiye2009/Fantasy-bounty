package client

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	CreateOrUpdateProfile(ctx context.Context, userID string, req *CreateClientProfileRequest) (*ClientProfile, error)
	GetProfile(ctx context.Context, userID string) (*ClientProfile, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateOrUpdateProfile(ctx context.Context, userID string, req *CreateClientProfileRequest) (*ClientProfile, error) {
	p := &ClientProfile{
		ID:          userID,
		UserID:      userID,
		CompanyName: req.CompanyName,
		ContactName: req.ContactName,
		Address:     req.Address,
		Industry:    req.Industry,
	}
	if err := s.repo.CreateOrUpdateProfile(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*ClientProfile, error) {
	p, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client profile not found")
		}
		return nil, err
	}
	return p, nil
}

