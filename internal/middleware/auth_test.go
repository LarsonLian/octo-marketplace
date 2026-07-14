package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
)

type stubResolver struct {
	identity model.Identity
	err      error
}

func (r stubResolver) Resolve(context.Context, string) (model.Identity, error) {
	return r.identity, r.err
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	identity, _ := IdentityFromContext(r.Context())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(identity.UID + "@" + SpaceIDFromContext(r.Context())))
}

func TestAuthDisabledUsesDevelopmentContext(t *testing.T) {
	authenticator := NewAuthenticator(false, nil, model.Identity{UID: "dev"}, "dev-space")
	recorder := httptest.NewRecorder()
	authenticator.Wrap(http.HandlerFunc(successHandler)).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "dev@dev-space" {
		t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}

func TestAuthEnabled(t *testing.T) {
	tests := []struct {
		name     string
		resolver stubResolver
		token    string
		spaceID  string
		want     int
	}{
		{name: "missing token", want: http.StatusUnauthorized},
		{name: "resolver unavailable", token: "t", spaceID: "s1", resolver: stubResolver{err: errors.New("down")}, want: http.StatusServiceUnavailable},
		{name: "invalid token", token: "t", spaceID: "s1", want: http.StatusUnauthorized},
		{name: "old server response", token: "t", spaceID: "s1", resolver: stubResolver{identity: model.Identity{UID: "u1"}}, want: http.StatusServiceUnavailable},
		{name: "space required", token: "t", resolver: stubResolver{identity: model.Identity{UID: "u1", ContextIncluded: true}}, want: http.StatusBadRequest},
		{name: "space forbidden", token: "t", spaceID: "s2", resolver: stubResolver{identity: model.Identity{UID: "u1", ContextIncluded: true, Spaces: []string{"s1"}}}, want: http.StatusForbidden},
		{name: "allowed", token: "t", spaceID: "s1", resolver: stubResolver{identity: model.Identity{UID: "u1", ContextIncluded: true, Spaces: []string{"s1"}}}, want: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := NewAuthenticator(true, tt.resolver, model.Identity{}, "")
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Token", tt.token)
			}
			if tt.spaceID != "" {
				req.Header.Set("X-Space-Id", tt.spaceID)
			}
			recorder := httptest.NewRecorder()
			authenticator.Wrap(http.HandlerFunc(successHandler)).ServeHTTP(recorder, req)
			if recorder.Code != tt.want {
				t.Fatalf("status=%d want=%d body=%s", recorder.Code, tt.want, recorder.Body.String())
			}
		})
	}
}

func TestOwnsBot(t *testing.T) {
	identity := model.Identity{UID: "u1", OwnedBotsBySpace: map[string][]string{"s1": {"bot-1"}}}
	ctx := withAuthContext(context.Background(), identity, "s1")
	if !OwnsBot(ctx, "bot-1") || OwnsBot(ctx, "bot-2") {
		t.Fatal("unexpected bot ownership result")
	}
}
