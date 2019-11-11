package pktline

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp"
)

// ReportStatus returns the unpacked status from a report status it
// returns a sideband aware muxed value if necessary.
func ReportStatus(enc *Encoder, status *packp.ReportStatus) {
	enc.EncodeString(fmt.Sprintf("unpack %s\n", status.UnpackStatus))
	for _, cs := range status.CommandStatuses {
		if cs.Error() != nil {
			enc.EncodeString(fmt.Sprintf("ng %s %s\n", cs.ReferenceName.String(), cs.Status))
			continue
		}
		enc.EncodeString(fmt.Sprintf("ok %s\n", cs.ReferenceName.String()))
	}
	enc.Sideband.Flush()
}
