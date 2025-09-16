package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	Port                 string
	NodeEnv              string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPass               string
	DBName               string
	JWTSecret            string
	RefreshTokenSecret   string
	AccessTokenExpires   string
	RefreshTokenExpires  string
	CookieSecure         bool
	CookieSameSite       string
	CookieDomain         string
}

func Load() AppConfig {
	secure := os.Getenv("COOKIE_SECURE") == "true"

	return AppConfig{
		Port:               get("PORT", "8080"),
		NodeEnv:            get("NODE_ENV", "development"),
		DBHost:             get("DB_HOST", "127.0.0.1"),
		DBPort:             get("DB_PORT", "3306"),
		DBUser:             get("DB_USER", "root"),
		DBPass:             get("DB_PASS", ""),
		DBName:             get("DB_NAME", "yeabue_coffee"),
		JWTSecret:          get("JWT_SECRET", "dev_access"),
		RefreshTokenSecret: get("REFRESH_TOKEN_SECRET", "dev_refresh"),
		AccessTokenExpires: get("ACCESS_TOKEN_EXPIRES", "1h"),
		RefreshTokenExpires:get("REFRESH_TOKEN_EXPIRES", "720h"),
		CookieSecure:       secure,
		CookieSameSite:     get("COOKIE_SAMESITE", "strict"),
		CookieDomain:       get("COOKIE_DOMAIN", "localhost"),
	}
}

func get(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func DSN(cfg AppConfig) string {
	// parseTime=true ทำให้ DATE/TIME map ถูกต้อง
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
}
