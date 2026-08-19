package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mirac61/lifelog/internal/config"
	"github.com/Mirac61/lifelog/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		slog.Error("open store", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("lifelog started", "database", cfg.DatabasePath, "address", cfg.ListenAddress)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	slog.Info("shutting down")
}
