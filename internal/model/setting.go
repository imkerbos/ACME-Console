package model

import (
	"time"

	"gorm.io/gorm"
)

// Setting 配置项键名常量
const (
	SettingACMEEmail    = "acme.email"    // ACME 账户邮箱（已废弃，改为申请时填写）
	SettingDNSResolvers = "dns.resolvers" // DNS 服务器列表，逗号分隔
	SettingDNSTimeout   = "dns.timeout"   // DNS 查询超时，如 "10s"
	SettingSiteTitle    = "site.title"    // 网站标题
	SettingSiteSubtitle = "site.subtitle" // 网站副标题
)

// Setting 系统配置表
type Setting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Description string    `gorm:"type:varchar(255)" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}

func MigrateSetting(db *gorm.DB) error {
	return db.AutoMigrate(&Setting{})
}

// 默认配置值
var defaultSettings = map[string]struct {
	Value       string
	Description string
}{
	SettingDNSResolvers: {"8.8.8.8:53,1.1.1.1:53", "DNS 服务器列表，逗号分隔"},
	SettingDNSTimeout:   {"10s", "DNS 查询超时时间"},
	SettingSiteTitle:    {"ACME Console", "网站标题"},
	SettingSiteSubtitle: {"证书管理系统", "网站副标题"},
}

// InitDefaultSettings 初始化默认配置
func InitDefaultSettings(db *gorm.DB) error {
	for key, def := range defaultSettings {
		var count int64
		// 使用反引号转义 key（MySQL 保留字）
		if err := db.Model(&Setting{}).Where("`key` = ?", key).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			setting := Setting{
				Key:         key,
				Value:       def.Value,
				Description: def.Description,
			}
			if err := db.Create(&setting).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
