package pktline

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

var _ ScanSideband = &SidebandScanner{}

func TestNewScanner(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr error
	}{
		{
			// https://github.com/git/git/blob/master/Documentation/technical/pack-protocol.txt (bottom)
			name: "from pack-protocol.txt [1]",
			data: "00677d1665144a3a975c05f1f43902ddaf084e784dbe 74730d410fcb6603ace96f1dc55ea6196122532d refs/heads/debug\n006874730d410fcb6603ace96f1dc55ea6196122532d 5a3f6be755bbb7deae50065988cbfa1ffa9ab68a refs/heads/master\n0000",
			want: []string{
				"7d1665144a3a975c05f1f43902ddaf084e784dbe 74730d410fcb6603ace96f1dc55ea6196122532d refs/heads/debug\n",
				"74730d410fcb6603ace96f1dc55ea6196122532d 5a3f6be755bbb7deae50065988cbfa1ffa9ab68a refs/heads/master\n",
			},
		},
		{
			name: "bad line length after parse",
			data: "00677d1665144a3a975c05f1f43902ddaf084e784dbe 74730d410fcb6603ace96f1dc55ea6196122532d refs/heads/debug\nxxxx74730d410fcb6603ace96f1dc55ea6196122532d 5a3f6be755bbb7deae50065988cbfa1ffa9ab68a refs/heads/master\n0000",
			want: []string{
				"7d1665144a3a975c05f1f43902ddaf084e784dbe 74730d410fcb6603ace96f1dc55ea6196122532d refs/heads/debug\n",
			},
			wantErr: ErrScanInvalidLineLength,
		},
		{
			name:    "bad line length before parse",
			data:    "xxxx7d1665144a3a975c05f1f43902ddaf084e784dbe 74730d410fcb6603ace96f1dc55ea6196122532d refs/heads/debug\n006874730d410fcb6603ace96f1dc55ea6196122532d 5a3f6be755bbb7deae50065988cbfa1ffa9ab68a refs/heads/master\n0000",
			want:    []string{},
			wantErr: ErrScanInvalidLineLength,
		},
	}

	for _, test := range tests {
		func(data io.Reader, want []string, wantErr error) {
			t.Run(test.name, func(t *testing.T) {
				scn := NewScanner(data)
				var lenHave int
				for scn.Scan() {
					lenHave++
					if lenHave > len(want) {
						t.Log("too many scanlines")
						t.Fatalf("have: %d want: %d", lenHave, len(want))
					}
					have := scn.Text()
					if have != want[lenHave-1] {
						t.Fatalf("want: %q have: %q", have, want[lenHave-1])
					}
				}
				if lenHave != len(want) {
					t.Log("not enough scanlines")
					t.Fatalf("have: %d want: %d", lenHave, len(want))
				}
				haveErr := scn.Err()
				if haveErr != wantErr {
					t.Fatalf("have: %q want: %q", haveErr, wantErr)
				}
			})
		}(strings.NewReader(test.data), test.want, test.wantErr)
	}
}

func TestNewSidebandScanner(t *testing.T) {
	type wantStrings struct{ kind, value string }

	tests := []struct {
		name    string
		data    string
		want    []wantStrings
		wantErr error
	}{
		{
			name: "from pack-protocol.txt [1]",
			data: "0010\x02Hello World0011\x01Data Data...0010\x02Hello Test\n0011\x01More Data...000a\x02Done.0000",
			want: []wantStrings{
				{"progress", "Hello World"},
				{"packfile", "Data Data..."},
				{"progress", "Hello Test\n"},
				{"packfile", "More Data..."},
				{"progress", "Done."},
			},
		},
	}

	for _, test := range tests {
		func(data io.Reader, want []wantStrings, wantErr error) {
			t.Run(test.name, func(t *testing.T) {
				scn := NewScanner(data, WithSidebandDemuxer)
				var lenHave int
				for scn.Scan() {
					lenHave++
					if lenHave > len(want) {
						t.Log("too many scanlines")
						t.Fatalf("have: %d want: %d", lenHave, len(want))
					}
					havePackfileKind, haveProgressKind := "packfile", "progress"
					haveProgressValue := scn.Sideband.ProgressText()
					if len(haveProgressValue) > 0 {
						if haveProgressKind != want[lenHave-1].kind {
							t.Fatalf("have: %q want: %q", haveProgressKind, want[lenHave-1].kind)
						}
						if haveProgressValue != want[lenHave-1].value {
							t.Fatalf("have: %q want: %q", haveProgressValue, want[lenHave-1].value)
						}
						continue
					}
					havePackfileValue := string(scn.Sideband.Packfile()) // packfile
					if len(havePackfileValue) > 0 {
						if havePackfileKind != want[lenHave-1].kind {
							t.Fatalf("have: %q want: %q", havePackfileKind, want[lenHave-1].kind)
						}
						if havePackfileValue != want[lenHave-1].value {
							t.Fatalf("have: %q want: %q", havePackfileValue, want[lenHave-1].value)
						}
					}
				}
				if lenHave != len(want) {
					t.Log("not enough scanlines")
					t.Fatalf("have: %d want: %d", lenHave, len(want))
				}
				haveErr := scn.Err()
				if haveErr != wantErr {
					t.Fatalf("have: %q want: %q", haveErr, wantErr)
				}
			})
		}(strings.NewReader(test.data), test.want, test.wantErr)
	}
}

func TestNewSidebandScannerToWriter(t *testing.T) {
	type wantStrings struct{ kind, value string }

	tests := []struct {
		name        string
		data        string
		want        []wantStrings
		wantWritten string
		wantErr     error
	}{
		{
			name: "from pack-protocol.txt [1]",
			data: "0010\x02Hello World0011\x01Data Data...0010\x02Hello Test\n0011\x01More Data...000a\x02Done.0000",
			want: []wantStrings{
				{"progress", "Hello World"},
				{"packfile", "Data Data..."},
				{"progress", "Hello Test\n"},
				{"packfile", "More Data..."},
				{"progress", "Done."},
			},
			wantWritten: "Hello WorldHello Test\nDone.",
		},
	}

	for _, test := range tests {
		func(data io.Reader, want []wantStrings, wantWritten string, wantErr error) {
			t.Run(test.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				scn := NewScanner(data, WithSidebandDemuxer, WithWriter(buf))
				var lenHave int
				for scn.Scan() {
					lenHave++
					if lenHave > len(want) {
						t.Log("too many scanlines")
						t.Fatalf("have: %d want: %d", lenHave, len(want))
					}
					havePackfileKind, haveProgressKind := "packfile", "progress"
					haveProgressValue := scn.Sideband.ProgressText()
					if len(haveProgressValue) > 0 {
						if haveProgressKind != want[lenHave-1].kind {
							t.Fatalf("have: %q want: %q", haveProgressKind, want[lenHave-1].kind)
						}
						if haveProgressValue != want[lenHave-1].value {
							t.Fatalf("have: %q want: %q", haveProgressValue, want[lenHave-1].value)
						}
						continue
					}
					havePackfileValue := string(scn.Sideband.Packfile()) // packfile
					if len(havePackfileValue) > 0 {
						if havePackfileKind != want[lenHave-1].kind {
							t.Fatalf("have: %q want: %q", havePackfileKind, want[lenHave-1].kind)
						}
						if havePackfileValue != want[lenHave-1].value {
							t.Fatalf("have: %q want: %q", havePackfileValue, want[lenHave-1].value)
						}
					}
				}
				if lenHave != len(want) {
					t.Log("not enough scanlines")
					t.Fatalf("have: %d want: %d", lenHave, len(want))
				}
				haveWritten := buf.String()
				if haveWritten != wantWritten {
					t.Fatalf("have: %q want: %q", haveWritten, wantWritten)
				}
				haveErr := scn.Err()
				if haveErr != wantErr {
					t.Fatalf("have: %q want: %q", haveErr, wantErr)
				}
			})
		}(strings.NewReader(test.data), test.want, test.wantWritten, test.wantErr)
	}
}
