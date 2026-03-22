package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/response"
)

type UserHandler struct{ userSvc service.UserService }

func NewUserHandler(userSvc service.UserService) *UserHandler { return &UserHandler{userSvc: userSvc} }

func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if req.Page == 0 { req.Page = 1 }
	if req.PageSize == 0 { req.PageSize = 20 }
	res, err := h.userSvc.List(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}
func (h *UserHandler) GetUser(c *gin.Context) {
	user, err := h.userSvc.GetByUUID(c.Request.Context(), c.Param("uuid"))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, user)
}
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	user, err := h.userSvc.Create(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.Created(c, user)
}
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	user, err := h.userSvc.Update(c.Request.Context(), c.Param("uuid"), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, user)
}
func (h *UserHandler) PatchUserStatus(c *gin.Context) {
	var req dto.PatchUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.userSvc.SetStatus(c.Request.Context(), c.Param("uuid"), req.Status); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) DeleteUser(c *gin.Context) {
	if err := h.userSvc.Delete(c.Request.Context(), c.Param("uuid")); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.userSvc.ResetPassword(c.Request.Context(), c.Param("uuid"), req.NewPassword); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) AssignRoles(c *gin.Context) {
	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.userSvc.AssignRoles(c.Request.Context(), c.Param("uuid"), req.RoleIDs); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) RemoveRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid role_id"); return }
	if err := h.userSvc.RemoveRole(c.Request.Context(), c.Param("uuid"), roleID); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	perms, err := h.userSvc.GetPermissions(c.Request.Context(), c.Param("uuid"))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, perms)
}
func (h *UserHandler) GetMe(c *gin.Context) {
	user, err := h.userSvc.GetByUUID(c.Request.Context(), CurrentUserUUID(c))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, user)
}
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	user, err := h.userSvc.Update(c.Request.Context(), CurrentUserUUID(c), req)
	if err != nil { response.Fail(c, err); return }
	response.OK(c, user)
}
func (h *UserHandler) UpdateMyPassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.userSvc.UpdatePassword(c.Request.Context(), CurrentUserUUID(c), req.OldPassword, req.NewPassword); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *UserHandler) EnableMFA(c *gin.Context) {
	res, err := h.userSvc.EnableMFA(c.Request.Context(), CurrentUserUUID(c))
	if err != nil { response.Fail(c, err); return }
	response.OK(c, res)
}
func (h *UserHandler) DisableMFA(c *gin.Context) {
	var req dto.VerifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.userSvc.DisableMFA(c.Request.Context(), CurrentUserUUID(c), req.Code); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
