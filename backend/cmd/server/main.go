// Command server starts the Go backend for the AI Payment Support System.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/api"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/auth"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/config"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/service"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Stop cleanly on Ctrl+C or a SIGTERM from the hosting platform.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var st store.Store
	if cfg.DatabaseURL == "" {
		log.Warn("DATABASE_URL not set: using in-memory store (data is lost on restart)")
		st = store.NewMemory()
	} else {
		db, err := store.OpenMySQL(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		st = db
	}
	defer st.Close()

	svc := service.New(st, auth.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL))
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		if err := svc.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
			return err
		}
		log.Info("admin account ready", "email", cfg.AdminEmail)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(svc, cfg.AllowedOrigins, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}
