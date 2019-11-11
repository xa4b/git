package cfgssh

import (
	"golang.org/x/crypto/ssh"
)

// ReceivePackHandler handles SSH calls to 'receive-pack'
func (s *Server) ReceivePackHandler(repoName string, rw ssh.Channel) {
	pack := s.git.NewReceivePack(repoName)
	defer pack.Cleanup()

	pack.DoSSH(rw)

	if pack.Err() != nil {
		return // no ExitCode returned
	}

	ExitCode(rw, 0)
}

// UploadPackHandler handles SSH calls to 'upload-pack'
func (s *Server) UploadPackHandler(repoName string, rw ssh.Channel) {
	pack := s.git.NewUploadPack(repoName)
	defer pack.Cleanup()

	pack.DoSSH(rw)

	if pack.Err() != nil {
		return // no ExitCode returned
	}

	ExitCode(rw, 0)
}

// ExitCode sends a specific exit status back to the SSH channel. If no code is
// sent then the channel closes with a error code of -1
func ExitCode(rw ssh.Channel, code uint32) {
	rw.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{uint32(0)}))
}
