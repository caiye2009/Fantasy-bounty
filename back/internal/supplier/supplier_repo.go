package supplier

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// Supplier
	GetSupplierByID(ctx context.Context, id string) (*Supplier, error)
	ListSuppliers(ctx context.Context, offset, limit int) ([]Supplier, int64, error)

	// UserSupplier
	GetUserSupplierByUsername(ctx context.Context, username string) (*UserSupplier, error)

	// SupplierApplication
	CreateApplication(ctx context.Context, app *SupplierApplication) error
	GetPendingApplicationByUsername(ctx context.Context, username string) (*SupplierApplication, error)
	GetLatestRejectedByUsername(ctx context.Context, username string) (*SupplierApplication, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ========== Supplier ==========

func (r *repository) GetSupplierByID(ctx context.Context, id string) (*Supplier, error) {
	var supplier Supplier
	err := r.db.WithContext(ctx).First(&supplier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *repository) ListSuppliers(ctx context.Context, offset, limit int) ([]Supplier, int64, error) {
	var suppliers []Supplier
	var total int64

	if err := r.db.WithContext(ctx).Model(&Supplier{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&suppliers).Error

	if err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

// ========== UserSupplier ==========

func (r *repository) GetUserSupplierByUsername(ctx context.Context, username string) (*UserSupplier, error) {
	var us UserSupplier
	err := r.db.WithContext(ctx).First(&us, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &us, nil
}

// ========== SupplierApplication ==========

func (r *repository) CreateApplication(ctx context.Context, app *SupplierApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *repository) GetPendingApplicationByUsername(ctx context.Context, username string) (*SupplierApplication, error) {
	var app SupplierApplication
	err := r.db.WithContext(ctx).
		Where("username = ? AND status = ?", username, ApplicationStatusPending).
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *repository) GetLatestRejectedByUsername(ctx context.Context, username string) (*SupplierApplication, error) {
	var app SupplierApplication
	err := r.db.WithContext(ctx).
		Where("username = ? AND status = ?", username, ApplicationStatusRejected).
		Order("created_at DESC").
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}
