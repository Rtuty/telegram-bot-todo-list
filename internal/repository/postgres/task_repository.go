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

// TaskRepositoryImpl реализует интерфейс TaskRepository
type TaskRepositoryImpl struct {
	db *Database
	sq squirrel.StatementBuilderType
}

// NewTaskRepository создает новый экземпляр TaskRepositoryImpl
func NewTaskRepository(db *Database) domain.TaskRepository {
	return &TaskRepositoryImpl{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create создает новую задачу
func (r *TaskRepositoryImpl) Create(ctx context.Context, task *domain.Task) error {
	query := r.sq.Insert("tasks").
		Columns("title", "description", "status", "priority", "user_id", "notify_at").
		Values(task.Title, task.Description, task.Status, task.Priority, task.UserID, task.NotifyAt).
		Suffix("RETURNING id, created_at, updated_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.db.DB.QueryRowContext(ctx, sql, args...).Scan(
		&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetByID получает задачу по ID
func (r *TaskRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Task, error) {
	query, args, err := r.sq.
		Select(
			"id", "title", "description", "status", "priority",
			"created_at", "updated_at", "completed_at", "notify_at", "user_id").
		From("tasks").
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.NotEq{"status": "deleted"}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	task := &domain.Task{}
	err = r.db.DB.QueryRowContext(ctx, query, args...).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
		&task.NotifyAt,
		&task.UserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// GetByUserID получает задачи пользователя по статусу
func (r *TaskRepositoryImpl) GetByUserID(ctx context.Context, userID int64, status domain.TaskStatus) ([]*domain.Task, error) {
	query, args, err := r.sq.
		Select(
			"id", "title", "description", "status", "priority",
			"created_at", "updated_at", "completed_at", "notify_at", "user_id").
		From("tasks").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Eq{"status": status}).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// GetAll получает все активные задачи пользователя
func (r *TaskRepositoryImpl) GetAll(ctx context.Context, userID int64) ([]*domain.Task, error) {
	query, args, err := r.sq.
		Select(
			"id", "title", "description", "status", "priority",
			"created_at", "updated_at", "completed_at", "notify_at", "user_id").
		From("tasks").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.NotEq{"status": "deleted"}).
		OrderBy(`
			CASE status 
				WHEN 'pending' THEN 1 
				WHEN 'completed' THEN 2 
			END,
			CASE priority 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			created_at DESC`).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// Update обновляет задачу
func (r *TaskRepositoryImpl) Update(ctx context.Context, task *domain.Task) error {
	task.UpdatedAt = time.Now()

	query, args, err := r.sq.
		Update("tasks").
		Set("title", task.Title).
		Set("description", task.Description).
		Set("status", task.Status).
		Set("priority", task.Priority).
		Set("updated_at", task.UpdatedAt).
		Set("completed_at", task.CompletedAt).
		Set("notify_at", task.NotifyAt).
		Where(squirrel.Eq{"id": task.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	if _, err = r.db.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Delete удаляет задачу (помечает как удаленную)
func (r *TaskRepositoryImpl) Delete(ctx context.Context, id int) error {
	query, args, err := r.sq.
		Update("tasks").
		Set("status", "deleted").
		Set("updated_at", "CURRENT_TIMESTAMP").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	if _, err = r.db.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// GetTasksForNotification получает задачи для отправки уведомлений
func (r *TaskRepositoryImpl) GetTasksForNotification(ctx context.Context, beforeTime time.Time) ([]*domain.Task, error) {
	query, args, err := r.sq.
		Select(
			"id", "title", "description", "status", "priority",
			"created_at", "updated_at", "completed_at", "notify_at", "user_id").
		From("tasks").
		Where(squirrel.NotEq{"notify_at": nil}).
		Where(squirrel.LtOrEq{"notify_at": beforeTime}).
		Where(squirrel.Eq{"status": "pending"}).
		OrderBy("notify_at ASC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification tasks: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// scanTasks сканирует строки и возвращает массив задач
func (r *TaskRepositoryImpl) scanTasks(rows *sql.Rows) ([]*domain.Task, error) {
	var tasks []*domain.Task

	for rows.Next() {
		task := &domain.Task{}

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
			&task.NotifyAt,
			&task.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return tasks, nil
}
