package cfgssh

import (
	"golang.org/x/crypto/ssh"
	"gopkg.xa4b.com/git/cfg"
)

// GitServer is the interface used to interact with the git server via SSH
type GitServer interface {
	NewReceivePack(repoName string) ReceivePacker
	NewUploadPack(repoName string) UploadPacker

	WithLogger(...interface{})
	WithPreReceiveHook(cfg.PreReceivePackHookFunc)
	WithPostReceiveHook(cfg.PostReceivePackHookFunc)
}

// ReceivePacker returns SSH requests for 'receive-pack'
type ReceivePacker interface {
	DoSSH(ssh.Channel) ReceivePacker
	Cleanup()
	Err() error
}

// UploadPacker returns HTTP requests for 'upload-pack'
type UploadPacker interface {
	DoSSH(ssh.Channel) UploadPacker
	Cleanup()
	Err() error
}
