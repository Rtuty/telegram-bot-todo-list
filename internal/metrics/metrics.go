package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Метрики для задач
	TasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todobot_tasks_total",
			Help: "Общее количество задач",
		},
		[]string{"status"},
	)

	TasksCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todobot_tasks_created_total",
			Help: "Количество созданных задач",
		},
	)

	TasksCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todobot_tasks_completed_total",
			Help: "Количество выполненных задач",
		},
	)

	// Метрики для пользователей
	UsersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todobot_users_total",
			Help: "Общее количество пользователей",
		},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "todobot_active_users",
			Help: "Количество активных пользователей",
		},
	)

	// Метрики для запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "todobot_request_duration_seconds",
			Help:    "Длительность обработки запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method"},
	)

	// Метрики для ошибок
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todobot_errors_total",
			Help: "Общее количество ошибок",
		},
		[]string{"type"},
	)
)
