package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"todolist/internal/usecase"
)

// CronScheduler представляет планировщик задач
type CronScheduler struct {
	cron                *cron.Cron
	notificationService *usecase.NotificationService
	authService         *usecase.AuthService
	logger              *zap.Logger
}

// NewCronScheduler создает новый экземпляр планировщика
func NewCronScheduler(
	notificationService *usecase.NotificationService,
	authService *usecase.AuthService,
	logger *zap.Logger,
) *CronScheduler {
	c := cron.New(cron.WithSeconds())

	return &CronScheduler{
		cron:                c,
		notificationService: notificationService,
		authService:         authService,
		logger:              logger,
	}
}

// Start запускает планировщик
func (s *CronScheduler) Start(ctx context.Context) error {
	s.logger.Info("starting cron scheduler...")

	// Отправка уведомлений каждую минуту
	_, err := s.cron.AddFunc("0 * * * * *", func() {
		s.sendNotifications(ctx)
	})
	if err != nil {
		return err
	}

	// Очистка истекших сессий каждые 30 минут
	_, err = s.cron.AddFunc("0 */30 * * * *", func() {
		s.cleanupSessions(ctx)
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info("cron scheduler started")

	// Ждем завершения контекста
	<-ctx.Done()

	s.logger.Info("stopping cron scheduler...")
	stopCtx := s.cron.Stop()
	<-stopCtx.Done()
	s.logger.Info("cron scheduler stopped")

	return ctx.Err()
}

// sendNotifications отправляет уведомления о задачах
func (s *CronScheduler) sendNotifications(ctx context.Context) {
	s.logger.Debug("checking for notifications to send...")

	if err := s.notificationService.SendTaskNotifications(ctx); err != nil {
		s.logger.Error("failed to send notifications", zap.Error(err))
	}
}

// cleanupSessions очищает истекшие сессии
func (s *CronScheduler) cleanupSessions(ctx context.Context) {
	s.logger.Debug("cleaning up expired sessions...")

	if err := s.authService.CleanupExpiredSessions(ctx); err != nil {
		s.logger.Error("failed to cleanup sessions", zap.Error(err))
	}
}
