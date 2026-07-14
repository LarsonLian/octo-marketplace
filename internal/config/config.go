package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	MySQLDSN          string
	OctoAPIURL        string
	APIPort           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func Load() Config {
	return Config{
		MySQLDSN:          env("MYSQL_DSN", ""),
		OctoAPIURL:        strings.TrimRight(env("OCTO_API_URL", ""), "/"),
		APIPort:           env("API_PORT", "8092"),
		ReadHeaderTimeout: envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:       envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
	}
}

func (c Config) ValidateAPI() error {
	if c.MySQLDSN == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}
	return validatePort(c.APIPort, "API_PORT")
}

func validatePort(value, name string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%s must be a valid TCP port", name)
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
