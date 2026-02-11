package service

import (
	"sync"
	"time"

	"github.com/imkerbos/ACME-Console/internal/model"
	"gorm.io/gorm"
)

// SettingService 配置服务，带内存缓存
type SettingService struct {
	db    *gorm.DB
	cache map[string]string
	mu    sync.RWMutex
}

// NewSettingService 创建配置服务
func NewSettingService(db *gorm.DB) *SettingService {
	svc := &SettingService{
		db:    db,
		cache: make(map[string]string),
	}
	// 初始化时加载所有配置到缓存
	svc.loadAll()
	return svc
}

// loadAll 加载所有配置到缓存
func (s *SettingService) loadAll() {
	var settings []model.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, setting := range settings {
		s.cache[setting.Key] = setting.Value
	}
}

// Get 获取配置值
func (s *SettingService) Get(key string) string {
	s.mu.RLock()
	if val, ok := s.cache[key]; ok {
		s.mu.RUnlock()
		return val
	}
	s.mu.RUnlock()

	// 缓存未命中，从数据库读取
	var setting model.Setting
	if err := s.db.Where("`key` = ?", key).First(&setting).Error; err != nil {
		return ""
	}

	s.mu.Lock()
	s.cache[key] = setting.Value
	s.mu.Unlock()

	return setting.Value
}

// GetWithDefault 获取配置值，如果不存在返回默认值
func (s *SettingService) GetWithDefault(key, defaultValue string) string {
	val := s.Get(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// Set 设置配置值
func (s *SettingService) Set(key, value string) error {
	var setting model.Setting
	err := s.db.Where("`key` = ?", key).First(&setting).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新配置
		setting = model.Setting{
			Key:   key,
			Value: value,
		}
		if err := s.db.Create(&setting).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// 更新现有配置
		if err := s.db.Model(&setting).Update("value", value).Error; err != nil {
			return err
		}
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()

	return nil
}

// SetMultiple 批量设置配置
func (s *SettingService) SetMultiple(settings map[string]string) error {
	for key, value := range settings {
		if err := s.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetAll 获取所有配置
func (s *SettingService) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	if err := s.db.Order("`key` ASC").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// Refresh 刷新缓存
func (s *SettingService) Refresh() {
	s.mu.Lock()
	s.cache = make(map[string]string)
	s.mu.Unlock()
	s.loadAll()
}

// GetACMEConfig 获取 ACME 相关配置
func (s *SettingService) GetACMEConfig() ACMESettings {
	return ACMESettings{
		DNSResolvers: s.GetWithDefault(model.SettingDNSResolvers, "8.8.8.8:53,1.1.1.1:53"),
		DNSTimeout:   s.GetWithDefault(model.SettingDNSTimeout, "10s"),
	}
}

// ACMESettings ACME 配置结构
type ACMESettings struct {
	DNSResolvers string `json:"dns_resolvers"`
	DNSTimeout   string `json:"dns_timeout"`
}

// SiteSettings 网站配置结构
type SiteSettings struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

// GetSiteConfig 获取网站配置
func (s *SettingService) GetSiteConfig() SiteSettings {
	return SiteSettings{
		Title:    s.GetWithDefault(model.SettingSiteTitle, "ACME Console"),
		Subtitle: s.GetWithDefault(model.SettingSiteSubtitle, "证书管理系统"),
	}
}

// GetDNSTimeout 获取 DNS 超时时间
func (s *SettingService) GetDNSTimeout() time.Duration {
	timeout := s.GetWithDefault(model.SettingDNSTimeout, "10s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return 10 * time.Second
	}
	return d
}
