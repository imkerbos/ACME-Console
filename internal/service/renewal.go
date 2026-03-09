package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/imkerbos/ACME-Console/internal/logger"
	"github.com/imkerbos/ACME-Console/internal/model"
	"gorm.io/gorm"
)

// RenewalService handles certificate auto-renewal logic
type RenewalService struct {
	db              *gorm.DB
	certSvc         *CertificateService
	notificationSvc *NotificationService
	settingSvc      *SettingService
}

// NewRenewalService creates a new RenewalService
func NewRenewalService(db *gorm.DB, certSvc *CertificateService, notificationSvc *NotificationService, settingSvc *SettingService) *RenewalService {
	return &RenewalService{
		db:              db,
		certSvc:         certSvc,
		notificationSvc: notificationSvc,
		settingSvc:      settingSvc,
	}
}

// CheckAndRenew is the main entry point called by the scheduler every 6 hours.
// It processes certificates through the renewal state machine.
func (s *RenewalService) CheckAndRenew() error {
	if !s.isRenewalEnabled() {
		logger.Info("Certificate renewal is disabled globally")
		return nil
	}

	maxAttempts := s.getMaxAttempts()

	// 1. idle + about to expire → initiate renewal
	if err := s.processIdleCertificates(); err != nil {
		logger.Error("Failed to process idle certificates", logger.Err(err))
	}

	// 2. pending → check DNS
	if err := s.processPendingCertificates(); err != nil {
		logger.Error("Failed to process pending certificates", logger.Err(err))
	}

	// 3. dns_ready → finalize
	if err := s.processDNSReadyCertificates(); err != nil {
		logger.Error("Failed to process dns_ready certificates", logger.Err(err))
	}

	// 4. failed + attempts < max → reset to idle for retry
	if err := s.processFailedCertificates(maxAttempts); err != nil {
		logger.Error("Failed to process failed certificates", logger.Err(err))
	}

	return nil
}

// processIdleCertificates finds certificates with auto_renew=true, renewal_status=idle,
// and expiring within renew_before_days. Initiates renewal by creating a new ACME order.
func (s *RenewalService) processIdleCertificates() error {
	var certs []model.Certificate
	now := time.Now()

	if err := s.db.Where(
		"auto_renew = ? AND renewal_status = ? AND status = ? AND expires_at IS NOT NULL AND expires_at <= ?",
		true, model.RenewalStatusIdle, model.CertificateStatusReady,
		now.AddDate(0, 0, 90), // broad filter, we check renew_before_days per cert
	).Find(&certs).Error; err != nil {
		return fmt.Errorf("failed to query idle certificates: %w", err)
	}

	for _, cert := range certs {
		if cert.ExpiresAt == nil {
			continue
		}
		daysUntilExpiry := int(time.Until(*cert.ExpiresAt).Hours() / 24)
		renewDays := cert.RenewBeforeDays
		if renewDays <= 0 {
			renewDays = 30
		}
		if daysUntilExpiry > renewDays {
			continue
		}

		logger.Info("Initiating renewal for certificate",
			logger.Uint("cert_id", cert.ID),
			logger.Int("days_until_expiry", daysUntilExpiry),
		)

		if err := s.initiateRenewal(&cert); err != nil {
			logger.Error("Failed to initiate renewal",
				logger.Uint("cert_id", cert.ID),
				logger.Err(err),
			)
			s.logRenewal(cert.ID, "initiated", "failed", err.Error(), cert.ExpiresAt, nil)
		}
	}

	return nil
}

// initiateRenewal creates a new ACME order for the certificate, generating new challenges.
// The certificate's main Status stays "ready" so it remains usable during renewal.
func (s *RenewalService) initiateRenewal(cert *model.Certificate) error {
	var domains []string
	if err := json.Unmarshal([]byte(cert.Domains), &domains); err != nil {
		return fmt.Errorf("failed to parse domains: %w", err)
	}

	// Call existing CreateOrder to generate new challenges
	if s.certSvc.useLego && s.certSvc.legoSvc != nil {
		if err := s.certSvc.legoSvc.CreateOrder(cert.ID, cert.Email, domains, string(cert.KeyType), cert.KeySize); err != nil {
			s.updateRenewalStatus(cert.ID, model.RenewalStatusFailed, true)
			return fmt.Errorf("failed to create renewal order: %w", err)
		}
	} else {
		return fmt.Errorf("lego service not available, cannot renew")
	}

	// Update renewal status to pending (waiting for DNS)
	now := time.Now()
	s.db.Model(cert).Updates(map[string]any{
		"renewal_status": model.RenewalStatusPending,
		"last_renewal_at": &now,
	})

	s.logRenewal(cert.ID, "initiated", "success", "Renewal order created, waiting for DNS update", cert.ExpiresAt, nil)

	// Send notification with new TXT records
	s.sendRenewalNotification(cert.ID, "renewal_started")

	return nil
}

// processPendingCertificates checks DNS for certificates waiting on TXT record updates.
// If DNS matches, advances to dns_ready. If pending > 7 days, resets to idle.
func (s *RenewalService) processPendingCertificates() error {
	var certs []model.Certificate
	if err := s.db.Where(
		"auto_renew = ? AND renewal_status = ?",
		true, model.RenewalStatusPending,
	).Find(&certs).Error; err != nil {
		return fmt.Errorf("failed to query pending certificates: %w", err)
	}

	for _, cert := range certs {
		// Check if pending too long (> 7 days) — order expired, reset
		if cert.LastRenewalAt != nil && time.Since(*cert.LastRenewalAt) > 7*24*time.Hour {
			logger.Info("Renewal pending too long, resetting",
				logger.Uint("cert_id", cert.ID),
			)
			s.updateRenewalStatus(cert.ID, model.RenewalStatusIdle, false)
			s.logRenewal(cert.ID, "reset", "success", "Order expired after 7 days, will retry", cert.ExpiresAt, nil)
			continue
		}

		// Check DNS using existing PreVerifyDNS
		_, allOK, err := s.certSvc.PreVerifyDNS(cert.ID)
		if err != nil {
			logger.Error("Failed to pre-verify DNS for renewal",
				logger.Uint("cert_id", cert.ID),
				logger.Err(err),
			)
			continue
		}

		if allOK {
			logger.Info("DNS verified for renewal, advancing to dns_ready",
				logger.Uint("cert_id", cert.ID),
			)
			s.updateRenewalStatus(cert.ID, model.RenewalStatusDNSReady, false)
			s.logRenewal(cert.ID, "dns_ready", "success", "DNS records verified", cert.ExpiresAt, nil)
		}
	}

	return nil
}

// processDNSReadyCertificates finalizes certificates whose DNS has been verified.
func (s *RenewalService) processDNSReadyCertificates() error {
	var certs []model.Certificate
	if err := s.db.Where(
		"auto_renew = ? AND renewal_status = ?",
		true, model.RenewalStatusDNSReady,
	).Find(&certs).Error; err != nil {
		return fmt.Errorf("failed to query dns_ready certificates: %w", err)
	}

	for _, cert := range certs {
		oldExpiresAt := cert.ExpiresAt

		if !s.certSvc.useLego || s.certSvc.legoSvc == nil {
			s.updateRenewalStatus(cert.ID, model.RenewalStatusFailed, true)
			s.logRenewal(cert.ID, "completed", "failed", "lego service not available", oldExpiresAt, nil)
			continue
		}

		if err := s.certSvc.legoSvc.FinalizeOrder(cert.ID); err != nil {
			logger.Error("Failed to finalize renewal",
				logger.Uint("cert_id", cert.ID), logger.Err(err),
			)
			s.updateRenewalStatus(cert.ID, model.RenewalStatusFailed, true)
			s.logRenewal(cert.ID, "completed", "failed", err.Error(), oldExpiresAt, nil)
			s.sendRenewalNotification(cert.ID, "renewal_failed")
			continue
		}

		renewed, err := s.certSvc.GetByID(cert.ID)
		if err != nil {
			logger.Error("Failed to reload renewed cert", logger.Err(err))
			continue
		}

		s.db.Model(&model.Certificate{}).Where("id = ?", cert.ID).Updates(map[string]any{
			"renewal_status":   model.RenewalStatusCompleted,
			"renewal_attempts": 0,
		})

		s.logRenewal(cert.ID, "completed", "success", "Certificate renewed", oldExpiresAt, renewed.ExpiresAt)
		s.sendRenewalNotification(cert.ID, "renewal_completed")
		logger.Info("Certificate renewed successfully", logger.Uint("cert_id", cert.ID))
	}

	return nil
}

// processFailedCertificates resets failed certificates to idle if under max attempts.
func (s *RenewalService) processFailedCertificates(maxAttempts int) error {
	var certs []model.Certificate
	if err := s.db.Where(
		"auto_renew = ? AND renewal_status = ? AND renewal_attempts < ?",
		true, model.RenewalStatusFailed, maxAttempts,
	).Find(&certs).Error; err != nil {
		return fmt.Errorf("failed to query failed certificates: %w", err)
	}

	for _, cert := range certs {
		logger.Info("Resetting failed renewal for retry",
			logger.Uint("cert_id", cert.ID),
			logger.Int("attempts", cert.RenewalAttempts),
		)
		s.updateRenewalStatus(cert.ID, model.RenewalStatusIdle, false)
		s.logRenewal(cert.ID, "reset", "success",
			fmt.Sprintf("Reset for retry (attempt %d/%d)", cert.RenewalAttempts, maxAttempts),
			cert.ExpiresAt, nil)
	}

	return nil
}

// updateRenewalStatus updates the renewal_status and optionally increments attempts.
func (s *RenewalService) updateRenewalStatus(certID uint, status model.RenewalStatus, incrementAttempts bool) {
	updates := map[string]any{"renewal_status": status}
	if incrementAttempts {
		s.db.Model(&model.Certificate{}).Where("id = ?", certID).
			Update("renewal_attempts", gorm.Expr("renewal_attempts + 1"))
	}
	s.db.Model(&model.Certificate{}).Where("id = ?", certID).Updates(updates)
}

// logRenewal creates a RenewalLog entry.
func (s *RenewalService) logRenewal(certID uint, action, status, message string, oldExpires, newExpires *time.Time) {
	entry := model.RenewalLog{
		CertificateID: certID,
		Action:        action,
		Status:        status,
		Message:       message,
		OldExpiresAt:  oldExpires,
		NewExpiresAt:  newExpires,
	}
	if err := s.db.Create(&entry).Error; err != nil {
		logger.Error("Failed to create renewal log", logger.Err(err))
	}
}

// sendRenewalNotification sends a renewal-related notification.
func (s *RenewalService) sendRenewalNotification(certID uint, eventType string) {
	if s.notificationSvc == nil {
		return
	}
	s.notificationSvc.SendRenewalNotification(certID, eventType)
}

// isRenewalEnabled checks the global renewal.enabled setting.
func (s *RenewalService) isRenewalEnabled() bool {
	val := s.settingSvc.Get(model.SettingRenewalEnabled)
	if val == "" {
		return true
	}
	return val == "true"
}

// getMaxAttempts reads the renewal.max_attempts setting.
func (s *RenewalService) getMaxAttempts() int {
	val := s.settingSvc.Get(model.SettingRenewalMaxAttempts)
	if val == "" {
		return 3
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 3
	}
	return n
}

// TriggerRenewal manually triggers renewal for a specific certificate.
func (s *RenewalService) TriggerRenewal(certID uint) error {
	var cert model.Certificate
	if err := s.db.First(&cert, certID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	if cert.Status != model.CertificateStatusReady {
		return fmt.Errorf("certificate is not in ready status")
	}

	s.db.Model(&cert).Updates(map[string]any{
		"renewal_status":   model.RenewalStatusIdle,
		"renewal_attempts": 0,
		"auto_renew":       true,
	})

	return s.initiateRenewal(&cert)
}
