package api

import (
	"bytes"
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"

	"github.com/Mirac61/lifelog/web"
)

type Server struct {
	db *sql.DB
}

func New(db *sql.DB) (*Server, error) {
	return &Server{db: db}, nil
}

func (s *Server) Routes() http.Handler {
	static, _ := fs.Sub(web.FS, "static")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /day", s.handleDay)
	mux.HandleFunc("GET /git/view", s.handleGitView)
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /git", s.handleGit)
	mux.HandleFunc("GET /todos", s.handleTodos)
	mux.HandleFunc("POST /todos", s.handleCreateTodo)
	mux.HandleFunc("POST /todos/{id}/{status}", s.handleTodo)
	mux.HandleFunc("DELETE /todos/{id}", s.handleDeleteTodo)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte("ok"))
}

func render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	var buf bytes.Buffer
	if err := c.Render(r.Context(), &buf); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
