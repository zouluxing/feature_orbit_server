package service

import (
	"context"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/repository"
	apperr "github.com/zouluxing/ums/pkg/errors"
)

type PermissionService interface {
	ListRoles(ctx context.Context) ([]dto.RoleInfo, error)
	CreateRole(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleInfo, error)
	DeleteRole(ctx context.Context, id int64) error
	SetRolePermissions(ctx context.Context, roleID int64, req dto.AssignPermissionsRequest) error
	ListPermissions(ctx context.Context) ([]dto.PermissionInfo, error)
	CheckPermission(ctx context.Context, userUUID, resource, action string) (bool, error)
}

type permissionService struct {
	roleRepo repository.RoleRepository
	permRepo repository.PermissionRepository
	userRepo repository.UserRepository
}

func NewPermissionService(roleRepo repository.RoleRepository, permRepo repository.PermissionRepository, userRepo repository.UserRepository) PermissionService {
	return &permissionService{roleRepo: roleRepo, permRepo: permRepo, userRepo: userRepo}
}

func (s *permissionService) ListRoles(ctx context.Context) ([]dto.RoleInfo, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil { return nil, apperr.ErrInternal }
	result := make([]dto.RoleInfo, 0, len(roles))
	for _, r := range roles {
		desc := ""; if r.Description != nil { desc = *r.Description }
		result = append(result, dto.RoleInfo{ID: r.ID, Name: r.Name, Description: desc, IsSystem: r.IsSystem})
	}
	return result, nil
}
func (s *permissionService) CreateRole(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleInfo, error) {
	role := &model.Role{Name: req.Name}
	if req.Description != "" { role.Description = &req.Description }
	if err := s.roleRepo.Create(ctx, role); err != nil { return nil, apperr.ErrInternal }
	return &dto.RoleInfo{ID: role.ID, Name: role.Name}, nil
}
func (s *permissionService) DeleteRole(ctx context.Context, id int64) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil || role == nil { return apperr.ErrRoleNotFound }
	if role.IsSystem { return apperr.ErrSystemRole }
	return s.roleRepo.Delete(ctx, id)
}
func (s *permissionService) SetRolePermissions(ctx context.Context, roleID int64, req dto.AssignPermissionsRequest) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil || role == nil { return apperr.ErrRoleNotFound }
	return s.roleRepo.SetPermissions(ctx, roleID, req.PermissionIDs)
}
func (s *permissionService) ListPermissions(ctx context.Context) ([]dto.PermissionInfo, error) {
	perms, err := s.permRepo.FindAll(ctx)
	if err != nil { return nil, apperr.ErrInternal }
	result := make([]dto.PermissionInfo, 0, len(perms))
	for _, p := range perms { result = append(result, dto.PermissionInfo{ID: p.ID, Resource: p.Resource, Action: p.Action}) }
	return result, nil
}
func (s *permissionService) CheckPermission(ctx context.Context, userUUID, resource, action string) (bool, error) {
	u, err := s.userRepo.FindByUUID(ctx, userUUID)
	if err != nil || u == nil { return false, apperr.ErrUserNotFound }
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action { return true, nil }
		}
	}
	return false, nil
}
