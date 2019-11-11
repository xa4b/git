package pktline

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var _ EncodeSideband = &noSidebandEncoder{}
var _ EncodeSideband = &SidebandEncoder{}

func TestSidebandBaseEncodeString(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		// {"basic", "hello world", "0010\x01hello world", nil, false},
		{"basic", "hello world", "0014\x01000fhello world", nil, false},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandMuxer)
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

func TestSidebandEncode(t *testing.T) {
	tests := []struct {
		name      string
		channel   SidebandChannel
		data      []byte
		want      []byte
		wantErr   error
		withFlush bool
	}{
		{"basic packfile", SidebandPackfile, []byte("hello world"), []byte("0014\x01000fhello world"), nil, false},
		{"basic progress", SidebandProgress, []byte("hello world"), []byte("0014\x02000fhello world"), nil, false},
		{"basic error", SidebandError, []byte("hello world"), []byte("0014\x03000fhello world"), nil, false},
		{"basic channel invalid", SidebandChannel(4), []byte("hello world"), nil, ErrSidebandChannelInvalid, false},
		{"basic with flush", SidebandProgress, []byte("hello world"), []byte("0014\x02000fhello world0009\x010000"), nil, true},
		{"basic line too long ", SidebandProgress, make([]byte, 1000), nil, ErrPktlineTooLong, false},
	}

	for _, test := range tests {
		func(c SidebandChannel, data, want []byte, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandMuxer)
				haveErr := enc.Sideband.Encode(c, test.data)
				if withFlush {
					enc.Sideband.Flush()
				}
				if haveErr != wantErr {
					t.Fatalf("have: %v want: %v", haveErr, wantErr)
				}
				have := buf.Bytes()
				if !bytes.Equal(have, want) {
					t.Fatalf("have: %q want: %q", have, want)
				}
			})
		}(test.channel, test.data, test.want, test.wantErr, test.withFlush)
	}
}

func TestSideband64kEncode(t *testing.T) {
	long1000Bytes := make([]byte, 1000)

	tests := []struct {
		name      string
		channel   SidebandChannel
		data      []byte
		want      []byte
		wantErr   error
		withFlush bool
	}{
		{"basic packfile", SidebandPackfile, []byte("hello world"), []byte("0014\x01000fhello world"), nil, false},
		{"basic progress", SidebandProgress, []byte("hello world"), []byte("0014\x02000fhello world"), nil, false},
		{"basic error", SidebandError, []byte("hello world"), []byte("0014\x03000fhello world"), nil, false},
		{"basic channel invalid", SidebandChannel(4), []byte("hello world"), nil, ErrSidebandChannelInvalid, false},
		{"basic with flush", SidebandProgress, []byte("hello world"), []byte("0014\x02000fhello world0009\x010000"), nil, true},
		{"basic line sideband long (valid)", SidebandProgress, long1000Bytes, append([]byte{'0', '3', 'f', '1', 0x02, '0', '3', 'e', 'c'}, long1000Bytes...), nil, false},
		{"basic line too long ", SidebandProgress, make([]byte, 65520), nil, ErrPktlineTooLong, false},
	}

	for _, test := range tests {
		func(c SidebandChannel, data, want []byte, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSideband64kMuxer)
				haveErr := enc.Sideband.Encode(c, test.data)
				if withFlush {
					enc.Sideband.Flush()
				}
				if haveErr != wantErr {
					t.Fatalf("have: %v want: %v", haveErr, wantErr)
				}
				have := buf.Bytes()
				if !bytes.Equal(have, want) {
					t.Fatalf("have: %q want: %q", have, want)
				}
			})
		}(test.channel, test.data, test.want, test.wantErr, test.withFlush)
	}
}

func TestSidebandEncodePackfile(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		want      []byte
		wantErr   error
		withFlush bool
	}{
		{"basic packfile", []byte("hello world"), []byte("0014\x01000fhello world"), nil, false},
		{"basic packfile with flush", []byte("hello world"), []byte("0014\x01000fhello world0009\x010000"), nil, true},
	}

	for _, test := range tests {
		func(data, want []byte, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandMuxer)
				haveErr := enc.Sideband.EncodePackfile(test.data)
				if withFlush {
					enc.Sideband.Flush()
				}
				if haveErr != wantErr {
					t.Fatalf("have: %v want: %v", haveErr, wantErr)
				}
				have := buf.Bytes()
				if !bytes.Equal(have, want) {
					t.Fatalf("have: %q want: %q", have, want)
				}
			})
		}(test.data, test.want, test.wantErr, test.withFlush)
	}
}

func TestSidebandEncodeProgress(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		{"basic progress", "hello world", "0010\x02hello world", nil, false},
		{"basic progress with flush", "hello world", "0010\x02hello world0000", nil, true},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandMuxer)
				haveErr := enc.Sideband.EncodeProgress(test.data)
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

func TestSidebandEncodeError(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		{"basic error", "hello world", "0010\x03hello world", nil, false},
		{"basic error with flush", "hello world", "0010\x03hello world0009\x010000", nil, true},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandMuxer)
				haveErr := enc.Sideband.EncodeError(test.data)
				if withFlush {
					enc.Sideband.Flush()
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

func TestSidebandSkipEncodeProgress(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      string
		wantErr   error
		withFlush bool
	}{
		{"basic progress", "hello world", "", nil, false},
		{"basic progress with flush", "hello world", "0000", nil, true},
	}

	for _, test := range tests {
		func(data, want string, wantErr error, withFlush bool) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				enc := NewEncoder(buf, WithSidebandSkip(WithSidebandMuxer, true))
				haveErr := enc.Sideband.EncodeProgress(test.data)
				if withFlush {
					enc.Sideband.Flush()
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

func TestSidebandNoEncode(t *testing.T) {
	data := "hello world"
	wantErr := ErrSidebandNotImplemented

	enc := NewEncoder(ioutil.Discard)

	{
		haveErr := enc.Sideband.EncodeString(SidebandPackfile, data)
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

	{
		haveErr := enc.Sideband.Encode(SidebandPackfile, []byte(data))
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

	{
		haveErr := enc.Sideband.EncodePackfile([]byte(data))
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

	{
		haveErr := enc.Sideband.EncodeProgress(data)
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

	{
		haveErr := enc.Sideband.EncodeError(data)
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

	{
		haveErr := enc.Sideband.Flush()
		if haveErr != wantErr {
			t.Fatalf("have: %v want: %v", haveErr, wantErr)
		}
	}

}

// enc := NewEncoder(w)
// enc.Encode()                  // <data-len><data> -->// encodes the byte (double encode sideband)
// enc.EncodeString()            // <data-len><data> -->// encodes the string (double encode sideband)
// enc.Write()                   // <data>           -->// a pass through writer, no encoding
// enc.WriteString()             // <data>
// end.Flush()                   // <zeros>

// enc.Sideband.Encode()         // <total-len>?<data-len><data>
// enc.Sideband.EncodeString()   // <total-len>?<data-len><data>
// enc.Sideband.Write()          // <data>
// enc.Sideband.WriteString()    // <data>
// enc.Sideband.Flush()          // <zeros>

// enc.Sideband.EncodePackfile() // <total-len>\1<data-len><data>
// enc.Sideband.EncodeProgress() // <total-len>\2<data-len><data>
// enc.Sideband.EncodeError()    // <total-len>\3<data-len><data>

// enc.WithSidebandCapability(0|1|64)
// enc.SidebandCapability() 0|1|64

// bui := NewBuilder()
// bui.Encode()       // <data-len><data>
// bui.EncodeString() // <data-len><data>
// bui.Write()
// bui.WriteString()
// bui.Flush()
// bui.FlushString()
