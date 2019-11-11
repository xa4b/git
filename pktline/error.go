package pktline

// strErr is a simple wrapper around a string so that we can use constants
// with our errors
type strErr string

// Error provides the error interface
func (e strErr) Error() string { return string(e) }

// errors we handle
const (
	ErrSidebandNotImplemented strErr = "sideband is not implemented"
	ErrSidebandChannelInvalid strErr = "sideband channel is invalid"

	ErrPktlineTooLong strErr = "pktline is too long to encode"

	ErrScanTooShort          strErr = "pktline scan was too short"
	ErrScanInvalidLineLength strErr = "pktline scan line length is invalid"
)
