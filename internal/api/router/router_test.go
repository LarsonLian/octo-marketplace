package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	marketmiddleware "github.com/Mininglamp-OSS/octo-marketplace/internal/middleware"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
	"github.com/gin-gonic/gin"
)

type stubPinger struct{ err error }

func (p stubPinger) PingContext(context.Context) error { return p.err }

func init() { gin.SetMode(gin.TestMode) }

func testAuthenticator() *marketmiddleware.Authenticator {
	return marketmiddleware.NewAuthenticator(false, nil, model.Identity{UID: "dev-user", Name: "Developer"}, "dev-space")
}

func testStorageConfig() StorageConfig {
	return StorageConfig{Driver: "local", LocalDir: "/tmp/marketplace-test-uploads", BaseURL: "http://127.0.0.1:8092", MaxMB: 20}
}

func TestHealthz(t *testing.T) {
	recorder := httptest.NewRecorder()
	Public(stubPinger{}, testAuthenticator(), testStorageConfig()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", recorder.Code, http.StatusOK)
	}
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "ready", want: http.StatusOK},
		{name: "database unavailable", err: errors.New("down"), want: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			Public(stubPinger{err: tt.err}, testAuthenticator(), testStorageConfig()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if recorder.Code != tt.want {
				t.Fatalf("status=%d want=%d", recorder.Code, tt.want)
			}
		})
	}
}

func TestSessionUsesDevelopmentIdentity(t *testing.T) {
	recorder := httptest.NewRecorder()
	Public(stubPinger{}, testAuthenticator(), testStorageConfig()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/session", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "{\"name\":\"Developer\",\"space_id\":\"dev-space\",\"uid\":\"dev-user\"}" {
		t.Fatalf("body=%q", body)
	}
}
