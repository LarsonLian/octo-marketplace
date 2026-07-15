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
	AuthEnabled       bool
	AuthCacheTTL      time.Duration
	AuthCacheCapacity int
	DevAuthUID        string
	DevAuthName       string
	DevSpaceID        string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	// Object storage (OSS/S3) configuration for skill file uploads.
	StorageDriver   string // "local" or "oss"
	LocalStorageDir string
	OSSEndpoint     string
	OSSBucket       string
	OSSAccessKey    string
	OSSSecretKey    string
	OSSPublicBase     string
	OSSRegion         string
	OSSPublicEndpoint string // external endpoint for presigned URLs
	MaxUploadMB       int
}

func Load() Config {
	return Config{
		MySQLDSN:          env("MYSQL_DSN", ""),
		OctoAPIURL:        strings.TrimRight(env("OCTO_API_URL", ""), "/"),
		APIPort:           env("API_PORT", "8092"),
		AuthEnabled:       envBool("AUTH_ENABLED", false),
		AuthCacheTTL:      envDuration("AUTH_CACHE_TTL", 30*time.Second),
		AuthCacheCapacity: envInt("AUTH_CACHE_CAPACITY", 10000),
		DevAuthUID:        env("DEV_AUTH_UID", "dev-user"),
		DevAuthName:       env("DEV_AUTH_NAME", "Developer"),
		DevSpaceID:        env("DEV_SPACE_ID", "dev-space"),
		ReadHeaderTimeout: envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:       envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),

		StorageDriver:   env("STORAGE_DRIVER", "local"),
		LocalStorageDir: env("LOCAL_STORAGE_DIR", "/tmp/marketplace-uploads"),
		OSSEndpoint:     env("OSS_ENDPOINT", ""),
		OSSBucket:       env("OSS_BUCKET", ""),
		OSSAccessKey:    env("OSS_ACCESS_KEY", ""),
		OSSSecretKey:    env("OSS_SECRET_KEY", ""),
		OSSPublicBase:     strings.TrimRight(env("OSS_PUBLIC_BASE_URL", ""), "/"),
		OSSRegion:         env("OSS_REGION", "us-east-1"),
		OSSPublicEndpoint: strings.TrimRight(env("OSS_PUBLIC_ENDPOINT", ""), "/"),
		MaxUploadMB:       envInt("MAX_UPLOAD_MB", 20),
	}
}

func (c Config) ValidateAPI() error {
	if c.MySQLDSN == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}
	if c.AuthEnabled && c.OctoAPIURL == "" {
		return fmt.Errorf("OCTO_API_URL is required when AUTH_ENABLED=true")
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

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
