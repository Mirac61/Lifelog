package api

import "net/http"

func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Active string
	}{"git"}

	s.renderPage(w, "git.html", data)
}
