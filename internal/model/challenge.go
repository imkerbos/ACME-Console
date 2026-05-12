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
	Type          string          `gorm:"type:varchar(20);not null;default:dns-01" json:"type"` // dns-01 or http-01
	TXTHost       string          `gorm:"type:varchar(255)" json:"txt_host,omitempty"` // _acme-challenge.example.com (DNS-01)
	TXTValue      string          `gorm:"type:varchar(255)" json:"txt_value,omitempty"` // DNS-01 TXT value
	HTTPPath      string          `gorm:"type:varchar(512)" json:"http_path,omitempty"`    // HTTP-01: /.well-known/acme-challenge/{token}
	HTTPContent   string          `gorm:"type:text" json:"http_content,omitempty"`         // HTTP-01: key authorization response body
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
