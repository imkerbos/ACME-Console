package model

import (
	"fmt"

	"github.com/imkerbos/ACME-Console/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// First, connect without database to create it if needed
	if err := ensureDatabase(cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure database: %w", err)
	}

	// Connect to the database
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate all models
	if err := MigrateUser(db); err != nil {
		return nil, err
	}
	if err := MigrateWorkspace(db); err != nil {
		return nil, err
	}
	if err := MigrateWorkspaceMember(db); err != nil {
		return nil, err
	}
	if err := MigrateACMEAccount(db); err != nil {
		return nil, err
	}

	// Fix legacy data before migrating Certificate table
	if err := fixLegacyCertificateData(db); err != nil {
		return nil, err
	}

	if err := MigrateCertificate(db); err != nil {
		return nil, err
	}
	if err := MigrateChallenge(db); err != nil {
		return nil, err
	}
	if err := MigrateSetting(db); err != nil {
		return nil, err
	}
	if err := MigrateNotification(db); err != nil {
		return nil, err
	}

	// Initialize default settings
	if err := InitDefaultSettings(db); err != nil {
		return nil, err
	}

	// Create default admin user
	if err := CreateDefaultAdmin(db); err != nil {
		return nil, err
	}

	return db, nil
}

// ensureDatabase creates the database if it doesn't exist
func ensureDatabase(cfg *config.DatabaseConfig) error {
	db, err := gorm.Open(mysql.Open(cfg.DSNWithoutDB()), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create database if not exists
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)
	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	// Close the connection
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// fixLegacyCertificateData fixes legacy certificate data before migration
func fixLegacyCertificateData(db *gorm.DB) error {
	// Check if certificates table exists
	if !db.Migrator().HasTable("certificates") {
		return nil // Table doesn't exist yet, nothing to fix
	}

	// Check if created_by column exists
	if !db.Migrator().HasColumn(&Certificate{}, "created_by") {
		return nil // Column doesn't exist yet, nothing to fix
	}

	// Update created_by = 0 to NULL for legacy certificates
	result := db.Exec("UPDATE certificates SET created_by = NULL WHERE created_by = 0")
	if result.Error != nil {
		return fmt.Errorf("failed to fix legacy certificate data: %w", result.Error)
	}

	return nil
}
