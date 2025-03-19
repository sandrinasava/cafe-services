# Этап сборки
FROM golang:1.20 AS builder

# Установка необходимых инструментов
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" \
    && curl -LO $PB_REL/download/v25.1/protoc-25.1-linux-x86_64.zip \
    && unzip protoc-25.1-linux-x86_64.zip -d $HOME/.local \
    && rm protoc-25.1-linux-x86_64.zip \
    && export PATH="$PATH:$HOME/.local/bin" \
    && echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc \
    && source ~/.bashrc

# Установка Go плагинов для protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Копирование модуля и зависимостей
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Генерация кода из .proto файлов
RUN protoc --go_out=. --go-grpc_out=. ./proto/auth.proto

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