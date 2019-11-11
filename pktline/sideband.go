package pktline

import (
	"bytes"
	"fmt"
	"io"
)

type SidebandChannel byte
type SidebandCapability int

const (
	SidebandPackfile SidebandChannel = 0x1
	SidebandProgress SidebandChannel = 0x2
	SidebandError    SidebandChannel = 0x3
)

const (
	SidebandNone SidebandCapability = 0
	Sideband     SidebandCapability = 1
	Sideband64k  SidebandCapability = 64
)

type EncodeSideband interface {
	Encode(SidebandChannel, []byte) error
	EncodeString(SidebandChannel, string) error
	EncodePackfile([]byte) error
	EncodeProgress(string) error
	EncodeError(string) error
	Flush() error

	Write([]byte) (int, error)
	WriteString(string) (int, error)
}

type ScanSideband interface {
	Scan() bool
	Bytes() []byte
	Text() string
	Err() error

	Packfile() []byte
	ProgressText() string
	ErrorText() string
}

type noSidebandEncoder struct{}

func (e *noSidebandEncoder) Write(_ []byte) (n int, err error)          { return n, e.Flush() }
func (e *noSidebandEncoder) WriteString(_ string) (n int, err error)    { return n, e.Flush() }
func (e *noSidebandEncoder) Encode(_ SidebandChannel, _ []byte) error   { return e.Flush() }
func (e *noSidebandEncoder) EncodeString(SidebandChannel, string) error { return e.Flush() }
func (e *noSidebandEncoder) EncodePackfile(_ []byte) error              { return e.Flush() }
func (e *noSidebandEncoder) EncodeProgress(_ string) error              { return e.Flush() }
func (e *noSidebandEncoder) EncodeError(_ string) error                 { return e.Flush() }
func (e *noSidebandEncoder) Flush() error                               { return ErrSidebandNotImplemented }

type SidebandEncoder struct {
	w io.Writer

	skipWrite bool
	maxLnLen  int

	buf func() (*bytes.Buffer, io.Writer)
}

func newSidebandEncoder(w io.Writer, lnLen int) *SidebandEncoder {
	return &SidebandEncoder{w: w, maxLnLen: lnLen}
}

func (enc *SidebandEncoder) setSkip(b bool) {
	enc.skipWrite = b
}

func (enc *SidebandEncoder) skip() bool {
	return enc.skipWrite
}

// Write is a pass-through method. All bytes should be previously encoded
func (enc *SidebandEncoder) Write(d []byte) (n int, err error) {
	return enc.w.Write(d)
}

// WriteString is a pass-through convenience method for a string to be passed to the underlining writer. No encoding is happening
func (enc *SidebandEncoder) WriteString(s string) (n int, err error) {
	return enc.Write([]byte(s))
}

// Encode takes a byte slice to encode for the following:
// If 'side-band' or 'side-band-64k' capabilities have been specified by
// the client, the server will send the packfile data multiplexed.
//
// Each packet starting with the packet-line length of the amount of data
// that follows, followed by a single byte specifying the sideband the
// following data is coming in on.
func (enc *SidebandEncoder) Encode(c SidebandChannel, b []byte) (err error) {
	if c < 1 || c > 3 {
		return ErrSidebandChannelInvalid
	}
	if enc.skipWrite {
		return nil // pretend to write...
	}

	return encode(enc.w, enc.maxLnLen, append([]byte{byte(c)}, append([]byte(fmt.Sprintf("%04x", 4+len(b))), b...)...))
}

func (enc *SidebandEncoder) Data(c SidebandChannel, b []byte) (err error) {
	if c < 1 || c > 3 {
		return ErrSidebandChannelInvalid
	}
	if enc.skipWrite {
		return nil // pretend to write...
	}

	return encode(enc.w, enc.maxLnLen, append([]byte{byte(c)}, b...))
}

// EncodeString takes a string to encode for the following:
// If 'side-band' or 'side-band-64k' capabilities have been specified by
// the client, the server will send the packfile data multiplexed.
//
// Each packet starting with the packet-line length of the amount of data
// that follows, followed by a single byte specifying the sideband the
// following data is coming in on.
func (enc *SidebandEncoder) EncodeString(c SidebandChannel, s string) error {
	return enc.Encode(c, []byte(s))
}

// EncodePackfile is a convenience method for Encode(Packfile, b)
func (enc *SidebandEncoder) EncodePackfile(b []byte) error {
	return enc.Encode(SidebandPackfile, b)
}

// EncodeProgress is a convenience method for EncodeString(Progress, s)
func (enc *SidebandEncoder) EncodeProgress(s string) error {
	return enc.Data(SidebandProgress, []byte(s))
}

// EncodeError is a convenience method for EncodeString(Error, s)
func (enc *SidebandEncoder) EncodeError(s string) error {
	return enc.Data(SidebandError, []byte(s))
}

// Flush sends 0000 to the underlining writer
func (enc *SidebandEncoder) Flush() error {
	if enc.skipWrite {
		return flush(enc.w)
	}

	// otherwise we are in sideband mode so we flush with encoded packfile mode
	_, err := enc.w.Write([]byte("0009\x010000")) // this is static...
	return err
}

type SidebandScanner struct {
	r io.Reader
	b []byte
	w io.Writer

	progressText string
	errorText    string

	canScan bool
	scanErr error
}

func newSidebandScanner(w io.Writer, r io.Reader) *SidebandScanner {
	return &SidebandScanner{
		r:       r,
		w:       w,
		canScan: true,
	}
}

func (scn *SidebandScanner) Scan() bool {
	scn.progressText, scn.errorText = "", "" // clear out data

	scn.b, scn.canScan, scn.scanErr = scan(scn.w, scn.r)
	if !scn.canScan || scn.scanErr != nil {
		return false
	}

	var c byte
	var sendToWriter bool
	c, scn.b = scn.b[0], scn.b[1:]

	switch c {
	case 2:
		scn.progressText = string(scn.b)
		sendToWriter = scn.w != nil // only send if Progress or Error
	case 3:
		scn.errorText = string(scn.b)
		sendToWriter = scn.w != nil // only send if Progress or Error
	}

	if sendToWriter {
		_, scn.scanErr = scn.w.Write(scn.b)
		if scn.scanErr != nil {
			return false
		}
	}

	return true
}

func (scn *SidebandScanner) Bytes() []byte { return scn.b }

func (scn *SidebandScanner) Text() string { return string(scn.b) }

func (scn *SidebandScanner) Err() error { return scn.scanErr }

func (scn *SidebandScanner) Packfile() []byte { return scn.b }

func (scn *SidebandScanner) ProgressText() string { return scn.progressText }

func (scn *SidebandScanner) ErrorText() string { return scn.errorText }
