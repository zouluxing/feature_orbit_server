package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/feature_orbit_server/internal/dto"
	"github.com/zouluxing/feature_orbit_server/internal/middleware"
	"github.com/zouluxing/feature_orbit_server/internal/service"
)

type FeatureHandler struct{ svc service.FeatureService }

func NewFeatureHandler(svc service.FeatureService) *FeatureHandler { return &FeatureHandler{svc: svc} }

func (h *FeatureHandler) ListFeatures(c *gin.Context) {
	var req dto.FeatureListRequest
	if err := c.ShouldBindQuery(&req); err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, err.Error())); return }
	if req.Page == 0 { req.Page = 1 }
	if req.PageSize == 0 { req.PageSize = 20 }
	res, err := h.svc.List(c.Request.Context(), req)
	if err != nil { c.JSON(http.StatusInternalServerError, apiErr(50001, "internal error")); return }
	c.JSON(http.StatusOK, okResp(res))
}

func (h *FeatureHandler) GetFeature(c *gin.Context) {
	id, err := parseU64(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, "invalid id")); return }
	info, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrFeatureNotFound) { c.JSON(http.StatusNotFound, apiErr(40401, "feature not found")); return }
		c.JSON(http.StatusInternalServerError, apiErr(50001, "internal error")); return
	}
	c.JSON(http.StatusOK, okResp(info))
}

func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok { c.JSON(http.StatusUnauthorized, apiErr(40004, "not authenticated")); return }
	var req dto.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, err.Error())); return }
	info, err := h.svc.Create(c.Request.Context(), claims.UserUUID, claims.Email, req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateSlug) { c.JSON(http.StatusConflict, apiErr(40901, "slug already exists")); return }
		c.JSON(http.StatusInternalServerError, apiErr(50001, "internal error")); return
	}
	c.JSON(http.StatusCreated, okResp(info))
}

func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok { c.JSON(http.StatusUnauthorized, apiErr(40004, "not authenticated")); return }
	id, err := parseU64(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, "invalid id")); return }
	var req dto.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, err.Error())); return }
	info, err := h.svc.Update(c.Request.Context(), id, claims.UserUUID, req)
	if err != nil {
		if errors.Is(err, service.ErrFeatureNotFound) { c.JSON(http.StatusNotFound, apiErr(40401, "not found")); return }
		if errors.Is(err, service.ErrForbidden) { c.JSON(http.StatusForbidden, apiErr(40005, "forbidden")); return }
		c.JSON(http.StatusInternalServerError, apiErr(50001, "internal error")); return
	}
	c.JSON(http.StatusOK, okResp(info))
}

func (h *FeatureHandler) DeleteFeature(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok { c.JSON(http.StatusUnauthorized, apiErr(40004, "not authenticated")); return }
	id, err := parseU64(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, apiErr(40000, "invalid id")); return }
	if err := h.svc.Delete(c.Request.Context(), id, claims.UserUUID); err != nil {
		if errors.Is(err, service.ErrFeatureNotFound) { c.JSON(http.StatusNotFound, apiErr(40401, "not found")); return }
		if errors.Is(err, service.ErrForbidden) { c.JSON(http.StatusForbidden, apiErr(40005, "forbidden")); return }
		c.JSON(http.StatusInternalServerError, apiErr(50001, "internal error")); return
	}
	c.JSON(http.StatusOK, okResp(nil))
}

func GetMe(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok { c.JSON(http.StatusUnauthorized, apiErr(40004, "not authenticated")); return }
	c.JSON(http.StatusOK, okResp(gin.H{"user_uuid": claims.UserUUID, "email": claims.Email, "roles": claims.Roles, "scopes": claims.Scopes}))
}

type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func okResp(data any) envelope          { return envelope{Code: 0, Message: "ok", Data: data} }
func apiErr(code int, msg string) envelope { return envelope{Code: code, Message: msg} }
func parseU64(s string) (uint64, error)  { return strconv.ParseUint(s, 10, 64) }
