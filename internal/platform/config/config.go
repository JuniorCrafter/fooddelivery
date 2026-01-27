package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string `env:"DATABASE_URL"`
	LocalDBURL  string `env:"LOCAL_DATABASE_URL"`
	AuthPort    string `env:"AUTH_PORT" envDefault:":8081"`
}

func Load() *Config {
	// 1. Загружаем файл.env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл.env не найден, используем переменные окружения")
	}

	cfg := Config{}
	// 2. Парсим переменные в структуру
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
	}

	// 3. Проверка: если мы запущены НЕ в Docker, используем LOCAL_DATABASE_URL
	// Простой способ проверить Docker — наличие файла /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err != nil {
		if cfg.LocalDBURL != "" {
			cfg.DatabaseURL = cfg.LocalDBURL
		}
	}

	return &cfg
}
