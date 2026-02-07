package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/service"
	"github.com/imkerbos/ACME-Console/internal/utils"
)

type ChallengeHandler struct {
	certSvc *service.CertificateService
}

func NewChallengeHandler(certSvc *service.CertificateService) *ChallengeHandler {
	return &ChallengeHandler{certSvc: certSvc}
}

// List handles GET /api/v1/certificates/:id/challenges
func (h *ChallengeHandler) List(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	challenges, err := h.certSvc.GetChallenges(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, challenges)
}

// Export handles GET /api/v1/certificates/:id/challenges/export
func (h *ChallengeHandler) Export(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	template, err := h.certSvc.ExportChallenges(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=dns-challenges.txt")
	c.String(http.StatusOK, template)
}
