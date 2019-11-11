package cfgssh

import "golang.org/x/crypto/ssh"

// HandlerFunc is an adapter to allow the use of ordinary functions as SSH handlers
type HandlerFunc func(string, ssh.Channel)

// Mux holds the handlers and a command associated with it.
type Mux struct {
	Handlers map[string]HandlerFunc

	NotFoundHandler HandlerFunc
}

// NewMux returns a new initialized Mux object
func NewMux() *Mux {
	return &Mux{Handlers: make(map[string]HandlerFunc)}
}

// HandlerFunc adds a handler with a command to the Mux
func (m *Mux) HandlerFunc(command string, fn HandlerFunc) { m.Handlers[command] = fn }
