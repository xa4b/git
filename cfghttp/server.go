package cfghttp /* import "gopkg.xa4b.com/git/cfghttp" */

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// Server holds the handlers, middleware and mux for
// requests that the git client can make via HTTP
type Server struct {
	mux *chi.Mux

	pathPrefix string
	git        GitServer

	middlewares []func(http.Handler) http.Handler
}

// NewServer returns a new server object that can be used as a mux with the http.Handler
func NewServer(gs GitServer, opts ...ServerOption) http.Handler {
	s := &Server{git: gs, mux: chi.NewRouter()}
	for _, optFn := range opts {
		optFn(s) // the GitServer object needs to be added before this... so some options can interact with it
	}

	s.mux.Use(s.middlewares...)
	s.mux.Get(fmt.Sprintf("/info/refs"), s.InfoRefsHandler)
	s.mux.Get(fmt.Sprintf("/{repoName}/info/refs"), s.InfoRefsHandler)

	s.mux.Post(fmt.Sprintf("/git-receive-pack"), s.ReceivePackHandler)
	s.mux.Post(fmt.Sprintf("/{repoName}/git-receive-pack"), s.ReceivePackHandler)

	s.mux.Post(fmt.Sprintf("/git-upload-pack"), s.UploadPackHandler)
	s.mux.Post(fmt.Sprintf("/{repoName}/git-upload-pack"), s.UploadPackHandler)

	return s
}

// ServeHTTP serves the internal muxer for the http handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
