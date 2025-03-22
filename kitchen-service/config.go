package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Port string `env:"PORT" env-default:"8082"`
	}

	Kafka struct {
		Broker   string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
		TopicIn  string `env:"KAFKA_TOPIC_IN" env-default:"new_orders"`
		TopicOut string `env:"KAFKA_TOPIC_OUT" env-default:"ready_orders"`
	}

	JWT struct {
		SecretKey string `env:"JWT_SECRET_KEY"`
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
