package pktline

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncoderStrings(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		{"basic", "hello world", "000fhello world", nil, false},
		{"with newline", "hello world\n", "0010hello world\n", nil, false},
		{"with flush", "hello world", "000fhello world0000", nil, true},
		{"error too long", strings.Repeat("*", 65517), "", ErrPktlineTooLong, false},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf)
				haveErr := enc.EncodeString(test.data)
				if withFlush {
					enc.Flush()
				}
				if haveErr != wantErr {
					t.Fatalf("have: %v want: %v", haveErr, wantErr)
				}
				have := buf.String()
				if have != want {
					t.Fatalf("have: %q want: %q", have, want)
				}
			})
		}(test.data, test.want, test.wantErr, test.withFlush)
	}
}
