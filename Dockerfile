# Этап сборки: Используем официальный образ Go для сборки приложения
FROM golang:1.24-alpine AS builder 
# Или golang:latest-alpine для самой новой версии

# Устанавливаем необходимые зависимости для сборки (если есть)
RUN apk add --no-cache git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod .
COPY go.sum .

# Загружаем все зависимости
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /my-task-app .

# Этап запуска: Используем облегченный образ для финального образа
FROM alpine:latest

# Устанавливаем необходимые зависимости для запуска (если есть, например, сертификаты)
RUN apk add --no-cache ca-certificates

# Копируем собранное приложение из образа builder
COPY --from=builder /my-task-app /my-task-app

# Открываем порт, который будет использовать приложение
EXPOSE 8080

# Определяем команду для запуска приложения
CMD ["/my-task-app"]