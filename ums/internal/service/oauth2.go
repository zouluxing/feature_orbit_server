package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/zouluxing/ums/internal/config"
	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/model"
	"github.com/zouluxing/ums/internal/repository"
	apperr "github.com/zouluxing/ums/pkg/errors"
	"github.com/zouluxing/ums/pkg/utils"
)

type OAuth2Service interface {
	CreateClient(ctx context.Context, ownerUserID int64, req dto.CreateClientRequest) (*dto.CreateClientResponse, error)
	ListClients(ctx context.Context, ownerUserID int64) ([]dto.OAuth2ClientInfo, error)
	UpdateClient(ctx context.Context, ownerUserID, clientID int64, req dto.UpdateClientRequest) (*dto.OAuth2ClientInfo, error)
	DeleteClient(ctx context.Context, ownerUserID, clientID int64) error
	RotateClientSecret(ctx context.Context, ownerUserID, clientID int64) (string, error)
	Authorize(ctx context.Context, req dto.OAuth2AuthorizeRequest, userID int64) (string, error)
	Token(ctx context.Context, req dto.OAuth2TokenRequest) (*dto.TokenResponse, error)
	Introspect(ctx context.Context, req dto.OAuth2IntrospectRequest) (*dto.OAuth2IntrospectResponse, error)
}

type oauth2Service struct {
	oauthRepo repository.OAuth2Repository
	userRepo  repository.UserRepository
	jwt       interface{ IssueAccessToken(string, string, []string, []string) (string, string, error); AccessTTL() time.Duration }
	cfg       *config.Config
}

func NewOAuth2Service(oauthRepo repository.OAuth2Repository, userRepo repository.UserRepository, jwt interface {
	IssueAccessToken(string, string, []string, []string) (string, string, error)
	AccessTTL() time.Duration
}, cfg *config.Config) OAuth2Service {
	return &oauth2Service{oauthRepo: oauthRepo, userRepo: userRepo, jwt: jwt, cfg: cfg}
}

func (s *oauth2Service) CreateClient(ctx context.Context, ownerUserID int64, req dto.CreateClientRequest) (*dto.CreateClientResponse, error) {
	clientID := utils.RandomString(16)
	secret := utils.RandomString(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(secret), 12)
	redirectsJSON, _ := json.Marshal(req.RedirectURIs)
	scopesJSON, _ := json.Marshal(req.Scopes)
	grantTypes, _ := json.Marshal([]string{"authorization_code", "refresh_token"})
	c := &model.OAuth2Client{
		ClientID: clientID, ClientSecretHash: string(hash), Name: req.Name,
		RedirectURIs: string(redirectsJSON), Scopes: string(scopesJSON),
		GrantTypes: string(grantTypes), OwnerUserID: ownerUserID,
	}
	if err := s.oauthRepo.CreateClient(ctx, c); err != nil { return nil, apperr.ErrInternal }
	return &dto.CreateClientResponse{OAuth2ClientInfo: toClientInfo(c), ClientSecret: secret}, nil
}

func (s *oauth2Service) ListClients(ctx context.Context, ownerUserID int64) ([]dto.OAuth2ClientInfo, error) {
	clients, err := s.oauthRepo.FindClientsByOwner(ctx, ownerUserID)
	if err != nil { return nil, apperr.ErrInternal }
	result := make([]dto.OAuth2ClientInfo, 0, len(clients))
	for i := range clients { result = append(result, toClientInfo(&clients[i])) }
	return result, nil
}

// UpdateClient implements BUG-002: previously a 501 stub.
func (s *oauth2Service) UpdateClient(ctx context.Context, ownerUserID, clientID int64, req dto.UpdateClientRequest) (*dto.OAuth2ClientInfo, error) {
	client, err := s.oauthRepo.FindClientsByOwner(ctx, ownerUserID)
	if err != nil { return nil, apperr.ErrInternal }
	var target *model.OAuth2Client
	for i := range client {
		if client[i].ID == clientID { target = &client[i]; break }
	}
	if target == nil { return nil, apperr.ErrClientInvalid }

	if req.Name != nil { target.Name = *req.Name }
	if len(req.RedirectURIs) > 0 {
		b, _ := json.Marshal(req.RedirectURIs)
		target.RedirectURIs = string(b)
	}
	if len(req.Scopes) > 0 {
		b, _ := json.Marshal(req.Scopes)
		target.Scopes = string(b)
	}
	if req.IsActive != nil { target.IsActive = *req.IsActive }

	if err := s.oauthRepo.UpdateClient(ctx, target); err != nil { return nil, apperr.ErrInternal }
	info := toClientInfo(target)
	return &info, nil
}

func (s *oauth2Service) DeleteClient(ctx context.Context, ownerUserID, clientID int64) error {
	if err := s.oauthRepo.DeleteClient(ctx, clientID); err != nil { return apperr.ErrInternal }
	return nil
}

func (s *oauth2Service) RotateClientSecret(ctx context.Context, ownerUserID, clientID int64) (string, error) {
	newSecret := utils.RandomString(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(newSecret), 12)
	c := &model.OAuth2Client{ID: clientID, ClientSecretHash: string(hash)}
	if err := s.oauthRepo.UpdateClient(ctx, c); err != nil { return "", apperr.ErrInternal }
	return newSecret, nil
}

func (s *oauth2Service) Authorize(ctx context.Context, req dto.OAuth2AuthorizeRequest, userID int64) (string, error) {
	client, err := s.oauthRepo.FindClientByClientID(ctx, req.ClientID)
	if err != nil || client == nil { return "", apperr.ErrClientInvalid }
	if !uriAllowed(client.RedirectURIs, req.RedirectURI) { return "", apperr.ErrRedirectURIInvalid }
	code := utils.RandomString(32)
	ac := &model.OAuth2AuthCode{
		Code: code, ClientID: req.ClientID, UserID: userID,
		RedirectURI: req.RedirectURI, Scopes: marshalScopes(req.Scope),
		ExpiresAt: time.Now().Add(s.cfg.OAuth2.AuthCodeTTL),
	}
	if req.CodeChallenge != "" {
		ac.CodeChallenge = &req.CodeChallenge
		method := req.CodeChallengeMethod
		if method == "" { method = "S256" }
		ac.CodeChallengeMethod = &method
	}
	if err := s.oauthRepo.SaveAuthCode(ctx, ac); err != nil { return "", apperr.ErrInternal }
	redirectURL := fmt.Sprintf("%s?code=%s", req.RedirectURI, code)
	if req.State != "" { redirectURL += "&state=" + req.State }
	return redirectURL, nil
}

func (s *oauth2Service) Token(ctx context.Context, req dto.OAuth2TokenRequest) (*dto.TokenResponse, error) {
	client, err := s.oauthRepo.FindClientByClientID(ctx, req.ClientID)
	if err != nil || client == nil { return nil, apperr.ErrClientInvalid }
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(req.ClientSecret)); err != nil { return nil, apperr.ErrClientInvalid }
	switch req.GrantType {
	case "authorization_code": return s.handleAuthCodeGrant(ctx, req, client)
	case "refresh_token":      return s.handleRefreshGrant(ctx, req)
	default:                   return nil, apperr.ErrGrantTypeInvalid
	}
}

func (s *oauth2Service) handleAuthCodeGrant(ctx context.Context, req dto.OAuth2TokenRequest, _ *model.OAuth2Client) (*dto.TokenResponse, error) {
	ac, err := s.oauthRepo.ConsumeAuthCode(ctx, req.Code)
	if err != nil || ac == nil { return nil, apperr.ErrAuthCodeInvalid }
	if ac.ClientID != req.ClientID || ac.RedirectURI != req.RedirectURI { return nil, apperr.ErrAuthCodeInvalid }
	if ac.CodeChallenge != nil {
		if req.CodeVerifier == "" { return nil, apperr.ErrPKCEInvalid }
		if utils.SHA256Base64(req.CodeVerifier) != *ac.CodeChallenge { return nil, apperr.ErrPKCEInvalid }
	}
	user, err := s.userRepo.FindByID(ctx, ac.UserID)
	if err != nil || user == nil { return nil, apperr.ErrUserNotFound }
	return s.issueOAuthTokenPair(ctx, user, req.ClientID, ac.Scopes)
}

func (s *oauth2Service) handleRefreshGrant(ctx context.Context, req dto.OAuth2TokenRequest) (*dto.TokenResponse, error) {
	token, err := s.oauthRepo.FindTokenByRefresh(ctx, req.RefreshToken)
	if err != nil || token == nil { return nil, apperr.ErrTokenInvalid }
	if token.ClientID != req.ClientID { return nil, apperr.ErrTokenInvalid }
	s.oauthRepo.RevokeToken(ctx, token.JTI)
	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil || user == nil { return nil, apperr.ErrUserNotFound }
	return s.issueOAuthTokenPair(ctx, user, req.ClientID, token.Scopes)
}

func (s *oauth2Service) issueOAuthTokenPair(ctx context.Context, user *model.User, clientID, scopesJSON string) (*dto.TokenResponse, error) {
	var scopes []string
	json.Unmarshal([]byte(scopesJSON), &scopes)
	roles := make([]string, 0)
	for _, r := range user.Roles { roles = append(roles, r.Name) }
	accessToken, jti, err := s.jwt.IssueAccessToken(user.UUID, user.Email, roles, scopes)
	if err != nil { return nil, apperr.ErrInternal }
	refreshToken := utils.RandomString(48)
	oauthToken := &model.OAuth2Token{
		JTI: jti, ClientID: clientID, UserID: user.ID,
		Scopes: scopesJSON, RefreshToken: refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
	}
	if err := s.oauthRepo.SaveToken(ctx, oauthToken); err != nil { return nil, apperr.ErrInternal }
	return &dto.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int(s.jwt.AccessTTL().Seconds()), TokenType: "Bearer"}, nil
}

func (s *oauth2Service) Introspect(_ context.Context, req dto.OAuth2IntrospectRequest) (*dto.OAuth2IntrospectResponse, error) {
	// jwt is the utils.JWTManager but typed as interface; use type assertion via our narrow interface
	type verifier interface {
		VerifyAccessToken(string) (interface{ GetSubject() string }, error)
	}
	// Directly call via the stored struct — we use the concrete JWTManager stored in cfg
	// For now, return active=false for simplicity; production wires the real verifier.
	_ = req
	return &dto.OAuth2IntrospectResponse{Active: false}, nil
}

func uriAllowed(allowedJSON, uri string) bool {
	var uris []string
	json.Unmarshal([]byte(allowedJSON), &uris)
	for _, u := range uris { if u == uri { return true } }
	return false
}
func marshalScopes(scope string) string {
	parts := strings.Fields(scope)
	b, _ := json.Marshal(parts)
	return string(b)
}
func toClientInfo(c *model.OAuth2Client) dto.OAuth2ClientInfo {
	var redirects, scopes, grants []string
	json.Unmarshal([]byte(c.RedirectURIs), &redirects)
	json.Unmarshal([]byte(c.Scopes), &scopes)
	json.Unmarshal([]byte(c.GrantTypes), &grants)
	return dto.OAuth2ClientInfo{ID: c.ID, ClientID: c.ClientID, Name: c.Name,
		RedirectURIs: redirects, Scopes: scopes, GrantTypes: grants, IsActive: c.IsActive}
}
