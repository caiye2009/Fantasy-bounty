package company

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// Company
	GetCompanyByID(ctx context.Context, id string) (*Company, error)
	ListCompanies(ctx context.Context, offset, limit int) ([]Company, int64, error)

	// AccountCompany
	GetAccountCompanyByAccountID(ctx context.Context, accountID string) (*AccountCompany, error)

	// CompanyApplication
	CreateApplication(ctx context.Context, app *CompanyApplication) error
	GetPendingApplicationByAccountID(ctx context.Context, accountID string) (*CompanyApplication, error)
	GetLatestRejectedByAccountID(ctx context.Context, accountID string) (*CompanyApplication, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ========== Company ==========

func (r *repository) GetCompanyByID(ctx context.Context, id string) (*Company, error) {
	var company Company
	err := r.db.WithContext(ctx).First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *repository) ListCompanies(ctx context.Context, offset, limit int) ([]Company, int64, error) {
	var companies []Company
	var total int64

	if err := r.db.WithContext(ctx).Model(&Company{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&companies).Error

	if err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

// ========== AccountCompany ==========

func (r *repository) GetAccountCompanyByAccountID(ctx context.Context, accountID string) (*AccountCompany, error) {
	var ac AccountCompany
	err := r.db.WithContext(ctx).First(&ac, "account_id = ?", accountID).Error
	if err != nil {
		return nil, err
	}
	return &ac, nil
}

// ========== CompanyApplication ==========

func (r *repository) CreateApplication(ctx context.Context, app *CompanyApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *repository) GetPendingApplicationByAccountID(ctx context.Context, accountID string) (*CompanyApplication, error) {
	var app CompanyApplication
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND status = ?", accountID, ApplicationStatusPending).
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *repository) GetLatestRejectedByAccountID(ctx context.Context, accountID string) (*CompanyApplication, error) {
	var app CompanyApplication
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND status = ?", accountID, ApplicationStatusRejected).
		Order("created_at DESC").
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}
