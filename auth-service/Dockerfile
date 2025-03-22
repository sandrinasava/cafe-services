
FROM golang:1.23.3-alpine AS builder

ENV GOFLAGS=""
ENV GOVERSION="go1.23.3"

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