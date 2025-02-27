// internal/config/config.go
package config

import (
	"os"
	"time"
)

type Config struct {
	Database struct {
		Host       string `json:"host"`
		Port       string `json:"port"`
		User       string `json:"user"`
		Password   string `json:"password"`
		Name       string `json:"name"`
		SSLMode    string `json:"sslmode"`
		SearchPath string `json:"schema"`
	} `json:"database"`
	Supra struct {
		Host   string `json:"host"`
		APIKey string `json:"api_key"`
	} `json:"supra"`
	JWT struct {
		Secret       string        `json:"secret"`
		ExpiryPeriod time.Duration `json:"expiry_period"`
	} `json:"jwt"`
	Server struct {
		Port         string        `json:"port"`
		ReadTimeout  time.Duration `json:"read_timeout"`
		WriteTimeout time.Duration `json:"write_timeout"`
	}
	Sendgrid struct {
		APIKey string `json:"api_key"`
		From   string `json:"from"`
	} `json:"sendgrid"`
	SMTP map[string]struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
	} `json:"smtp"`
	BaseURL string `json:"base_url"`
}

func Load() *Config {
	cfg := &Config{}

	// Database configuration
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.Name = getEnv("DB_NAME", "myapp")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.SearchPath = getEnv("DB_SCHEMA", "public")

	// Supra host
	cfg.Supra.Host = getEnv("SUPRA_HOST", "http://localhost:4780")

	// JWT configuration
	cfg.JWT.Secret = getEnv("JWT_SECRET", "your-secret-key")
	cfg.JWT.ExpiryPeriod = time.Hour * 24

	// Sendgrid configuration
	cfg.Sendgrid.APIKey = getEnv("SENDGRID_API_KEY", "")
	cfg.Sendgrid.From = getEnv("SENDGRID_FROM", "")

	// Server configuration
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.ReadTimeout = time.Second * 15
	cfg.Server.WriteTimeout = time.Second * 15

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
