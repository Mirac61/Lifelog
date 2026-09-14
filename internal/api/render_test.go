package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mirac61/lifelog/web/templ/git"
)

func TestGitPageRendersOwnBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/git", nil)
	rec := httptest.NewRecorder()
	render(rec, req, git.Git(git.GitPage{Year: 2026}))

	body := rec.Body.String()
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, "git-jahr") {
		t.Errorf("git body missing:\n%s", body)
	}
	if strings.Contains(body, "personal data warehouse") {
		t.Errorf("dashboard body leaked into /git:\n%s", body)
	}
	if strings.Contains(body, "render failed") {
		t.Errorf("partial render written to client:\n%s", body)
	}
}
