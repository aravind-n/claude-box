package egress

import "strings"

// Mode is the egress enforcement posture, mirroring smokescreen's vocabulary.
type Mode string

const (
	ModeEnforce Mode = "enforce" // allow only allowlisted hosts (default, fail-closed)
	ModeOpen    Mode = "open"    // allow any public host
	ModeReport  Mode = "report"  // allow any public host, but log what enforce would deny
)

// parseMode maps free-form file contents to a Mode, defaulting to the
// fail-closed ModeEnforce for anything unrecognized or empty.
func parseMode(s string) Mode {
	switch Mode(strings.TrimSpace(strings.ToLower(s))) {
	case ModeOpen:
		return ModeOpen
	case ModeReport:
		return ModeReport
	default:
		return ModeEnforce
	}
}
