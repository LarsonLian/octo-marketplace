package router

import (
	"context"
	"encoding/json"
	"net/http"

	marketmiddleware "github.com/Mininglamp-OSS/octo-marketplace/internal/middleware"
)

type Pinger interface {
	PingContext(context.Context) error
}

func Public(database Pinger, authenticator *marketmiddleware.Authenticator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", jsonStatus(http.StatusOK, "ok"))
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.PingContext(r.Context()); err != nil {
			writeStatus(w, http.StatusServiceUnavailable, "not_ready")
			return
		}
		writeStatus(w, http.StatusOK, "ready")
	})
	session := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := marketmiddleware.IdentityFromContext(r.Context())
		if !ok {
			writeStatus(w, http.StatusInternalServerError, "error")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uid":      identity.UID,
			"name":     identity.Name,
			"space_id": marketmiddleware.SpaceIDFromContext(r.Context()),
		})
	})
	mux.Handle("GET /api/v1/session", authenticator.Wrap(session))
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
