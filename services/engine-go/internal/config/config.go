package config

import (
	"fmt"
	"net/url"
	"os"

	"sourcecraft.dev/benzo/testengine/pkg/log"

	"github.com/BurntSushi/toml"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	defaultGRPCPort = 50036
)

// LoggerConfig определяет конфигурацию логирования приложения
type LoggerConfig struct {
	Level      string `toml:"level"`
	TimeFormat string `toml:"time_format"`
}

type GRPCConfig struct {
	Port uint16 `toml:"port"`
}

type ConnectionConfig struct {
	Host     string `toml:"host" env:"HOST" env-required:"true"`
	Port     uint16 `toml:"port" env:"PORT" env-required:"true" env-default:"5432"`
	User     string `toml:"user" env:"USER" env-required:"true"`
	Password string `toml:"password" env:"PASSWORD" env-required:"true"`
	Database string `toml:"database" env:"DATABASE" env-required:"true"`
}

type PostgresConfig struct {
	ConnectionConfig
	SSLMode string `toml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
}

type Config struct {
	Log       LoggerConfig
	Port      uint16         `env:"SERVICE_PORT" env-required:"true"`
	Postgres  PostgresConfig `env-prefix:"ENGINE_PG_"`
	CertsPath string         `env:"CERTS_PATH" env-required:"true"`
}

func defaultConfig() Config {
	return Config{
		Log: LoggerConfig{
			Level:      "info",
			TimeFormat: "2006-01-02 15:04:05",
		},
		Port: defaultGRPCPort,
	}
}

func (c *Config) ApplyDefaults() {
	if c.Port == 0 {
		c.Port = defaultGRPCPort
	}
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()

	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to decode toml: %w", err)
		}
	} else {
		l := log.New(cfg.Log.Level, cfg.Log.TimeFormat)
		l.Warn("Config file not found or path is wrong, relying on environment variables")
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func applyEnvVariables(cfg *Config) error {
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	return cleanenv.ReadEnv(cfg)
}

func (pgCfg PostgresConfig) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(pgCfg.User, pgCfg.Password),
		Host:   fmt.Sprintf("%s:%d", pgCfg.Host, pgCfg.Port),
		Path:   pgCfg.Database,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String()
}
