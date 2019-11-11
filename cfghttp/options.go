package cfghttp

import (
	"io"
	"net/http"

	"gopkg.xa4b.com/git/cfg"
)

// ServerOption provides functional options for Server objects in cfghttp
type ServerOption func(*Server)

// WithPathPrefix adds a http path prefix before any of the http path patterns
func WithPathPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.pathPrefix = prefix
	}
}

// WithMiddleware adds any middleware to the HTTP server (i.e. can be used for auth)
func WithMiddleware(wares ...func(http.Handler) http.Handler) ServerOption {
	return func(s *Server) {
		s.middlewares = wares
	}
}

// WithLogger adds a logger to the library. If *log.Logger is used then
// both debug and info logs will be displayed. Wrap a *log.Logger in
// DebugLogger or InfoLogger to display just the logs for one level.
func WithLogger(logger ...interface{}) ServerOption {
	return func(s *Server) {
		s.git.WithLogger(logger...)
	}
}

// WithPreReceiveHook adds a pre-receive hook to the handler.
// Return a nil error to indicate pre-receive hook success.
// Return a non-nil error to reject the pre-receive hook.
// (Note the limitation of pre-receive hooks at: https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks#_code_pre_receive_code)
func WithPreReceiveHook(fn func(io.Writer, *cfg.PreReceivePackHookData) (string, *cfg.ReceivePackHookError)) ServerOption {
	return func(s *Server) {
		s.git.WithPreReceiveHook(fn)
	}
}

// WithPostReceiveHook adds a post-receive hook to the handler.
// Return a nil error to indicate post-receive hook success.
// Return a non-nil error to reject the post-receive hook.
// (Note the limitation of pre-receive hooks at: https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks#_code_post_receive_code)
func WithPostReceiveHook(fn func(io.Writer, *cfg.PostReceivePackHookData)) ServerOption {
	return func(s *Server) {
		s.git.WithPostReceiveHook(fn)
	}
}
