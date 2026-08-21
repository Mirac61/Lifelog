package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)


func TestGitPageRendersOwnBody(t *testing.T) {
	s, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, httptest.NewRequest("GET", "/git", nil))

	body := rec.Body.String()
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, "Coming soon") {
		t.Errorf("git body missing:\n%s", body)
	}
	if strings.Contains(body, "personal data warehouse") {
		t.Errorf("dashboard body leaked into /git:\n%s", body)
	}
	if strings.Contains(body, "render failed") {
		t.Errorf("partial render written to client:\n%s", body)
	}
}
