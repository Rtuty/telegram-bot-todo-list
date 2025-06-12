package domain

import (
	"context"
	"time"
)

// TaskRepository определяет интерфейс для работы с задачами
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id int) (*Task, error)
	GetByUserID(ctx context.Context, userID int64, status TaskStatus) ([]*Task, error)
	GetAll(ctx context.Context, userID int64) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id int) error
	GetTasksForNotification(ctx context.Context, beforeTime time.Time) ([]*Task, error)
}

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
}

// SessionRepository определяет интерфейс для работы с сессиями
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, telegramID int64) error
	CleanupExpired(ctx context.Context) error
}

// NoteRepository определяет интерфейс для работы с заметками
type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByID(ctx context.Context, id int) (*Note, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Note, error)
	GetByCategory(ctx context.Context, userID int64, category NoteCategory) ([]*Note, error)
	GetByType(ctx context.Context, userID int64, noteType NoteType) ([]*Note, error)
	GetFavorites(ctx context.Context, userID int64) ([]*Note, error)
	Search(ctx context.Context, userID int64, query string) ([]*Note, error)
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id int) error
}
