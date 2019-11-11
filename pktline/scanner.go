package pktline

import (
	"bytes"
	"io"
	"strconv"
)

// Scanner is the object that allows scanning of pktline data
type Scanner struct {
	r io.Reader
	b []byte
	w io.Writer

	Sideband ScanSideband

	canScan bool
	scanErr error
}

// NewScanner returns a scanner object that takes functional options
func NewScanner(r io.Reader, opts ...ScannerOption) *Scanner {
	scn := &Scanner{r: r, canScan: true}
	for _, optFn := range opts {
		optFn(scn)
	}
	return scn
}

// Scan is true while there are items to scan
func (scn *Scanner) Scan() bool {
	if scn.canScan == false {
		return false
	}

	if scn.Sideband != nil {
		scn.b = scn.Sideband.Bytes()
		scn.canScan = scn.Sideband.Scan()
		scn.scanErr = scn.Err()
		return scn.canScan
	}

	scn.b, scn.canScan, scn.scanErr = scan(scn.w, scn.r)
	return scn.canScan
}

// Bytes returns the bytes of the line scanned
func (scn *Scanner) Bytes() []byte {
	if scn.scanErr != nil {
		return nil
	}
	return scn.b
}

// Text returns the string text of the line scanned
func (scn *Scanner) Text() string {
	if scn.scanErr != nil {
		return ""
	}
	return string(scn.b)
}

// Err returns any errors that occurred while scanning
func (scn *Scanner) Err() error {
	return scn.scanErr
}

// scan scans a pktline according to the spec
func scan(w io.Writer, r io.Reader) (b []byte, ok bool, err error) {
	var lnlenStr [4]byte
	r.Read(lnlenStr[:])

	lnlen, _ := strconv.ParseUint(string(lnlenStr[:]), 16, 16) // we check for 0 on the next line

	// check if we have 0's to stop scanning
	if lnlen == 0 {
		if bytes.Equal(lnlenStr[:], []byte("0000")) {
			return nil, false, nil
		}
		err = ErrScanInvalidLineLength
		return
	}

	lnlen -= 4 // account for the len bytes

	var n int
	b = make([]byte, lnlen)
	n, err = r.Read(b)
	if err != nil {
		return
	}
	if uint64(n) != lnlen {
		err = ErrScanTooShort
		return
	}

	ok = true
	return
}
