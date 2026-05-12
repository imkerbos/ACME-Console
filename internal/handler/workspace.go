package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/service"
	"github.com/imkerbos/ACME-Console/internal/utils"
)

type WorkspaceHandler struct {
	svc *service.WorkspaceService
}

func NewWorkspaceHandler(svc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc}
}

// List handles GET /api/v1/workspaces
func (h *WorkspaceHandler) List(c *gin.Context) {
	userID := utils.GetUserID(c)

	workspaces, err := h.svc.ListUserWorkspaces(userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, workspaces)
}

// Create handles POST /api/v1/workspaces
func (h *WorkspaceHandler) Create(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req service.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	workspace, err := h.svc.Create(userID, &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, workspace)
}

// Get handles GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) Get(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	workspace, err := h.svc.GetByID(workspaceID, userID)
	if err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		if err == service.ErrWorkspaceNotFound {
			response.NotFound(c, "workspace not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, workspace)
}

// Update handles PUT /api/v1/workspaces/:id
func (h *WorkspaceHandler) Update(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	var req service.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.svc.Update(workspaceID, userID, &req); err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, "workspace updated successfully")
}

// Delete handles DELETE /api/v1/workspaces/:id
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	if err := h.svc.Delete(workspaceID, userID); err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		if err == service.ErrWorkspaceNotFound {
			response.NotFound(c, "workspace not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, "workspace deleted successfully")
}

// ListMembers handles GET /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	members, err := h.svc.ListMembers(workspaceID, userID)
	if err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, members)
}

// AddMember handles POST /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	var req service.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.svc.AddMember(workspaceID, userID, &req); err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		if err == service.ErrUserAlreadyMember {
			response.BadRequest(c, "user is already a member")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, gin.H{"message": "member added successfully"})
}

// UpdateMember handles PUT /api/v1/workspaces/:id/members/:userId
func (h *WorkspaceHandler) UpdateMember(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	memberUserIDStr := c.Param("userId")
	memberUserIDInt, err := strconv.ParseUint(memberUserIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	memberUserID := uint(memberUserIDInt)

	var req service.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.svc.UpdateMember(workspaceID, userID, memberUserID, &req); err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		if err == service.ErrMemberNotFound {
			response.NotFound(c, "member not found")
			return
		}
		if err == service.ErrCannotChangeOwnerRole {
			response.BadRequest(c, "cannot change owner role")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, "member updated successfully")
}

// RemoveMember handles DELETE /api/v1/workspaces/:id/members/:userId
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	userID := utils.GetUserID(c)
	workspaceID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid workspace id")
		return
	}

	memberUserIDStr := c.Param("userId")
	memberUserIDInt, err := strconv.ParseUint(memberUserIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	memberUserID := uint(memberUserIDInt)

	if err := h.svc.RemoveMember(workspaceID, userID, memberUserID); err != nil {
		if err == service.ErrWorkspaceAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		if err == service.ErrMemberNotFound {
			response.NotFound(c, "member not found")
			return
		}
		if err == service.ErrCannotRemoveOwner {
			response.BadRequest(c, "cannot remove workspace owner")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, "member removed successfully")
}
