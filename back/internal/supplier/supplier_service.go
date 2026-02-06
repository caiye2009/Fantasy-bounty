package supplier

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	// Supplier
	GetSupplier(ctx context.Context, id string) (*Supplier, error)
	ListSuppliers(ctx context.Context, page, pageSize int) ([]Supplier, int64, error)

	// Application - 用户操作
	ApplySupplier(ctx context.Context, username string, req *ApplySupplierRequest, licenseImage string) (*SupplierApplication, error)
	GetMySupplierStatus(ctx context.Context, username string) (*MySupplierStatus, error)

	// OCR识别
	RecognizeLicense(ctx context.Context, imagePath string) (*OCRResult, error)

	// 供应商认证校验（供其他模块调用）
	IsUserVerified(ctx context.Context, username string) (bool, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ========== Supplier ==========

func (s *service) GetSupplier(ctx context.Context, id string) (*Supplier, error) {
	supplier, err := s.repo.GetSupplierByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supplier not found")
		}
		return nil, err
	}
	return supplier, nil
}

func (s *service) ListSuppliers(ctx context.Context, page, pageSize int) ([]Supplier, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.ListSuppliers(ctx, offset, pageSize)
}

// ========== Application - 用户操作 ==========

func (s *service) ApplySupplier(ctx context.Context, username string, req *ApplySupplierRequest, licenseImage string) (*SupplierApplication, error) {
	// 检查是否已经认证过供应商
	_, err := s.repo.GetUserSupplierByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("您已认证过供应商，无需重复申请")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 检查是否有待审核的申请
	_, err = s.repo.GetPendingApplicationByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("您已有待审核的申请，请等待审核结果")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建申请
	app := &SupplierApplication{
		Username:             username,
		Name:                 req.Name,
		BusinessLicenseNo:    req.BusinessLicenseNo,
		BusinessLicenseImage: licenseImage,
		Status:               ApplicationStatusPending,
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *service) GetMySupplierStatus(ctx context.Context, username string) (*MySupplierStatus, error) {
	status := &MySupplierStatus{
		HasVerifiedSupplier: false,
	}

	// 检查是否已认证供应商
	us, err := s.repo.GetUserSupplierByUsername(ctx, username)
	if err == nil {
		// 已认证，获取供应商信息
		supplier, err := s.repo.GetSupplierByID(ctx, us.SupplierID)
		if err == nil {
			status.HasVerifiedSupplier = true
			status.Supplier = supplier
			return status, nil
		}
	}

	// 检查是否有待审核的申请
	pendingApp, err := s.repo.GetPendingApplicationByUsername(ctx, username)
	if err == nil {
		status.PendingApplication = pendingApp
	}

	// 获取最近一次被拒绝的申请
	rejectedApp, err := s.repo.GetLatestRejectedByUsername(ctx, username)
	if err == nil {
		status.LatestRejected = rejectedApp
	}

	return status, nil
}

// ========== OCR识别 ==========

// RecognizeLicense 识别营业执照（当前为假函数，返回固定测试数据）
// TODO: 后期替换为真实OCR服务调用
func (s *service) RecognizeLicense(ctx context.Context, imagePath string) (*OCRResult, error) {
	return &OCRResult{
		CompanyName:       "深圳市测试科技有限公司",
		BusinessLicenseNo: "91440300MA5FEXAMPLE",
		LegalPerson:       "张三",
		RegisteredCapital: "1000万人民币",
		EstablishDate:     "2020-01-15",
		BusinessScope:     "技术开发、技术咨询、技术服务；软件开发；信息系统集成服务",
		Address:           "深圳市南山区科技园南区A栋101",
	}, nil
}

// ========== 供应商认证校验 ==========

func (s *service) IsUserVerified(ctx context.Context, username string) (bool, error) {
	_, err := s.repo.GetUserSupplierByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
