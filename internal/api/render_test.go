package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mirac61/lifelog/web/templ/git"
	"github.com/Mirac61/lifelog/web/templ/todo"
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

// Pins the preview to the parse the create handler runs.
func TestCapturePreviewMirrorsParse(t *testing.T) {
	const input = "Wäsche aufhängen 15m #haushalt @09.09.2026"
	preview := todo.NewPreview(parseTodoInput(input), "2026-09-07", "2026-09-07")

	req := httptest.NewRequest("POST", "/todos/parse", nil)
	rec := httptest.NewRecorder()
	render(rec, req, todo.CapturePreview(preview))

	body := rec.Body.String()
	for _, want := range []string{"Wäsche aufhängen", "15", "min", "#haushalt", "Mi 9.9.", "Kommende"} {
		if !strings.Contains(body, want) {
			t.Errorf("preview missing %q:\n%s", want, body)
		}
	}
	// Die Tokens duerfen nicht im Titel stehenbleiben.
	if strings.Contains(body, "15m") || strings.Contains(body, "@09.09.2026") {
		t.Errorf("raw tokens leaked into the title:\n%s", body)
	}
	if strings.Contains(body, "Erkannt wird") {
		t.Errorf("hint shown although something was typed:\n%s", body)
	}
}

func TestCapturePreviewEmptyExplainsSyntax(t *testing.T) {
	req := httptest.NewRequest("POST", "/todos/parse", nil)
	rec := httptest.NewRecorder()
	render(rec, req, todo.CapturePreview(todo.Preview{Empty: true}))

	if body := rec.Body.String(); !strings.Contains(body, "Erkannt wird") {
		t.Errorf("empty capture should name its syntax:\n%s", body)
	}
}
