package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"todolist/config"
	"todolist/internal/handler/telegram"
	"todolist/internal/repository/postgres"
	"todolist/internal/scheduler"
	"todolist/internal/usecase"
)

func main() {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Подключение к базе данных
	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Создание таблиц
	if err = db.CreateTables(); err != nil {
		logger.Fatal("failed to create tables", zap.Error(err))
	}

	// Инициализация репозиториев
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	taskRepo := postgres.NewTaskRepository(db)
	noteRepo := postgres.NewNoteRepository(db)

	// Инициализация сервисов
	authService := usecase.NewAuthService(userRepo, sessionRepo, cfg, logger)
	taskService := usecase.NewTaskService(taskRepo, logger)
	noteService := usecase.NewNoteService(noteRepo)

	// Инициализация телеграм бота
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		logger.Fatal("failed to create bot", zap.Error(err))
	}
	bot.Debug = cfg.Bot.Debug

	logger.Info("bot authorized", zap.String("username", bot.Self.UserName))

	// Инициализация сервиса уведомлений
	notificationService := usecase.NewNotificationService(bot, taskService, logger)

	// Инициализация обработчика телеграм бота
	telegramHandler := telegram.NewBot(bot, authService, taskService, noteService, notificationService, cfg, logger)

	// Инициализация планировщика
	cronScheduler := scheduler.NewCronScheduler(notificationService, authService, logger)

	// Создание контекста для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка системных сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Запуск телеграм бота
	wg.Add(1)
	go func() {
		defer wg.Done()
		if thErr := telegramHandler.Start(ctx); thErr != nil && !errors.Is(thErr, context.Canceled) {
			logger.Error("telegram bot error", zap.Error(thErr))
		}
	}()

	// Запуск планировщика
	wg.Add(1)
	go func() {
		defer wg.Done()
		if crErr := cronScheduler.Start(ctx); crErr != nil && !errors.Is(crErr, context.Canceled) {
			logger.Error("scheduler error", zap.Error(crErr))
		}
	}()

	// Запуск HTTP-сервера для метрик
	wg.Add(1)
	go func() {
		defer wg.Done()
		http.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{
			Addr: ":8080",
		}
		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("metrics server error", zap.Error(err))
			}
		}()
		<-ctx.Done()
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			logger.Error("metrics server shutdown error", zap.Error(err))
		}
	}()

	logger.Info("application started")

	// Ожидание сигнала завершения
	<-sigChan
	logger.Info("shutdown signal received")

	// Graceful shutdown
	cancel()
	wg.Wait()

	logger.Info("application stopped")
}
