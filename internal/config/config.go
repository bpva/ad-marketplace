package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	HTTP     HTTP
	Postgres Postgres
	Telegram Telegram
	JWT      JWT
	Logger   Logger
}

type Logger struct {
	Level        string `yaml:"level" env-default:"info"`
	OTLPEnabled  bool   `yaml:"otlp_enabled" env:"OTLP_ENABLED" env-default:"false"`
	OTLPEndpoint string `yaml:"otlp_endpoint" env:"OTLP_ENDPOINT" env-default:"https://otlp-gateway-prod-eu-west-2.grafana.net/otlp"`
	OTLPAuth     string `env:"OTLP_AUTH"`
}

type HTTP struct {
	Port            string        `yaml:"port" env:"HTTP_PORT" env-default:"8090"`
	PrivatePort     string        `yaml:"private_port" env:"HTTP_PRIVATE_PORT" env-default:"8091"`
	FrontendURL     string        `env:"FRONTEND_URL" env-default:"http://localhost:1313"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"30s"`
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
