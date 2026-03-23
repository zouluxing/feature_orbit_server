package model

import "time"

type FeatureStatus string

const (
	FeatureStatusDraft    FeatureStatus = "draft"
	FeatureStatusActive   FeatureStatus = "active"
	FeatureStatusArchived FeatureStatus = "archived"
)

type Feature struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug        string        `gorm:"type:varchar(128);uniqueIndex;not null" json:"slug"`
	Name        string        `gorm:"type:varchar(256);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Status      FeatureStatus `gorm:"type:varchar(16);not null;default:'draft'" json:"status"`
	OwnerUUID   string        `gorm:"type:varchar(36);not null;index" json:"owner_uuid"`
	OwnerEmail  string        `gorm:"type:varchar(255);not null" json:"owner_email"`
	Tags        string        `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	CreatedAt   time.Time     `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time    `gorm:"index" json:"-"`
}

func (Feature) TableName() string { return "features" }
