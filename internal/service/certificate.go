package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/pagination"
	"gorm.io/gorm"
)

type CertificateService struct {
	db       *gorm.DB
	acmeSvc  *AcmeShService  // Legacy mock service (deprecated)
	legoSvc  *LegoService    // Real ACME service
	useLego  bool            // Whether to use real ACME (lego) or mock
}

func NewCertificateService(db *gorm.DB, acmeSvc *AcmeShService) *CertificateService {
	return &CertificateService{
		db:      db,
		acmeSvc: acmeSvc,
		useLego: false,
	}
}

// NewCertificateServiceWithLego creates a CertificateService with real ACME support
func NewCertificateServiceWithLego(db *gorm.DB, legoSvc *LegoService) *CertificateService {
	return &CertificateService{
		db:      db,
		legoSvc: legoSvc,
		useLego: true,
	}
}

type CreateCertificateRequest struct {
	Domains     []string `json:"domains" binding:"required,min=1"`
	Email       string   `json:"email" binding:"required,email"`                 // 申请人邮箱
	KeyType     string   `json:"key_type" binding:"omitempty,oneof=RSA ECC"`     // 可选，默认 RSA
	KeySize     int      `json:"key_size,omitempty"`                             // 可选，根据 KeyType 自动设置
	WorkspaceID *uint    `json:"workspace_id,omitempty"`                         // 可选，NULL=私有证书
}

type CertificateResponse struct {
	ID         uint                `json:"id"`
	Domains    []string            `json:"domains"`
	KeyType    string              `json:"key_type"`
	Status     string              `json:"status"`
	IssuedAt   *string             `json:"issued_at,omitempty"`
	ExpiresAt  *string             `json:"expires_at,omitempty"`
	CreatedAt  string              `json:"created_at"`
	Challenges []ChallengeResponse `json:"challenges,omitempty"`
}

type ChallengeResponse struct {
	ID       uint   `json:"id"`
	Domain   string `json:"domain"`
	TXTHost  string `json:"txt_host"`
	TXTValue string `json:"txt_value"`
	Status   string `json:"status"`
}

func (s *CertificateService) Create(req *CreateCertificateRequest, userID uint) (*model.Certificate, error) {
	// 设置默认值
	if req.KeyType == "" {
		req.KeyType = "RSA" // 默认 RSA
	}
	if req.KeySize == 0 {
		if req.KeyType == "ECC" {
			req.KeySize = 256 // ECC 默认 P-256
		} else {
			req.KeySize = 2048 // RSA 默认 2048
		}
	}

	// 规范化域名：通配符域名自动包含根域名
	domains := normalizeDomains(req.Domains)

	domainsJSON, err := json.Marshal(domains)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal domains: %w", err)
	}

	createdByID := userID
	cert := &model.Certificate{
		Email:       req.Email,
		Domains:     string(domainsJSON),
		KeyType:     model.KeyType(req.KeyType),
		KeySize:     req.KeySize,
		Status:      model.CertificateStatusPending,
		WorkspaceID: req.WorkspaceID,
		CreatedBy:   &createdByID,
	}

	if err := s.db.Create(cert).Error; err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	if s.useLego && s.legoSvc != nil {
		// Use real ACME flow - create order and get challenges from CA
		if err := s.legoSvc.CreateOrder(cert.ID, req.Email, domains, req.KeyType, req.KeySize); err != nil {
			// If ACME fails, we might still have challenges stored
			// Reload to see what we have
			if reloadErr := s.db.Preload("Challenges").First(cert, cert.ID).Error; reloadErr != nil {
				return nil, fmt.Errorf("failed to create ACME order: %w", err)
			}
			// If we have challenges, return the cert with pending status
			if len(cert.Challenges) > 0 {
				return cert, nil
			}
			return nil, fmt.Errorf("failed to create ACME order: %w", err)
		}
	} else {
		// Use mock challenges (legacy behavior)
		challenges, err := s.acmeSvc.GenerateChallenges(cert.ID, domains)
		if err != nil {
			return nil, fmt.Errorf("failed to generate challenges: %w", err)
		}

		for _, ch := range challenges {
			if err := s.db.Create(&ch).Error; err != nil {
				return nil, fmt.Errorf("failed to create challenge: %w", err)
			}
		}
	}

	// Reload with challenges
	if err := s.db.Preload("Challenges").First(cert, cert.ID).Error; err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *CertificateService) List() ([]model.Certificate, error) {
	var certs []model.Certificate
	if err := s.db.Preload("Challenges").Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

func (s *CertificateService) ListPaginated(params pagination.Params) (*pagination.Result[model.Certificate], error) {
	var certs []model.Certificate
	var total int64

	// Count total
	if err := s.db.Model(&model.Certificate{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated records
	if err := s.db.Preload("Challenges").
		Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.Limit()).
		Find(&certs).Error; err != nil {
		return nil, err
	}

	result := pagination.NewResult(certs, total, params)
	return &result, nil
}

// ListPaginatedWithFilter lists certificates with workspace and user filtering
func (s *CertificateService) ListPaginatedWithFilter(params pagination.Params, userID uint, workspaceID *uint) (*pagination.Result[model.Certificate], error) {
	var certs []model.Certificate
	var total int64

	query := s.db.Model(&model.Certificate{})

	// Apply filters
	if workspaceID != nil {
		// Filter by workspace
		query = query.Where("workspace_id = ?", *workspaceID)
	} else {
		// Show private certificates (created by user) and workspace certificates (user is member)
		// Get user's workspace IDs
		var memberWorkspaceIDs []uint
		s.db.Model(&model.WorkspaceMember{}).
			Where("user_id = ?", userID).
			Pluck("workspace_id", &memberWorkspaceIDs)

		if len(memberWorkspaceIDs) > 0 {
			// Private certificates OR workspace certificates user has access to
			query = query.Where("(workspace_id IS NULL AND created_by = ?) OR workspace_id IN ?", userID, memberWorkspaceIDs)
		} else {
			// Only private certificates
			query = query.Where("workspace_id IS NULL AND created_by = ?", userID)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated records
	if err := query.Preload("Challenges").
		Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.Limit()).
		Find(&certs).Error; err != nil {
		return nil, err
	}

	result := pagination.NewResult(certs, total, params)
	return &result, nil
}

func (s *CertificateService) GetByID(id uint) (*model.Certificate, error) {
	var cert model.Certificate
	if err := s.db.Preload("Challenges").First(&cert, id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

func (s *CertificateService) Verify(id uint) (*model.Certificate, error) {
	cert, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.useLego && s.legoSvc != nil {
		// Use real ACME verification
		if err := s.legoSvc.FinalizeOrder(id); err != nil {
			cert.Status = model.CertificateStatusFailed
			s.db.Save(cert)
			return nil, fmt.Errorf("verification failed: %w", err)
		}
		// Reload to get updated status
		return s.GetByID(id)
	}

	// Mock verification (legacy behavior)
	verified, err := s.acmeSvc.VerifyChallenges(cert)
	if err != nil {
		return nil, err
	}

	if verified {
		now := time.Now()
		expiresAt := now.AddDate(0, 3, 0) // Mock: 3 months validity

		cert.Status = model.CertificateStatusReady
		cert.IssuedAt = &now
		cert.ExpiresAt = &expiresAt

		// Update all challenges to verified
		for i := range cert.Challenges {
			cert.Challenges[i].Status = model.ChallengeStatusVerified
			s.db.Save(&cert.Challenges[i])
		}
	} else {
		cert.Status = model.CertificateStatusFailed
	}

	if err := s.db.Save(cert).Error; err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *CertificateService) GetChallenges(certID uint) ([]model.Challenge, error) {
	var challenges []model.Challenge
	if err := s.db.Where("certificate_id = ?", certID).Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (s *CertificateService) ExportChallenges(certID uint) (string, error) {
	challenges, err := s.GetChallenges(certID)
	if err != nil {
		return "", err
	}

	return s.acmeSvc.ExportTXTTemplate(challenges), nil
}

// PreVerifyDNS checks if DNS TXT records are correctly set up
func (s *CertificateService) PreVerifyDNS(certID uint) ([]DNSCheckResult, bool, error) {
	if s.legoSvc == nil {
		return nil, false, fmt.Errorf("lego service not configured")
	}

	results, allOK, err := s.legoSvc.PreVerifyDNS(certID)
	if err != nil {
		return nil, false, err
	}

	// Convert to our response type
	dnsResults := make([]DNSCheckResult, len(results))
	for i, r := range results {
		dnsResults[i] = DNSCheckResult{
			Domain:        r.Domain,
			TXTHost:       r.TXTHost,
			ExpectedValue: r.ExpectedValue,
			FoundValues:   r.FoundValues,
			Matched:       r.Matched,
			Error:         r.Error,
		}
	}

	return dnsResults, allOK, nil
}

// DNSCheckResult represents the result of a DNS check
type DNSCheckResult struct {
	Domain        string   `json:"domain"`
	TXTHost       string   `json:"txt_host"`
	ExpectedValue string   `json:"expected_value"`
	FoundValues   []string `json:"found_values"`
	Matched       bool     `json:"matched"`
	Error         string   `json:"error,omitempty"`
}

// GetCertificateBundle returns the certificate in the specified format
func (s *CertificateService) GetCertificateBundle(certID uint, format, password string) ([]byte, string, error) {
	if s.legoSvc != nil {
		return s.legoSvc.GetCertificateBundle(certID, DownloadFormat(format), password)
	}

	// Mock mode: return mock certificate bundle
	return s.getMockCertificateBundle(certID, format)
}

// getMockCertificateBundle generates a mock certificate bundle for testing
func (s *CertificateService) getMockCertificateBundle(certID uint, format string) ([]byte, string, error) {
	var cert model.Certificate
	if err := s.db.First(&cert, certID).Error; err != nil {
		return nil, "", fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status != model.CertificateStatusReady {
		return nil, "", fmt.Errorf("certificate is not ready")
	}

	// Parse domains
	var domains []string
	if err := json.Unmarshal([]byte(cert.Domains), &domains); err != nil {
		domains = []string{"example.com"}
	}

	// Generate mock certificate content (include all domains)
	mockCert := fmt.Sprintf(`-----BEGIN CERTIFICATE-----
MOCK CERTIFICATE FOR TESTING
Domains: %s
Key Type: %s
Status: %s
Created: %s
-----END CERTIFICATE-----
`, strings.Join(domains, ", "), cert.KeyType, cert.Status, cert.CreatedAt.Format(time.RFC3339))

	mockKey := `-----BEGIN PRIVATE KEY-----
MOCK PRIVATE KEY FOR TESTING
This is a mock certificate generated for testing purposes.
Do not use in production.
-----END PRIVATE KEY-----
`

	switch format {
	case "pem":
		return []byte(mockCert), "certificate.pem", nil

	case "fullchain":
		return []byte(mockCert), "fullchain.pem", nil

	case "zip":
		var buf bytes.Buffer
		w := zip.NewWriter(&buf)

		// Create per-domain directories
		for _, domain := range domains {
			dir := sanitizeDomainDir(domain)

			certFile, _ := w.Create(dir + "/certificate.pem")
			certFile.Write([]byte(mockCert))

			chainFile, _ := w.Create(dir + "/fullchain.pem")
			chainFile.Write([]byte(mockCert))

			keyFile, _ := w.Create(dir + "/private.key")
			keyFile.Write([]byte(mockKey))
		}

		readmeFile, _ := w.Create("README.txt")
		readmeFile.Write([]byte(fmt.Sprintf(`MOCK CERTIFICATE BUNDLE

This is a mock certificate bundle for testing purposes.
To use real certificates, configure ACME settings in the admin panel.

Domains: %s

Each domain directory contains:
- certificate.pem: Certificate file
- fullchain.pem: Full certificate chain
- private.key: Private key file
`, strings.Join(domains, ", "))))

		w.Close()
		return buf.Bytes(), "certificate.zip", nil

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// normalizeDomains 规范化域名列表
// - 通配符域名 *.example.com 自动添加根域名 example.com
// - 去重并保持顺序（根域名放在通配符前面）
func normalizeDomains(domains []string) []string {
	seen := make(map[string]bool)
	var result []string

	// 第一遍：收集所有需要添加的根域名
	var rootsToAdd []string
	for _, domain := range domains {
		if strings.HasPrefix(domain, "*.") {
			rootDomain := strings.TrimPrefix(domain, "*.")
			if !seen[rootDomain] {
				// 检查用户是否已经手动添加了根域名
				hasRoot := false
				for _, d := range domains {
					if d == rootDomain {
						hasRoot = true
						break
					}
				}
				if !hasRoot {
					rootsToAdd = append(rootsToAdd, rootDomain)
					seen[rootDomain] = true
				}
			}
		}
	}

	// 第二遍：构建最终列表（根域名在前，保持原顺序）
	seen = make(map[string]bool) // 重置

	for _, domain := range domains {
		if seen[domain] {
			continue
		}

		// 如果是通配符域名，先添加对应的根域名
		if strings.HasPrefix(domain, "*.") {
			rootDomain := strings.TrimPrefix(domain, "*.")
			if !seen[rootDomain] {
				result = append(result, rootDomain)
				seen[rootDomain] = true
			}
		}

		result = append(result, domain)
		seen[domain] = true
	}

	return result
}

// Delete deletes a certificate and its associated challenges
func (s *CertificateService) Delete(id uint) error {
	// Check if certificate exists
	var cert model.Certificate
	if err := s.db.First(&cert, id).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	// Delete associated challenges first (due to foreign key constraint)
	if err := s.db.Where("certificate_id = ?", id).Delete(&model.Challenge{}).Error; err != nil {
		return fmt.Errorf("failed to delete challenges: %w", err)
	}

	// Delete the certificate
	if err := s.db.Delete(&cert).Error; err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}
