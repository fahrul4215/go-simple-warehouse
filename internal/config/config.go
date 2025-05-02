package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBUser    string
	DBPass    string
	DBHost    string
	DBName    string
	JWTSecret string

	AdminUser string
	AdminPass string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBUser:    os.Getenv("DB_USER"),
		DBPass:    os.Getenv("DB_PASS"),
		DBHost:    os.Getenv("DB_HOST"),
		DBName:    os.Getenv("DB_NAME"),
		JWTSecret: os.Getenv("JWT_SECRET"),

		AdminUser: os.Getenv("ADMIN_USER"),
		AdminPass: os.Getenv("ADMIN_PASS"),
	}

	if cfg.DBUser == "" || cfg.DBPass == "" ||
		cfg.DBHost == "" || cfg.DBName == "" ||
		cfg.JWTSecret == "" || cfg.AdminUser == "" ||
		cfg.AdminPass == "" {
		return nil, fmt.Errorf("one or more required env vars missing")
	}
	return cfg, nil
}
