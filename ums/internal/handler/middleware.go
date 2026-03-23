package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/service"
	apperr "github.com/zouluxing/ums/pkg/errors"
	"github.com/zouluxing/ums/pkg/response"
	"github.com/zouluxing/ums/pkg/utils"
)

const (
	ContextKeyUserUUID = "user_uuid"
	ContextKeyRoles    = "roles"
	ContextKeyClaims   = "claims"
)

func JWTAuth(jwt *utils.JWTManager, cache service.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") { response.Fail(c, apperr.ErrTokenInvalid); c.Abort(); return }
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwt.VerifyAccessToken(tokenStr)
		if err != nil { response.Fail(c, apperr.ErrTokenExpired); c.Abort(); return }
		if cache != nil {
			if revoked, _ := cache.Exists(c.Request.Context(), "ums:blacklist:"+claims.ID); revoked {
				response.Fail(c, apperr.ErrTokenInvalid); c.Abort(); return
			}
		}
		c.Set(ContextKeyUserUUID, claims.Subject)
		c.Set(ContextKeyRoles, claims.Roles)
		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	required := make(map[string]bool, len(roles))
	for _, r := range roles { required[r] = true }
	return func(c *gin.Context) {
		userRoles, _ := c.Get(ContextKeyRoles)
		if s, ok := userRoles.([]string); ok {
			for _, r := range s { if required[r] { c.Next(); return } }
		}
		response.Fail(c, apperr.ErrForbidden); c.Abort()
	}
}

func CurrentUserUUID(c *gin.Context) string {
	v, _ := c.Get(ContextKeyUserUUID)
	if s, ok := v.(string); ok { return s }; return ""
}

func JWTAuthWithUserID(jwt *utils.JWTManager, cache service.CacheService) gin.HandlerFunc {
	base := JWTAuth(jwt, cache)
	return func(c *gin.Context) {
		base(c); if c.IsAborted() { return }
		c.Set("user_id_int64", int64(0)); c.Next()
	}
}
