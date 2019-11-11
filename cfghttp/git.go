package cfghttp

import (
	"net/http"

	"gopkg.xa4b.com/git/cfg"
)

// GitServer is the interface used to interact with the git server via HTTP
type GitServer interface {
	NewInfoRefs(repoName string) InfoRefser
	NewReceivePack(repoName string) ReceivePacker
	NewUploadPack(repoName string) UploadPacker

	WithLogger(...interface{})
	WithPreReceiveHook(cfg.PreReceivePackHookFunc)
	WithPostReceiveHook(cfg.PostReceivePackHookFunc)
}

// InfoRefser returns HTTP requests for '/info/ref'
type InfoRefser interface {
	DoHTTP(http.ResponseWriter, *http.Request) InfoRefser
	Err() error
}

// ReceivePacker returns HTTP requests for '/receive-pack'
type ReceivePacker interface {
	DoHTTP(http.ResponseWriter, *http.Request) ReceivePacker
	Cleanup()
	Err() error
}

// UploadPacker returns HTTP requests for '/upload-pack'
type UploadPacker interface {
	DoHTTP(http.ResponseWriter, *http.Request) UploadPacker
	Cleanup()
	Err() error
}
