package web

import "net/http"

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	hits, err := s.Svc.Search(q)
	if err != nil {
		httpFromErr(w, err)
		return
	}
	if hits == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
		return
	}
	writeJSON(w, http.StatusOK, hits)
}
