package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/pagination"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/service"
	"github.com/imkerbos/ACME-Console/internal/utils"
)

type CertificateHandler struct {
	svc *service.CertificateService
}

func NewCertificateHandler(svc *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{svc: svc}
}

// Create handles POST /api/v1/certificates
func (h *CertificateHandler) Create(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req service.CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.svc.Create(&req, userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, resp)
}

// List handles GET /api/v1/certificates
func (h *CertificateHandler) List(c *gin.Context) {
	userID := utils.GetUserID(c)
	params := pagination.ParseFromContext(c)

	// Check for workspace_id filter
	workspaceIDStr := c.Query("workspace_id")
	var workspaceID *uint
	if workspaceIDStr != "" {
		id, err := strconv.ParseUint(workspaceIDStr, 10, 32)
		if err == nil {
			wid := uint(id)
			workspaceID = &wid
		}
	}

	result, err := h.svc.ListPaginatedWithFilter(params, userID, workspaceID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// Get handles GET /api/v1/certificates/:id
func (h *CertificateHandler) Get(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	cert, err := h.svc.GetByID(id)
	if err != nil {
		response.NotFound(c, "certificate not found")
		return
	}

	response.Success(c, cert)
}

// Verify handles POST /api/v1/certificates/:id/verify
func (h *CertificateHandler) Verify(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	cert, err := h.svc.Verify(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, cert)
}

// PreVerify handles POST /api/v1/certificates/:id/pre-verify
// Checks if DNS TXT records are correctly set up before triggering CA verification
func (h *CertificateHandler) PreVerify(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	results, allOK, err := h.svc.PreVerifyDNS(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"results":   results,
		"all_ok":    allOK,
		"ready":     allOK,
		"message":   getPreVerifyMessage(allOK),
	})
}

func getPreVerifyMessage(allOK bool) string {
	if allOK {
		return "All DNS records verified. Ready to request certificate from CA."
	}
	return "Some DNS records are not yet propagated. Please wait and try again."
}

func getContentType(format string) string {
	switch format {
	case "pem", "fullchain":
		return "application/x-pem-file"
	case "pfx":
		return "application/x-pkcs12"
	case "zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// Download handles GET /api/v1/certificates/:id/download
// Query params:
//   - format: pem, fullchain, pfx, zip (default: pem)
//   - password: password for PFX format (default: changeit)
func (h *CertificateHandler) Download(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	cert, err := h.svc.GetByID(id)
	if err != nil {
		response.NotFound(c, "certificate not found")
		return
	}

	if cert.Status != model.CertificateStatusReady {
		response.BadRequest(c, "certificate is not ready for download")
		return
	}

	format := c.DefaultQuery("format", "pem")
	password := c.Query("password")

	// For legacy API compatibility, return JSON if no format specified
	if format == "pem" && c.Query("format") == "" {
		response.Success(c, gin.H{
			"cert":  cert.CertPEM,
			"chain": cert.ChainPEM,
		})
		return
	}

	// Get certificate bundle in requested format
	data, filename, err := h.svc.GetCertificateBundle(id, format, password)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Set appropriate content type
	contentType := getContentType(format)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, contentType, data)
}

// Delete handles DELETE /api/v1/certificates/:id
func (h *CertificateHandler) Delete(c *gin.Context) {
	id, err := utils.ParseID(c)
	if err != nil {
		response.BadRequest(c, "invalid certificate id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "certificate deleted successfully")
}
