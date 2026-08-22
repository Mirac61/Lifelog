package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Mirac61/lifelog/internal/github"
	"github.com/Mirac61/lifelog/internal/store"
)

func syncYear(ctx context.Context, db *sql.DB, client *github.Client, year int) error {
	item, err := client.FetchContributions(ctx, year)
	if err != nil {
		return fmt.Errorf("fetch contributions: %w", err)
	}
	if err := store.InsertRaw(ctx, db, "github", item); err != nil {
		return fmt.Errorf("insert raw payload: %w", err)
	}
	slog.Info("sync complete", "external_id", item.ExternalID, "bytes", len(item.Payload))

	repoItem, err := client.FetchRepoCommits(ctx, year)
	if err != nil {
		return fmt.Errorf("fetchrepo contributions: %w", err)
	}
	if err := store.InsertRaw(ctx, db, "github", repoItem); err != nil {
		return fmt.Errorf("insert repo payload: %w", err)
	}
	slog.Info("sync complete", "external_id", repoItem.ExternalID, "bytes", len(repoItem.Payload))

	prItem, err := client.FetchPRContributions(ctx, year)
	if err != nil {
		return fmt.Errorf("fetch pr contributions: %w", err)
	}
	if err := store.InsertRaw(ctx, db, "github", prItem); err != nil {
		return fmt.Errorf("insert pr payload: %w", err)
	}
	slog.Info("sync complete", "external_id", prItem.ExternalID, "bytes", len(prItem.Payload))
	return nil
}

func normalizeAll(ctx context.Context, db *sql.DB) error {
	list, err := store.ListRaw(ctx, db, "github")
	if err != nil {
		return fmt.Errorf("fetching Raw list: %w", err)
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
			return fmt.Errorf("fetching events: %w", err)
		}
		if err := store.InsertEvents(ctx, db, "github", row.ID, events); err != nil {
			return fmt.Errorf("insert events: %w", err)
		}
		total += len(events)
	}
	slog.Info("normalize complete", "events", total)
	return nil
}
