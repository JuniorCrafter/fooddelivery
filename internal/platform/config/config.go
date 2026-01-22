package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv      string `env:"APP_ENV,required"`
	HTTPPort    string `env:"HTTP_PORT,required"`
	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTSecret  string        `env:"JWT_SECRET,required"`
	AccessTTL  time.Duration `env:"ACCESS_TTL,required"`
	RefreshTTL time.Duration `env:"REFRESH_TTL,required"`

	BcryptCost int `env:"BCRYPT_COST" envDefault:"12"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
