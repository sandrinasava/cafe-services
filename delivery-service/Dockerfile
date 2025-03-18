FROM golang:1.23.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o delivery-service .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/delivery-service .

CMD ["./delivery-service"]