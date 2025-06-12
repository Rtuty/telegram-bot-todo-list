package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"todolist/internal/domain"

	"github.com/Masterminds/squirrel"
)

// UserRepositoryImpl реализует интерфейс UserRepository
type UserRepositoryImpl struct {
	db *Database
	sq squirrel.StatementBuilderType
}

// NewUserRepository создает новый экземпляр UserRepositoryImpl
func NewUserRepository(db *Database) domain.UserRepository {
	return &UserRepositoryImpl{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create создает нового пользователя
func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	query := r.sq.Insert("users").
		Columns("telegram_id", "username", "first_name", "last_name").
		Values(user.TelegramID, user.Username, user.FirstName, user.LastName).
		Suffix("RETURNING id, created_at, updated_at, last_login_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.db.DB.QueryRowContext(ctx, sql, args...).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.IsActive = true
	return nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *UserRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query, args, err := r.sq.
		Select(
			"id", "telegram_id", "username", "first_name", "last_name",
			"is_active", "created_at", "updated_at", "last_login_at").
		From("users").
		Where(squirrel.Eq{"telegram_id": telegramID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	user := &domain.User{}
	err = r.db.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update обновляет пользователя
func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	query, args, err := r.sq.
		Update("users").
		Set("username", user.Username).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("is_active", user.IsActive).
		Set("updated_at", "CURRENT_TIMESTAMP").
		Set("last_login_at", user.LastLoginAt).
		Where(squirrel.Eq{"id": user.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	if _, err = r.db.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
