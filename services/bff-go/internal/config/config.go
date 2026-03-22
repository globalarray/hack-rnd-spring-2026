package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort                  string
	LogLevel                  string
	DefaultClientReportFormat string
	PublicBaseURL             string
	Engine                    GRPCClientConfig
	Analytics                 GRPCClientConfig
	Auth                      GRPCClientConfig
	SMTP                      SMTPConfig
}

type GRPCClientConfig struct {
	Address        string
	Insecure       bool
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
	ServerName     string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:                  getEnv("PORT", "8080"),
		LogLevel:                  strings.ToLower(getEnv("LOG_LEVEL", "info")),
		DefaultClientReportFormat: getEnv("DEFAULT_CLIENT_REPORT_FORMAT", "client_docx"),
		PublicBaseURL:             getEnv("PUBLIC_BASE_URL", "http://localhost:3000"),
		Engine: GRPCClientConfig{
			Address:        getEnv("ENGINE_ADDR", "test-engine:50036"),
			Insecure:       getEnvBool("ENGINE_INSECURE", false),
			CACertPath:     getEnv("ENGINE_CA_CERT_PATH", getEnv("CA_CERT_PATH", "")),
			ClientCertPath: getEnv("ENGINE_CLIENT_CERT_PATH", getEnv("CLIENT_CERT_PATH", "")),
			ClientKeyPath:  getEnv("ENGINE_CLIENT_KEY_PATH", getEnv("CLIENT_KEY_PATH", "")),
			ServerName:     getEnv("ENGINE_SERVER_NAME", ""),
		},
		Analytics: GRPCClientConfig{
			Address:        getEnv("ANALYTICS_ADDR", "analytics-python:50051"),
			Insecure:       getEnvBool("ANALYTICS_INSECURE", true),
			CACertPath:     getEnv("ANALYTICS_CA_CERT_PATH", ""),
			ClientCertPath: getEnv("ANALYTICS_CLIENT_CERT_PATH", ""),
			ClientKeyPath:  getEnv("ANALYTICS_CLIENT_KEY_PATH", ""),
			ServerName:     getEnv("ANALYTICS_SERVER_NAME", ""),
		},
		Auth: GRPCClientConfig{
			Address:        getEnv("AUTH_ADDR", "auth-go:50037"),
			Insecure:       getEnvBool("AUTH_INSECURE", false),
			CACertPath:     getEnv("AUTH_CA_CERT_PATH", getEnv("CA_CERT_PATH", "")),
			ClientCertPath: getEnv("AUTH_CLIENT_CERT_PATH", getEnv("CLIENT_CERT_PATH", "")),
			ClientKeyPath:  getEnv("AUTH_CLIENT_KEY_PATH", getEnv("CLIENT_KEY_PATH", "")),
			ServerName:     getEnv("AUTH_SERVER_NAME", getEnv("ENGINE_SERVER_NAME", "")),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			UseTLS:   getEnvBool("SMTP_USE_TLS", false),
		},
	}

	if strings.TrimSpace(cfg.HTTPPort) == "" {
		return Config{}, fmt.Errorf("PORT must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}
