package model

import (
	"time"

	"gorm.io/gorm"
)

type CertificateStatus string

const (
	CertificateStatusPending CertificateStatus = "pending"
	CertificateStatusReady   CertificateStatus = "ready"
	CertificateStatusFailed  CertificateStatus = "failed"
)

type KeyType string

const (
	KeyTypeRSA KeyType = "RSA"
	KeyTypeECC KeyType = "ECC"
)

type Certificate struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	AccountID     *uint             `gorm:"index" json:"account_id,omitempty"`       // Foreign key to ACMEAccount
	WorkspaceID   *uint             `gorm:"index" json:"workspace_id,omitempty"`     // Foreign key to Workspace (NULL=private)
	CreatedBy     *uint             `gorm:"index" json:"created_by,omitempty"`       // User who created this certificate (NULL for legacy certs)
	Email         string            `gorm:"type:varchar(255)" json:"email,omitempty"` // 申请人邮箱
	Domains       string            `gorm:"type:json;not null" json:"domains"`       // JSON array: ["example.com", "*.example.com"]
	KeyType       KeyType           `gorm:"type:varchar(10);not null" json:"key_type"`
	KeySize       int               `gorm:"default:2048" json:"key_size"`            // RSA: 2048/4096, ECC: 256/384
	Status        CertificateStatus `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	OrderURL      string            `gorm:"type:varchar(512)" json:"order_url,omitempty"`       // ACME order URL
	CertPEM       string            `gorm:"type:text" json:"cert_pem,omitempty"`
	KeyPEM        string            `gorm:"type:text" json:"-"`                                 // Encrypted, never expose in JSON
	ChainPEM      string            `gorm:"type:text" json:"chain_pem,omitempty"`
	IssuerCertPEM string            `gorm:"type:text" json:"-"`                                 // Issuer certificate
	SerialNumber  string            `gorm:"type:varchar(64)" json:"serial_number,omitempty"`    // Certificate serial number
	Fingerprint   string            `gorm:"type:varchar(64)" json:"fingerprint,omitempty"`      // SHA-256 fingerprint
	IssuedAt      *time.Time        `json:"issued_at,omitempty"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Account       *ACMEAccount      `gorm:"foreignKey:AccountID" json:"-"`
	Workspace     *Workspace        `gorm:"foreignKey:WorkspaceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"workspace,omitempty"`
	Creator       *User             `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"creator,omitempty"`
	Challenges    []Challenge       `gorm:"foreignKey:CertificateID" json:"challenges,omitempty"`
}

func (Certificate) TableName() string {
	return "certificates"
}

func MigrateCertificate(db *gorm.DB) error {
	return db.AutoMigrate(&Certificate{})
}
