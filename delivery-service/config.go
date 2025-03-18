package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Port string `env:"PORT" env-default:"8083"`
	} `env-prefix:"SERVER"`

	Kafka struct {
		Broker string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
		Topic  string `env:"KAFKA_TOPIC" env-default:"ready_orders"`
	} `env-prefix:"KAFKA"`

	Postgres struct {
		Host     string `env:"POSTGRES_HOST" env-default:"postgres"`
		Port     string `env:"POSTGRES_PORT" env-default:"5432"`
		User     string `env:"POSTGRES_USER" env-default:"user"`
		Password string `env:"POSTGRES_PASSWORD" env-default:"password"`
		DBName   string `env:"POSTGRES_DB" env-default:"delivery_db"`
	} `env-prefix:"POSTGRES"`

	JWT struct {
		SecretKey string `env:"JWT_SECRET_KEY" env-required:"true"`
	} `env-prefix:"JWT"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
		return nil, err
	}
	return &cfg, nil
}
