package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	syncYear := flag.Int("year", time.Now().Year(), "year to sync")
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
		item, err := client.FetchContributions(ctx, *syncYear)
		if err != nil {
			slog.Error("fetch contributions", "error", err)
			os.Exit(1)
		}
		if err := store.InsertRaw(ctx, db, "github", item); err != nil {
			slog.Error("insert raw payload", "error", err)
			os.Exit(1)
		}
		slog.Info("sync complete", "external_id", item.ExternalID, "bytes", len(item.Payload))

		repoItem, err := client.FetchRepoCommits(ctx, *syncYear)
		if err != nil {
			slog.Error("fetch repo contributions", "error", err)
			os.Exit(1)
		}
		if err := store.InsertRaw(ctx, db, "github", repoItem); err != nil {
			slog.Error("insert repo payload", "error", err)
			os.Exit(1)
		}
		slog.Info("sync complete", "external_id", repoItem.ExternalID, "bytes", len(repoItem.Payload))
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
			normalize := github.NormalizeContributions
			switch {
			case strings.HasPrefix(row.ExternalID, "repo-commits:"):
				normalize = github.NormalizeRepoContributions
			case strings.HasPrefix(row.ExternalID, "pull-requests:"):
				normalize = github.NormalizePRContributions
			case strings.HasPrefix(row.ExternalID, "contributions:"):
			default:
				slog.Warn("skip unknown raw payload", "external_id", row.ExternalID)
				continue
			}

			events, err := normalize(row.Payload)
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

	slog.Info("lifelog started", "database", cfg.DatabasePath, "address", cfg.ListenAddress)
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}
	slog.Info("shutting down")
}
