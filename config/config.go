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

const (
	_botTokenKey   = "BOT_TOKEN"
	_botDebugKey   = "BOT_DEBUG"
	_botTimeoutKey = "BOT_TIMEOUT"

	_dbHostKey     = "DB_HOST"
	_dbPortKey     = "DB_PORT"
	_dbUserKey     = "DB_USER"
	_dbPasswordKey = "DB_PASSWORD"
	_dbNameKey     = "DB_NAME"
	_dbSSLModeKey  = "DB_SSLMODE"

	_authPasswordKey    = "AUTH_PASSWORD"
	_authSessionTimeout = "AUTH_SESSION_TIMEOUT"
)

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл если он существует
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		Bot: BotConfig{
			Token:   getEnv(_botTokenKey, ""),
			Debug:   getEnvBool(_botDebugKey, false),
			Timeout: getEnvDuration(_botTimeoutKey, 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv(_dbHostKey, "localhost"),
			Port:     getEnvInt(_dbPortKey, 5432),
			User:     getEnv(_dbUserKey, "todobot"),
			Password: getEnv(_dbPasswordKey, "password"),
			Database: getEnv(_dbNameKey, "todolist"),
			SSLMode:  getEnv(_dbSSLModeKey, "disable"),
		},
		Auth: AuthConfig{
			Password:       getEnv(_authPasswordKey, "password123"),
			SessionTimeout: getEnvDuration(_authSessionTimeout, 24*time.Hour),
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
