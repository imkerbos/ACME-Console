package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/service"
	"github.com/imkerbos/ACME-Console/internal/utils"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List handles GET /api/v1/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	var workspaceID, certificateID *uint

	if wsID := c.Query("workspace_id"); wsID != "" {
		id := utils.ParseQueryUint(c, "workspace_id")
		if id > 0 {
			workspaceID = &id
		}
	}

	if certID := c.Query("certificate_id"); certID != "" {
		id := utils.ParseQueryUint(c, "certificate_id")
		if id > 0 {
			certificateID = &id
		}
	}

	configs, err := h.svc.ListConfigs(workspaceID, certificateID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, configs)
}

// Create handles POST /api/v1/notifications
func (h *NotificationHandler) Create(c *gin.Context) {
	var req service.CreateNotificationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	config, err := h.svc.CreateConfig(&req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, config)
}

// Get handles GET /api/v1/notifications/:id
func (h *NotificationHandler) Get(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	config, err := h.svc.GetConfig(id)
	if err != nil {
		response.NotFound(c, "notification config not found")
		return
	}

	response.Success(c, config)
}

// Update handles PUT /api/v1/notifications/:id
func (h *NotificationHandler) Update(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	var req service.CreateNotificationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.svc.UpdateConfig(id, &req); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "notification config updated successfully")
}

// Delete handles DELETE /api/v1/notifications/:id
func (h *NotificationHandler) Delete(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.svc.DeleteConfig(id); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "notification config deleted successfully")
}

// ListLogs handles GET /api/v1/certificates/:id/notification-logs
func (h *NotificationHandler) ListLogs(c *gin.Context) {
	certID, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	limit := utils.ParseQueryInt(c, "limit", 50)

	logs, err := h.svc.ListLogs(certID, limit)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, logs)
}

// Test handles POST /api/v1/notifications/:id/test
func (h *NotificationHandler) Test(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.svc.TestWebhook(id); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "test notification sent successfully")
}
