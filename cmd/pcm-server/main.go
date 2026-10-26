// Package main runs the pcm HTTP server.
//
//	@title			pcm API
//	@version		1.0
//	@description	Personal context manager: store and resolve coding contexts.
//	@BasePath		/
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/rsbh/personal-context-manager/docs/swagger"
	"github.com/rsbh/personal-context-manager/internal/config"
	"github.com/rsbh/personal-context-manager/internal/gitsync"
	"github.com/rsbh/personal-context-manager/internal/server"
	"github.com/rsbh/personal-context-manager/internal/store"
)

const shutdownTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dir, err := config.DefaultDir()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		log.Fatal(err)
	}
	st, err := store.New(cfg.Path)
	if err != nil {
		log.Fatal(err)
	}

	// Read remote on each git call, so pcm config set remote needs no restart.
	remote := func() (string, error) {
		c, err := config.Load(dir)
		return c.Remote, err
	}
	git := gitsync.New(cfg.Path, remote)

	opts := []server.Option{server.WithGit(git)}
	if cfg.Swagger {
		opts = append(opts, server.WithSwagger())
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.NewRouter(st, opts...),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s, repo %s", addr, cfg.Path)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
