package model

import (
	"time"

	"gorm.io/gorm"
)

// WorkspaceRole constants
const (
	WorkspaceRoleOwner  = "owner"
	WorkspaceRoleAdmin  = "admin"
	WorkspaceRoleMember = "member"
)

type WorkspaceMember struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WorkspaceID uint      `gorm:"not null;index:idx_workspace_user,unique" json:"workspace_id"`
	UserID      uint      `gorm:"not null;index:idx_workspace_user,unique" json:"user_id"`
	Role        string    `gorm:"type:varchar(20);not null;default:member" json:"role"` // owner/admin/member
	CreatedAt   time.Time `json:"created_at"`

	// Relations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

func MigrateWorkspaceMember(db *gorm.DB) error {
	return db.AutoMigrate(&WorkspaceMember{})
}
