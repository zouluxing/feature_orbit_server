package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/zouluxing/ums/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id int64) (*model.User, error)
	FindByUUID(ctx context.Context, uuid string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
	SoftDelete(ctx context.Context, uuid string) error
	List(ctx context.Context, email string, status *int8, offset, limit int) ([]model.User, int64, error)
	IncrFailedLogin(ctx context.Context, id int64) (int, error)
	ResetFailedLogin(ctx context.Context, id int64) error
	SetStatus(ctx context.Context, uuid string, status model.UserStatus) error
	UpdateLastLogin(ctx context.Context, id int64, at time.Time) error
}

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}
func (r *userRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Preload("Roles.Permissions").First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &u, err
}
func (r *userRepo) FindByUUID(ctx context.Context, uuid string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("uuid = ? AND deleted_at IS NULL", uuid).Preload("Roles.Permissions").First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &u, err
}
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).Preload("Roles.Permissions").First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &u, err
}
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &u, err
}
func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
func (r *userRepo) SoftDelete(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("uuid = ? AND deleted_at IS NULL", uuid).Update("deleted_at", time.Now()).Error
}
func (r *userRepo) List(ctx context.Context, email string, status *int8, offset, limit int) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("deleted_at IS NULL")
	if email != "" { q = q.Where("email ILIKE ?", "%"+email+"%") }
	if status != nil { q = q.Where("status = ?", *status) }
	var total int64
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	var users []model.User
	err := q.Preload("Roles").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}
func (r *userRepo) IncrFailedLogin(ctx context.Context, id int64) (int, error) {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).UpdateColumn("failed_login_count", gorm.Expr("failed_login_count + 1")).Error; err != nil { return 0, err }
	var u model.User
	if err := r.db.WithContext(ctx).Select("failed_login_count").First(&u, id).Error; err != nil { return 0, err }
	return u.FailedLoginCount, nil
}
func (r *userRepo) ResetFailedLogin(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("failed_login_count", 0).Error
}
func (r *userRepo) SetStatus(ctx context.Context, uuid string, status model.UserStatus) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("uuid = ? AND deleted_at IS NULL", uuid).Update("status", status).Error
}
func (r *userRepo) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login_at", at).Error
}
