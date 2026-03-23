package service

import (
	"context"
	"fmt"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/repository"
	apperr "github.com/zouluxing/ums/pkg/errors"
	"github.com/zouluxing/ums/pkg/utils"
)

// MFAEncryptor is satisfied by pkg/crypto.AESEncryptor.
// Defined as an interface here to keep the service layer testable without
// importing the crypto package directly (avoids test key management).
type MFAEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type UserService interface {
	GetByUUID(ctx context.Context, uuid string) (*dto.UserInfo, error)
	List(ctx context.Context, req dto.UserListRequest) (*dto.UserListResponse, error)
	Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserInfo, error)
	Update(ctx context.Context, uuid string, req dto.UpdateUserRequest) (*dto.UserInfo, error)
	UpdatePassword(ctx context.Context, uuid, oldPass, newPass string) error
	ResetPassword(ctx context.Context, uuid, newPass string) error
	SetStatus(ctx context.Context, uuid string, status int8) error
	Delete(ctx context.Context, uuid string) error
	AssignRoles(ctx context.Context, uuid string, roleIDs []int64) error
	RemoveRole(ctx context.Context, uuid string, roleID int64) error
	GetPermissions(ctx context.Context, uuid string) ([]dto.PermissionInfo, error)
	EnableMFA(ctx context.Context, uuid string) (*dto.EnableMFAResponse, error)
	DisableMFA(ctx context.Context, uuid, code string) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	// enc is optional: when nil, MFA secrets are stored in plaintext (dev-only).
	// Production deployments MUST set UMS_MFA_ENCRYPTION_KEY.
	enc MFAEncryptor
}

// NewUserService creates a UserService.
// Pass a non-nil enc (pkg/crypto.AESEncryptor) in production to encrypt MFA secrets at rest.
// Passing nil is accepted only for development; a warning is logged at startup by main.go.
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, enc MFAEncryptor) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo, enc: enc}
}

func (s *userService) GetByUUID(ctx context.Context, uuid string) (*dto.UserInfo, error) {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil { return nil, apperr.ErrInternal }
	if u == nil { return nil, apperr.ErrUserNotFound }
	return toUserInfo(u), nil
}
func (s *userService) List(ctx context.Context, req dto.UserListRequest) (*dto.UserListResponse, error) {
	page := utils.DefaultInt(req.Page, 1)
	size := utils.DefaultInt(req.PageSize, 20)
	users, total, err := s.userRepo.List(ctx, req.Email, req.Status, utils.Offset(page, size), size)
	if err != nil { return nil, apperr.ErrInternal }
	list := make([]dto.UserInfo, 0, len(users))
	for i := range users { list = append(list, *toUserInfo(&users[i])) }
	return &dto.UserListResponse{List: list, Total: total, Page: page, PageSize: size}, nil
}
func (s *userService) Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserInfo, error) {
	if u, _ := s.userRepo.FindByEmail(ctx, req.Email); u != nil { return nil, apperr.ErrDuplicateUser }
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	u := &model.User{UUID: utils.NewUUID(), Email: req.Email, Username: req.Username, PasswordHash: string(hash), Status: model.UserStatusActive}
	if err := s.userRepo.Create(ctx, u); err != nil { return nil, apperr.ErrInternal }
	if len(req.RoleIDs) > 0 { s.roleRepo.AssignToUser(ctx, u.ID, req.RoleIDs) }
	u, _ = s.userRepo.FindByUUID(ctx, u.UUID)
	return toUserInfo(u), nil
}
func (s *userService) Update(ctx context.Context, uuid string, req dto.UpdateUserRequest) (*dto.UserInfo, error) {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return nil, apperr.ErrUserNotFound }
	if req.Username != nil { u.Username = *req.Username }
	if req.Phone != nil { u.Phone = req.Phone }
	if req.AvatarURL != nil { u.AvatarURL = req.AvatarURL }
	if err := s.userRepo.Update(ctx, u); err != nil { return nil, apperr.ErrInternal }
	return toUserInfo(u), nil
}
func (s *userService) UpdatePassword(ctx context.Context, uuid, oldPass, newPass string) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPass)); err != nil { return apperr.ErrInvalidCredentials }
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), 12)
	u.PasswordHash = string(hash)
	return s.userRepo.Update(ctx, u)
}
func (s *userService) ResetPassword(ctx context.Context, uuid, newPass string) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), 12)
	u.PasswordHash = string(hash)
	return s.userRepo.Update(ctx, u)
}
func (s *userService) SetStatus(ctx context.Context, uuid string, status int8) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	return s.userRepo.SetStatus(ctx, uuid, model.UserStatus(status))
}
func (s *userService) Delete(ctx context.Context, uuid string) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	return s.userRepo.SoftDelete(ctx, uuid)
}
func (s *userService) AssignRoles(ctx context.Context, uuid string, roleIDs []int64) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	return s.roleRepo.AssignToUser(ctx, u.ID, roleIDs)
}
func (s *userService) RemoveRole(ctx context.Context, uuid string, roleID int64) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	return s.roleRepo.RemoveFromUser(ctx, u.ID, roleID)
}
func (s *userService) GetPermissions(ctx context.Context, uuid string) ([]dto.PermissionInfo, error) {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return nil, apperr.ErrUserNotFound }
	seen := map[int64]bool{}
	var result []dto.PermissionInfo
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if !seen[perm.ID] { seen[perm.ID] = true; result = append(result, dto.PermissionInfo{ID: perm.ID, Resource: perm.Resource, Action: perm.Action}) }
		}
	}
	return result, nil
}

// EnableMFA generates a TOTP secret, encrypts it with AES-256-GCM (if enc
// is configured), persists the ciphertext, and returns the provisioning URI.
func (s *userService) EnableMFA(ctx context.Context, uuid string) (*dto.EnableMFAResponse, error) {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return nil, apperr.ErrUserNotFound }
	if u.MFAEnabled { return nil, apperr.ErrMFAAlreadyEnabled }

	key, err := totp.Generate(totp.GenerateOpts{Issuer: "UMS", AccountName: u.Email})
	if err != nil { return nil, apperr.ErrInternal }

	plainSecret := key.Secret()
	storedSecret := plainSecret

	// Encrypt the TOTP secret at rest when an encryptor is available.
	if s.enc != nil {
		encrypted, err := s.enc.Encrypt(plainSecret)
		if err != nil { return nil, apperr.ErrInternal }
		storedSecret = encrypted
	}

	u.MFASecret = &storedSecret
	if err := s.userRepo.Update(ctx, u); err != nil { return nil, apperr.ErrInternal }

	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", key.URL())
	return &dto.EnableMFAResponse{Secret: plainSecret, OTPAuthURI: key.URL(), QRCodeURL: qrURL}, nil
}

// DisableMFA verifies the TOTP code against the (decrypted) secret, then
// clears the MFA fields.
func (s *userService) DisableMFA(ctx context.Context, uuid, code string) error {
	u, err := s.userRepo.FindByUUID(ctx, uuid)
	if err != nil || u == nil { return apperr.ErrUserNotFound }
	if !u.MFAEnabled || u.MFASecret == nil { return apperr.ErrMFACodeInvalid }

	secret := *u.MFASecret
	// Decrypt if an encryptor is configured.
	if s.enc != nil {
		decrypted, err := s.enc.Decrypt(secret)
		if err != nil { return apperr.ErrInternal }
		secret = decrypted
	}

	if !totp.Validate(code, secret) { return apperr.ErrMFACodeInvalid }
	u.MFAEnabled = false
	u.MFASecret = nil
	return s.userRepo.Update(ctx, u)
}

func toUserInfo(u *model.User) *dto.UserInfo {
	roles := make([]string, 0, len(u.Roles))
	var perms []string
	seen := map[string]bool{}
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
		for _, p := range r.Permissions {
			key := p.Resource + ":" + p.Action
			if !seen[key] { seen[key] = true; perms = append(perms, key) }
		}
	}
	return &dto.UserInfo{UUID: u.UUID, Email: u.Email, Username: u.Username, Phone: u.Phone,
		AvatarURL: u.AvatarURL, Status: int8(u.Status), MFAEnabled: u.MFAEnabled, Roles: roles, Permissions: perms}
}
