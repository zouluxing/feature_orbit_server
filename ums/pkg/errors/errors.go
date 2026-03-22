package errors

import "net/http"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func New(code, status int, message string) *AppError {
	return &AppError{Code: code, Status: status, Message: message}
}

var (
	ErrInvalidCredentials = New(40001, http.StatusUnauthorized, "账号或密码错误")
	ErrAccountLocked      = New(40002, http.StatusUnauthorized, "账号已锁定，请30分钟后重试")
	ErrTokenExpired       = New(40003, http.StatusUnauthorized, "Token 已过期")
	ErrTokenInvalid       = New(40004, http.StatusUnauthorized, "Token 无效或已吊销")
	ErrForbidden          = New(40005, http.StatusForbidden, "权限不足")
	ErrMFACodeInvalid     = New(40006, http.StatusBadRequest, "MFA 验证码错误")
	ErrDuplicateUser      = New(40007, http.StatusConflict, "邮箱或用户名已存在")
	ErrUserNotFound       = New(40008, http.StatusNotFound, "用户不存在")
	ErrClientInvalid      = New(40009, http.StatusBadRequest, "OAuth2 client_id 无效")
	ErrRedirectURIInvalid = New(40010, http.StatusBadRequest, "redirect_uri 不在白名单")
	ErrRoleNotFound       = New(40011, http.StatusNotFound, "角色不存在")
	ErrSystemRole         = New(40012, http.StatusForbidden, "系统角色不可删除")
	ErrAuthCodeInvalid    = New(40013, http.StatusBadRequest, "授权码无效或已过期")
	ErrGrantTypeInvalid   = New(40014, http.StatusBadRequest, "不支持的 grant_type")
	ErrMFARequired        = New(40015, http.StatusUnauthorized, "需要 MFA 验证")
	ErrMFAAlreadyEnabled  = New(40016, http.StatusBadRequest, "MFA 已开启")
	ErrPKCEInvalid        = New(40017, http.StatusBadRequest, "PKCE code_verifier 校验失败")
	ErrInternal           = New(50001, http.StatusInternalServerError, "服务内部错误")
)
