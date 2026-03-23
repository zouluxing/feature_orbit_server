package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/zouluxing/ums/internal/model"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]model.Role, error)
	FindByID(ctx context.Context, id int64) (*model.Role, error)
	Create(ctx context.Context, r *model.Role) error
	Delete(ctx context.Context, id int64) error
	AssignToUser(ctx context.Context, userID int64, roleIDs []int64) error
	RemoveFromUser(ctx context.Context, userID int64, roleID int64) error
	SetPermissions(ctx context.Context, roleID int64, permIDs []int64) error
}

type roleRepo struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository { return &roleRepo{db: db} }

func (r *roleRepo) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Preload("Permissions").Find(&roles).Error
	return roles, err
}
func (r *roleRepo) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Preload("Permissions").First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }; return &role, err
}
func (r *roleRepo) Create(ctx context.Context, role *model.Role) error { return r.db.WithContext(ctx).Create(role).Error }
func (r *roleRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Role{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error
}
func (r *roleRepo) AssignToUser(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil { return err }
		for _, rid := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil { return err }
		}
		return nil
	})
}
func (r *roleRepo) RemoveFromUser(ctx context.Context, userID int64, roleID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error
}
func (r *roleRepo) SetPermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil { return err }
		for _, pid := range permIDs {
			if err := tx.Create(&model.RolePermission{RoleID: roleID, PermissionID: pid}).Error; err != nil { return err }
		}
		return nil
	})
}

type PermissionRepository interface {
	FindAll(ctx context.Context) ([]model.Permission, error)
	FindByIDs(ctx context.Context, ids []int64) ([]model.Permission, error)
}

type permRepo struct{ db *gorm.DB }

func NewPermissionRepository(db *gorm.DB) PermissionRepository { return &permRepo{db: db} }

func (p *permRepo) FindAll(ctx context.Context) ([]model.Permission, error) {
	var perms []model.Permission
	return perms, p.db.WithContext(ctx).Find(&perms).Error
}
func (p *permRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.Permission, error) {
	var perms []model.Permission
	return perms, p.db.WithContext(ctx).Where("id IN ?", ids).Find(&perms).Error
}

type OAuth2Repository interface {
	CreateClient(ctx context.Context, c *model.OAuth2Client) error
	FindClientByClientID(ctx context.Context, clientID string) (*model.OAuth2Client, error)
	FindClientsByOwner(ctx context.Context, ownerUserID int64) ([]model.OAuth2Client, error)
	UpdateClient(ctx context.Context, c *model.OAuth2Client) error
	DeleteClient(ctx context.Context, id int64) error
	SaveAuthCode(ctx context.Context, code *model.OAuth2AuthCode) error
	ConsumeAuthCode(ctx context.Context, code string) (*model.OAuth2AuthCode, error)
	SaveToken(ctx context.Context, t *model.OAuth2Token) error
	FindTokenByRefresh(ctx context.Context, refreshToken string) (*model.OAuth2Token, error)
	RevokeToken(ctx context.Context, jti string) error
	RevokeAllClientTokens(ctx context.Context, clientID string) error
}

type oauth2Repo struct{ db *gorm.DB }

func NewOAuth2Repository(db *gorm.DB) OAuth2Repository { return &oauth2Repo{db: db} }

func (r *oauth2Repo) CreateClient(ctx context.Context, c *model.OAuth2Client) error { return r.db.WithContext(ctx).Create(c).Error }
func (r *oauth2Repo) FindClientByClientID(ctx context.Context, clientID string) (*model.OAuth2Client, error) {
	var c model.OAuth2Client
	err := r.db.WithContext(ctx).Where("client_id = ? AND deleted_at IS NULL AND is_active = true", clientID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }; return &c, err
}
func (r *oauth2Repo) FindClientsByOwner(ctx context.Context, ownerUserID int64) ([]model.OAuth2Client, error) {
	var clients []model.OAuth2Client
	return clients, r.db.WithContext(ctx).Where("owner_user_id = ? AND deleted_at IS NULL", ownerUserID).Find(&clients).Error
}
func (r *oauth2Repo) UpdateClient(ctx context.Context, c *model.OAuth2Client) error { return r.db.WithContext(ctx).Save(c).Error }
func (r *oauth2Repo) DeleteClient(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.OAuth2Client{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
func (r *oauth2Repo) SaveAuthCode(ctx context.Context, code *model.OAuth2AuthCode) error { return r.db.WithContext(ctx).Create(code).Error }
func (r *oauth2Repo) ConsumeAuthCode(ctx context.Context, code string) (*model.OAuth2AuthCode, error) {
	var ac model.OAuth2AuthCode
	err := r.db.WithContext(ctx).Where("code = ? AND expires_at > ?", code, time.Now()).First(&ac).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	if err != nil { return nil, err }
	r.db.WithContext(ctx).Delete(&ac)
	return &ac, nil
}
func (r *oauth2Repo) SaveToken(ctx context.Context, t *model.OAuth2Token) error { return r.db.WithContext(ctx).Create(t).Error }
func (r *oauth2Repo) FindTokenByRefresh(ctx context.Context, refreshToken string) (*model.OAuth2Token, error) {
	var t model.OAuth2Token
	err := r.db.WithContext(ctx).Where("refresh_token = ? AND revoked_at IS NULL AND expires_at > ?", refreshToken, time.Now()).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }; return &t, err
}
func (r *oauth2Repo) RevokeToken(ctx context.Context, jti string) error {
	return r.db.WithContext(ctx).Model(&model.OAuth2Token{}).Where("jti = ?", jti).Update("revoked_at", time.Now()).Error
}
func (r *oauth2Repo) RevokeAllClientTokens(ctx context.Context, clientID string) error {
	return r.db.WithContext(ctx).Model(&model.OAuth2Token{}).Where("client_id = ? AND revoked_at IS NULL", clientID).Update("revoked_at", time.Now()).Error
}
