package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"todolist/config"
	"todolist/internal/domain"

	"go.uber.org/zap"
)

// AuthService предоставляет методы для авторизации
type AuthService struct {
	userRepository    domain.UserRepository
	sessionRepository domain.SessionRepository
	config            *config.Config
	logger            *zap.Logger
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(
	userRepository domain.UserRepository,
	sessionRepository domain.SessionRepository,
	config *config.Config,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		config:            config,
		logger:            logger,
	}
}

// Login выполняет авторизацию пользователя
func (s *AuthService) Login(ctx context.Context, telegramID int64, username, firstName, lastName, password string) (*domain.User, error) {
	// Проверяем пароль
	if password != s.config.Auth.Password {
		s.logger.Warn("invalid password attempt", zap.Int64("telegram_id", telegramID))
		return nil, fmt.Errorf("неверный пароль")
	}

	// Получаем или создаем пользователя
	user, err := s.userRepository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Создаем нового пользователя
			user = &domain.User{
				TelegramID: telegramID,
				Username:   username,
				FirstName:  firstName,
				LastName:   lastName,
				IsActive:   true,
			}

			if err := s.userRepository.Create(ctx, user); err != nil {
				s.logger.Error("failed to create user", zap.Error(err))
				return nil, fmt.Errorf("не удалось создать пользователя")
			}

			s.logger.Info("new user created", zap.Int64("telegram_id", telegramID))
		} else {
			s.logger.Error("failed to get user", zap.Error(err))
			return nil, fmt.Errorf("ошибка получения пользователя")
		}
	} else {
		// Обновляем информацию о пользователе
		user.Username = username
		user.FirstName = firstName
		user.LastName = lastName
		user.LastLoginAt = time.Now()

		if err := s.userRepository.Update(ctx, user); err != nil {
			s.logger.Error("failed to update user", zap.Error(err))
		}
	}

	// Создаем сессию
	session := &domain.Session{
		UserID:     user.ID,
		TelegramID: telegramID,
		IsActive:   true,
		ExpiresAt:  time.Now().Add(s.config.Auth.SessionTimeout),
	}

	if err := s.sessionRepository.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
		return nil, fmt.Errorf("ошибка создания сессии")
	}

	s.logger.Info("user logged in", zap.Int64("user_id", user.ID), zap.Int64("telegram_id", telegramID))
	return user, nil
}

// IsAuthenticated проверяет, авторизован ли пользователь
func (s *AuthService) IsAuthenticated(ctx context.Context, telegramID int64) (*domain.User, error) {
	session, err := s.sessionRepository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена")
	}

	if !session.IsValid() {
		s.logger.Info("invalid session", zap.Int64("telegram_id", telegramID))
		return nil, fmt.Errorf("сессия истекла или недействительна")
	}

	user, err := s.userRepository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))
		return nil, fmt.Errorf("пользователь не найден")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("пользователь неактивен")
	}

	return user, nil
}

// Logout выполняет выход пользователя из системы
func (s *AuthService) Logout(ctx context.Context, telegramID int64) error {
	if err := s.sessionRepository.Delete(ctx, telegramID); err != nil {
		s.logger.Error("failed to delete session", zap.Error(err))
		return fmt.Errorf("ошибка выхода из системы")
	}

	s.logger.Info("user logged out", zap.Int64("telegram_id", telegramID))
	return nil
}

// CleanupExpiredSessions удаляет истекшие сессии
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.sessionRepository.CleanupExpired(ctx); err != nil {
		s.logger.Error("failed to cleanup expired sessions", zap.Error(err))
		return err
	}

	s.logger.Info("expired sessions cleaned up")
	return nil
}
