package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imkerbos/ACME-Console/internal/model"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CheckAndNotify checks for expiring certificates and sends notifications
func (s *NotificationService) CheckAndNotify() error {
	// Get all enabled notification configs
	var configs []model.NotificationConfig
	if err := s.db.Where("enabled = ?", true).Find(&configs).Error; err != nil {
		return fmt.Errorf("failed to load notification configs: %w", err)
	}

	for _, config := range configs {
		if err := s.processConfig(&config); err != nil {
			// Log error but continue processing other configs
			fmt.Printf("Error processing config %d: %v\n", config.ID, err)
		}
	}

	return nil
}

func (s *NotificationService) processConfig(config *model.NotificationConfig) error {
	// Get certificates to check
	var certs []model.Certificate
	query := s.db.Where("status = ? AND expires_at IS NOT NULL", model.CertificateStatusReady)

	if config.CertificateID != nil {
		// Specific certificate
		query = query.Where("id = ?", *config.CertificateID)
	} else if config.WorkspaceID != nil {
		// Workspace-level: all certificates in this workspace
		query = query.Where("workspace_id = ?", *config.WorkspaceID)
	} else {
		// Personal certificates (workspace_id IS NULL)
		query = query.Where("workspace_id IS NULL")
	}

	if err := query.Find(&certs).Error; err != nil {
		return fmt.Errorf("failed to load certificates: %w", err)
	}

	for _, cert := range certs {
		if cert.ExpiresAt == nil {
			continue
		}

		// Calculate days until expiry
		daysUntilExpiry := int(time.Until(*cert.ExpiresAt).Hours() / 24)

		// Check if we should notify
		if daysUntilExpiry <= config.NotifyDays && daysUntilExpiry >= 0 {
			// Check if already notified for this period
			if s.hasRecentNotification(cert.ID, config.ID, config.NotifyDays) {
				continue
			}

			// Send notification
			if err := s.sendNotification(config, &cert, daysUntilExpiry); err != nil {
				s.logNotification(cert.ID, config.ID, config.NotifyDays, "failed", err.Error())
			} else {
				s.logNotification(cert.ID, config.ID, config.NotifyDays, "success", "")
			}
		}
	}

	return nil
}

func (s *NotificationService) hasRecentNotification(certID, configID uint, notifyDays int) bool {
	var count int64
	// Check if notified in the last 23 hours (to avoid duplicate daily notifications)
	since := time.Now().Add(-23 * time.Hour)
	s.db.Model(&model.NotificationLog{}).
		Where("certificate_id = ? AND config_id = ? AND notify_days = ? AND notified_at > ? AND status = ?",
			certID, configID, notifyDays, since, "success").
		Count(&count)
	return count > 0
}

func (s *NotificationService) logNotification(certID, configID uint, notifyDays int, status, errorMsg string) {
	log := &model.NotificationLog{
		CertificateID: certID,
		ConfigID:      configID,
		NotifyDays:    notifyDays,
		Status:        status,
		ErrorMessage:  errorMsg,
		NotifiedAt:    time.Now(),
	}
	s.db.Create(log)
}

func (s *NotificationService) sendNotification(config *model.NotificationConfig, cert *model.Certificate, daysLeft int) error {
	switch config.Type {
	case model.NotificationTypeGeneric:
		return s.sendGenericWebhook(config, cert, daysLeft)
	case model.NotificationTypeTelegram:
		return s.sendTelegramNotification(config, cert, daysLeft)
	case model.NotificationTypeLark:
		return s.sendLarkNotification(config, cert, daysLeft)
	default:
		return fmt.Errorf("unsupported notification type: %s", config.Type)
	}
}

// sendGenericWebhook sends a generic webhook notification
func (s *NotificationService) sendGenericWebhook(config *model.NotificationConfig, cert *model.Certificate, daysLeft int) error {
	domains := s.parseDomains(cert.Domains)

	// Determine urgency level
	urgency := "low"
	if daysLeft <= 3 {
		urgency = "critical"
	} else if daysLeft <= 7 {
		urgency = "high"
	} else if daysLeft <= 14 {
		urgency = "medium"
	}

	payload := map[string]interface{}{
		"event":      "certificate_expiring",
		"cert_id":    cert.ID,
		"domains":    domains,
		"days_left":  daysLeft,
		"expires_at": cert.ExpiresAt.Format(time.RFC3339),
		"status":     cert.Status,
		"urgency":    urgency,
		"message":    fmt.Sprintf("Certificate for %s will expire in %d days", domains[0], daysLeft),
		"timestamp":  time.Now().Unix(),
	}

	return s.sendHTTPPost(config.WebhookURL, payload)
}

// sendTelegramNotification sends a Telegram notification
func (s *NotificationService) sendTelegramNotification(config *model.NotificationConfig, cert *model.Certificate, daysLeft int) error {
	domains := s.parseDomains(cert.Domains)
	domainsText := ""
	for _, d := range domains {
		domainsText += fmt.Sprintf("  <code>%s</code>\n", d)
	}

	// Determine alert emoji based on urgency
	alertEmoji := "🟡"
	urgencyText := "Medium"
	if daysLeft <= 3 {
		alertEmoji = "🔴"
		urgencyText = "Critical"
	} else if daysLeft <= 7 {
		alertEmoji = "🟠"
		urgencyText = "High"
	} else if daysLeft > 14 {
		alertEmoji = "🟢"
		urgencyText = "Low"
	}

	// Beautiful Telegram message with HTML format
	message := fmt.Sprintf(
		"<b>🔐 SSL Certificate Expiry Alert</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━\n\n"+
			"%s <b>%s Priority</b>\n\n"+
			"⏰ <b>Expires in:</b> <code>%d days</code>\n"+
			"📅 <b>Expiry Date:</b> <code>%s</code>\n"+
			"🆔 <b>Certificate:</b> <code>#%d</code>\n\n"+
			"🌐 <b>Domains:</b>\n%s\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"<i>🤖 ACME Console</i>",
		alertEmoji,
		urgencyText,
		daysLeft,
		cert.ExpiresAt.Format("2006-01-02 15:04"),
		cert.ID,
		domainsText,
	)

	payload := map[string]any{
		"text":       message,
		"parse_mode": "HTML",
	}

	// Extract chat_id from config if provided
	var webhookConfig map[string]any
	if config.WebhookConfig != "" {
		if err := json.Unmarshal([]byte(config.WebhookConfig), &webhookConfig); err != nil {
			return fmt.Errorf("failed to parse webhook config: %w", err)
		}
		if chatID, ok := webhookConfig["chat_id"]; ok {
			payload["chat_id"] = chatID
		} else {
			return fmt.Errorf("chat_id is required for Telegram notifications")
		}
	} else {
		return fmt.Errorf("webhook config is required for Telegram notifications")
	}

	return s.sendHTTPPost(config.WebhookURL, payload)
}

// sendLarkNotification sends a Lark (Feishu) notification
func (s *NotificationService) sendLarkNotification(config *model.NotificationConfig, cert *model.Certificate, daysLeft int) error {
	domains := s.parseDomains(cert.Domains)
	domainsText := ""
	for i, d := range domains {
		if i > 0 {
			domainsText += "\\n"
		}
		domainsText += d
	}

	// Determine card color and urgency
	cardColor := "blue"
	urgencyText := "普通"
	urgencyEmoji := "🟡"
	if daysLeft <= 3 {
		cardColor = "red"
		urgencyText = "紧急"
		urgencyEmoji = "🔴"
	} else if daysLeft <= 7 {
		cardColor = "orange"
		urgencyText = "重要"
		urgencyEmoji = "🟠"
	} else if daysLeft > 14 {
		cardColor = "green"
		urgencyText = "提醒"
		urgencyEmoji = "🟢"
	}

	// Beautiful Lark card message format
	card := map[string]any{
		"config": map[string]any{
			"wide_screen_mode": true,
		},
		"header": map[string]any{
			"title": map[string]any{
				"tag":     "plain_text",
				"content": "🔐 SSL 证书过期提醒",
			},
			"template": cardColor,
		},
		"elements": []any{
			// Urgency banner
			map[string]any{
				"tag": "div",
				"text": map[string]any{
					"tag":     "lark_md",
					"content": fmt.Sprintf("%s **%s** - 证书将在 **%d 天**后过期", urgencyEmoji, urgencyText, daysLeft),
				},
			},
			// Divider
			map[string]any{
				"tag": "hr",
			},
			// Certificate details
			map[string]any{
				"tag": "div",
				"fields": []any{
					map[string]any{
						"is_short": false,
						"text": map[string]any{
							"tag":     "lark_md",
							"content": fmt.Sprintf("**🌐 域名列表**\\n%s", domainsText),
						},
					},
				},
			},
			// Divider
			map[string]any{
				"tag": "hr",
			},
			// Expiry info
			map[string]any{
				"tag": "div",
				"fields": []any{
					map[string]any{
						"is_short": true,
						"text": map[string]any{
							"tag":     "lark_md",
							"content": fmt.Sprintf("**📅 过期时间**\\n%s", cert.ExpiresAt.Format("2006-01-02 15:04")),
						},
					},
					map[string]any{
						"is_short": true,
						"text": map[string]any{
							"tag":     "lark_md",
							"content": fmt.Sprintf("**⏰ 剩余天数**\\n%d 天", daysLeft),
						},
					},
					map[string]any{
						"is_short": true,
						"text": map[string]any{
							"tag":     "lark_md",
							"content": fmt.Sprintf("**🆔 证书 ID**\\n#%d", cert.ID),
						},
					},
					map[string]any{
						"is_short": true,
						"text": map[string]any{
							"tag":     "lark_md",
							"content": fmt.Sprintf("**📊 状态**\\n%s", cert.Status),
						},
					},
				},
			},
			// Divider
			map[string]any{
				"tag": "hr",
			},
			// Footer note
			map[string]any{
				"tag": "note",
				"elements": []any{
					map[string]any{
						"tag":     "plain_text",
						"content": "🤖 ACME Console 自动提醒",
					},
				},
			},
		},
	}

	payload := map[string]any{
		"msg_type": "interactive",
		"card":     card,
	}

	return s.sendHTTPPost(config.WebhookURL, payload)
}

func (s *NotificationService) parseDomains(domainsJSON string) []string {
	var domains []string
	json.Unmarshal([]byte(domainsJSON), &domains)
	return domains
}

func (s *NotificationService) sendHTTPPost(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// CRUD operations for notification configs

type CreateNotificationConfigRequest struct {
	WorkspaceID   *uint  `json:"workspace_id,omitempty"`
	CertificateID *uint  `json:"certificate_id,omitempty"`
	Type          string `json:"type" binding:"required,oneof=generic telegram lark"`
	WebhookURL    string `json:"webhook_url" binding:"required,url"`
	WebhookConfig string `json:"webhook_config,omitempty"`
	NotifyDays    int    `json:"notify_days" binding:"min=1,max=90"`
	Enabled       bool   `json:"enabled"`
}

func (s *NotificationService) CreateConfig(req *CreateNotificationConfigRequest) (*model.NotificationConfig, error) {
	if req.NotifyDays == 0 {
		req.NotifyDays = 7 // Default to 7 days
	}

	config := &model.NotificationConfig{
		WorkspaceID:   req.WorkspaceID,
		CertificateID: req.CertificateID,
		Type:          model.NotificationType(req.Type),
		WebhookURL:    req.WebhookURL,
		WebhookConfig: req.WebhookConfig,
		NotifyDays:    req.NotifyDays,
		Enabled:       req.Enabled,
	}

	if err := s.db.Create(config).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification config: %w", err)
	}

	return config, nil
}

func (s *NotificationService) ListConfigs(workspaceID, certificateID *uint) ([]model.NotificationConfig, error) {
	var configs []model.NotificationConfig
	query := s.db.Model(&model.NotificationConfig{})

	if workspaceID != nil {
		query = query.Where("workspace_id = ?", *workspaceID)
	} else {
		query = query.Where("workspace_id IS NULL")
	}
	if certificateID != nil {
		query = query.Where("certificate_id = ?", *certificateID)
	} else {
		query = query.Where("certificate_id IS NULL")
	}

	if err := query.Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}

	return configs, nil
}

func (s *NotificationService) GetConfig(id uint) (*model.NotificationConfig, error) {
	var config model.NotificationConfig
	if err := s.db.First(&config, id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *NotificationService) UpdateConfig(id uint, req *CreateNotificationConfigRequest) error {
	updates := map[string]interface{}{
		"type":           req.Type,
		"webhook_url":    req.WebhookURL,
		"webhook_config": req.WebhookConfig,
		"notify_days":    req.NotifyDays,
		"enabled":        req.Enabled,
	}

	return s.db.Model(&model.NotificationConfig{}).Where("id = ?", id).Updates(updates).Error
}

func (s *NotificationService) DeleteConfig(id uint) error {
	return s.db.Delete(&model.NotificationConfig{}, id).Error
}

// ListLogs returns notification logs for a certificate
func (s *NotificationService) ListLogs(certificateID uint, limit int) ([]model.NotificationLog, error) {
	if limit == 0 {
		limit = 50
	}

	var logs []model.NotificationLog
	if err := s.db.Where("certificate_id = ?", certificateID).
		Order("notified_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// TestWebhook sends a test notification
func (s *NotificationService) TestWebhook(configID uint) error {
	config, err := s.GetConfig(configID)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Create a mock certificate for testing
	mockCert := &model.Certificate{
		ID:      999999,
		Domains: `["example.com","*.example.com"]`,
		Status:  model.CertificateStatusReady,
	}
	now := time.Now().Add(7 * 24 * time.Hour)
	mockCert.ExpiresAt = &now

	return s.sendNotification(config, mockCert, 7)
}
