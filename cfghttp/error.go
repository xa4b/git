package cfghttp

import (
	"errors"
	"fmt"
)

// all provided errors
const (
	ErrSession         strErr = "%s session: %v"
	ErrSessionAdvRefs  strErr = "%s session advertised references: %v"
	ErrPackDecode      strErr = "pack decode: %v"
	ErrPackScanAdvRefs strErr = "pack scan [1] advertised references: %v"

	ErrReceivePack       strErr = "bad receive pack: %v"
	ErrUploadPackRequest strErr = "bad upload pack: %v"
	ErrUploadPack        strErr = "bad upload pack: %v"

	ErrTransportEndpoint strErr = "repo [%s] endpoint invalid: %v"
	ErrNoServiceFound    strErr = "no service found"
	ErrEmptyHookData     strErr = "empty receive-pack hook data"
)

// strErr provides an error wrapper for strings with an option to
// provide formatting values. It is used for error constants that
// have built-in formatting directives. So we can provide a base
// string constant that can be comparable by type or 'sentinel' value.
type strErr string

func (e strErr) Error() string { return string(e) }

// F captures the values for an error string formatting. This is a
// separate method so an error can be matched with its base
// formatting directives.
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
	// otherwise we have some nil item, but a valid err, so pass the err along
	if hasNil && !hasErr {
		return nil
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

func (e fmtErr) Error() string { return fmt.Sprintf(e.err.Error(), e.v...) }

// Unwrap is a method to help unwrap errors to the base error for go1.13
func (e fmtErr) Unwrap() error { return errors.Unwrap(e.err) }
