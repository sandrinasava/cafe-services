FROM golang:1.23.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o kitchen-service .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/kitchen-service .

CMD ["./kitchen-service"]