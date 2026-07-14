package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/auth"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
)

type contextKey string

const (
	identityKey contextKey = "marketplace.identity"
	spaceKey    contextKey = "marketplace.space_id"
)

type Authenticator struct {
	enabled     bool
	resolver    auth.Resolver
	devIdentity model.Identity
	devSpaceID  string
}

func NewAuthenticator(enabled bool, resolver auth.Resolver, devIdentity model.Identity, devSpaceID string) *Authenticator {
	return &Authenticator{
		enabled:     enabled,
		resolver:    resolver,
		devIdentity: devIdentity,
		devSpaceID:  devSpaceID,
	}
}

func (a *Authenticator) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			spaceID := strings.TrimSpace(r.Header.Get("X-Space-Id"))
			if spaceID == "" {
				spaceID = a.devSpaceID
			}
			next.ServeHTTP(w, r.WithContext(withAuthContext(r.Context(), a.devIdentity, spaceID)))
			return
		}

		token := requestToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "err.marketplace.authentication_required", "Authentication is required.")
			return
		}
		if a.resolver == nil {
			writeError(w, http.StatusServiceUnavailable, "err.marketplace.auth_unavailable", "Authentication service is unavailable.")
			return
		}
		identity, err := a.resolver.Resolve(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "err.marketplace.auth_unavailable", "Authentication service is unavailable.")
			return
		}
		if identity.UID == "" {
			writeError(w, http.StatusUnauthorized, "err.marketplace.invalid_token", "Invalid or expired token.")
			return
		}
		if !identity.ContextIncluded {
			writeError(w, http.StatusServiceUnavailable, "err.marketplace.auth_context_unavailable", "Authorization context is unavailable.")
			return
		}

		spaceID := strings.TrimSpace(r.Header.Get("X-Space-Id"))
		if spaceID == "" {
			writeError(w, http.StatusBadRequest, "err.marketplace.space_required", "X-Space-Id header is required.")
			return
		}
		if !contains(identity.Spaces, spaceID) {
			writeError(w, http.StatusForbidden, "err.marketplace.space_forbidden", "Access to this Space is forbidden.")
			return
		}

		next.ServeHTTP(w, r.WithContext(withAuthContext(r.Context(), identity, spaceID)))
	})
}

func IdentityFromContext(ctx context.Context) (model.Identity, bool) {
	identity, ok := ctx.Value(identityKey).(model.Identity)
	return identity, ok
}

func SpaceIDFromContext(ctx context.Context) string {
	spaceID, _ := ctx.Value(spaceKey).(string)
	return spaceID
}

func OwnsBot(ctx context.Context, botID string) bool {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return false
	}
	spaceID := SpaceIDFromContext(ctx)
	return contains(identity.OwnedBotsBySpace[spaceID], botID)
}

func withAuthContext(ctx context.Context, identity model.Identity, spaceID string) context.Context {
	ctx = context.WithValue(ctx, identityKey, identity)
	return context.WithValue(ctx, spaceKey, spaceID)
}

func requestToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("Token")); token != "" {
		return token
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authorization) > 7 && strings.EqualFold(authorization[:7], "Bearer ") {
		return strings.TrimSpace(authorization[7:])
	}
	return ""
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":        code,
			"message":     message,
			"http_status": status,
		},
	})
}
