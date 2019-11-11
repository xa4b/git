package pktline

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		{"basic", "hello world", "000fhello world", nil, false},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				build := NewBuilder()
				build.EncodeString(test.data)
				if withFlush {
					build.Flush()
				}
				have := build.String()
				if have != want {
					t.Fatalf("have: %q want: %q", have, want)
				}
			})
		}(test.data, test.want, test.wantErr, test.withFlush)
	}
}
