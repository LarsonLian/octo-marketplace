package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/auth"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
	"github.com/gin-gonic/gin"
)

const (
	identityKey = "marketplace.identity"
	spaceKey    = "marketplace.space_id"
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

func (a *Authenticator) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.enabled {
			spaceID := strings.TrimSpace(c.GetHeader("X-Space-Id"))
			if spaceID == "" {
				spaceID = a.devSpaceID
			}
			setAuthContext(c, a.devIdentity, spaceID)
			c.Next()
			return
		}

		token := requestToken(c)
		if token == "" {
			abortError(c, http.StatusUnauthorized, "err.marketplace.authentication_required", "Authentication is required.")
			return
		}
		if a.resolver == nil {
			abortError(c, http.StatusServiceUnavailable, "err.marketplace.auth_unavailable", "Authentication service is unavailable.")
			return
		}
		identity, err := a.resolver.Resolve(c.Request.Context(), token)
		if err != nil {
			abortError(c, http.StatusServiceUnavailable, "err.marketplace.auth_unavailable", "Authentication service is unavailable.")
			return
		}
		if identity.UID == "" {
			abortError(c, http.StatusUnauthorized, "err.marketplace.invalid_token", "Invalid or expired token.")
			return
		}
		if !identity.ContextIncluded {
			abortError(c, http.StatusServiceUnavailable, "err.marketplace.auth_context_unavailable", "Authorization context is unavailable.")
			return
		}

		spaceID := strings.TrimSpace(c.GetHeader("X-Space-Id"))
		if spaceID == "" {
			abortError(c, http.StatusBadRequest, "err.marketplace.space_required", "X-Space-Id header is required.")
			return
		}
		if !contains(identity.Spaces, spaceID) {
			abortError(c, http.StatusForbidden, "err.marketplace.space_forbidden", "Access to this Space is forbidden.")
			return
		}

		setAuthContext(c, identity, spaceID)
		c.Next()
	}
}

func Identity(c *gin.Context) (model.Identity, bool) {
	value, ok := c.Get(identityKey)
	if !ok {
		return model.Identity{}, false
	}
	identity, ok := value.(model.Identity)
	return identity, ok
}

func SpaceID(c *gin.Context) string {
	value, _ := c.Get(spaceKey)
	spaceID, _ := value.(string)
	return spaceID
}

func OwnsBot(c *gin.Context, botID string) bool {
	identity, ok := Identity(c)
	if !ok {
		return false
	}
	return contains(identity.OwnedBotsBySpace[SpaceID(c)], botID)
}

func setAuthContext(c *gin.Context, identity model.Identity, spaceID string) {
	c.Set(identityKey, identity)
	c.Set(spaceKey, spaceID)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIdentityKey{}, identity))
}

type requestIdentityKey struct{}

func requestToken(c *gin.Context) string {
	if token := strings.TrimSpace(c.GetHeader("Token")); token != "" {
		return token
	}
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
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

func abortError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":        code,
			"message":     message,
			"http_status": status,
		},
	})
}
