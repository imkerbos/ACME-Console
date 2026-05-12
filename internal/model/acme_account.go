package model

import (
	"time"

	"gorm.io/gorm"
)

// ACMEAccount stores ACME account credentials for a Certificate Authority
type ACMEAccount struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_email_ca" json:"email"`
	CAURL        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_email_ca" json:"ca_url"` // e.g., https://acme-v02.api.letsencrypt.org/directory
	PrivateKey   string    `gorm:"type:text;not null" json:"-"`                                      // Encrypted PEM-encoded private key
	Registration string    `gorm:"type:text" json:"-"`                                               // JSON-encoded registration resource
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ACMEAccount) TableName() string {
	return "acme_accounts"
}

func MigrateACMEAccount(db *gorm.DB) error {
	return db.AutoMigrate(&ACMEAccount{})
}
