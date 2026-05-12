package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
	}

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if got := cfg.DSN(); got != expected {
		t.Errorf("DSN() = %s, want %s", got, expected)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: "127.0.0.1"
  port: 9090

database:
  host: "dbhost"
  port: 3307
  user: "dbuser"
  password: "dbpass"
  dbname: "testdb"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Database.Host != "dbhost" {
		t.Errorf("Database.Host = %s, want dbhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 3307 {
		t.Errorf("Database.Port = %d, want 3307", cfg.Database.Port)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}
