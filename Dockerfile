# Используем официальный образ Go для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модуля Go
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем CA сертификаты и timezone
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для запуска приложения
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарный файл из builder стейджа
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Меняем владельца файлов
RUN chown -R appuser:appgroup /root

# Переключаемся на непривилегированного пользователя
USER appuser

# Указываем порт (если будет веб интерфейс)
EXPOSE 8080

# Команда запуска
CMD ["./main"] 