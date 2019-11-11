package cfgssh

import (
	"errors"
	"fmt"
)

// strErr is a simple type that will convert a string
// to an error. This is used so that we can add errors
// as constants to the package
type strErr string

// Error returns the error string
func (e strErr) Error() string { return string(e) }

// F captures the values for string formating of an error
// the two are seperate so that an error can be matched
// with its base formmating directives.
func (e strErr) F(v ...interface{}) error {
	var hasErr, hasNil bool
	for _, vv := range v {
		switch err := vv.(type) {
		case error:
			if err == nil {
				return nil
			}
			hasErr = true
		case nil:
			hasNil = true
		}
	}

	// if there is no error object, and we have a nil, then the err is nil
	if !hasErr && hasNil {
		return nil // so we pass along nil err as expected
	}

	return fmtErr{err: fmt.Errorf("%w", e), v: v}
}

// fmtErr is for errors that will be formatted. It hold the
// formatting values in a field so they can be added when the
// error is stringfied. Otherwise the underlining error without
// formatting can be matched.
type fmtErr struct {
	err error
	v   []interface{}
}

// Error returns the string of the error
func (e fmtErr) Error() string { return fmt.Sprintf(e.err.Error(), e.v...) }

// Unwrap is a method to help unwrap errors to
// the base error for go1.13
func (e fmtErr) Unwrap() error { return errors.Unwrap(e.err) }

// returns all of the errors
const (
	ErrSession        strErr = "%s session: %v"
	ErrSessionAdvRefs strErr = "%s session advertised references: %v"

	ErrAdvRefsEncode   strErr = "%s advertied references encode: %v"
	ErrPackDecode      strErr = "pack decode: %v"
	ErrPackScanAdvRefs strErr = "pack scan [1] advertised references: %v"

	ErrResponseEncode strErr = "%s response encode: %v"
	ErrRequestDecode  strErr = "%s request decode: %v"

	ErrReceivePack strErr = "bad receive pack: %v"
	ErrUploadPack  strErr = "bad upload pack: %v"

	ErrTransportEndpoint strErr = "repo [%s] endpoint invalid: %v"
	ErrEmptyHookData     strErr = "empty receive-pack hook data"
)
