package config

import (
	"os"
	"strconv"
)

// Config конфигурация приложения
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Email    EmailConfig    `json:"email"`
	App      AppConfig      `json:"app"`
}

// ServerConfig настройки сервера
type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Mode string `json:"mode"` // development, production
}

// DatabaseConfig настройки базы данных
type DatabaseConfig struct {
	MongoURI     string `json:"mongo_uri"`
	DatabaseName string `json:"database_name"`
}

// EmailConfig настройки email
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
}

// AppConfig настройки приложения
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	JWTSecret   string `json:"jwt_secret"`
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8081"),
			Mode: getEnv("SERVER_MODE", "development"),
		},
		Database: DatabaseConfig{
			MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
			DatabaseName: getEnv("DATABASE_NAME", "billing_system"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@billing.com"),
			FromName:     getEnv("FROM_NAME", "Billing System"),
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Billing System"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-here"),
		},
	}
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(name string, fallback int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
