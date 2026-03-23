package model

import "time"

type UserStatus int8

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
	UserStatusLocked   UserStatus = 3
)

// User GORM 模型。
// uniqueIndex 的约束名显式指定，与 migrations/000001_init.up.sql 中的命名保持一致，
// 避免 GORM AutoMigrate 生成 uni_xxx 与 SQL 中 uq_xxx 冲突。
type User struct {
	ID               int64      `gorm:"primaryKey;autoIncrement"`
	UUID             string     `gorm:"type:varchar(36);uniqueIndex:uq_users_uuid;not null"`
	Email            string     `gorm:"type:varchar(255);uniqueIndex:uq_users_email;not null"`
	Username         string     `gorm:"type:varchar(64);uniqueIndex:uq_users_username;not null"`
	PasswordHash     string     `gorm:"type:varchar(255);not null"`
	Phone            *string    `gorm:"type:varchar(32)"`
	AvatarURL        *string    `gorm:"type:text"`
	Status           UserStatus `gorm:"not null;default:1"`
	MFAEnabled       bool       `gorm:"not null;default:false"`
	MFASecret        *string    `gorm:"type:varchar(255)"` // AES-GCM ciphertext in production
	FailedLoginCount int        `gorm:"not null;default:0"`
	LastLoginAt      *time.Time
	Extra            *string    `gorm:"type:jsonb"`
	CreatedAt        time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt        *time.Time `gorm:"index"`
	Roles            []Role     `gorm:"many2many:user_roles;"`
}

func (User) TableName() string { return "users" }
