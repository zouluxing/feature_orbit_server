package dto

import (
	"time"

	"github.com/zouluxing/feature_orbit_server/internal/model"
)

type FeatureInfo struct {
	ID          uint64              `json:"id"`
	Slug        string              `json:"slug"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Status      model.FeatureStatus `json:"status"`
	OwnerUUID   string              `json:"owner_uuid"`
	OwnerEmail  string              `json:"owner_email"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type CreateFeatureRequest struct {
	Slug        string `json:"slug"        binding:"required,min=2,max=128,alphanum"`
	Name        string `json:"name"        binding:"required,min=2,max=256"`
	Description string `json:"description" binding:"max=2048"`
}

type UpdateFeatureRequest struct {
	Name        *string              `json:"name"        binding:"omitempty,min=2,max=256"`
	Description *string              `json:"description" binding:"omitempty,max=2048"`
	Status      *model.FeatureStatus `json:"status"      binding:"omitempty,oneof=draft active archived"`
}

type FeatureListRequest struct {
	OwnerUUID string              `form:"owner_uuid"`
	Status    model.FeatureStatus `form:"status"`
	Page      int                 `form:"page"      binding:"min=1"`
	PageSize  int                 `form:"page_size" binding:"min=1,max=100"`
}

type FeatureListResponse struct {
	List     []FeatureInfo `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func ToFeatureInfo(f *model.Feature) FeatureInfo {
	return FeatureInfo{ID: f.ID, Slug: f.Slug, Name: f.Name, Description: f.Description,
		Status: f.Status, OwnerUUID: f.OwnerUUID, OwnerEmail: f.OwnerEmail,
		CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt}
}
