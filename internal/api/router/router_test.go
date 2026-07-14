package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct{ err error }

func (p stubPinger) PingContext(context.Context) error { return p.err }

func TestHealthz(t *testing.T) {
	recorder := httptest.NewRecorder()
	Public(stubPinger{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
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
			Public(stubPinger{err: tt.err}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if recorder.Code != tt.want {
				t.Fatalf("status=%d want=%d", recorder.Code, tt.want)
			}
		})
	}
}
