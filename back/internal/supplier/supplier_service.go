package supplier

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	// 创建/更新供应商档案
	CreateOrUpdateProfile(ctx context.Context, userID string, req *CreateSupplierRequest) (*SupplierProfile, error)
	
	// 获取供应商信息
	GetSupplierProfile(ctx context.Context, userID string) (*SupplierProfile, error)
	
	// 获取供应商完整信息
	GetSupplierFullInfo(ctx context.Context, userID string) (*SupplierFullInfo, error)
	
	// 更新机器能力
	UpdateCapabilities(ctx context.Context, userID string, capabilities map[string]int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateOrUpdateProfile 创建/更新供应商档案
func (s *service) CreateOrUpdateProfile(ctx context.Context, userID string, req *CreateSupplierRequest) (*SupplierProfile, error) {
	// 将能力map转换为JSON
	capabilitiesJSON, err := json.Marshal(req.Capabilities)
	if err != nil {
		return nil, errors.New("机器能力格式错误")
	}

	profile := &SupplierProfile{
		UserID:      userID,
		CompanyType: req.CompanyType,
		CompanyName: req.CompanyName,
		Capabilities: capabilitiesJSON,
	}

	if err := s.repo.CreateOrUpdateProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// GetSupplierProfile 获取供应商信息
func (s *service) GetSupplierProfile(ctx context.Context, userID string) (*SupplierProfile, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supplier profile not found")
		}
		return nil, err
	}
	return profile, nil
}

// GetSupplierFullInfo 获取供应商完整信息
func (s *service) GetSupplierFullInfo(ctx context.Context, userID string) (*SupplierFullInfo, error) {
	// 获取供应商档案
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有profile时返回空的完整信息
			return &SupplierFullInfo{
				Info:          nil,
				Qualification: nil,
				Capabilities:  make(map[string]int),
			}, nil
		}
		return nil, err
	}

	// 解析capabilities JSON
	var capabilities map[string]int
	if len(profile.Capabilities) > 0 {
		if err := json.Unmarshal(profile.Capabilities, &capabilities); err != nil {
			capabilities = make(map[string]int)
		}
	} else {
		capabilities = make(map[string]int)
	}

	// 构建完整信息
	fullInfo := &SupplierFullInfo{
		Info: &SupplierInfo{
			ID:          profile.ID,
			UserID:      profile.UserID,
			CompanyType: profile.CompanyType,
			CompanyName: profile.CompanyName,
			CreatedAt:   profile.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   profile.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
		Qualification: &SupplierQualification{},
		Capabilities:  capabilities,
	}

	return fullInfo, nil
}

// UpdateCapabilities 更新机器能力（增量更新）
func (s *service) UpdateCapabilities(ctx context.Context, userID string, newCapabilities map[string]int) error {
	// 获取现有capabilities
	existingCapabilities := make(map[string]int)
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	
	if profile != nil && len(profile.Capabilities) > 0 {
		if err := json.Unmarshal(profile.Capabilities, &existingCapabilities); err != nil {
			existingCapabilities = make(map[string]int)
		}
	}

	// 增量更新：只更新传入的字段，保留其他字段
	for key, value := range newCapabilities {
		existingCapabilities[key] = value
	}

	return s.repo.UpdateCapabilities(ctx, userID, existingCapabilities)
}
