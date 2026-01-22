package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTP
	Postgres Postgres
	Telegram Telegram
}

type HTTP struct {
	Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB" env-default:"ad_marketplace"`
}

type Telegram struct {
	BotToken string `env:"BOT_TOKEN" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
