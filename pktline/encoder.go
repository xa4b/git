package pktline

import (
	"fmt"
	"io"
)

// MaxLength is The maximum length of a pkt-line's data component is 65516 bytes.
// Implementations MUST NOT send pkt-line whose length exceeds 65520
// (65516 bytes of payload + 4 bytes of length data).
const MaxLength = 65516 // the data portion

// Encoder is the underlining structure to encoding data to the
// added writer
type Encoder struct {
	w io.Writer

	Sideband EncodeSideband

	maxLnLen int
}

// NewEncoder returns an encoder that will encode packets directly to
// the underlining writer
func NewEncoder(w io.Writer, opts ...EncoderOption) *Encoder {
	enc := &Encoder{
		w:        w,
		Sideband: &noSidebandEncoder{},
		maxLnLen: MaxLength,
	}
	for _, optFn := range opts {
		optFn(enc)
	}
	return enc
}

// WithSidebandCapability takes a sideband value and sets up the sideband
// encoding if the value is Sideband or Sideband64k. If the value is empty
// then no side-band is used. Calls to Sideband encoding will result in an error
func (enc *Encoder) WithSidebandCapability(cap SidebandCapability) *Encoder {
	if cap == SidebandNone {
		if capEnc, ok := enc.Sideband.(interface{ setSkip(bool) }); ok {
			capEnc.setSkip(true)
		}
	}
	return enc
}

// SidebandCapability returns if the pktline encoder has any
// sideband capabilities initiated.
func (enc *Encoder) SidebandCapability() SidebandCapability {
	if capEnc, ok := enc.Sideband.(*SidebandEncoder); ok {
		if capEnc.maxLnLen > 5000 {
			return Sideband64k
		}
		return Sideband
	}
	return SidebandNone
}

// Write writes bytes directly to the underlining writer. No encoding
// is preformed. It is expected the bytes written this way will have
// be previously encoded.
func (enc *Encoder) Write(d []byte) (n int, err error) {
	return enc.w.Write(d)
}

// WriteString writes the string to the underlining writer. No encoding
// is preformed. It is expected the bytes written this way will have
// be previously encoded.
func (enc *Encoder) WriteString(s string) (n int, err error) {
	return enc.Write([]byte(s))
}

// Encode encodes bytes to the underlining writer. If a sideband is
// initiated, then the bytes will be encoded as a sideband mux'd packet line
func (enc *Encoder) Encode(b []byte) (err error) {
	if capEnc, ok := enc.Sideband.(interface{ skip() bool }); ok && !capEnc.skip() {
		return enc.Sideband.EncodePackfile(b)
	}
	return encode(enc.w, enc.maxLnLen, b)
}

// EncodeString encodes a string in the pkline format to the underlining writer.
// If a sideband is used then the string will be muxed as sideband or sidebadn64k
func (enc *Encoder) EncodeString(s string) (err error) { return enc.Encode([]byte(s)) }

// Flush sends 0000 to the undrtlining writer
func (enc *Encoder) Flush() {
	flush(enc.w)
}

// flush writes the 0000 packets to the writer
func flush(w io.Writer) error {
	_, err := w.Write([]byte("0000"))
	return err
}

// encode takes the writer and the max bytes to write. It encodes the
// line length and bytes in the pktline format to the underlining writer
func encode(w io.Writer, max int, b []byte) (err error) {
	if len(b) > max {
		return ErrPktlineTooLong
	}

	_, err = w.Write(append([]byte(fmt.Sprintf("%04x", len(b)+4)), b...))
	if err != nil {
		return err
	}

	return nil
}
