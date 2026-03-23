package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/response"
)

type OAuth2Handler struct {
	svc     service.OAuth2Service
	authSvc service.AuthService
}

func NewOAuth2Handler(svc service.OAuth2Service, authSvc service.AuthService) *OAuth2Handler {
	return &OAuth2Handler{svc: svc, authSvc: authSvc}
}

func (h *OAuth2Handler) Authorize(c *gin.Context) {
	var req dto.OAuth2AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.BadRequest(c, err.Error()); return }
	userIDVal, exists := c.Get("user_id_int64")
	if !exists { response.Fail(c, errTokenInvalid()); return }
	userID, _ := userIDVal.(int64)
	redirectURL, err := h.svc.Authorize(c.Request.Context(), req, userID)
	if err != nil { response.Fail(c, err); return }
	c.Redirect(http.StatusFound, redirectURL)
}

func (h *OAuth2Handler) Token(c *gin.Context) {
	var req dto.OAuth2TokenRequest
	if err := c.ShouldBind(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.svc.Token(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *OAuth2Handler) Introspect(c *gin.Context) {
	var req dto.OAuth2IntrospectRequest
	if err := c.ShouldBind(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.svc.Introspect(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *OAuth2Handler) ListClients(c *gin.Context) {
	ownerID := clientOwnerID(c)
	clients, err := h.svc.ListClients(c.Request.Context(), ownerID)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, clients)
}

func (h *OAuth2Handler) CreateClient(c *gin.Context) {
	ownerID := clientOwnerID(c)
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.svc.CreateClient(c.Request.Context(), ownerID, req)
	if err != nil { response.Fail(c, err); return }
	response.Created(c, res)
}

func (h *OAuth2Handler) UpdateClient(c *gin.Context) {
	ownerID := clientOwnerID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid client id"); return }
	var req dto.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.svc.UpdateClient(c.Request.Context(), ownerID, id, req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *OAuth2Handler) DeleteClient(c *gin.Context) {
	ownerID := clientOwnerID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid client id"); return }
	if err := h.svc.DeleteClient(c.Request.Context(), ownerID, id); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}

func (h *OAuth2Handler) RotateSecret(c *gin.Context) {
	ownerID := clientOwnerID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid client id"); return }
	newSecret, err := h.svc.RotateClientSecret(c.Request.Context(), ownerID, id)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, gin.H{"client_secret": newSecret})
}

func clientOwnerID(c *gin.Context) int64 {
	v, _ := c.Get("user_id_int64")
	id, _ := v.(int64)
	return id
}

func errTokenInvalid() error {
	return &tokenErr{}
}

type tokenErr struct{}
func (e *tokenErr) Error() string { return "token invalid" }
