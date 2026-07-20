package vhrn

import (
	"fmt"
	"strings"
)

// runFlags holds the wrapper-owned flags consumed before the agent's own args.
type runFlags struct {
	openNet    bool     // --open-net: drop the egress guard this run
	extraAllow []string // --allow: session additions to the allowlist
	rest       []string // everything forwarded to the agent verbatim
}

// parseRunFlags consumes wrapper flags up front then forwards the rest verbatim,
// mirroring vhrn.sh's loop: --open-net / --allow[=]<d,d> are read, `--` stops flag
// reading, and the first unrecognized token ends parsing (so agent flags pass
// through untouched).
func parseRunFlags(args []string) (runFlags, error) {
	var f runFlags
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--open-net":
			f.openNet = true
			i++
		case a == "--allow":
			i++
			if i >= len(args) {
				return f, fmt.Errorf("--allow needs a domain")
			}
			f.extraAllow = append(f.extraAllow, splitDomains(args[i])...)
			i++
		case strings.HasPrefix(a, "--allow="):
			f.extraAllow = append(f.extraAllow, splitDomains(strings.TrimPrefix(a, "--allow="))...)
			i++
		case a == "--":
			f.rest = append(f.rest, args[i+1:]...)
			return f, nil
		default:
			f.rest = append(f.rest, args[i:]...)
			return f, nil
		}
	}
	return f, nil
}

// splitDomains splits a comma-separated --allow value, dropping empty fields.
func splitDomains(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
