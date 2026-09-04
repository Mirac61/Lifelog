package github

import "testing"

const repoContributionsPayload = `{
  "viewer": {
    "contributionsCollection": {
      "startedAt": "2025-01-01T00:00:00Z",
      "endedAt": "2025-12-31T23:59:59Z",
      "commitContributionsByRepository": [
        {
          "repository": {
            "name": "Mirac61",
            "nameWithOwner": "Mirac61/Mirac61",
            "pushedAt": "2026-08-02T10:00:00Z"
          },
          "contributions": {"totalCount": 2}
        }
      ]
    }
  }
}`

func TestNormalizeRepoContributionsDatesByWindow(t *testing.T) {
	events, err := NormalizeRepoContributions([]byte(repoContributionsPayload))
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if got, want := events[0].LocalDate, "2025-01-01"; got != want {
		t.Errorf("LocalDate = %q, want %q (pushedAt darf das Datum nicht setzen)", got, want)
	}
	if got, want := events[0].Value, 2.0; got != want {
		t.Errorf("Value = %v, want %v", got, want)
	}
}

func TestNormalizeRepoContributionsRejectsMissingWindow(t *testing.T) {
	_, err := NormalizeRepoContributions([]byte(`{"viewer":{"contributionsCollection":{}}}`))
	if err == nil {
		t.Fatal("want error für Payload ohne startedAt, got nil")
	}
}
