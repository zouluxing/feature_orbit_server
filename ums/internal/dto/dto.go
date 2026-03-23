package dto

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email,max=255"`
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}
type RegisterResponse struct {
	UserUUID string `json:"user_uuid"`
	Email    string `json:"email"`
}
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
	MFACode  string `json:"mfa_code"`
}
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type UserInfo struct {
	UUID        string   `json:"uuid"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Phone       *string  `json:"phone,omitempty"`
	AvatarURL   *string  `json:"avatar_url,omitempty"`
	Status      int8     `json:"status"`
	MFAEnabled  bool     `json:"mfa_enabled"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
type UpdateUserRequest struct {
	Username  *string `json:"username"   binding:"omitempty,min=3,max=64"`
	Phone     *string `json:"phone"      binding:"omitempty,max=32"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url,max=512"`
}
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}
type PatchUserStatusRequest struct {
	Status int8 `json:"status" binding:"required,oneof=1 2 3"`
}
type CreateUserRequest struct {
	Email    string  `json:"email"    binding:"required,email"`
	Username string  `json:"username" binding:"required,min=3,max=64"`
	Password string  `json:"password" binding:"required,min=8,max=64"`
	RoleIDs  []int64 `json:"role_ids"`
}
type UserListRequest struct {
	Email    string `form:"email"`
	Status   *int8  `form:"status"`
	Page     int    `form:"page"      binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}
type UserListResponse struct {
	List     []UserInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
type RoleInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsSystem    bool   `json:"is_system"`
}
type CreateRoleRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=256"`
}
type PermissionInfo struct {
	ID       int64  `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}
type AssignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}
type OAuth2ClientInfo struct {
	ID           int64    `json:"id"`
	ClientID     string   `json:"client_id"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	GrantTypes   []string `json:"grant_types"`
	IsActive     bool     `json:"is_active"`
}
type CreateClientRequest struct {
	Name         string   `json:"name"          binding:"required,min=2,max=128"`
	RedirectURIs []string `json:"redirect_uris" binding:"required,min=1"`
	Scopes       []string `json:"scopes"        binding:"required,min=1"`
}
type CreateClientResponse struct {
	OAuth2ClientInfo
	ClientSecret string `json:"client_secret"`
}
type UpdateClientRequest struct {
	Name         *string  `json:"name"          binding:"omitempty,min=2,max=128"`
	RedirectURIs []string `json:"redirect_uris" binding:"omitempty,min=1"`
	Scopes       []string `json:"scopes"        binding:"omitempty,min=1"`
	IsActive     *bool    `json:"is_active"`
}
type OAuth2AuthorizeRequest struct {
	ClientID            string `form:"client_id"             binding:"required"`
	ResponseType        string `form:"response_type"         binding:"required,eq=code"`
	RedirectURI         string `form:"redirect_uri"          binding:"required,url"`
	Scope               string `form:"scope"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge"`
	CodeChallengeMethod string `form:"code_challenge_method"`
}
type OAuth2TokenRequest struct {
	GrantType    string `form:"grant_type"    binding:"required"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"     binding:"required"`
	ClientSecret string `form:"client_secret"`
	RefreshToken string `form:"refresh_token"`
	CodeVerifier string `form:"code_verifier"`
}
type OAuth2IntrospectRequest struct {
	Token         string `form:"token"           binding:"required"`
	TokenTypeHint string `form:"token_type_hint"`
}
type OAuth2IntrospectResponse struct {
	Active   bool   `json:"active"`
	Sub      string `json:"sub,omitempty"`
	Email    string `json:"email,omitempty"`
	Scope    string `json:"scope,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Exp      int64  `json:"exp,omitempty"`
}
type EnableMFAResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otp_auth_uri"`
	QRCodeURL  string `json:"qr_code_url"`
}
type VerifyMFARequest struct {
	Code string `json:"code" binding:"required,len=6"`
}
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
