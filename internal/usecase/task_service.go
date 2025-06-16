package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"todolist/internal/domain"
	"todolist/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// TaskService предоставляет методы для работы с задачами
type TaskService struct {
	taskRepository domain.TaskRepository
	logger         *zap.Logger
}

// NewTaskService создает новый экземпляр TaskService
func NewTaskService(taskRepository domain.TaskRepository, logger *zap.Logger) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		logger:         logger,
	}
}

// CreateTask создает новую задачу
func (s *TaskService) CreateTask(ctx context.Context, userID int64, title, description string, priority domain.TaskPriority) (*domain.Task, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("create_task", "POST"))
	defer timer.ObserveDuration()

	if strings.TrimSpace(title) == "" {
		metrics.ErrorsTotal.WithLabelValues("empty_title").Inc()
		return nil, fmt.Errorf("название задачи не может быть пустым")
	}

	task := &domain.Task{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      domain.TaskStatusPending,
		Priority:    priority,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.taskRepository.Create(ctx, task); err != nil {
		metrics.ErrorsTotal.WithLabelValues("create_task_db_error").Inc()
		s.logger.Error("failed to create task", zap.Error(err))
		return nil, fmt.Errorf("ошибка создания задачи")
	}

	metrics.TasksCreated.Inc()
	metrics.TasksTotal.WithLabelValues("pending").Inc()
	s.logger.Info("task created", zap.Int("task_id", task.ID), zap.Int64("user_id", userID))
	return task, nil
}

// GetTasks получает все задачи пользователя
func (s *TaskService) GetTasks(ctx context.Context, userID int64) ([]*domain.Task, error) {
	tasks, err := s.taskRepository.GetAll(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get tasks", zap.Error(err))
		return nil, fmt.Errorf("ошибка получения задач")
	}

	return tasks, nil
}

// GetTasksByStatus получает задачи пользователя по статусу
func (s *TaskService) GetTasksByStatus(ctx context.Context, userID int64, status domain.TaskStatus) ([]*domain.Task, error) {
	tasks, err := s.taskRepository.GetByUserID(ctx, userID, status)
	if err != nil {
		s.logger.Error("failed to get tasks by status", zap.Error(err))
		return nil, fmt.Errorf("ошибка получения задач")
	}

	return tasks, nil
}

// GetTaskByID получает задачу по ID
func (s *TaskService) GetTaskByID(ctx context.Context, taskID int, userID int64) (*domain.Task, error) {
	task, err := s.taskRepository.GetByID(ctx, taskID)
	if err != nil {
		s.logger.Error("failed to get task", zap.Error(err))
		return nil, fmt.Errorf("задача не найдена")
	}

	if task.UserID != userID {
		return nil, fmt.Errorf("задача не принадлежит пользователю")
	}

	return task, nil
}

// CompleteTask помечает задачу как выполненную
func (s *TaskService) CompleteTask(ctx context.Context, taskID int, userID int64) (*domain.Task, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("complete_task", "POST"))
	defer timer.ObserveDuration()

	task, err := s.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("complete_task_not_found").Inc()
		return nil, err
	}

	if task.IsCompleted() {
		metrics.ErrorsTotal.WithLabelValues("task_already_completed").Inc()
		return nil, fmt.Errorf("задача уже выполнена")
	}

	task.Complete()

	if err := s.taskRepository.Update(ctx, task); err != nil {
		metrics.ErrorsTotal.WithLabelValues("complete_task_db_error").Inc()
		s.logger.Error("failed to complete task", zap.Error(err))
		return nil, fmt.Errorf("ошибка завершения задачи")
	}

	metrics.TasksCompleted.Inc()
	metrics.TasksTotal.WithLabelValues("pending").Dec()
	metrics.TasksTotal.WithLabelValues("completed").Inc()
	s.logger.Info("task completed", zap.Int("task_id", taskID), zap.Int64("user_id", userID))
	return task, nil
}

// DeleteTask удаляет задачу
func (s *TaskService) DeleteTask(ctx context.Context, taskID int, userID int64) error {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("delete_task", "DELETE"))
	defer timer.ObserveDuration()

	task, err := s.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("delete_task_not_found").Inc()
		return err
	}

	if task.IsDeleted() {
		metrics.ErrorsTotal.WithLabelValues("task_already_deleted").Inc()
		return fmt.Errorf("задача уже удалена")
	}

	if err := s.taskRepository.Delete(ctx, taskID); err != nil {
		metrics.ErrorsTotal.WithLabelValues("delete_task_db_error").Inc()
		s.logger.Error("failed to delete task", zap.Error(err))
		return fmt.Errorf("ошибка удаления задачи")
	}

	metrics.TasksTotal.WithLabelValues(task.Status.String()).Dec()
	s.logger.Info("task deleted", zap.Int("task_id", taskID), zap.Int64("user_id", userID))
	return nil
}

// UpdateTask обновляет задачу
func (s *TaskService) UpdateTask(ctx context.Context, taskID int, userID int64, title, description string, priority domain.TaskPriority) (*domain.Task, error) {
	task, err := s.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	if task.IsCompleted() || task.IsDeleted() {
		return nil, fmt.Errorf("нельзя редактировать завершенную или удаленную задачу")
	}

	if strings.TrimSpace(title) != "" {
		task.Title = strings.TrimSpace(title)
	}
	task.Description = strings.TrimSpace(description)
	task.Priority = priority
	task.UpdatedAt = time.Now()

	if err := s.taskRepository.Update(ctx, task); err != nil {
		s.logger.Error("failed to update task", zap.Error(err))
		return nil, fmt.Errorf("ошибка обновления задачи")
	}

	s.logger.Info("task updated", zap.Int("task_id", taskID), zap.Int64("user_id", userID))
	return task, nil
}

// SetTaskNotification устанавливает уведомление для задачи
func (s *TaskService) SetTaskNotification(ctx context.Context, taskID int, userID int64, notifyAt time.Time) (*domain.Task, error) {
	task, err := s.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	if task.IsCompleted() || task.IsDeleted() {
		return nil, fmt.Errorf("нельзя установить уведомление для завершенной или удаленной задачи")
	}

	if notifyAt.Before(time.Now()) {
		return nil, fmt.Errorf("время уведомления должно быть в будущем")
	}

	task.SetNotification(notifyAt)

	if err := s.taskRepository.Update(ctx, task); err != nil {
		s.logger.Error("failed to set notification", zap.Error(err))
		return nil, fmt.Errorf("ошибка установки уведомления")
	}

	s.logger.Info("notification set", zap.Int("task_id", taskID), zap.Time("notify_at", notifyAt))
	return task, nil
}

// GetTasksForNotification получает задачи для отправки уведомлений
func (s *TaskService) GetTasksForNotification(ctx context.Context) ([]*domain.Task, error) {
	tasks, err := s.taskRepository.GetTasksForNotification(ctx, time.Now())
	if err != nil {
		s.logger.Error("failed to get notification tasks", zap.Error(err))
		return nil, err
	}

	return tasks, nil
}

// ParseTaskIDFromText извлекает ID задачи из текста
func (s *TaskService) ParseTaskIDFromText(text string) (int, error) {
	// Ищем числа в тексте
	parts := strings.Fields(text)
	for _, part := range parts {
		if id, err := strconv.Atoi(part); err == nil && id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf("не удалось найти ID задачи в тексте")
}

// FormatTaskList форматирует список задач для отображения
func (s *TaskService) FormatTaskList(tasks []*domain.Task) string {
	if len(tasks) == 0 {
		return "📝 Задач нет"
	}

	var result strings.Builder
	result.WriteString("📝 Ваши задачи:\n\n")

	for _, task := range tasks {
		status := "⏳"
		if task.IsCompleted() {
			status = "✅"
		}

		priority := ""
		switch task.Priority {
		case domain.TaskPriorityHigh:
			priority = "🔴"
		case domain.TaskPriorityMedium:
			priority = "🟡"
		case domain.TaskPriorityLow:
			priority = "🟢"
		}

		result.WriteString(fmt.Sprintf("%s %s [%d] %s\n", status, priority, task.ID, task.Title))

		if task.Description != "" {
			result.WriteString(fmt.Sprintf("   💬 %s\n", task.Description))
		}

		if task.NotifyAt != nil {
			result.WriteString(fmt.Sprintf("   ⏰ %s\n", task.NotifyAt.Format("02.01.2006 15:04")))
		}

		result.WriteString("\n")
	}

	return result.String()
}

// FormatTask форматирует одну задачу для отображения
func (s *TaskService) FormatTask(task *domain.Task) string {
	status := "⏳ Не выполнена"
	if task.IsCompleted() {
		status = "✅ Выполнена"
	}

	priority := ""
	switch task.Priority {
	case domain.TaskPriorityHigh:
		priority = "🔴 Высокий"
	case domain.TaskPriorityMedium:
		priority = "🟡 Средний"
	case domain.TaskPriorityLow:
		priority = "🟢 Низкий"
	}

	result := fmt.Sprintf("📋 Задача [%d]\n\n", task.ID)
	result += fmt.Sprintf("📌 Название: %s\n", task.Title)

	if task.Description != "" {
		result += fmt.Sprintf("💬 Описание: %s\n", task.Description)
	}

	result += fmt.Sprintf("📊 Статус: %s\n", status)
	result += fmt.Sprintf("🎯 Приоритет: %s\n", priority)
	result += fmt.Sprintf("📅 Создана: %s\n", task.CreatedAt.Format("02.01.2006 15:04"))

	if task.NotifyAt != nil {
		result += fmt.Sprintf("⏰ Уведомление: %s\n", task.NotifyAt.Format("02.01.2006 15:04"))
	}

	if task.CompletedAt != nil {
		result += fmt.Sprintf("✅ Завершена: %s\n", task.CompletedAt.Format("02.01.2006 15:04"))
	}

	return result
}
