# Этап сборки
FROM golang:1.23.3 AS builder

# Копирование модуля и зависимостей
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o myproject .

# Этап запуска
FROM alpine:latest

# Установка необходимых зависимостей
RUN apk add --no-cache ca-certificates

# Копирование собранного приложения из этапа сборки
COPY --from=builder /app/myproject /app/myproject

# Установка рабочей директории
WORKDIR /app

# Запуск приложения
CMD ["./myproject"]