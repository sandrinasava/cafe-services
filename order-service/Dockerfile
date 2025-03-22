FROM golang:1.23.3-alpine AS builder

WORKDIR /app

ENV GOFLAGS=""
ENV GOVERSION="go1.23.3"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/order-service .

CMD ["./order-service"]