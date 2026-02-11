package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/service"
)

type SettingHandler struct {
	svc *service.SettingService
}

func NewSettingHandler(svc *service.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

// List 获取所有配置
// GET /api/v1/settings
func (h *SettingHandler) List(c *gin.Context) {
	settings, err := h.svc.GetAll()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, settings)
}

// GetACME 获取 ACME 相关配置
// GET /api/v1/settings/acme
func (h *SettingHandler) GetACME(c *gin.Context) {
	config := h.svc.GetACMEConfig()
	response.Success(c, config)
}

// UpdateACMERequest ACME 配置更新请求
type UpdateACMERequest struct {
	DNSResolvers *string `json:"dns_resolvers"`
	DNSTimeout   *string `json:"dns_timeout"`
}

// UpdateACME 更新 ACME 相关配置
// PUT /api/v1/settings/acme
func (h *SettingHandler) UpdateACME(c *gin.Context) {
	var req UpdateACMERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.DNSResolvers != nil {
		if err := h.svc.Set(model.SettingDNSResolvers, *req.DNSResolvers); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	if req.DNSTimeout != nil {
		if err := h.svc.Set(model.SettingDNSTimeout, *req.DNSTimeout); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	// 返回更新后的配置
	config := h.svc.GetACMEConfig()
	response.Success(c, config)
}

// GetSite 获取网站配置（公开接口，无需认证）
// GET /api/v1/settings/site
func (h *SettingHandler) GetSite(c *gin.Context) {
	config := h.svc.GetSiteConfig()
	response.Success(c, config)
}

// UpdateSiteRequest 网站配置更新请求
type UpdateSiteRequest struct {
	Title    *string `json:"title"`
	Subtitle *string `json:"subtitle"`
}

// UpdateSite 更新网站配置
// PUT /api/v1/admin/settings/site
func (h *SettingHandler) UpdateSite(c *gin.Context) {
	var req UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Title != nil {
		if err := h.svc.Set(model.SettingSiteTitle, *req.Title); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	if req.Subtitle != nil {
		if err := h.svc.Set(model.SettingSiteSubtitle, *req.Subtitle); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	config := h.svc.GetSiteConfig()
	response.Success(c, config)
}

// UpdateRequest 通用配置更新请求
type UpdateRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// Update 批量更新配置
// PUT /api/v1/settings
func (h *SettingHandler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.svc.SetMultiple(req.Settings); err != nil {
		response.InternalError(c, err)
		return
	}

	settings, err := h.svc.GetAll()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, settings)
}
