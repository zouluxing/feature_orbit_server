package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/zouluxing/ums/internal/config"
	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/repository"
	apperr "github.com/zouluxing/ums/pkg/errors"
	"github.com/zouluxing/ums/pkg/utils"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	Logout(ctx context.Context, jti string) error
}

type authService struct {
	userRepo repository.UserRepository
	cache    CacheService
	jwt      *utils.JWTManager
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cache CacheService, jwt *utils.JWTManager, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cache: cache, jwt: jwt, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if existing, err := s.userRepo.FindByEmail(ctx, req.Email); err != nil {
		return nil, apperr.ErrInternal
	} else if existing != nil {
		return nil, apperr.ErrDuplicateUser
	}
	if existing, err := s.userRepo.FindByUsername(ctx, req.Username); err != nil {
		return nil, apperr.ErrInternal
	} else if existing != nil {
		return nil, apperr.ErrDuplicateUser
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	u := &model.User{UUID: utils.NewUUID(), Email: req.Email, Username: req.Username, PasswordHash: string(hash), Status: model.UserStatusActive}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, apperr.ErrInternal
	}
	return &dto.RegisterResponse{UserUUID: u.UUID, Email: u.Email}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	ipKey := fmt.Sprintf("ums:ratelimit:login:%s", req.Email)
	locked, err := s.cache.Exists(ctx, fmt.Sprintf("ums:lock:%s", req.Email))
	if err != nil {
		return nil, apperr.ErrInternal
	}
	if locked {
		return nil, apperr.ErrAccountLocked
	}
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	if user == nil {
		s.cache.Incr(ctx, ipKey, time.Duration(s.cfg.RateLimit.LoginWindowSecs)*time.Second)
		return nil, apperr.ErrInvalidCredentials
	}
	if user.Status == model.UserStatusLocked {
		return nil, apperr.ErrAccountLocked
	}
	if user.Status == model.UserStatusDisabled {
		return nil, apperr.ErrForbidden
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		count, _ := s.userRepo.IncrFailedLogin(ctx, user.ID)
		if count >= s.cfg.RateLimit.LoginMaxAttempts {
			s.userRepo.SetStatus(ctx, user.UUID, model.UserStatusLocked)
			s.cache.Set(ctx, fmt.Sprintf("ums:lock:%s", req.Email), "1", time.Duration(s.cfg.RateLimit.LockDurationMins)*time.Minute)
			return nil, apperr.ErrAccountLocked
		}
		return nil, apperr.ErrInvalidCredentials
	}
	if user.MFAEnabled {
		if req.MFACode == "" { return nil, apperr.ErrMFARequired }
		if !verifyTOTP(user.MFASecret, req.MFACode) { return nil, apperr.ErrMFACodeInvalid }
	}
	s.userRepo.ResetFailedLogin(ctx, user.ID)
	s.cache.Del(ctx, ipKey)
	s.userRepo.UpdateLastLogin(ctx, user.ID, time.Now())
	return s.issueTokenPair(ctx, user, nil)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	rhash := utils.SHA256Base64(refreshToken)
	revoked, _ := s.cache.Exists(ctx, "ums:revoked:refresh:"+rhash)
	if revoked { return nil, apperr.ErrTokenInvalid }
	claims, err := s.jwt.VerifyAccessToken(refreshToken)
	if err != nil { return nil, apperr.ErrTokenInvalid }
	user, err := s.userRepo.FindByUUID(ctx, claims.Subject)
	if err != nil || user == nil { return nil, apperr.ErrUserNotFound }
	if user.Status != model.UserStatusActive { return nil, apperr.ErrForbidden }
	s.cache.Set(ctx, "ums:revoked:refresh:"+rhash, "1", s.cfg.JWT.RefreshTTL)
	return s.issueTokenPair(ctx, user, nil)
}

func (s *authService) Logout(ctx context.Context, jti string) error {
	return s.cache.Set(ctx, "ums:blacklist:"+jti, "1", s.cfg.JWT.AccessTTL+time.Minute)
}

func (s *authService) issueTokenPair(_ context.Context, user *model.User, scopes []string) (*dto.TokenResponse, error) {
	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles { roles = append(roles, r.Name) }
	accessToken, _, err := s.jwt.IssueAccessToken(user.UUID, user.Email, roles, scopes)
	if err != nil { return nil, apperr.ErrInternal }
	refreshToken, _, err := s.jwt.IssueAccessToken(user.UUID, user.Email, roles, scopes)
	if err != nil { return nil, apperr.ErrInternal }
	return &dto.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int(s.cfg.JWT.AccessTTL.Seconds()), TokenType: "Bearer"}, nil
}

func verifyTOTP(secret *string, code string) bool {
	if secret == nil || *secret == "" { return false }
	return totp.Validate(code, *secret)
}
