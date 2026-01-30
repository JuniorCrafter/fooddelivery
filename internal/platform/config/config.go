package config

import (
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string `env:"DATABASE_URL"`
	LocalDBURL  string `env:"LOCAL_DATABASE_URL"`
	AuthPort    string `env:"AUTH_PORT" envDefault:":8081"`
	RabbitMQURL string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
	}

	// Умная замена хостов для локального запуска
	if _, err := os.Stat("/.dockerenv"); err != nil {
		// Если мы не в докере, меняем имена контейнеров на localhost
		cfg.DatabaseURL = strings.Replace(cfg.DatabaseURL, "@postgres:", "@localhost:", 1)
		cfg.RabbitMQURL = strings.Replace(cfg.RabbitMQURL, "@rabbitmq:", "@localhost:", 1)
	}

	return &cfg
}
