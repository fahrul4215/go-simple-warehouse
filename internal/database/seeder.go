// internal/database/seeder.go
package database

import (
	"errors"
	"go-simple-warehouse/internal/auth"
	"go-simple-warehouse/internal/config"

	"gorm.io/gorm"
)

func SeedSuperAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminUser == "" || cfg.AdminPass == "" {
		return nil
	}

	var existing auth.User
	err := db.Where("username = ?", cfg.AdminUser).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	admin := auth.User{
		Username: cfg.AdminUser,
		Password: cfg.AdminPass,
	}
	return db.Create(&admin).Error
}
