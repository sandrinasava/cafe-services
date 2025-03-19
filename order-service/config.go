package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Port string `env:"PORT" env-default:"8081"`
	} `env-prefix:"SERVER"`

	Kafka struct {
		Broker string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
		Topic  string `env:"KAFKA_TOPIC" env-default:"new_orders"`
	} `env-prefix:"KAFKA"`

	Redis struct {
		Host string `env:"REDIS_HOST" env-default:"redis:6379"`
	} `env-prefix:"REDIS"`

	AuthService struct {
		Address string `env:"AUTH_SERVICE_ADDRESS" env-default:"auth-service:50051"`
	} `env-prefix:"AUTH_SERVICE"`

	DB struct {
		DSN string `env:"DB_DSN" env-default:"postgres://user:password@localhost:5432/orders?sslmode=disable"`
	} `env-prefix:"DATABASE"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Не удалось найти конфгурации: %v", err)
		return nil, err
	}
	return &cfg, nil
}
