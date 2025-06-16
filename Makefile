.PHONY: lint build run test clean docker-build docker-up docker-down

# Переменные
BINARY_NAME=todobot
LINT_VERSION=v1.55.2

# Команды для линтера
lint:
	@echo "Установка golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $(LINT_VERSION); \
	fi
	@echo "Запуск линтера..."
	@golangci-lint run ./...

# Команды для сборки и запуска
build:
	@echo "Сборка приложения..."
	@go build -o $(BINARY_NAME) ./cmd/main.go

run: build
	@echo "Запуск приложения..."
	@./$(BINARY_NAME)

test:
	@echo "Запуск тестов..."
	@go test -v ./...

clean:
	@echo "Очистка..."
	@rm -f $(BINARY_NAME)
	@go clean

# Docker команды
docker-build:
	@echo "Сборка Docker образа..."
	@docker-compose build

docker-up:
	@echo "Запуск Docker контейнеров..."
	@docker-compose up -d

docker-down:
	@echo "Остановка Docker контейнеров..."
	@docker-compose down

# Команда для проверки и исправления форматирования кода
fmt:
	@echo "Форматирование кода..."
	@go fmt ./...

# Команда для проверки зависимостей
deps:
	@echo "Проверка и обновление зависимостей..."
	@go mod tidy
	@go mod verify

# Команда для проверки безопасности зависимостей
security:
	@echo "Проверка безопасности зависимостей..."
	@go list -json -m all | nancy sleuth

# Команда для генерации документации
docs:
	@echo "Генерация документации..."
	@godoc -http=:6060

# Команда для проверки всего проекта
check: fmt lint test deps security

# Команда для полной сборки и запуска
all: check build docker-build docker-up 