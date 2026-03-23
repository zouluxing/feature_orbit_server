package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/response"
)

type UserHandler struct{ svc service.UserService }

func NewUserHandler(svc service.UserService) *UserHandler { return &UserHandler{svc: svc} }

func (h *UserHandler) GetMe(c *gin.Context) {
	uuid := CurrentUserUUID(c)
	info, err := h.svc.GetByUUID(c.Request.Context(), uuid)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, info)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	uuid := CurrentUserUUID(c)
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	info, err := h.svc.Update(c.Request.Context(), uuid, req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, info)
}

func (h *UserHandler) UpdateMyPassword(c *gin.Context) {
	uuid := CurrentUserUUID(c)
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.UpdatePassword(c.Request.Context(), uuid, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err); return
	}
	response.OK(c, nil)
}

func (h *UserHandler) EnableMFA(c *gin.Context) {
	uuid := CurrentUserUUID(c)
	res, err := h.svc.EnableMFA(c.Request.Context(), uuid)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *UserHandler) DisableMFA(c *gin.Context) {
	uuid := CurrentUserUUID(c)
	var req dto.VerifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.DisableMFA(c.Request.Context(), uuid, req.Code); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.BadRequest(c, err.Error()); return }
	res, err := h.svc.List(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	info, err := h.svc.GetByUUID(c.Request.Context(), c.Param("uuid"))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, info)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	info, err := h.svc.Create(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.Created(c, info)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	info, err := h.svc.Update(c.Request.Context(), c.Param("uuid"), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, info)
}

func (h *UserHandler) PatchUserStatus(c *gin.Context) {
	var req dto.PatchUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.SetStatus(c.Request.Context(), c.Param("uuid"), req.Status); err != nil {
		response.Fail(c, err); return
	}
	response.OK(c, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("uuid")); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.ResetPassword(c.Request.Context(), c.Param("uuid"), req.NewPassword); err != nil {
		response.Fail(c, err); return
	}
	response.OK(c, nil)
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.AssignRoles(c.Request.Context(), c.Param("uuid"), req.RoleIDs); err != nil {
		response.Fail(c, err); return
	}
	response.OK(c, nil)
}

func (h *UserHandler) RemoveRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid role_id"); return }
	if err := h.svc.RemoveRole(c.Request.Context(), c.Param("uuid"), roleID); err != nil {
		response.Fail(c, err); return
	}
	response.OK(c, nil)
}

func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	perms, err := h.svc.GetPermissions(c.Request.Context(), c.Param("uuid"))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, gin.H{"permissions": perms})
}

func currentUserID(c *gin.Context) int64 {
	v, _ := c.Get("user_id_int64")
	id, _ := v.(int64)
	return id
}

func ParseInt64Param(c *gin.Context, key string) (int64, bool) {
	v, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid "+key)
		return 0, false
	}
	return v, true
}

func unused_currentUserID() int64 { _ = currentUserID; return 0 }
func init() { _ = unused_currentUserID }

// ensure unused function references compile
var _ = func() { _ = ParseInt64Param; _ = currentUserID }

// Needed by handler_test - exported alias
func GetCurrentUserID(c *gin.Context) int64 {
	v, _ := c.Get(ContextKeyUserUUID)
	if s, ok := v.(string); ok && s != "" { return 0 }
	v2, _ := c.Get("user_id_int64")
	id, _ := v2.(int64)
	return id
}

// keep import used
var _ = http.StatusOK
