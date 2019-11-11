package cfghttp

import (
	"net/http"

	"github.com/go-chi/chi"
)

// InfoRefsHandler handles HTTP requests for 'info-refs/'
func (s *Server) InfoRefsHandler(w http.ResponseWriter, r *http.Request) {
	refs := s.git.NewInfoRefs(chi.URLParam(r, "repoName"))
	refs.DoHTTP(w, r)

	if refs.Err() != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// ReceivePackHandler handles HTTP requests for 'receive-pack/'
func (s *Server) ReceivePackHandler(w http.ResponseWriter, r *http.Request) {
	pack := s.git.NewReceivePack(chi.URLParam(r, "repoName"))
	defer pack.Cleanup()

	pack.DoHTTP(w, r)

	if pack.Err() != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// UploadPackHandler handles HTTP requests for 'upload-pack/'
func (s *Server) UploadPackHandler(w http.ResponseWriter, r *http.Request) {
	pack := s.git.NewUploadPack(chi.URLParam(r, "repoName"))
	defer pack.Cleanup()

	pack.DoHTTP(w, r)

	if pack.Err() != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
