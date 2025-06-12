package domain

import (
	"time"
)

// TaskStatus представляет статус задачи
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusDeleted   TaskStatus = "deleted"
)

// TaskPriority представляет приоритет задачи
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// Task представляет задачу в системе
type Task struct {
	ID          int          `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at" db:"completed_at"`
	NotifyAt    *time.Time   `json:"notify_at" db:"notify_at"`
	UserID      int64        `json:"user_id" db:"user_id"`
}

// IsCompleted проверяет, завершена ли задача
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsDeleted проверяет, удалена ли задача
func (t *Task) IsDeleted() bool {
	return t.Status == TaskStatusDeleted
}

// CanNotify проверяет, нужно ли отправить уведомление
func (t *Task) CanNotify() bool {
	return t.NotifyAt != nil &&
		t.NotifyAt.Before(time.Now()) &&
		!t.IsCompleted() &&
		!t.IsDeleted()
}

// Complete помечает задачу как выполненную
func (t *Task) Complete() {
	t.Status = TaskStatusCompleted
	t.UpdatedAt = time.Now()
	now := time.Now()
	t.CompletedAt = &now
}

// Delete помечает задачу как удаленную
func (t *Task) Delete() {
	t.Status = TaskStatusDeleted
	t.UpdatedAt = time.Now()
}

// SetNotification устанавливает время уведомления
func (t *Task) SetNotification(notifyAt time.Time) {
	t.NotifyAt = &notifyAt
	t.UpdatedAt = time.Now()
}
