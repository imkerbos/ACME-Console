package model

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType represents the type of notification channel
type NotificationType string

const (
	NotificationTypeGeneric  NotificationType = "generic"  // Generic webhook with custom JSON
	NotificationTypeTelegram NotificationType = "telegram" // Telegram bot
	NotificationTypeLark     NotificationType = "lark"     // Lark (Feishu)
)

// NotificationConfig stores webhook configuration for certificate expiry notifications
type NotificationConfig struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	WorkspaceID   *uint            `gorm:"index" json:"workspace_id,omitempty"`   // NULL = personal certificate config
	CertificateID *uint            `gorm:"index" json:"certificate_id,omitempty"` // NULL = workspace-level config
	Type          NotificationType `gorm:"type:varchar(20);not null" json:"type"`
	WebhookURL    string           `gorm:"type:varchar(512);not null" json:"webhook_url"`
	WebhookConfig string           `gorm:"type:json" json:"webhook_config,omitempty"` // Extra config (secret, chat_id, etc.)
	NotifyDays    int              `gorm:"default:7" json:"notify_days"`              // Notify N days before expiry
	Enabled       bool             `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`

	// Relations
	Workspace   *Workspace   `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	Certificate *Certificate `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
}

func (NotificationConfig) TableName() string {
	return "notification_configs"
}

// NotificationLog records sent notifications to avoid duplicates
type NotificationLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CertificateID uint      `gorm:"not null;index" json:"certificate_id"`
	ConfigID      uint      `gorm:"not null;index" json:"config_id"` // Which config was used
	NotifyDays    int       `gorm:"not null" json:"notify_days"`     // Notified N days before expiry
	Status        string    `gorm:"type:varchar(20);not null" json:"status"` // success/failed
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	NotifiedAt    time.Time `gorm:"not null;index" json:"notified_at"`

	// Relations
	Certificate *Certificate        `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
	Config      *NotificationConfig `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
}

func (NotificationLog) TableName() string {
	return "notification_logs"
}

func MigrateNotification(db *gorm.DB) error {
	if err := db.AutoMigrate(&NotificationConfig{}); err != nil {
		return err
	}
	return db.AutoMigrate(&NotificationLog{})
}
