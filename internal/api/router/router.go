package router

import (
	"context"
	"encoding/json"
	"net/http"
)

type Pinger interface {
	PingContext(context.Context) error
}

func Public(database Pinger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", jsonStatus(http.StatusOK, "ok"))
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.PingContext(r.Context()); err != nil {
			writeStatus(w, http.StatusServiceUnavailable, "not_ready")
			return
		}
		writeStatus(w, http.StatusOK, "ready")
	})
	return mux
}

func jsonStatus(status int, value string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeStatus(w, status, value)
	}
}

func writeStatus(w http.ResponseWriter, status int, value string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": value})
}
