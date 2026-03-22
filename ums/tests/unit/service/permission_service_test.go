package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/service"
	apperr "github.com/zouluxing/ums/pkg/errors"
)

type mockPermRepo struct{ mock.Mock }

func (m *mockPermRepo) FindAll(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Permission), args.Error(1)
}
func (m *mockPermRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.Permission, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func TestListRoles(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	desc := "admin role"
	rr.On("FindAll", mock.Anything).Return([]model.Role{
		{ID: 1, Name: "admin", Description: &desc, IsSystem: true},
		{ID: 2, Name: "viewer"},
	}, nil)
	roles, err := service.NewPermissionService(rr, pr, ur).ListRoles(context.Background())
	require.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Name)
	assert.True(t, roles[0].IsSystem)
}

func TestCreateRole_Success(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	rr.On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)
	role, err := service.NewPermissionService(rr, pr, ur).CreateRole(context.Background(), dto.CreateRoleRequest{Name: "moderator"})
	require.NoError(t, err)
	assert.Equal(t, "moderator", role.Name)
}

func TestDeleteRole_SystemRoleBlocked(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	rr.On("FindByID", mock.Anything, int64(1)).Return(&model.Role{ID: 1, Name: "admin", IsSystem: true}, nil)
	err := service.NewPermissionService(rr, pr, ur).DeleteRole(context.Background(), 1)
	assert.ErrorIs(t, err, apperr.ErrSystemRole)
}

func TestDeleteRole_Success(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	rr.On("FindByID", mock.Anything, int64(5)).Return(&model.Role{ID: 5, Name: "custom", IsSystem: false}, nil)
	rr.On("Delete", mock.Anything, int64(5)).Return(nil)
	require.NoError(t, service.NewPermissionService(rr, pr, ur).DeleteRole(context.Background(), 5))
}

func TestSetRolePermissions(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	rr.On("FindByID", mock.Anything, int64(2)).Return(&model.Role{ID: 2, Name: "editor"}, nil)
	rr.On("SetPermissions", mock.Anything, int64(2), []int64{1, 3, 5}).Return(nil)
	err := service.NewPermissionService(rr, pr, ur).SetRolePermissions(context.Background(), 2, dto.AssignPermissionsRequest{PermissionIDs: []int64{1, 3, 5}})
	require.NoError(t, err)
}

func TestListPermissions(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	pr.On("FindAll", mock.Anything).Return([]model.Permission{
		{ID: 1, Resource: "user", Action: "read"},
		{ID: 2, Resource: "user", Action: "write"},
	}, nil)
	perms, err := service.NewPermissionService(rr, pr, ur).ListPermissions(context.Background())
	require.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestCheckPermission_Granted(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	u := &model.User{UUID: "u1", Roles: []model.Role{{Name: "editor", Permissions: []model.Permission{{Resource: "article", Action: "write"}}}}}
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	ok, err := service.NewPermissionService(rr, pr, ur).CheckPermission(context.Background(), "u1", "article", "write")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckPermission_Denied(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	u := &model.User{UUID: "u1", Roles: []model.Role{{Name: "viewer", Permissions: []model.Permission{{Resource: "user", Action: "read"}}}}}
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	ok, err := service.NewPermissionService(rr, pr, ur).CheckPermission(context.Background(), "u1", "user", "delete")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckPermission_UserNotFound(t *testing.T) {
	rr, pr, ur := new(mockRoleRepo), new(mockPermRepo), new(mockUserRepo)
	ur.On("FindByUUID", mock.Anything, "ghost").Return(nil, nil)
	_, err := service.NewPermissionService(rr, pr, ur).CheckPermission(context.Background(), "ghost", "user", "read")
	assert.ErrorIs(t, err, apperr.ErrUserNotFound)
}
