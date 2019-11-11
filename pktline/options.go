package pktline

import "io"

// EncoderOption are functional options for encoding
type EncoderOption func(*Encoder)

// ScannerOption are functional options for scanning pktline data
type ScannerOption func(*Scanner)

// WithSidebandMuxer adds the Sideband muxer to the encoder that older
// clients can use
func WithSidebandMuxer(enc *Encoder) {
	enc.Sideband = newSidebandEncoder(enc.w, 999)
}

// WithSideband64kMuxer adds the Sideband64k muxer to the encoder that
// newer clients can use
func WithSideband64kMuxer(enc *Encoder) {
	enc.Sideband = newSidebandEncoder(enc.w, 65519)
}

// WithSidebandDemuxer adds the Sideband demuxer to the scanner that
// older clients can use
func WithSidebandDemuxer(scn *Scanner) {
	scn.Sideband = newSidebandScanner(scn.w, scn.r)
}

// WithSideband64kDemuxer adds the Sideband64k demuxer to the scanner
// that newer clients can use
func WithSideband64kDemuxer(scn *Scanner) {
	scn.Sideband = newSidebandScanner(scn.w, scn.r)
}

// WithWriter adds a writer that the scanner can write it's output
// to.
func WithWriter(w io.Writer) ScannerOption {
	return func(scn *Scanner) {
		scn.w = w
		if sbScn, ok := scn.Sideband.(*SidebandScanner); ok {
			sbScn.w = w
		}
	}
}

// WithSidebandSkip allows for you to wrap the Sideband option with
// a skip option. That is if the client doesn't offer a sideband
// capability then the sideband information will be skipped. Otherwise
// there will be an error that there is no sideband information
// being written to the encoder.
func WithSidebandSkip(optFn EncoderOption, ok bool) EncoderOption {
	return func(enc *Encoder) {
		optFn(enc)
		if sbEnc, ok := enc.Sideband.(*SidebandEncoder); ok {
			sbEnc.skipWrite = ok
		}
	}
}
