package database

import (
	"fmt"
	"go-simple-warehouse/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
