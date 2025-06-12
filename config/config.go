package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения
type Config struct {
	Bot      BotConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

// BotConfig содержит настройки телеграм бота
type BotConfig struct {
	Token   string
	Debug   bool
	Timeout time.Duration
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// AuthConfig содержит настройки авторизации
type AuthConfig struct {
	Password       string
	SessionTimeout time.Duration
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл если он существует
	_ = godotenv.Load()

	cfg := &Config{
		Bot: BotConfig{
			Token:   getEnv("BOT_TOKEN", ""),
			Debug:   getEnvBool("BOT_DEBUG", false),
			Timeout: getEnvDuration("BOT_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "todobot"),
			Password: getEnv("DB_PASSWORD", "password"),
			Database: getEnv("DB_NAME", "todolist"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			Password:       getEnv("AUTH_PASSWORD", "password123"),
			SessionTimeout: getEnvDuration("AUTH_SESSION_TIMEOUT", 24*time.Hour),
		},
	}

	return cfg, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает целочисленное значение переменной окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool получает булево значение переменной окружения
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration получает duration значение переменной окружения
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
