package api

import (
	"bytes"
	"database/sql"
	"html/template"
	"io/fs"
	"net/http"
	"path"

	"github.com/Mirac61/lifelog/web"
)

type Server struct {
	db       *sql.DB
	pages    map[string]*template.Template
	partials *template.Template
}

func New(db *sql.DB) (*Server, error) {
	partials, err := template.ParseFS(web.FS, "templates/partials/*.html")
	if err != nil {
		return nil, err
	}

	names, err := fs.Glob(web.FS, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	// One template set per page: every page defines "body", so a shared set
	// would let them silently overwrite each other.
	pages := make(map[string]*template.Template, len(names))
	for _, name := range names {
		t, err := template.ParseFS(web.FS, "templates/layout.html", "templates/partials/*.html", name)
		if err != nil {
			return nil, err
		}
		pages[path.Base(name)] = t
	}

	return &Server{db: db, pages: pages, partials: partials}, nil
}

func (s *Server) Routes() http.Handler {
	static, _ := fs.Sub(web.FS, "static")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /heatmap", s.handleHeatmap)
	mux.HandleFunc("GET /day", s.handleDay)
	mux.HandleFunc("GET /git/view", s.handleGitView)
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /git", s.handleGit)
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

func (s *Server) renderPage(w http.ResponseWriter, page string, data any) {
	t := s.pages[page]
	if t == nil {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		return
	}
	render(w, t, "layout.html", data)
}

func (s *Server) renderPartial(w http.ResponseWriter, name string, data any) {
	render(w, s.partials, name, data)
}

func render(w http.ResponseWriter, t *template.Template, name string, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
