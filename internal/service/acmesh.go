package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/imkerbos/ACME-Console/internal/model"
)

// AcmeShService handles integration with acme.sh CLI
// For MVP, this provides mock implementations
type AcmeShService struct {
	// In real implementation, this would contain acme.sh path and config
}

func NewAcmeShService() *AcmeShService {
	return &AcmeShService{}
}

// GenerateChallenges creates DNS-01 challenge records for the given domains
// For MVP, this generates mock challenge tokens
func (s *AcmeShService) GenerateChallenges(certID uint, domains []string) ([]model.Challenge, error) {
	var challenges []model.Challenge

	for _, domain := range domains {
		// Generate a mock challenge token
		token, err := generateRandomToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		// For wildcard domains, remove the *. prefix for the TXT host
		baseDomain, _ := strings.CutPrefix(domain, "*.")

		challenge := model.Challenge{
			CertificateID: certID,
			Domain:        domain,
			TXTHost:       fmt.Sprintf("_acme-challenge.%s", baseDomain),
			TXTValue:      token,
			Status:        model.ChallengeStatusPending,
		}

		challenges = append(challenges, challenge)
	}

	return challenges, nil
}

// VerifyChallenges checks if all DNS-01 challenges are properly configured
// For MVP, this always returns true (mock implementation)
func (s *AcmeShService) VerifyChallenges(cert *model.Certificate) (bool, error) {
	// In real implementation:
	// 1. Query DNS for each TXT record
	// 2. Compare with expected values
	// 3. Call acme.sh to verify with CA

	// MVP: Always return success
	return true, nil
}

// ExportTXTTemplate generates a human-readable template for DNS TXT records
func (s *AcmeShService) ExportTXTTemplate(challenges []model.Challenge) string {
	var sb strings.Builder

	sb.WriteString("# DNS TXT Records for ACME Challenge\n")
	sb.WriteString("# Add these records to your DNS configuration\n")
	sb.WriteString("#\n")
	sb.WriteString("# Format: HOST TTL IN TXT \"VALUE\"\n")
	sb.WriteString("#\n\n")

	for _, ch := range challenges {
		sb.WriteString(fmt.Sprintf("# Domain: %s\n", ch.Domain))
		sb.WriteString(fmt.Sprintf("%s. 300 IN TXT \"%s\"\n\n", ch.TXTHost, ch.TXTValue))
	}

	sb.WriteString("# After adding these records, wait for DNS propagation (usually 5-10 minutes)\n")
	sb.WriteString("# Then trigger verification via the API\n")

	return sb.String()
}

// generateRandomToken creates a base64url-encoded random token
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
