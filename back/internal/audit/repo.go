package audit

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines the persistence interface for audit logs.
type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new audit log repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
