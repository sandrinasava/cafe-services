# Используем официальный образ Go для сборки
FROM golang:1.20 AS builder

# Устанавливаем необходимые пакеты
RUN apt-get update && apt-get install -y \
    unzip \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-x86_64.zip && \
    unzip protoc-21.12-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/protoc /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/ && \
    rm -rf protoc3 protoc-21.12-linux-x86_64.zip

# Устанавливаем плагин protoc-gen-go
RUN go install github.com/golang/protobuf/protoc-gen-go@latest

# Устанавливаем плагин protoc-gen-go-grpc
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем остальные файлы
COPY . .

# Генерируем Go-код из auth.proto
RUN protoc --go_out=./proto --go-grpc_out=./proto proto/auth.proto

# Сборка бинарного файла
RUN go build -o auth-service main.go

# Используем минимальный образ для запуска
FROM alpine:latest

# Устанавливаем необходимые библиотеки
RUN apk add --no-cache ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем сгенерированный бинарный файл
COPY --from=builder /app/auth-service .

# Копируем сгенерированные файлы
COPY --from=builder /app/proto ./proto

# Копируем конфигурационные файлы
COPY --from=builder /app/config.go .

# Копируем init.sql
COPY --from=builder /app/init.sql .

# Устанавливаем порт
EXPOSE 50051

# Запуск сервиса
CMD ["./auth-service"]