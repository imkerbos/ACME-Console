package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/imkerbos/ACME-Console/internal/acme"
	internalCrypto "github.com/imkerbos/ACME-Console/internal/crypto"
	"github.com/imkerbos/ACME-Console/internal/model"
	"gorm.io/gorm"
	"software.sslmate.com/src/go-pkcs12"
	officialAcme "golang.org/x/crypto/acme"
)

// LegoService handles ACME certificate operations
type LegoService struct {
	db         *gorm.DB
	settingSvc *SettingService
	encryptor  *internalCrypto.Encryptor
}

// NewLegoServiceWithSettings creates a new LegoService with database-based settings
func NewLegoServiceWithSettings(db *gorm.DB, settingSvc *SettingService, encryptor *internalCrypto.Encryptor) *LegoService {
	return &LegoService{
		db:         db,
		settingSvc: settingSvc,
		encryptor:  encryptor,
	}
}

// CreateOrder creates a new certificate order with the ACME CA.
// This generates a private key, creates an order, and stores challenges for user DNS setup.
func (s *LegoService) CreateOrder(certID uint, email string, domains []string, keyType string, keySize int) error {
	return s.CreateOrderWithOptions(certID, email, domains, keyType, keySize, string(model.ChallengeTypeDNS01), "")
}

// CreateOrderWithOptions creates a new certificate order with challenge type and CA environment options.
func (s *LegoService) CreateOrderWithOptions(certID uint, email string, domains []string, keyType string, keySize int, challengeType string, caEnv string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	if challengeType == "" {
		challengeType = string(model.ChallengeTypeDNS01)
	}

	// Get or create ACME account
	account, err := s.getOrCreateAccount(ctx, email, caEnv)
	if err != nil {
		s.setError(certID, fmt.Sprintf("Failed to get/create ACME account: %v", err))
		return fmt.Errorf("failed to get/create ACME account: %w", err)
	}

	// Generate certificate private key
	kt := acme.KeyType(keyType)
	if keySize == 0 {
		keySize = acme.GetDefaultKeySize(kt)
	}
	if err := acme.ValidateKeySize(kt, keySize); err != nil {
		s.setError(certID, fmt.Sprintf("Invalid key size: %v", err))
		return fmt.Errorf("invalid key size: %w", err)
	}

	certKey, err := acme.GeneratePrivateKey(kt, keySize)
	if err != nil {
		s.setError(certID, fmt.Sprintf("Failed to generate private key: %v", err))
		return fmt.Errorf("failed to generate certificate key: %w", err)
	}

	// Encode and encrypt the private key
	keyPEM, err := acme.EncodePrivateKeyPEM(certKey)
	if err != nil {
		s.setError(certID, fmt.Sprintf("Failed to encode private key: %v", err))
		return fmt.Errorf("failed to encode private key: %w", err)
	}
	encryptedKey, err := s.encryptor.Encrypt(keyPEM)
	if err != nil {
		s.setError(certID, fmt.Sprintf("Failed to encrypt private key: %v", err))
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create ACME client
	client, err := s.createClientFromAccount(account)
	if err != nil {
		s.setError(certID, fmt.Sprintf("Failed to create ACME client: %v", err))
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Create order with CA
	order, err := client.CreateOrder(ctx, domains)
	if err != nil {
		errMsg := s.formatACMEError(err)
		s.setError(certID, errMsg)
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Get challenges from authorizations
	var challenges []model.Challenge
	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzURL)
		if err != nil {
			s.setError(certID, fmt.Sprintf("Failed to get authorization: %v", err))
			return fmt.Errorf("failed to get authorization: %w", err)
		}

		switch challengeType {
		case string(model.ChallengeTypeHTTP01):
			ch, err := s.buildHTTP01Challenge(client, authz, authzURL, certID)
			if err != nil {
				s.setError(certID, err.Error())
				return err
			}
			challenges = append(challenges, ch)

		default: // dns-01
			ch, err := s.buildDNS01Challenge(client, authz, authzURL, certID)
			if err != nil {
				s.setError(certID, err.Error())
				return err
			}
			challenges = append(challenges, ch)
		}
	}

	// Save challenges to database
	if err := s.saveChallenges(certID, challenges); err != nil {
		s.setError(certID, fmt.Sprintf("Failed to save challenges: %v", err))
		return fmt.Errorf("failed to save challenges: %w", err)
	}

	// Compute order expiration (Let's Encrypt orders expire in ~7 days)
	var orderExpiresAt *time.Time
	if !order.Expires.IsZero() {
		orderExpiresAt = &order.Expires
	} else {
		// Default to 7 days from now if CA doesn't provide expiry
		t := time.Now().Add(7 * 24 * time.Hour)
		orderExpiresAt = &t
	}

	// Update certificate with account, key, and order info
	if err := s.db.Model(&model.Certificate{}).Where("id = ?", certID).Updates(map[string]any{
		"account_id":       account.ID,
		"key_pem":          encryptedKey,
		"key_size":         keySize,
		"order_url":        order.URI,
		"order_expires_at": orderExpiresAt,
		"error_message":    "", // Clear previous error on success
		"status":           model.CertificateStatusPending,
	}).Error; err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	return nil
}

// buildDNS01Challenge creates a DNS-01 challenge model from an authorization.
func (s *LegoService) buildDNS01Challenge(client *acme.ClientV2, authz *officialAcme.Authorization, authzURL string, certID uint) (model.Challenge, error) {
	var dns01Challenge *officialAcme.Challenge
	for _, ch := range authz.Challenges {
		if ch.Type == "dns-01" {
			dns01Challenge = ch
			break
		}
	}

	if dns01Challenge == nil {
		return model.Challenge{}, fmt.Errorf("no DNS-01 challenge found for domain %s", authz.Identifier.Value)
	}

	txtValue, err := client.DNS01ChallengeRecord(dns01Challenge.Token)
	if err != nil {
		return model.Challenge{}, fmt.Errorf("failed to compute TXT value: %w", err)
	}

	domain := authz.Identifier.Value
	txtHost := "_acme-challenge." + strings.TrimPrefix(domain, "*.")

	return model.Challenge{
		CertificateID: certID,
		Domain:        domain,
		Type:          "dns-01",
		TXTHost:       txtHost,
		TXTValue:      txtValue,
		Token:         dns01Challenge.Token,
		AuthzURL:      authzURL,
		ChallengeURL:  dns01Challenge.URI,
		Status:        model.ChallengeStatusPending,
	}, nil
}

// buildHTTP01Challenge creates an HTTP-01 challenge model from an authorization.
func (s *LegoService) buildHTTP01Challenge(client *acme.ClientV2, authz *officialAcme.Authorization, authzURL string, certID uint) (model.Challenge, error) {
	var http01Challenge *officialAcme.Challenge
	for _, ch := range authz.Challenges {
		if ch.Type == "http-01" {
			http01Challenge = ch
			break
		}
	}

	if http01Challenge == nil {
		return model.Challenge{}, fmt.Errorf("no HTTP-01 challenge found for domain %s (wildcard domains only support DNS-01)", authz.Identifier.Value)
	}

	httpPath := client.HTTP01ChallengePath(http01Challenge.Token)
	httpContent, err := client.HTTP01ChallengeResponse(http01Challenge.Token)
	if err != nil {
		return model.Challenge{}, fmt.Errorf("failed to compute HTTP-01 response: %w", err)
	}

	return model.Challenge{
		CertificateID: certID,
		Domain:        authz.Identifier.Value,
		Type:          "http-01",
		HTTPPath:      httpPath,
		HTTPContent:   httpContent,
		Token:         http01Challenge.Token,
		AuthzURL:      authzURL,
		ChallengeURL:  http01Challenge.URI,
		Status:        model.ChallengeStatusPending,
	}, nil
}

// RetryOrder re-creates an ACME order for a failed certificate.
// It generates new challenges while preserving the certificate metadata.
func (s *LegoService) RetryOrder(certID uint) error {
	var cert model.Certificate
	if err := s.db.First(&cert, certID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	var domains []string
	if err := json.Unmarshal([]byte(cert.Domains), &domains); err != nil {
		return fmt.Errorf("failed to parse domains: %w", err)
	}

	// Re-create the order with existing parameters
	return s.CreateOrderWithOptions(certID, cert.Email, domains, string(cert.KeyType), cert.KeySize, string(cert.ChallengeMode), string(cert.CAEnv))
}

// PreVerifyDNS checks if DNS TXT records are correctly set up for all challenges.
func (s *LegoService) PreVerifyDNS(certID uint) ([]acme.DNSCheckResult, bool, error) {
	var challenges []model.Challenge
	if err := s.db.Where("certificate_id = ? AND type = ?", certID, "dns-01").Find(&challenges).Error; err != nil {
		return nil, false, fmt.Errorf("failed to get challenges: %w", err)
	}

	if len(challenges) == 0 {
		return nil, false, fmt.Errorf("no DNS-01 challenges found for certificate %d", certID)
	}

	checks := make([]struct {
		Domain        string
		TXTHost       string
		ExpectedValue string
	}, len(challenges))

	for i, ch := range challenges {
		checks[i] = struct {
			Domain        string
			TXTHost       string
			ExpectedValue string
		}{
			Domain:        ch.Domain,
			TXTHost:       ch.TXTHost,
			ExpectedValue: ch.TXTValue,
		}
	}

	// Dynamic timeout based on challenge count: base 30s + 10s per challenge, max 180s
	timeout := time.Duration(30+10*len(challenges)) * time.Second
	if timeout > 180*time.Second {
		timeout = 180 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dnsChecker := s.getDNSChecker()
	results := dnsChecker.CheckMultipleTXTRecords(ctx, checks)

	// Update challenge DNS check status
	now := time.Now()
	for i, result := range results {
		updates := map[string]any{
			"dns_checked_at": &now,
			"dns_check_ok":   result.Matched,
		}
		s.db.Model(&model.Challenge{}).Where("id = ?", challenges[i].ID).Updates(updates)
	}

	allMatched := acme.AllMatched(results)
	return results, allMatched, nil
}

// FinalizeOrder completes the certificate order after DNS verification.
func (s *LegoService) FinalizeOrder(certID uint) error {
	var cert model.Certificate
	if err := s.db.Preload("Challenges").First(&cert, certID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status == model.CertificateStatusReady {
		return nil // Already finalized
	}

	if cert.OrderURL == "" {
		s.setError(certID, "No ACME order URL found. Please click Retry to create a new order.")
		return fmt.Errorf("no order URL found for certificate")
	}

	// Check if order has expired
	if cert.OrderExpiresAt != nil && time.Now().After(*cert.OrderExpiresAt) {
		errMsg := "ACME order has expired. Please click Retry to create a new order."
		s.setError(certID, errMsg)
		s.db.Model(&cert).Update("status", model.CertificateStatusFailed)
		return fmt.Errorf("order has expired at %s", cert.OrderExpiresAt.Format(time.RFC3339))
	}

	// Get account
	if cert.AccountID == nil {
		s.setError(certID, "No ACME account associated with this certificate.")
		return fmt.Errorf("no account associated with certificate")
	}

	// Dynamic timeout based on domain count: base 120s + 30s per domain, max 600s
	var domains []string
	if err := json.Unmarshal([]byte(cert.Domains), &domains); err != nil {
		return fmt.Errorf("failed to parse domains: %w", err)
	}
	domainCount := len(domains)
	if domainCount < 1 {
		domainCount = 1
	}
	timeout := time.Duration(120+30*domainCount) * time.Second
	if timeout > 600*time.Second {
		timeout = 600 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var account model.ACMEAccount
	if err := s.db.First(&account, *cert.AccountID).Error; err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// Create ACME client
	client, err := s.createClientFromAccount(&account)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Accept all challenges (tell CA to verify)
	for _, ch := range cert.Challenges {
		if ch.ChallengeURL == "" {
			continue
		}

		challenge := &officialAcme.Challenge{
			URI:   ch.ChallengeURL,
			Token: ch.Token,
		}

		_, err := client.AcceptChallenge(ctx, challenge)
		if err != nil {
			errMsg := s.formatACMEError(err)
			s.setError(certID, fmt.Sprintf("Challenge failed for %s: %s", ch.Domain, errMsg))
			s.db.Model(&cert).Update("status", model.CertificateStatusFailed)
			return fmt.Errorf("failed to accept challenge for %s: %w", ch.Domain, err)
		}
	}

	// Wait for order to be ready
	order, err := client.WaitOrder(ctx, cert.OrderURL)
	if err != nil {
		// Fetch per-domain authorization details to get specific failure reasons
		errMsg := s.getAuthzErrorDetails(ctx, client, cert.Challenges)
		if errMsg == "" {
			errMsg = s.formatACMEError(err)
		}
		s.setError(certID, errMsg)
		s.db.Model(&cert).Update("status", model.CertificateStatusFailed)
		return fmt.Errorf("failed to wait for order: %w", err)
	}

	if order.Status != officialAcme.StatusReady {
		errMsg := fmt.Sprintf("Unexpected order status: %s (expected ready)", order.Status)
		s.setError(certID, errMsg)
		s.db.Model(&cert).Update("status", model.CertificateStatusFailed)
		return fmt.Errorf("order is not ready, status: %s", order.Status)
	}

	// Decrypt private key
	keyPEM, err := s.encryptor.Decrypt(cert.KeyPEM)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	certKey, err := acme.DecodePrivateKeyPEM(keyPEM)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	// Create CSR (domains already parsed above for timeout calculation)
	csr, err := createCSR(domains, certKey)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	// Finalize order and get certificate
	certChain, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr)
	if err != nil {
		errMsg := s.formatACMEError(err)
		s.setError(certID, fmt.Sprintf("Failed to issue certificate: %s", errMsg))
		s.db.Model(&cert).Update("status", model.CertificateStatusFailed)
		return fmt.Errorf("failed to create order cert: %w", err)
	}

	if len(certChain) == 0 {
		s.setError(certID, "No certificate returned from CA.")
		return fmt.Errorf("no certificate returned from CA")
	}

	// Encode certificates to PEM
	var certPEM, chainPEM, issuerPEM bytes.Buffer
	for i, certDER := range certChain {
		block := &pem.Block{Type: "CERTIFICATE", Bytes: certDER}
		pem.Encode(&chainPEM, block)
		if i == 0 {
			pem.Encode(&certPEM, block)
		} else {
			pem.Encode(&issuerPEM, block)
		}
	}

	// Parse certificate to get metadata
	certInfo, err := parseCertificateV2(certChain[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Update certificate record
	now := time.Now()
	updates := map[string]any{
		"cert_pem":        certPEM.String(),
		"chain_pem":       chainPEM.String(),
		"issuer_cert_pem": issuerPEM.String(),
		"serial_number":   certInfo.SerialNumber,
		"fingerprint":     certInfo.Fingerprint,
		"issued_at":       &now,
		"expires_at":      &certInfo.NotAfter,
		"status":          model.CertificateStatusReady,
		"error_message":   "", // Clear error on success
	}

	if err := s.db.Model(&cert).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	// Update challenges to verified
	if err := s.db.Model(&model.Challenge{}).Where("certificate_id = ?", certID).Updates(map[string]any{
		"status":       model.ChallengeStatusVerified,
		"validated_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("failed to update challenges: %w", err)
	}

	return nil
}

// RevokeCertificate revokes a previously issued certificate with the CA.
func (s *LegoService) RevokeCertificate(certID uint) error {
	var cert model.Certificate
	if err := s.db.First(&cert, certID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status != model.CertificateStatusReady {
		return fmt.Errorf("can only revoke certificates in ready status")
	}

	if cert.CertPEM == "" {
		return fmt.Errorf("no certificate PEM found")
	}

	if cert.AccountID == nil {
		return fmt.Errorf("no account associated with certificate")
	}

	var account model.ACMEAccount
	if err := s.db.First(&account, *cert.AccountID).Error; err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	client, err := s.createClientFromAccount(&account)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.RevokeCert(ctx, []byte(cert.CertPEM)); err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}

	// Update certificate status
	now := time.Now()
	s.db.Model(&cert).Updates(map[string]any{
		"status":        model.CertificateStatusRevoked,
		"error_message": fmt.Sprintf("Certificate revoked at %s", now.Format("2006-01-02 15:04:05")),
		"auto_renew":    false,
	})

	return nil
}

// IsOrderExpired checks if the ACME order for a certificate has expired.
func (s *LegoService) IsOrderExpired(certID uint) (bool, *time.Time) {
	var cert model.Certificate
	if err := s.db.Select("order_expires_at").First(&cert, certID).Error; err != nil {
		return false, nil
	}
	if cert.OrderExpiresAt == nil {
		return false, nil
	}
	return time.Now().After(*cert.OrderExpiresAt), cert.OrderExpiresAt
}

// DownloadFormat represents the available certificate download formats
type DownloadFormat string

const (
	DownloadFormatPEM       DownloadFormat = "pem"
	DownloadFormatFullChain DownloadFormat = "fullchain"
	DownloadFormatPFX       DownloadFormat = "pfx"
	DownloadFormatZIP       DownloadFormat = "zip"
)

// GetCertificateBundle returns the certificate in the specified format.
func (s *LegoService) GetCertificateBundle(certID uint, format DownloadFormat, password string) ([]byte, string, error) {
	var cert model.Certificate
	if err := s.db.First(&cert, certID).Error; err != nil {
		return nil, "", fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status != model.CertificateStatusReady {
		return nil, "", fmt.Errorf("certificate is not ready")
	}

	// Decrypt private key
	keyPEM, err := s.encryptor.Decrypt(cert.KeyPEM)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	switch format {
	case DownloadFormatPEM:
		return []byte(cert.CertPEM), "certificate.pem", nil

	case DownloadFormatFullChain:
		return []byte(cert.ChainPEM), "fullchain.pem", nil

	case DownloadFormatPFX:
		if password == "" {
			password = "changeit"
		}
		pfxData, err := createPFXV2([]byte(cert.CertPEM), keyPEM, password)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create PFX: %w", err)
		}
		return pfxData, "certificate.pfx", nil

	case DownloadFormatZIP:
		zipData, err := createZipBundleV2(cert, keyPEM)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create ZIP: %w", err)
		}
		return zipData, "certificate.zip", nil

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// getOrCreateAccount gets an existing ACME account or creates a new one.
func (s *LegoService) getOrCreateAccount(ctx context.Context, email string, caEnv string) (*model.ACMEAccount, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	caURL := acme.GetCAURL(caEnv)

	// Try to find existing account
	var account model.ACMEAccount
	err := s.db.Where("email = ? AND ca_url = ?", email, caURL).First(&account).Error
	if err == nil {
		return &account, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Create new account
	accountKey, err := acme.GeneratePrivateKey(acme.KeyTypeECC, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate account key: %w", err)
	}

	keyPEM, err := acme.EncodePrivateKeyPEM(accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode account key: %w", err)
	}

	encryptedKey, err := s.encryptor.Encrypt(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt account key: %w", err)
	}

	// Register with CA
	client := acme.NewClientV2WithCA(accountKey.(crypto.Signer), email, caURL)
	if err := client.Register(ctx); err != nil {
		return nil, fmt.Errorf("failed to register account: %w", err)
	}

	account = model.ACMEAccount{
		Email:      email,
		CAURL:      caURL,
		PrivateKey: encryptedKey,
	}

	if err := s.db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return &account, nil
}

func (s *LegoService) createClientFromAccount(account *model.ACMEAccount) (*acme.ClientV2, error) {
	// Decrypt account key
	keyPEM, err := s.encryptor.Decrypt(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt account key: %w", err)
	}

	accountKey, err := acme.DecodePrivateKeyPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode account key: %w", err)
	}

	return acme.NewClientV2WithCA(accountKey.(crypto.Signer), account.Email, account.CAURL), nil
}

func (s *LegoService) saveChallenges(certID uint, challenges []model.Challenge) error {
	// Delete existing challenges for this certificate
	if err := s.db.Where("certificate_id = ?", certID).Delete(&model.Challenge{}).Error; err != nil {
		return err
	}

	// Insert new challenges
	for i := range challenges {
		challenges[i].CertificateID = certID
		if err := s.db.Create(&challenges[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *LegoService) getDNSChecker() *acme.DNSChecker {
	settings := s.settingSvc.GetACMEConfig()
	resolvers := acme.ParseResolvers(settings.DNSResolvers)
	timeout, _ := time.ParseDuration(settings.DNSTimeout)
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return acme.NewDNSChecker(resolvers, timeout)
}

// setError updates the error_message field on a certificate.
func (s *LegoService) setError(certID uint, msg string) {
	s.db.Model(&model.Certificate{}).Where("id = ?", certID).Update("error_message", msg)
}

// getAuthzErrorDetails fetches authorization details from ACME and builds a
// user-friendly error message listing per-domain failure reasons.
func (s *LegoService) getAuthzErrorDetails(ctx context.Context, client *acme.ClientV2, challenges []model.Challenge) string {
	var details []string
	for _, ch := range challenges {
		if ch.AuthzURL == "" {
			continue
		}
		authz, err := client.GetAuthorization(ctx, ch.AuthzURL)
		if err != nil {
			continue
		}
		if authz.Status == "valid" {
			continue
		}
		// Find the challenge error for our challenge type
		for _, ac := range authz.Challenges {
			if ac.Type != ch.Type {
				continue
			}
			if ac.Error != nil {
				detail := ac.Error.Error()
				if acmeErr, ok := ac.Error.(*officialAcme.Error); ok && acmeErr.Detail != "" {
					detail = acmeErr.Detail
				}
				details = append(details, fmt.Sprintf("%s: %s", ch.Domain, detail))
			} else if ac.Status == "invalid" {
				details = append(details, fmt.Sprintf("%s: validation failed", ch.Domain))
			}
			break
		}
	}

	if len(details) == 0 {
		return ""
	}

	return "Domain validation failed:\n" + strings.Join(details, "\n")
}

// formatACMEError extracts a user-friendly error message from ACME errors.
func (s *LegoService) formatACMEError(err error) string {
	if err == nil {
		return ""
	}

	// Check for rate limit errors
	if acmeErr, ok := err.(*officialAcme.Error); ok {
		switch acmeErr.StatusCode {
		case 429:
			return fmt.Sprintf("CA rate limit: %s. Please try again later.", acmeErr.Detail)
		default:
			if acmeErr.Detail != "" {
				return acmeErr.Detail
			}
		}
	}

	return err.Error()
}

// Helper functions

func createCSR(domains []string, key crypto.PrivateKey) ([]byte, error) {
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domains[0],
		},
		DNSNames: domains,
	}

	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("key does not implement crypto.Signer")
	}

	return x509.CreateCertificateRequest(rand.Reader, template, signer)
}

type certInfoV2 struct {
	SerialNumber string
	Fingerprint  string
	NotAfter     time.Time
}

func parseCertificateV2(certDER []byte) (*certInfoV2, error) {
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	fingerprint := sha256.Sum256(cert.Raw)

	return &certInfoV2{
		SerialNumber: formatSerialNumberV2(cert.SerialNumber),
		Fingerprint:  hex.EncodeToString(fingerprint[:]),
		NotAfter:     cert.NotAfter,
	}, nil
}

func formatSerialNumberV2(serial *big.Int) string {
	return strings.ToUpper(hex.EncodeToString(serial.Bytes()))
}

func createPFXV2(certPEM, keyPEM []byte, password string) ([]byte, error) {
	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	key, err := acme.DecodePrivateKeyPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	pfxData, err := pkcs12.Modern.Encode(key, cert, nil, password)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PFX: %w", err)
	}

	return pfxData, nil
}

// zipDirEntry represents a merged directory entry for the ZIP bundle.
// Wildcard + root domain pairs (e.g. *.example.com + example.com) are merged
// into a single directory named after the root domain.
type zipDirEntry struct {
	DirName string   // directory name in ZIP
	Domains []string // domains covered by this directory
}

// mergeDomainsForZip groups domains for ZIP directory creation.
// *.example.com and example.com are merged into one "example.com" directory.
// Standalone domains (e.g. api.other.com) get their own directory.
func mergeDomainsForZip(domains []string) []zipDirEntry {
	// Build a set for quick lookup
	domainSet := make(map[string]bool, len(domains))
	for _, d := range domains {
		domainSet[d] = true
	}

	seen := make(map[string]bool)
	var entries []zipDirEntry

	for _, domain := range domains {
		if seen[domain] {
			continue
		}

		if strings.HasPrefix(domain, "*.") {
			root := strings.TrimPrefix(domain, "*.")
			// Merge wildcard + root into one entry
			seen[domain] = true
			seen[root] = true
			entry := zipDirEntry{DirName: root}
			if domainSet[root] {
				entry.Domains = []string{root, domain}
			} else {
				entry.Domains = []string{domain}
			}
			entries = append(entries, entry)
		} else {
			// Check if this root domain's wildcard was already processed
			if seen[domain] {
				continue
			}
			wildcard := "*." + domain
			if domainSet[wildcard] {
				// Will be handled when we process the wildcard
				continue
			}
			// Standalone domain
			seen[domain] = true
			entries = append(entries, zipDirEntry{
				DirName: domain,
				Domains: []string{domain},
			})
		}
	}

	return entries
}

func createZipBundleV2(cert model.Certificate, keyPEM []byte) ([]byte, error) {
	// Parse domains
	var allDomains []string
	if err := json.Unmarshal([]byte(cert.Domains), &allDomains); err != nil {
		allDomains = []string{"certificate"}
	}

	// For a combined (SAN) certificate, all domains share the same cert/key files.
	// Use only the first domain as the directory name to avoid redundant copies.
	entries := mergeDomainsForZip(allDomains)
	dirName := "certificate"
	if len(entries) > 0 {
		dirName = entries[0].DirName
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Create a single directory with the certificate files
	certFile, err := w.Create(dirName + "/certificate.pem")
	if err != nil {
		return nil, err
	}
	if _, err := certFile.Write([]byte(cert.CertPEM)); err != nil {
		return nil, err
	}

	chainFile, err := w.Create(dirName + "/fullchain.pem")
	if err != nil {
		return nil, err
	}
	if _, err := chainFile.Write([]byte(cert.ChainPEM)); err != nil {
		return nil, err
	}

	keyFile, err := w.Create(dirName + "/private.key")
	if err != nil {
		return nil, err
	}
	if _, err := keyFile.Write(keyPEM); err != nil {
		return nil, err
	}

	// Add README at root level listing all covered domains
	readmeFile, err := w.Create("README.txt")
	if err != nil {
		return nil, err
	}

	var domainList strings.Builder
	for _, d := range allDomains {
		domainList.WriteString("  - " + d + "\n")
	}

	readme := fmt.Sprintf(`SSL Certificate Bundle
======================

This is a multi-domain (SAN) certificate covering %d domains:
%s
All domains share the same certificate files, located in:
  %s/

Directory contents:
  - certificate.pem: SSL certificate
  - fullchain.pem:   Certificate + intermediate CA chain
  - private.key:     Private key (keep this secure!)

Issued:  %s
Expires: %s

For Nginx:
  ssl_certificate     /path/to/%s/fullchain.pem;
  ssl_certificate_key /path/to/%s/private.key;

For Apache:
  SSLCertificateFile      /path/to/%s/certificate.pem
  SSLCertificateKeyFile   /path/to/%s/private.key
  SSLCertificateChainFile /path/to/%s/fullchain.pem
`, len(allDomains), domainList.String(), dirName, cert.IssuedAt, cert.ExpiresAt,
		dirName, dirName, dirName, dirName, dirName)

	if _, err := readmeFile.Write([]byte(readme)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
