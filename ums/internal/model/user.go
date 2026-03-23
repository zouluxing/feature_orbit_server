package model

import "time"

type UserStatus int8

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
	UserStatusLocked   UserStatus = 3
)

type User struct {
	ID               int64      `gorm:"primaryKey;autoIncrement"`
	UUID             string     `gorm:"type:uuid;uniqueIndex;not null"`
	Email            string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username         string     `gorm:"type:varchar(64);uniqueIndex;not null"`
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
