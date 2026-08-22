package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mirac61/lifelog/internal/api"
	"github.com/Mirac61/lifelog/internal/config"
	"github.com/Mirac61/lifelog/internal/github"
	"github.com/Mirac61/lifelog/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	syncMode := flag.Bool("sync", false, "this is a bool argument")
	normalizeMode := flag.Bool("normalize", false, "rebuild events from stored payloads")
	year := flag.Int("year", time.Now().Year(), "year to sync")
	backfill := flag.Bool("backfill", false, "sync every year GitHub knows")
	flag.Parse()

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
	if err := store.Migrate(db); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := github.New(cfg.GitHubToken)

	if *syncMode {
		if err := syncYear(ctx, db, client, *year); err != nil {
			slog.Error("sync", "error", err)
			os.Exit(1)
		}
		return
	}

	if *normalizeMode {
		if err := normalizeAll(ctx, db); err != nil {
			slog.Error("normalize", "error", err)
			os.Exit(1)
		}
		return
	}

	if *backfill {
		years, err := client.FetchYears(ctx)
		if err != nil {
			slog.Error("backfill", "error", err)
			os.Exit(1)
		}
		for _, y := range years {
			if err := syncYear(ctx, db, client, y); err != nil {
				slog.Error("sync backfill", "error", err)
				continue
			}
		}
		if err := normalizeAll(ctx, db); err != nil {
			slog.Error("normalize backfill", "error", err)
			os.Exit(1)
		}
		return
	}

	apiServer, err := api.New(db)
	if err != nil {
		slog.Error("init api", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: apiServer.Routes(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
		}
	}()

	go func() {
		if err := syncYear(ctx, db, client, time.Now().Year()); err != nil {
			slog.Error("startup sync", "error", err)
			return
		}
		if err := normalizeAll(ctx, db); err != nil {
			slog.Error("startup normalize", "error", err)
		}
	}()

	slog.Info("lifelog started", "database", cfg.DatabasePath, "address", cfg.ListenAddress)
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}
	slog.Info("shutting down")
}
