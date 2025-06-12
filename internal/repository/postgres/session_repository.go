package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"todolist/internal/domain"

	"github.com/Masterminds/squirrel"
)

// SessionRepositoryImpl реализует интерфейс SessionRepository
type SessionRepositoryImpl struct {
	db *Database
	sq squirrel.StatementBuilderType
}

// NewSessionRepository создает новый экземпляр SessionRepositoryImpl
func NewSessionRepository(db *Database) domain.SessionRepository {
	return &SessionRepositoryImpl{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create создает новую сессию
func (r *SessionRepositoryImpl) Create(ctx context.Context, session *domain.Session) error {
	query, args, err := r.sq.
		Insert("sessions").
		Columns("user_id", "telegram_id", "is_active", "expires_at").
		Values(session.UserID, session.TelegramID, session.IsActive, session.ExpiresAt).
		Suffix(`ON CONFLICT (telegram_id) 
			DO UPDATE SET 
				user_id = EXCLUDED.user_id,
				is_active = EXCLUDED.is_active,
				created_at = CURRENT_TIMESTAMP,
				expires_at = EXCLUDED.expires_at
			RETURNING created_at`).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.db.DB.QueryRowContext(ctx, query, args...).Scan(&session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByTelegramID получает сессию по Telegram ID
func (r *SessionRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Session, error) {
	query, args, err := r.sq.
		Select(
			"user_id",
			"telegram_id",
			"is_active",
			"created_at",
			"expires_at",
		).
		From("sessions").
		Where(squirrel.Eq{"telegram_id": telegramID}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	session := &domain.Session{}
	err = r.db.DB.QueryRowContext(ctx, query, args...).Scan(
		&session.UserID,
		&session.TelegramID,
		&session.IsActive,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// Update обновляет сессию
func (r *SessionRepositoryImpl) Update(ctx context.Context, session *domain.Session) error {
	query, args, err := r.sq.
		Update("sessions").
		Set("is_active", session.IsActive).
		Set("expires_at", session.ExpiresAt).
		Where(squirrel.Eq{"telegram_id": session.TelegramID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete удаляет сессию
func (r *SessionRepositoryImpl) Delete(ctx context.Context, telegramID int64) error {
	query, args, err := r.sq.
		Delete("sessions").
		Where(squirrel.Eq{"telegram_id": telegramID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// CleanupExpired удаляет истекшие сессии
func (r *SessionRepositoryImpl) CleanupExpired(ctx context.Context) error {
	query, args, err := r.sq.
		Delete("sessions").
		Where(squirrel.Lt{"expires_at": time.Now()}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}
