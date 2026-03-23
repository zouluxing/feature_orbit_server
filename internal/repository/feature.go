package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/zouluxing/feature_orbit_server/internal/model"
)

type FeatureRepository interface {
	Create(ctx context.Context, f *model.Feature) error
	FindByID(ctx context.Context, id uint64) (*model.Feature, error)
	FindBySlug(ctx context.Context, slug string) (*model.Feature, error)
	List(ctx context.Context, ownerUUID string, status model.FeatureStatus, offset, limit int) ([]model.Feature, int64, error)
	Update(ctx context.Context, f *model.Feature) error
	SoftDelete(ctx context.Context, id uint64) error
}

type featureRepo struct{ db *gorm.DB }

func NewFeatureRepository(db *gorm.DB) FeatureRepository { return &featureRepo{db: db} }

func (r *featureRepo) Create(ctx context.Context, f *model.Feature) error {
	return r.db.WithContext(ctx).Create(f).Error
}
func (r *featureRepo) FindByID(ctx context.Context, id uint64) (*model.Feature, error) {
	var f model.Feature
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &f, err
}
func (r *featureRepo) FindBySlug(ctx context.Context, slug string) (*model.Feature, error) {
	var f model.Feature
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	return &f, err
}
func (r *featureRepo) List(ctx context.Context, ownerUUID string, status model.FeatureStatus, offset, limit int) ([]model.Feature, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Feature{}).Where("deleted_at IS NULL")
	if ownerUUID != "" { q = q.Where("owner_uuid = ?", ownerUUID) }
	if status != "" { q = q.Where("status = ?", status) }
	var total int64
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	var features []model.Feature
	err := q.Offset(offset).Limit(limit).Order("created_at DESC").Find(&features).Error
	return features, total, err
}
func (r *featureRepo) Update(ctx context.Context, f *model.Feature) error {
	return r.db.WithContext(ctx).Save(f).Error
}
func (r *featureRepo) SoftDelete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Feature{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error
}
