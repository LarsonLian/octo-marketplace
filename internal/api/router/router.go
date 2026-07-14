package router

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/http"

	categoryhandler "github.com/Mininglamp-OSS/octo-marketplace/internal/api/handler/category"
	skillhandler "github.com/Mininglamp-OSS/octo-marketplace/internal/api/handler/skill"
	marketmiddleware "github.com/Mininglamp-OSS/octo-marketplace/internal/middleware"
	categoryrepo "github.com/Mininglamp-OSS/octo-marketplace/internal/repository/category"
	skillrepo "github.com/Mininglamp-OSS/octo-marketplace/internal/repository/skill"
	categorysvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/category"
	skillsvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/skill"
	"github.com/gin-gonic/gin"
)

type Pinger interface {
	PingContext(context.Context) error
}

func Public(database Pinger, authenticator *marketmiddleware.Authenticator) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if err := database.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(authenticator.Handler())
	v1.GET("/session", func(c *gin.Context) {
		identity, ok := marketmiddleware.Identity(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"uid":      identity.UID,
			"name":     identity.Name,
			"space_id": marketmiddleware.SpaceID(c),
		})
	})

	// Wire up skill marketplace handlers
	db, ok := database.(*sql.DB)
	if ok {
		catRepo := categoryrepo.New(db)
		skRepo := skillrepo.New(db)

		catSvc := categorysvc.New(catRepo)
		skSvc := skillsvc.New(skRepo, catRepo, generateID)

		categoryhandler.New(catSvc).Register(v1)
		skillhandler.New(skSvc).Register(v1)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,Token,X-Space-Id,X-Request-Id")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// generateID produces a UUID v4 string.
func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

