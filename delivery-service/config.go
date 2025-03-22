package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Kafka struct {
		Broker string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
		Topic  string `env:"KAFKA_TOPIC_READY" env-default:"ready_orders"`
	}

	DB struct {
		DSN string `env:"DB_DSN" env-default:"postgres://username:password@localhost:5432/postgres"`
	}
	Redis struct {
		Host     string `env:"REDIS_HOST" env-default:"redis:6379"`
		Password string `env:"REDIS_PASSWORD" env-default:"defaultpassword"`
	}
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
		return nil, err
	}
	return &cfg, nil
}
