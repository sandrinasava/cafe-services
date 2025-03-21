package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Port string `env:"PORT" env-default:"8081"`
	}

	Kafka struct {
		Broker string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
		Topic  string `env:"KAFKA_TOPIC" env-default:"new_orders"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST" env-default:"redis:6379"`
		Password string `env:"REDIS_PASSWORD" env-default:"defaultpassword"`
	}

	AuthService struct {
		Address string `env:"AUTH_SERVICE_ADDRESS" env-default:"auth-service:50051"`
	}

	DB struct {
		DSN string `env:"DB_DSN" env-default:"postgres://username:password@localhost:5432/postgres"`
	}
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Не удалось найти конфгурации: %v", err)
		return nil, err
	}
	return &cfg, nil
}
