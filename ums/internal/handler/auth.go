package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/response"
	"github.com/zouluxing/ums/pkg/utils"
)

type AuthHandler struct {
	authSvc service.AuthService
	jwt     *utils.JWTManager
}

func NewAuthHandler(authSvc service.AuthService, jwt *utils.JWTManager) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, jwt: jwt}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.authSvc.Register(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.Created(c, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.authSvc.Login(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	claims, err := h.jwt.VerifyAccessToken(tokenStr)
	if err != nil { response.Fail(c, err); return }
	if err := h.authSvc.Logout(c.Request.Context(), claims.ID); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}

func (h *AuthHandler) JWKS(c *gin.Context) {
	response.OK(c, dto.JWKSResponse{Keys: []dto.JWK{{
		Kty: "RSA", Use: "sig", Kid: h.jwt.KeyID(), Alg: "RS256",
		N: h.jwt.PublicKeyN(), E: h.jwt.PublicKeyE(),
	}}})
}
