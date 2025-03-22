package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBDSN     string `env:"DB_DSN" env-default:"postgres://username:password@localhost:5432/postgres"`
	Port      string `env:"AUTH_PORT" env-default:"auth-service:50051"`
	JwtSecret string `env:"JWT_SECRET" env-default:"your_generated_secret"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Не удалось найти конфгурации: %v", err)
		return nil, err
	}
	return &cfg, nil
}
