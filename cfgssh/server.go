package cfgssh

import (
	"fmt"
	logg "log"
	"os"
	"runtime/debug"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.xa4b.com/git/pktline"
)

// Server holds the handlers for requests
// that the git client can make via SSH
type Server struct {
	git GitServer

	logPrefix string
	log       log
}

// NewServer returns a function that can wrap a SSH ssh.NewServer() function. It will
// then handle all requests for SSH
func NewServer(gs GitServer, opts ...ServerOption) func(*ssh.ServerConn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) error {
	return func(c *ssh.ServerConn, ch <-chan ssh.NewChannel, r <-chan *ssh.Request, err error) error {
		if err != nil {
			return err
		}

		s := &Server{git: gs}
		for _, optFn := range opts {
			optFn(s)
		}

		mux := NewMux()
		mux.HandlerFunc("git-receive-pack", HandlerFunc(s.ReceivePackHandler))
		mux.HandlerFunc("git-upload-pack", HandlerFunc(s.UploadPackHandler))

		s.ServeSSH(ch, r, mux)

		return nil
	}
}

// ServeSSH provides the internal handling for SSH channels accepted on a connection
func (s *Server) ServeSSH(chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, mux *Mux) {
	// handle panics so the whole thing doesn't crash if there is one.
	defer func() {
		if rvr := recover(); rvr != nil {
			fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
			debug.PrintStack()
		}
	}()

	go ssh.DiscardRequests(reqs)

	for ch := range chans {
		if t := ch.ChannelType(); t != "session" {
			ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		conn, reqs, err := ch.Accept()
		if err != nil {
			s.log.Info(s.logPrefix, "session accept channel error: %v", err)
			return
		}
		defer conn.Close()

		// go func(in <-chan *ssh.Request) {
		for req := range reqs {
			if t := req.Type; t != "exec" {
				s.log.Debugf("unknown request type: %s", t)
				continue
			}

			data := struct{ Payload string }{}
			if err := ssh.Unmarshal(req.Payload, &data); err != nil {
				s.log.Info(s.logPrefix, "payload unmarshal error: %v", err)
				continue
			}

			cmd := strings.SplitN(data.Payload, " ", 2)
			if len(cmd) != 2 {
				s.log.Info(s.logPrefix, "invalid payload (looking for git-receive-pack or git-upload-pack): %q", data.Payload)
				return
			}

			repoName := strings.TrimLeft(strings.Trim(cmd[1], "'"), "/")

			if handler, ok := mux.Handlers[cmd[0]]; ok {
				handler(repoName, conn)

				return
			}

			if mux.NotFoundHandler != nil {
				mux.NotFoundHandler(repoName, conn)
			}

			s.log.Info(s.logPrefix, "'exec' handler [%s] for %s was not found\r\n", cmd[0], repoName)
			pktline.NewEncoder(conn).Flush()
			return
		}

		fmt.Fprint(conn, "request type 'exec' was not found\r\n")
		// }(reqs)
	}
}

// WithLogger takes in logger/s to display debug and info logs for the server object
func (s *Server) WithLogger(logger ...interface{}) {
	for _, l := range logger {
		switch v := l.(type) {
		case DebugLogger:
			s.log.debug = v
		case InfoLogger:
			s.log.info = v
		case *logg.Logger:
			s.log.debug, s.log.info = v, v
		default:
			logg.Printf("invalid logger %T passed", v)
		}
	}
}
