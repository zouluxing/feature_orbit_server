package model

import "time"

// FeatureStatus represents the lifecycle state of a feature flag.
type FeatureStatus string

const (
	FeatureStatusDraft    FeatureStatus = "draft"
	FeatureStatusActive   FeatureStatus = "active"
	FeatureStatusArchived FeatureStatus = "archived"
)

// Feature is the core domain model for feature_orbit_server.
// A Feature represents a product feature / feature flag managed by the platform.
type Feature struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug        string        `gorm:"type:varchar(128);uniqueIndex;not null" json:"slug"`
	Name        string        `gorm:"type:varchar(256);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Status      FeatureStatus `gorm:"type:varchar(16);not null;default:'draft'" json:"status"`
	OwnerUUID   string        `gorm:"type:varchar(36);not null;index" json:"owner_uuid"`  // UMS user UUID
	OwnerEmail  string        `gorm:"type:varchar(255);not null" json:"owner_email"`
	Tags        string        `gorm:"type:jsonb;not null;default:'[]'" json:"-"`           // JSON array, serialised separately
	CreatedAt   time.Time     `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time    `gorm:"index" json:"-"`
}

func (Feature) TableName() string { return "features" }
