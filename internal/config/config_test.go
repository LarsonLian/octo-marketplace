package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("MYSQL_DSN", "test-dsn")
	t.Setenv("API_PORT", "")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "")
	cfg := Load()
	if cfg.APIPort != "8092" {
		t.Fatalf("APIPort=%q want=8092", cfg.APIPort)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout=%v want=5s", cfg.ReadHeaderTimeout)
	}
}

func TestValidateAPI(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "valid", cfg: Config{MySQLDSN: "dsn", APIPort: "8092"}},
		{name: "missing dsn", cfg: Config{APIPort: "8092"}, wantErr: true},
		{name: "invalid port", cfg: Config{MySQLDSN: "dsn", APIPort: "0"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ValidateAPI(); (got != nil) != tt.wantErr {
				t.Fatalf("ValidateAPI() error=%v wantErr=%v", got, tt.wantErr)
			}
		})
	}
}

func TestInvalidDurationFallsBack(t *testing.T) {
	t.Setenv("MYSQL_DSN", "test-dsn")
	t.Setenv("HTTP_READ_TIMEOUT", "invalid")
	if got := Load().ReadTimeout; got != 15*time.Second {
		t.Fatalf("ReadTimeout=%v want=15s", got)
	}
}
