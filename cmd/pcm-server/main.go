package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rsbh/personal-context-manager/internal/config"
	"github.com/rsbh/personal-context-manager/internal/gitsync"
	"github.com/rsbh/personal-context-manager/internal/search"
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

	idx, needRebuild, err := search.Open(filepath.Join(dir, "index"))
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Close()
	if needRebuild {
		// In the background, so the server starts at once. Search sees
		// the old (empty) index until the rebuild ends.
		go func() {
			n, err := server.RebuildIndex(st, idx)
			if err != nil {
				log.Printf("index rebuild: %v", err)
				return
			}
			log.Printf("index rebuilt: %d contexts", n)
		}()
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.NewRouter(st, server.WithGit(git), server.WithSearch(idx)),
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
