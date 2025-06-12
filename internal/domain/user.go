package domain

import "time"

// User представляет пользователя системы
type User struct {
	ID          int64     `json:"id" db:"id"`
	TelegramID  int64     `json:"telegram_id" db:"telegram_id"`
	Username    string    `json:"username" db:"username"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at" db:"last_login_at"`
}

// Session представляет сессию пользователя
type Session struct {
	UserID     int64     `json:"user_id" db:"user_id"`
	TelegramID int64     `json:"telegram_id" db:"telegram_id"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
}

// IsExpired проверяет, истекла ли сессия
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid проверяет, действительна ли сессия
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}
