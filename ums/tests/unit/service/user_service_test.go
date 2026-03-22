package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/service"
	apperr "github.com/zouluxing/ums/pkg/errors"
	"github.com/zouluxing/ums/pkg/utils"
)

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) FindAll(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx); return args.Get(0).([]model.Role), args.Error(1)
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	args := m.Called(ctx, id)
	if r, ok := args.Get(0).(*model.Role); ok { return r, args.Error(1) }
	return nil, args.Error(1)
}
func (m *mockRoleRepo) Create(ctx context.Context, r *model.Role) error { return m.Called(ctx, r).Error(0) }
func (m *mockRoleRepo) Delete(ctx context.Context, id int64) error { return m.Called(ctx, id).Error(0) }
func (m *mockRoleRepo) AssignToUser(ctx context.Context, userID int64, roleIDs []int64) error {
	return m.Called(ctx, userID, roleIDs).Error(0)
}
func (m *mockRoleRepo) RemoveFromUser(ctx context.Context, userID int64, roleID int64) error {
	return m.Called(ctx, userID, roleID).Error(0)
}
func (m *mockRoleRepo) SetPermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	return m.Called(ctx, roleID, permIDs).Error(0)
}

func testBcryptHash(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

func userWithHash(t *testing.T, uuid, plain string) *model.User {
	return &model.User{UUID: uuid, Email: uuid + "@x.com", Username: uuid,
		PasswordHash: testBcryptHash(t, plain), Status: model.UserStatusActive}
}

func TestGetByUUID_Found(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	u := &model.User{UUID: utils.NewUUID(), Email: "alice@x.com", Username: "alice",
		Status: model.UserStatusActive, Roles: []model.Role{{Name: "viewer"}}}
	ur.On("FindByUUID", mock.Anything, u.UUID).Return(u, nil)
	info, err := service.NewUserService(ur, rr).GetByUUID(context.Background(), u.UUID)
	require.NoError(t, err)
	assert.Equal(t, u.Email, info.Email)
	assert.Equal(t, []string{"viewer"}, info.Roles)
}

func TestGetByUUID_NotFound(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "ghost").Return(nil, nil)
	_, err := service.NewUserService(ur, rr).GetByUUID(context.Background(), "ghost")
	assert.ErrorIs(t, err, apperr.ErrUserNotFound)
}

func TestCreate_Success(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	created := &model.User{UUID: "uuid-bob", Email: "bob@x.com", Username: "bob", Status: model.UserStatusActive}
	ur.On("FindByEmail", mock.Anything, "bob@x.com").Return(nil, nil)
	ur.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	ur.On("FindByUUID", mock.Anything, mock.Anything).Return(created, nil)
	info, err := service.NewUserService(ur, rr).Create(context.Background(), dto.CreateUserRequest{Email: "bob@x.com", Username: "bob", Password: "pass12345"})
	require.NoError(t, err)
	assert.Equal(t, "bob@x.com", info.Email)
}

func TestCreate_DuplicateEmail(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByEmail", mock.Anything, "alice@x.com").Return(&model.User{}, nil)
	_, err := service.NewUserService(ur, rr).Create(context.Background(), dto.CreateUserRequest{Email: "alice@x.com", Username: "alice2", Password: "pass12345"})
	assert.ErrorIs(t, err, apperr.ErrDuplicateUser)
}

func TestCreate_WithRoleAssignment(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByEmail", mock.Anything, "carol@x.com").Return(nil, nil)
	ur.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(1).(*model.User); u.ID = 99; u.UUID = "uuid-99"
	})
	rr.On("AssignToUser", mock.Anything, int64(99), []int64{1, 2}).Return(nil)
	ur.On("FindByUUID", mock.Anything, "uuid-99").Return(&model.User{UUID: "uuid-99", Email: "carol@x.com", Username: "carol", Status: model.UserStatusActive}, nil)
	info, err := service.NewUserService(ur, rr).Create(context.Background(), dto.CreateUserRequest{Email: "carol@x.com", Username: "carol", Password: "pass12345", RoleIDs: []int64{1, 2}})
	require.NoError(t, err)
	assert.Equal(t, "carol@x.com", info.Email)
}

func TestUpdate_Success(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	u := &model.User{UUID: "u1", Email: "a@x.com", Username: "old", Status: model.UserStatusActive}
	newName := "new_name"
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	ur.On("Update", mock.Anything, mock.MatchedBy(func(u *model.User) bool { return u.Username == "new_name" })).Return(nil)
	info, err := service.NewUserService(ur, rr).Update(context.Background(), "u1", dto.UpdateUserRequest{Username: &newName})
	require.NoError(t, err)
	assert.Equal(t, "new_name", info.Username)
}

func TestUpdatePassword_Success(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	u := userWithHash(t, "u1", "old-pass")
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	ur.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := service.NewUserService(ur, rr).UpdatePassword(context.Background(), "u1", "old-pass", "new-pass-123")
	require.NoError(t, err)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("new-pass-123")))
}

func TestUpdatePassword_WrongOldPassword(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "u1").Return(userWithHash(t, "u1", "correct"), nil)
	err := service.NewUserService(ur, rr).UpdatePassword(context.Background(), "u1", "wrong", "new")
	assert.ErrorIs(t, err, apperr.ErrInvalidCredentials)
}

func TestResetPassword_Success(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	u := userWithHash(t, "u1", "old")
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	ur.On("Update", mock.Anything, mock.Anything).Return(nil)
	require.NoError(t, service.NewUserService(ur, rr).ResetPassword(context.Background(), "u1", "brand-new"))
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("brand-new")))
}

func TestSetStatus(t *testing.T) {
	for _, tc := range []struct{ name string; status int8 }{{"disable", 2}, {"lock", 3}, {"enable", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			ur, rr := new(mockUserRepo), new(mockRoleRepo)
			ur.On("FindByUUID", mock.Anything, "u1").Return(&model.User{UUID: "u1"}, nil)
			ur.On("SetStatus", mock.Anything, "u1", model.UserStatus(tc.status)).Return(nil)
			require.NoError(t, service.NewUserService(ur, rr).SetStatus(context.Background(), "u1", tc.status))
		})
	}
}

func TestDelete_SoftDeletes(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "u1").Return(&model.User{UUID: "u1"}, nil)
	ur.On("SoftDelete", mock.Anything, "u1").Return(nil)
	require.NoError(t, service.NewUserService(ur, rr).Delete(context.Background(), "u1"))
	ur.AssertCalled(t, "SoftDelete", mock.Anything, "u1")
}

func TestDelete_NotFound(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "ghost").Return(nil, nil)
	assert.ErrorIs(t, service.NewUserService(ur, rr).Delete(context.Background(), "ghost"), apperr.ErrUserNotFound)
}

func TestGetPermissions_Deduplicates(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	shared := model.Permission{ID: 1, Resource: "user", Action: "read"}
	u := &model.User{UUID: "u1", Roles: []model.Role{
		{Name: "editor", Permissions: []model.Permission{shared, {ID: 2, Resource: "article", Action: "write"}}},
		{Name: "viewer", Permissions: []model.Permission{shared}},
	}}
	ur.On("FindByUUID", mock.Anything, "u1").Return(u, nil)
	perms, err := service.NewUserService(ur, rr).GetPermissions(context.Background(), "u1")
	require.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestAssignRoles(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "u10").Return(&model.User{ID: 10, UUID: "u10"}, nil)
	rr.On("AssignToUser", mock.Anything, int64(10), []int64{1, 3}).Return(nil)
	require.NoError(t, service.NewUserService(ur, rr).AssignRoles(context.Background(), "u10", []int64{1, 3}))
}

func TestRemoveRole(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	ur.On("FindByUUID", mock.Anything, "u10").Return(&model.User{ID: 10, UUID: "u10"}, nil)
	rr.On("RemoveFromUser", mock.Anything, int64(10), int64(3)).Return(nil)
	require.NoError(t, service.NewUserService(ur, rr).RemoveRole(context.Background(), "u10", 3))
}

func TestList_DefaultPagination(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	users := []model.User{{UUID: "u1", Email: "a@x.com", Username: "a"}, {UUID: "u2", Email: "b@x.com", Username: "b"}}
	ur.On("List", mock.Anything, "", (*int8)(nil), 0, 20).Return(users, int64(2), nil)
	res, err := service.NewUserService(ur, rr).List(context.Background(), dto.UserListRequest{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
	assert.Len(t, res.List, 2)
}

func TestList_FilterByEmail(t *testing.T) {
	ur, rr := new(mockUserRepo), new(mockRoleRepo)
	users := []model.User{{UUID: "u1", Email: "alice@x.com", Username: "alice"}}
	ur.On("List", mock.Anything, "alice", (*int8)(nil), 0, 10).Return(users, int64(1), nil)
	res, err := service.NewUserService(ur, rr).List(context.Background(), dto.UserListRequest{Email: "alice", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
}
