package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role constants
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string     `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string     `gorm:"type:varchar(100)" json:"nickname"`
	Email     string     `gorm:"type:varchar(100)" json:"email"`
	Role      string     `gorm:"type:varchar(20);default:user" json:"role"` // admin, user
	Status    int        `gorm:"default:1" json:"status"`                   // 1: active, 0: disabled
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// CreateDefaultAdmin creates a default admin user if no users exist
func CreateDefaultAdmin(db *gorm.DB) error {
	var admin User
	result := db.Where("username = ?", "admin").First(&admin)

	if result.Error == nil {
		// Admin exists, ensure role is set to admin
		if admin.Role != RoleAdmin {
			return db.Model(&admin).Update("role", RoleAdmin).Error
		}
		return nil
	}

	// No admin user, create one
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}

	newAdmin := &User{
		Username: "admin",
		Nickname: "Administrator",
		Role:     RoleAdmin,
		Status:   1,
	}
	if err := newAdmin.SetPassword("admin123"); err != nil {
		return err
	}

	return db.Create(newAdmin).Error
}
