package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `env:"ENV" env-default:"local"`
	HTTP     HTTP     `yaml:"http"`
	Postgres Postgres `yaml:"postgres"`
	Telegram Telegram `yaml:"telegram"`
	JWT      JWT      `yaml:"jwt"`
	Logger   Logger   `yaml:"logger"`
}

type Logger struct {
	Level       string `yaml:"level" yaml-default:"info"`
	OTLPEnabled bool   `env:"OTLP_ENABLED" env-default:"false" yaml:"otlp_enabled"`
}

type HTTP struct {
	Port            string        `env:"HTTP_PORT" env-default:"8090" yaml:"port"`
	PrivatePort     string        `env:"HTTP_PRIVATE_PORT" env-default:"8091" yaml:"private_port"`
	FrontendURL     string        `env:"FRONTEND_URL" env-default:"http://localhost:1313"`
	ShutdownTimeout time.Duration `env-default:"30s" yaml:"shutdown_timeout"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB" env-default:"ad_marketplace"`
}

type Telegram struct {
	BotToken   string `env:"BOT_TOKEN" env-required:"true"`
	BaseURL    string `env:"VITE_API_URL"`
	MiniAppURL string `env:"FRONTEND_URL"`

	APIId   int    `env:"TG_API_ID"`
	APIHash string `env:"TG_API_HASH"`

	RetryDelay time.Duration `yaml:"retry_delay" env:"TG_RETRY_DELAY" env-default:"10s"`
	MaxRetries int           `yaml:"max_retries" env:"TG_MAX_RETRIES" env-default:"5"`
}

type JWT struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
