# Билд стадии
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /pubsub-server ./cmd/server/main.go

# Финальная стадия
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder стадии
COPY --from=builder /pubsub-server /app/pubsub-server
# Копируем .env файл
COPY .env /app/.env

# Команда запуска
CMD ["/app/pubsub-server"]