package model

import (
	"time"

	"gorm.io/gorm"
)

type ChallengeStatus string

const (
	ChallengeStatusPending  ChallengeStatus = "pending"
	ChallengeStatusVerified ChallengeStatus = "verified"
	ChallengeStatusFailed   ChallengeStatus = "failed"
)

type Challenge struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	CertificateID uint            `gorm:"index;not null" json:"certificate_id"`
	Domain        string          `gorm:"type:varchar(255);not null" json:"domain"`
	TXTHost       string          `gorm:"type:varchar(255);not null" json:"txt_host"` // _acme-challenge.example.com
	TXTValue      string          `gorm:"type:varchar(255);not null" json:"txt_value"`
	Token        string          `gorm:"type:varchar(255)" json:"-"`  // ACME challenge token
	KeyAuth      string          `gorm:"type:varchar(255)" json:"-"`  // ACME key authorization (for lego)
	AuthzURL     string          `gorm:"type:varchar(512)" json:"-"`  // ACME authorization URL
	ChallengeURL string          `gorm:"type:varchar(512)" json:"-"`  // ACME challenge URL
	Status       ChallengeStatus `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	ValidatedAt   *time.Time      `json:"validated_at,omitempty"`
	ErrorMessage  string          `gorm:"type:text" json:"error_message,omitempty"`
	DNSCheckedAt  *time.Time      `json:"dns_checked_at,omitempty"`
	DNSCheckOK    bool            `gorm:"default:false" json:"dns_check_ok"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (Challenge) TableName() string {
	return "challenges"
}

func MigrateChallenge(db *gorm.DB) error {
	return db.AutoMigrate(&Challenge{})
}
