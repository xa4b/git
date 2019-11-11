package cfgssh

import (
	"io"

	"gopkg.xa4b.com/git/cfg"
)

// ServerOption is the optional function type used for cfghttp
type ServerOption func(*Server)

// WithLogger adds a logger to the library. If *log.Logger is used then
// both debug and info logs will be populated. wrap a *log.Logger in
// DebugLogger or InfoLogger to just view the logs for one level
func WithLogger(l ...interface{}) ServerOption {
	return func(s *Server) {
		s.git.WithLogger(l...)
	}
}

// WithPreReceiveHook adds a pre-recieve hook to the handler, sending nil to the handler will sucessfully execute the git commad
// send a non nil error to reject the recieve. All writes the the writer will be
// done with newlines, otherwise it may be cut off.
func WithPreReceiveHook(fn func(io.Writer, *cfg.PreReceivePackHookData) (string, *cfg.ReceivePackHookError)) ServerOption {
	return func(s *Server) {
		s.git.WithPreReceiveHook(fn)
	}
}

// WithPostReceiveHook adds a post-recieve hook to the handler, sending nil to the handler will sucessfully execute the git commad
// send a non nil error to reject the recieve. All writes the the writer will be
// done with newlines, otherwise it may be cut off.
func WithPostReceiveHook(fn func(io.Writer, *cfg.PostReceivePackHookData)) ServerOption {
	return func(s *Server) {
		s.git.WithPostReceiveHook(fn)
	}
}
