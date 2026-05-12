package model

import (
	"time"

	"gorm.io/gorm"
)

// RenewalLog records certificate renewal audit trail
type RenewalLog struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	CertificateID uint       `gorm:"not null;index" json:"certificate_id"`
	Action        string     `gorm:"type:varchar(30);not null" json:"action"`  // initiated, dns_ready, completed, failed, reset
	Status        string     `gorm:"type:varchar(20);not null" json:"status"`  // success, failed
	Message       string     `gorm:"type:text" json:"message,omitempty"`
	OldExpiresAt  *time.Time `json:"old_expires_at,omitempty"`
	NewExpiresAt  *time.Time `json:"new_expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	Certificate *Certificate `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
}

func (RenewalLog) TableName() string {
	return "renewal_logs"
}

func MigrateRenewalLog(db *gorm.DB) error {
	return db.AutoMigrate(&RenewalLog{})
}
