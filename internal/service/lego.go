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
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Get or create ACME account
	account, err := s.getOrCreateAccount(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get/create ACME account: %w", err)
	}

	// Generate certificate private key
	kt := acme.KeyType(keyType)
	if keySize == 0 {
		keySize = acme.GetDefaultKeySize(kt)
	}
	if err := acme.ValidateKeySize(kt, keySize); err != nil {
		return fmt.Errorf("invalid key size: %w", err)
	}

	certKey, err := acme.GeneratePrivateKey(kt, keySize)
	if err != nil {
		return fmt.Errorf("failed to generate certificate key: %w", err)
	}

	// Encode and encrypt the private key
	keyPEM, err := acme.EncodePrivateKeyPEM(certKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}
	encryptedKey, err := s.encryptor.Encrypt(keyPEM)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create ACME client
	client, err := s.createClientFromAccount(account)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Create order with CA
	order, err := client.CreateOrder(ctx, domains)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Get challenges from authorizations
	var challenges []model.Challenge
	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzURL)
		if err != nil {
			return fmt.Errorf("failed to get authorization: %w", err)
		}

		// Find DNS-01 challenge
		var dns01Challenge *officialAcme.Challenge
		for _, ch := range authz.Challenges {
			if ch.Type == "dns-01" {
				dns01Challenge = ch
				break
			}
		}

		if dns01Challenge == nil {
			return fmt.Errorf("no DNS-01 challenge found for domain %s", authz.Identifier.Value)
		}

		// Compute TXT record value using the client (which has the account key)
		txtValue, err := client.DNS01ChallengeRecord(dns01Challenge.Token)
		if err != nil {
			return fmt.Errorf("failed to compute TXT value: %w", err)
		}

		// Determine TXT host
		domain := authz.Identifier.Value
		txtHost := "_acme-challenge." + strings.TrimPrefix(domain, "*.")

		challenge := model.Challenge{
			CertificateID: certID,
			Domain:        domain,
			TXTHost:       txtHost,
			TXTValue:      txtValue,
			Token:         dns01Challenge.Token,
			AuthzURL:      authzURL,
			ChallengeURL:  dns01Challenge.URI,
			Status:        model.ChallengeStatusPending,
		}
		challenges = append(challenges, challenge)
	}

	// Save challenges to database
	if err := s.saveChallenges(certID, challenges); err != nil {
		return fmt.Errorf("failed to save challenges: %w", err)
	}

	// Update certificate with account, key, and order info
	if err := s.db.Model(&model.Certificate{}).Where("id = ?", certID).Updates(map[string]any{
		"account_id": account.ID,
		"key_pem":    encryptedKey,
		"key_size":   keySize,
		"order_url":  order.URI,
	}).Error; err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	return nil
}

// PreVerifyDNS checks if DNS TXT records are correctly set up for all challenges.
func (s *LegoService) PreVerifyDNS(certID uint) ([]acme.DNSCheckResult, bool, error) {
	var challenges []model.Challenge
	if err := s.db.Where("certificate_id = ?", certID).Find(&challenges).Error; err != nil {
		return nil, false, fmt.Errorf("failed to get challenges: %w", err)
	}

	if len(challenges) == 0 {
		return nil, false, fmt.Errorf("no challenges found for certificate %d", certID)
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var cert model.Certificate
	if err := s.db.Preload("Challenges").First(&cert, certID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status == model.CertificateStatusReady {
		return nil // Already finalized
	}

	if cert.OrderURL == "" {
		return fmt.Errorf("no order URL found for certificate")
	}

	// Get account
	if cert.AccountID == nil {
		return fmt.Errorf("no account associated with certificate")
	}

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
			return fmt.Errorf("failed to accept challenge for %s: %w", ch.Domain, err)
		}
	}

	// Wait for order to be ready
	order, err := client.WaitOrder(ctx, cert.OrderURL)
	if err != nil {
		s.db.Model(&cert).Updates(map[string]any{
			"status": model.CertificateStatusFailed,
		})
		return fmt.Errorf("failed to wait for order: %w", err)
	}

	if order.Status != officialAcme.StatusReady {
		s.db.Model(&cert).Updates(map[string]any{
			"status": model.CertificateStatusFailed,
		})
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

	// Get domains from certificate
	var domains []string
	if err := json.Unmarshal([]byte(cert.Domains), &domains); err != nil {
		return fmt.Errorf("failed to parse domains: %w", err)
	}

	// Create CSR
	csr, err := createCSR(domains, certKey)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	// Finalize order and get certificate
	certChain, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr)
	if err != nil {
		s.db.Model(&cert).Updates(map[string]any{
			"status": model.CertificateStatusFailed,
		})
		return fmt.Errorf("failed to create order cert: %w", err)
	}

	if len(certChain) == 0 {
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
func (s *LegoService) getOrCreateAccount(ctx context.Context, email string) (*model.ACMEAccount, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	caURL := acme.LetsEncryptProductionV2

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
	client := acme.NewClientV2(accountKey.(crypto.Signer), email)
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

	return acme.NewClientV2(accountKey.(crypto.Signer), account.Email), nil
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

func createZipBundleV2(cert model.Certificate, keyPEM []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Add certificate
	certFile, err := w.Create("certificate.pem")
	if err != nil {
		return nil, err
	}
	if _, err := certFile.Write([]byte(cert.CertPEM)); err != nil {
		return nil, err
	}

	// Add fullchain
	chainFile, err := w.Create("fullchain.pem")
	if err != nil {
		return nil, err
	}
	if _, err := chainFile.Write([]byte(cert.ChainPEM)); err != nil {
		return nil, err
	}

	// Add private key
	keyFile, err := w.Create("private.key")
	if err != nil {
		return nil, err
	}
	if _, err := keyFile.Write(keyPEM); err != nil {
		return nil, err
	}

	// Add README
	readmeFile, err := w.Create("README.txt")
	if err != nil {
		return nil, err
	}
	readme := fmt.Sprintf(`SSL Certificate Bundle
======================

Files included:
- certificate.pem: Your SSL certificate
- fullchain.pem: Certificate + intermediate CA certificates
- private.key: Your private key (keep this secure!)

Domains: %s
Issued: %s
Expires: %s

For Nginx:
  ssl_certificate /path/to/fullchain.pem;
  ssl_certificate_key /path/to/private.key;

For Apache:
  SSLCertificateFile /path/to/certificate.pem
  SSLCertificateKeyFile /path/to/private.key
  SSLCertificateChainFile /path/to/fullchain.pem
`, cert.Domains, cert.IssuedAt, cert.ExpiresAt)
	if _, err := readmeFile.Write([]byte(readme)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
