package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mirac61/lifelog/internal/config"
	"github.com/Mirac61/lifelog/internal/github"
	"github.com/Mirac61/lifelog/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	syncMode := flag.Bool("sync", false, "this is a bool argument")
	normalizeMode := flag.Bool("normalize", false, "rebuild events from stored payloads")
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

	if *syncMode {
		client := github.New(cfg.GitHubToken)
		item, err := client.FetchContributions(ctx, time.Now().Year())
		if err != nil {
			slog.Error("fetch contributions", "error", err)
			os.Exit(1)
		}
		if err := store.InsertRaw(ctx, db, "github", item); err != nil {
			slog.Error("insert raw payload", "error", err)
			os.Exit(1)
		}
		slog.Info("sync complete", "external_id", item.ExternalID, "bytes", len(item.Payload))
		return
	}

	if *normalizeMode {
		list, err := store.ListRaw(ctx, db, "github")
		if err != nil {
			slog.Error("fetching Raw list", "error", err)
			os.Exit(1)
		}
		total := 0
		for _, row := range list {
			events, err := github.NormalizeContributions(row.Payload)
			if err != nil {
				slog.Error("fetching events", "error", err)
				os.Exit(1)
			}
			if err := store.InsertEvents(ctx, db, "github", row.ID, events); err != nil {
				slog.Error("insert events", "error", err)
				os.Exit(1)
			}
			total += len(events)
		}
		slog.Info("normalize complete", "events", total)
		return
	}

	slog.Info("lifelog started", "database", cfg.DatabasePath, "address", cfg.ListenAddress)
	<-ctx.Done()
	slog.Info("shutting down")
}
