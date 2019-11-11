package pktline

import "strings"

// Builder wraps the standard library string builder
// and is a way to build encoded messages that can
// be sent as sideband data
type Builder struct {
	*strings.Builder
}

// NewBuilder returns a new builder this doesn't require
// the use of Encode or EncodeString to init the builder
func NewBuilder() *Builder {
	return &Builder{new(strings.Builder)}
}

// Encode encodes the bytes b to the builder
func (d *Builder) Encode(b []byte) {
	d.EncodeString(string(b))
}

// EncodeString encodes the string s to the builder
func (d *Builder) EncodeString(s string) {
	if d.Builder == nil {
		d.Builder = new(strings.Builder)
	}
	encode(d.Builder, 65516, []byte(s))
}

// FlushString returns the fully built string
// with a trailing flush encoded.
func (d *Builder) FlushString() string {
	d.Flush()
	return d.String()
}

// Flush adds a non-encoded flush to the string
func (d *Builder) Flush() {
	if d.Builder == nil {
		d.Builder = new(strings.Builder)
	}
	d.Builder.WriteString("0000")
}
