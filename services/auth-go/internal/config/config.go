package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const defaultGRPCPort = "50037"

type Config struct {
	Env            string               `toml:"env" yaml:"env" env:"APP_ENV" env-default:"development"`
	Port           string               `toml:"port" yaml:"port" env:"SERVICE_PORT" env-default:"50037"`
	CertsPath      string               `toml:"certs_path" yaml:"certs_path" env:"CERTS_PATH" env-default:"/etc/certs"`
	MigrationsPath string               `toml:"migrations_path" yaml:"migrations_path" env:"MIGRATIONS_PATH" env-default:"./internal/migrations/init.sql"`
	AccessSecret   string               `env:"ACCESS_SECRET_KEY" env-required:"true"`
	RefreshSecret  string               `env:"REFRESH_SECRET_KEY" env-required:"true"`
	Database       DatabaseConfig       `toml:"database" yaml:"database" env-prefix:"AUTH_PG_"`
	BootstrapAdmin BootstrapAdminConfig `toml:"bootstrap_admin" yaml:"bootstrap_admin" env-prefix:"AUTH_BOOTSTRAP_ADMIN_"`
}

type DatabaseConfig struct {
	Host     string `toml:"host" yaml:"host" env:"HOST" env-default:"postgres_auth"`
	Port     string `toml:"port" yaml:"port" env:"PORT" env-default:"5432"`
	User     string `toml:"user" yaml:"user" env:"USER" env-default:"hack"`
	Password string `toml:"password" yaml:"password" env:"PASSWORD" env-default:"hack"`
	Name     string `toml:"name" yaml:"name" env:"DATABASE" env-default:"authdb"`
	SSLMode  string `toml:"ssl_mode" yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
}

type BootstrapAdminConfig struct {
	Email       string `toml:"email" yaml:"email" env:"EMAIL" env-default:"admin@profdnk.local"`
	Password    string `toml:"password" yaml:"password" env:"PASSWORD" env-default:"admin12345"`
	FullName    string `toml:"full_name" yaml:"full_name" env:"FULL_NAME" env-default:"System Administrator"`
	Phone       string `toml:"phone" yaml:"phone" env:"PHONE" env-default:"+70000000000"`
	AccessUntil string `toml:"access_until" yaml:"access_until" env:"ACCESS_UNTIL" env-default:"2099-12-31"`
	Role        string `toml:"role" yaml:"role" env:"ROLE" env-default:"admin"`
}

func Load(path string) (Config, error) {
	cfg := Config{
		Port: defaultGRPCPort,
	}

	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := cleanenv.ReadConfig(path, &cfg); err != nil {
				return Config{}, fmt.Errorf("read file config: %w", err)
			}
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env config: %w", err)
	}

	if cfg.Port == "" {
		cfg.Port = defaultGRPCPort
	}

	return cfg, nil
}
