package router

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/http"

	categoryhandler "github.com/Mininglamp-OSS/octo-marketplace/internal/api/handler/category"
	skillhandler "github.com/Mininglamp-OSS/octo-marketplace/internal/api/handler/skill"
	uploadhandler "github.com/Mininglamp-OSS/octo-marketplace/internal/api/handler/upload"
	marketmiddleware "github.com/Mininglamp-OSS/octo-marketplace/internal/middleware"
	categoryrepo "github.com/Mininglamp-OSS/octo-marketplace/internal/repository/category"
	skillrepo "github.com/Mininglamp-OSS/octo-marketplace/internal/repository/skill"
	categorysvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/category"
	parsesvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/parse"
	skillsvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/skill"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/storage"
	"github.com/gin-gonic/gin"
)

type Pinger interface {
	PingContext(context.Context) error
}

// StorageConfig holds configuration for the storage layer.
type StorageConfig struct {
	Driver            string // "local" or "oss"
	LocalDir          string
	BaseURL           string
	MaxMB             int
	OSSEndpoint       string
	OSSBucket         string
	OSSAccessKey      string
	OSSSecretKey      string
	OSSRegion         string
	OSSPublicEndpoint string // external endpoint for presigned URLs (e.g. http://localhost:29000)
}

func Public(database Pinger, authenticator *marketmiddleware.Authenticator, storageCfg StorageConfig) *gin.Engine {
	return publicWithOptions(database, authenticator, storageCfg, authenticator.AuthEnabled())
}

func publicWithOptions(database Pinger, authenticator *marketmiddleware.Authenticator, storageCfg StorageConfig, authEnabled bool) *gin.Engine {
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

		// Initialize storage first (needed by skill service)
		var store storage.Storage
		var localStorage *storage.LocalStorage
		switch storageCfg.Driver {
		case "local":
			ls := storage.NewLocal(storageCfg.LocalDir, storageCfg.BaseURL)
			store = ls
			localStorage = ls
		case "oss":
			oss, err := storage.NewOSS(storage.OSSConfig{
				Endpoint:       storageCfg.OSSEndpoint,
				Bucket:         storageCfg.OSSBucket,
				AccessKey:      storageCfg.OSSAccessKey,
				SecretKey:      storageCfg.OSSSecretKey,
				Region:         storageCfg.OSSRegion,
				PublicEndpoint: storageCfg.OSSPublicEndpoint,
			})
			if err != nil {
				panic("storage driver oss: " + err.Error())
			}
			store = oss
		default:
			panic("unsupported STORAGE_DRIVER: " + storageCfg.Driver)
		}

		catSvc := categorysvc.New(catRepo)
		skSvc := skillsvc.New(skRepo, catRepo, store, generateID)

		catH := categoryhandler.New(catSvc)
		catH.Register(v1)
		catH.RegisterAdmin(v1, generateID, authEnabled)
		skillhandler.New(skSvc).Register(v1)

		// Wire up upload/parse/download handlers

		parseRepo := parsesvc.NewRepo(db)
		worker := parsesvc.NewWorker(store, parseRepo, db)
		pSvc := parsesvc.NewService(store, parseRepo, worker, generateID, storageCfg.MaxMB)

		uploadH := uploadhandler.New(pSvc, skSvc, localStorage)
		uploadH.Register(v1)
		uploadH.RegisterLocalProxy(r)
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

// PublicWithDB is a convenience wrapper that creates the engine with a *sql.DB
// directly (useful for integration tests with sqlmock).
func PublicWithDB(db *sql.DB, authenticator *marketmiddleware.Authenticator, storageCfg StorageConfig) *gin.Engine {
	return Public(db, authenticator, storageCfg)
}

// PublicWithDBAndAuth is a test helper that overrides the authEnabled flag for admin routes.
// This allows testing admin authorization independently of the authenticator middleware mode.
func PublicWithDBAndAuth(db *sql.DB, authenticator *marketmiddleware.Authenticator, storageCfg StorageConfig, authEnabled bool) *gin.Engine {
	return publicWithOptions(db, authenticator, storageCfg, authEnabled)
}
