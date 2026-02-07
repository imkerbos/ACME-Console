package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	ACME       ACMEConfig       `mapstructure:"acme"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

type ACMEConfig struct {
	DNS DNSConfig `mapstructure:"dns"`
}

type DNSConfig struct {
	Resolvers string `mapstructure:"resolvers"` // Comma-separated DNS servers, e.g., "8.8.8.8:53,1.1.1.1:53"
	Timeout   string `mapstructure:"timeout"`   // DNS query timeout, e.g., "10s"
}

type EncryptionConfig struct {
	MasterKey string `mapstructure:"master_key"` // 32-byte hex-encoded key for AES-256-GCM
}

type JWTConfig struct {
	Secret   string `mapstructure:"secret"`
	ExpireHours int `mapstructure:"expire_hours"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// DSNWithoutDB returns DSN without database name for creating database
func (d *DatabaseConfig) DSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port)
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Environment variable overrides
	viper.SetEnvPrefix("ACME")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
