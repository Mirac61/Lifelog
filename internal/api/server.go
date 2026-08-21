package api

import (
	"database/sql"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Mirac61/lifelog/web"
)

type Server struct {
	db        *sql.DB
	templates *template.Template
}

func New(db *sql.DB) (*Server, error) {
	tmpl, err := template.ParseFS(web.FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{db: db, templates: tmpl}, nil
}

func (s *Server) Routes() http.Handler {
	static, _ := fs.Sub(web.FS, "static")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /heatmap", s.handleHeatmap)
	mux.HandleFunc("GET /day", s.handleDay)
	mux.HandleFunc("GET /", s.handleIndex)
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
