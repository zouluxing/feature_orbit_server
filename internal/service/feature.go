package service

import (
	"context"
	"fmt"

	"github.com/zouluxing/feature_orbit_server/internal/dto"
	"github.com/zouluxing/feature_orbit_server/internal/model"
	"github.com/zouluxing/feature_orbit_server/internal/repository"
)

// ErrFeatureNotFound is returned when a feature lookup finds no row.
var ErrFeatureNotFound = fmt.Errorf("feature not found")

// ErrDuplicateSlug is returned when the slug is already taken.
var ErrDuplicateSlug = fmt.Errorf("feature slug already exists")

// ErrForbidden is returned when the caller does not own the resource.
var ErrForbidden = fmt.Errorf("forbidden: not the owner")

// FeatureService manages Feature business logic.
type FeatureService interface {
	List(ctx context.Context, req dto.FeatureListRequest) (*dto.FeatureListResponse, error)
	GetByID(ctx context.Context, id uint64) (*dto.FeatureInfo, error)
	Create(ctx context.Context, ownerUUID, ownerEmail string, req dto.CreateFeatureRequest) (*dto.FeatureInfo, error)
	Update(ctx context.Context, id uint64, callerUUID string, req dto.UpdateFeatureRequest) (*dto.FeatureInfo, error)
	Delete(ctx context.Context, id uint64, callerUUID string) error
}

type featureService struct {
	repo repository.FeatureRepository
}

// NewFeatureService constructs a FeatureService.
func NewFeatureService(repo repository.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

func (s *featureService) List(ctx context.Context, req dto.FeatureListRequest) (*dto.FeatureListResponse, error) {
	page := defaultInt(req.Page, 1)
	size := defaultInt(req.PageSize, 20)
	features, total, err := s.repo.List(ctx, req.OwnerUUID, req.Status, (page-1)*size, size)
	if err != nil {
		return nil, fmt.Errorf("feature list: %w", err)
	}
	list := make([]dto.FeatureInfo, 0, len(features))
	for i := range features {
		list = append(list, dto.ToFeatureInfo(&features[i]))
	}
	return &dto.FeatureListResponse{List: list, Total: total, Page: page, PageSize: size}, nil
}

func (s *featureService) GetByID(ctx context.Context, id uint64) (*dto.FeatureInfo, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature get: %w", err)
	}
	if f == nil {
		return nil, ErrFeatureNotFound
	}
	info := dto.ToFeatureInfo(f)
	return &info, nil
}

func (s *featureService) Create(ctx context.Context, ownerUUID, ownerEmail string, req dto.CreateFeatureRequest) (*dto.FeatureInfo, error) {
	// Guard: slug must be unique
	if existing, err := s.repo.FindBySlug(ctx, req.Slug); err != nil {
		return nil, fmt.Errorf("feature create: check slug: %w", err)
	} else if existing != nil {
		return nil, ErrDuplicateSlug
	}
	f := &model.Feature{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		Status:      model.FeatureStatusDraft,
		OwnerUUID:   ownerUUID,
		OwnerEmail:  ownerEmail,
		Tags:        "[]",
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("feature create: %w", err)
	}
	info := dto.ToFeatureInfo(f)
	return &info, nil
}

func (s *featureService) Update(ctx context.Context, id uint64, callerUUID string, req dto.UpdateFeatureRequest) (*dto.FeatureInfo, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature update: %w", err)
	}
	if f == nil {
		return nil, ErrFeatureNotFound
	}
	if f.OwnerUUID != callerUUID {
		return nil, ErrForbidden
	}
	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Description != nil {
		f.Description = *req.Description
	}
	if req.Status != nil {
		f.Status = *req.Status
	}
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, fmt.Errorf("feature update save: %w", err)
	}
	info := dto.ToFeatureInfo(f)
	return &info, nil
}

func (s *featureService) Delete(ctx context.Context, id uint64, callerUUID string) error {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("feature delete: %w", err)
	}
	if f == nil {
		return ErrFeatureNotFound
	}
	if f.OwnerUUID != callerUUID {
		return ErrForbidden
	}
	return s.repo.SoftDelete(ctx, id)
}

func defaultInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
