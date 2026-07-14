package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/api/router"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/config"
	marketdb "github.com/Mininglamp-OSS/octo-marketplace/internal/db"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateAPI(); err != nil {
		log.Fatal(err)
	}
	database, err := marketdb.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if n, err := marketdb.RunMigrations(database); err != nil {
		log.Fatalf("[main] migration failed: %v", err)
	} else if n > 0 {
		log.Printf("[main] applied %d migration(s)", n)
	}

	publicServer := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router.Public(database),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	go serve("public", publicServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = publicServer.Shutdown(ctx)
}

func serve(name string, server *http.Server) {
	log.Printf("[%s] listening on %s", name, server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[%s] %v", name, err)
	}
}
