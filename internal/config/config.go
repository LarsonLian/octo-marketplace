package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MySQLDSN   string
	OctoAPIURL string
	APIPort    string
}

func Load() Config {
	return Config{
		MySQLDSN:   env("MYSQL_DSN", ""),
		OctoAPIURL: strings.TrimRight(env("OCTO_API_URL", ""), "/"),
		APIPort:    env("API_PORT", "8092"),
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
