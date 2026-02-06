package audit

import "time"

// AuditLog represents a single audit log entry persisted to the database.
type AuditLog struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RequestID  string    `gorm:"type:varchar(36);index"`
	Username   string    `gorm:"type:varchar(20);index"`
	Action     string    `gorm:"type:varchar(100);index"`
	Resource   string    `gorm:"type:varchar(50)"`
	ResourceID string    `gorm:"type:varchar(36)"`
	Method     string    `gorm:"type:varchar(10)"`
	Path       string    `gorm:"type:varchar(500)"`
	StatusCode int
	ClientIP   string    `gorm:"type:varchar(45)"`
	UserAgent  string    `gorm:"type:varchar(500)"`
	Duration   int64     // milliseconds
	Detail     string    `gorm:"type:jsonb"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index"`
}
