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
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
		return nil, err
	}
	return &cfg, nil
}
