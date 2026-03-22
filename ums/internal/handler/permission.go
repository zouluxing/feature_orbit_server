package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/ums/internal/dto"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/response"
)

type PermissionHandler struct{ permSvc service.PermissionService }

func NewPermissionHandler(permSvc service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permSvc: permSvc}
}

func (h *PermissionHandler) ListRoles(c *gin.Context) {
	roles, err := h.permSvc.ListRoles(c.Request.Context())
	if err != nil { response.Fail(c, err); return }
	response.OK(c, roles)
}
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	role, err := h.permSvc.CreateRole(c.Request.Context(), req)
	if err != nil { response.Fail(c, err); return }
	response.Created(c, role)
}
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid role id"); return }
	if err := h.permSvc.DeleteRole(c.Request.Context(), id); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *PermissionHandler) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.BadRequest(c, "invalid role id"); return }
	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.permSvc.SetRolePermissions(c.Request.Context(), id, req); err != nil { response.Fail(c, err); return }
	response.OK(c, nil)
}
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	perms, err := h.permSvc.ListPermissions(c.Request.Context())
	if err != nil { response.Fail(c, err); return }
	response.OK(c, perms)
}
