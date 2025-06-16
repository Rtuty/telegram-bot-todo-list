package usecase

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"todolist/internal/domain"
	"todolist/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// NotificationService предоставляет методы для отправки уведомлений
type NotificationService struct {
	bot         *tgbotapi.BotAPI
	taskService *TaskService
	logger      *zap.Logger
}

// NewNotificationService создает новый экземпляр NotificationService
func NewNotificationService(bot *tgbotapi.BotAPI, taskService *TaskService, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		bot:         bot,
		taskService: taskService,
		logger:      logger,
	}
}

// SendTaskNotifications отправляет уведомления о задачах
func (s *NotificationService) SendTaskNotifications(ctx context.Context) error {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("send_notifications", "POST"))
	defer timer.ObserveDuration()

	tasks, err := s.taskService.GetTasksForNotification(ctx)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("get_notification_tasks_error").Inc()
		s.logger.Error("failed to get notification tasks", zap.Error(err))
		return err
	}

	successCount := 0
	for _, task := range tasks {
		if err := s.sendTaskNotification(task); err != nil {
			metrics.ErrorsTotal.WithLabelValues("send_notification_error").Inc()
			s.logger.Error("failed to send notification",
				zap.Int("task_id", task.ID),
				zap.Int64("user_id", task.UserID),
				zap.Error(err))
			continue
		}
		successCount++

		// Очищаем время уведомления, чтобы не отправлять повторно
		task.NotifyAt = nil
		if err := s.taskService.taskRepository.Update(ctx, task); err != nil {
			metrics.ErrorsTotal.WithLabelValues("clear_notification_time_error").Inc()
			s.logger.Error("failed to clear notification time",
				zap.Int("task_id", task.ID),
				zap.Error(err))
		}
	}

	if len(tasks) > 0 {
		s.logger.Info("notifications sent", zap.Int("count", successCount), zap.Int("total", len(tasks)))
	}

	return nil
}

// sendTaskNotification отправляет уведомление о конкретной задаче
func (s *NotificationService) sendTaskNotification(task *domain.Task) error {
	message := fmt.Sprintf("⏰ Напоминание о задаче!\n\n")
	message += fmt.Sprintf("📌 %s\n", task.Title)

	if task.Description != "" {
		message += fmt.Sprintf("💬 %s\n", task.Description)
	}

	message += fmt.Sprintf("\n🆔 Задача [%d]", task.ID)

	msg := tgbotapi.NewMessage(task.UserID, message)

	// Добавляем inline клавиатуру для быстрых действий
	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{
					Text:         "✅ Выполнить",
					CallbackData: &[]string{fmt.Sprintf("complete_%d", task.ID)}[0],
				},
				tgbotapi.InlineKeyboardButton{
					Text:         "📋 Подробнее",
					CallbackData: &[]string{fmt.Sprintf("show_%d", task.ID)}[0],
				},
			},
		},
	}
	msg.ReplyMarkup = keyboard

	_, err := s.bot.Send(msg)
	return err
}

// SendMessage отправляет сообщение пользователю
func (s *NotificationService) SendMessage(userID int64, text string) error {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("send_message", "POST"))
	defer timer.ObserveDuration()

	msg := tgbotapi.NewMessage(userID, text)
	_, err := s.bot.Send(msg)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("send_message_error").Inc()
		s.logger.Error("failed to send message",
			zap.Int64("user_id", userID),
			zap.Error(err))
	}
	return err
}
