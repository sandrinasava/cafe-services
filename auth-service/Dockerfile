
FROM golang:1.23.3-alpine AS builder

# Проверяем версию Go
RUN go version

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myproject -ldflags="-w -s" .

FROM golang:1.23.3-alpine

WORKDIR /app

COPY --from=builder /app/myproject /app/myproject

RUN chmod +x /app/myproject

CMD ["./myproject"]