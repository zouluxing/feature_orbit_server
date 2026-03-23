package model

import "time"

type FeatureStatus string

const (
	FeatureStatusDraft    FeatureStatus = "draft"
	FeatureStatusActive   FeatureStatus = "active"
	FeatureStatusArchived FeatureStatus = "archived"
)

// Feature GORM 模型。
// uniqueIndex 约束名显式指定，与 migrations/001_create_features.up.sql 中的命名保持一致，
// 避免 GORM AutoMigrate 生成 uni_features_slug 与 SQL 中 uq_features_slug 冲突。
type Feature struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement"                              json:"id"`
	Slug        string        `gorm:"type:varchar(128);uniqueIndex:uq_features_slug;not null" json:"slug"`
	Name        string        `gorm:"type:varchar(256);not null"                            json:"name"`
	Description string        `gorm:"type:text"                                             json:"description"`
	Status      FeatureStatus `gorm:"type:varchar(16);not null;default:'draft'"              json:"status"`
	OwnerUUID   string        `gorm:"type:varchar(36);not null;index"                        json:"owner_uuid"`
	OwnerEmail  string        `gorm:"type:varchar(255);not null"                            json:"owner_email"`
	Tags        string        `gorm:"type:jsonb;not null;default:'[]'"                      json:"-"`
	CreatedAt   time.Time     `gorm:"not null;autoCreateTime"                               json:"created_at"`
	UpdatedAt   time.Time     `gorm:"not null;autoUpdateTime"                               json:"updated_at"`
	DeletedAt   *time.Time    `gorm:"index"                                                 json:"-"`
}

func (Feature) TableName() string { return "features" }
