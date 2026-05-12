package model

import (
	"time"

	"gorm.io/gorm"
)

type Workspace struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	OwnerID     uint      `gorm:"not null;index" json:"owner_id"`
	Status      int       `gorm:"default:1" json:"status"` // 1=active, 0=archived
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Owner   *User               `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []WorkspaceMember   `gorm:"foreignKey:WorkspaceID" json:"members,omitempty"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func MigrateWorkspace(db *gorm.DB) error {
	return db.AutoMigrate(&Workspace{})
}
