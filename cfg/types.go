package cfg /* import "gopkg.xa4b.com/git/cfg" */

import (
	"io"
)

// ReceivePackData holds the receive-pack data that's sent to the 'pre' and 'post' receive-pack hooks
type ReceivePackData struct{ OldHash, NewHash, RefName string }

// ReceivePackHookData holds all of the data needed to interact with the receive-pack hooks. This data cannot be changed, it is read-only.
type ReceivePackHookData struct {
	RepoName string
	Refs     []ReceivePackData
}

// PreReceivePackHookData is a wrapper around the ReceivePackHookData for the pre-receive-pack hook.
type PreReceivePackHookData struct {
	ReceivePackHookData
}

// PostReceivePackHookData is a wrapper around the ReceivePackHookData for the post-receive-pack hook. It also includes any push-options that may have been included.
type PostReceivePackHookData struct {
	ReceivePackHookData
	PushOptions map[string]string
}

// ReceivePackHookError represents 60 characters of a string that will be displayed as rejected text to git, use the RejectText() method to manipulate this string.
type ReceivePackHookError [60]rune

// String returns the string version of the error
func (e *ReceivePackHookError) String() string { return string(e[:]) }

// Error satisfies the error interface
func (e *ReceivePackHookError) Error() string { return e.String() }

// RejectText accepts a string that will be written to the git output when rejected. Only the first 60 characters of a string will be used.
func (e *ReceivePackHookError) RejectText(s string) { copy(e[:], []rune(s)) }

// PreReceivePackHookFunc a type describing the pre-receive-pack hook callback
type PreReceivePackHookFunc func(io.Writer, *PreReceivePackHookData) (string, *ReceivePackHookError)

// PostReceivePackHookFunc a type describing the post-receive-pack hook callback
type PostReceivePackHookFunc func(io.Writer, *PostReceivePackHookData)
