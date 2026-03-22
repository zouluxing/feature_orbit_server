package model

import "time"

type Role struct {
	ID          int64        `gorm:"primaryKey;autoIncrement"`
	Name        string       `gorm:"type:varchar(64);uniqueIndex;not null"`
	Description *string      `gorm:"type:text"`
	IsSystem    bool         `gorm:"not null;default:false"`
	CreatedAt   time.Time    `gorm:"not null;autoCreateTime"`
	DeletedAt   *time.Time   `gorm:"index"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          int64   `gorm:"primaryKey;autoIncrement"`
	Resource    string  `gorm:"type:varchar(64);not null;uniqueIndex:uq_resource_action"`
	Action      string  `gorm:"type:varchar(32);not null;uniqueIndex:uq_resource_action"`
	Description *string `gorm:"type:text"`
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	UserID    int64     `gorm:"primaryKey"`
	RoleID    int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	RoleID       int64 `gorm:"primaryKey"`
	PermissionID int64 `gorm:"primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }

type OAuth2Client struct {
	ID               int64      `gorm:"primaryKey;autoIncrement"`
	ClientID         string     `gorm:"type:varchar(64);uniqueIndex;not null"`
	ClientSecretHash string     `gorm:"type:varchar(255);not null"`
	Name             string     `gorm:"type:varchar(128);not null"`
	RedirectURIs     string     `gorm:"type:jsonb;not null"`
	Scopes           string     `gorm:"type:jsonb;not null"`
	GrantTypes       string     `gorm:"type:jsonb;not null"`
	IsActive         bool       `gorm:"not null;default:true"`
	OwnerUserID      int64      `gorm:"not null;index"`
	CreatedAt        time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"not null;autoUpdateTime"`
	DeletedAt        *time.Time `gorm:"index"`
}

func (OAuth2Client) TableName() string { return "oauth2_clients" }

type OAuth2AuthCode struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"`
	Code                string    `gorm:"type:varchar(128);uniqueIndex;not null"`
	ClientID            string    `gorm:"type:varchar(64);not null;index"`
	UserID              int64     `gorm:"not null"`
	RedirectURI         string    `gorm:"type:text;not null"`
	Scopes              string    `gorm:"type:jsonb;not null"`
	CodeChallenge       *string   `gorm:"type:varchar(128)"`
	CodeChallengeMethod *string   `gorm:"type:varchar(8)"`
	ExpiresAt           time.Time `gorm:"not null;index"`
	CreatedAt           time.Time `gorm:"not null;autoCreateTime"`
}

func (OAuth2AuthCode) TableName() string { return "oauth2_authorization_codes" }

type OAuth2Token struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	JTI          string     `gorm:"type:varchar(64);uniqueIndex;not null"`
	ClientID     string     `gorm:"type:varchar(64);not null;index"`
	UserID       int64      `gorm:"not null;index"`
	Scopes       string     `gorm:"type:jsonb;not null"`
	RefreshToken string     `gorm:"type:varchar(512);uniqueIndex;not null"`
	ExpiresAt    time.Time  `gorm:"not null"`
	RevokedAt    *time.Time
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime"`
}

func (OAuth2Token) TableName() string { return "oauth2_tokens" }
