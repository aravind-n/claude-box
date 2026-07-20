package egress

import (
	"os"
	"strings"
	"sync"
	"time"
)

// Verdict is the outcome of a policy check for one destination host.
type Verdict struct {
	Allow  bool // permit the connection
	Logged bool // would be denied under enforce; the caller should record it
	Mode   Mode // the mode the decision was made under
}

// Policy is the hot-reloaded egress policy: an allowlist and a mode, each read
// from a host-controlled file and refreshed whenever that file changes, so the
// host can retune egress without restarting the proxy. It is safe for
// concurrent use, and reloads never disturb connections already established.
type Policy struct {
	allowPath, modePath string

	mu       sync.Mutex
	allow    []string
	allowMod time.Time
	mode     Mode
	modeMod  time.Time
}

// NewPolicy returns a Policy backed by the given allowlist and mode files.
func NewPolicy(allowPath, modePath string) *Policy {
	return &Policy{allowPath: allowPath, modePath: modePath}
}

// refresh reloads whichever backing files changed since the last check. The
// allowlist is replaced wholesale, never mutated in place. Caller holds mu.
func (p *Policy) refresh() {
	if fi, err := os.Stat(p.modePath); err == nil {
		if m := fi.ModTime(); !m.Equal(p.modeMod) {
			p.modeMod = m
			p.mode = parseMode(firstLine(p.modePath))
		}
	} else {
		p.mode = ModeEnforce // no mode file => fail closed
	}

	if fi, err := os.Stat(p.allowPath); err == nil {
		if m := fi.ModTime(); !m.Equal(p.allowMod) {
			p.allowMod = m
			p.allow = loadAllow(p.allowPath)
		}
	} else {
		p.allow = nil // no allowlist => deny all under enforce
	}
}

// Mode returns the current enforcement mode.
func (p *Policy) Mode() Mode {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refresh()
	return p.mode
}

// Check reports whether egress to host is permitted under the current policy.
func (p *Policy) Check(host string) Verdict {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refresh()
	matched := hostAllowed(host, p.allow)
	switch p.mode {
	case ModeOpen:
		return Verdict{Allow: true, Logged: false, Mode: p.mode}
	case ModeReport:
		return Verdict{Allow: true, Logged: !matched, Mode: p.mode}
	default: // enforce
		return Verdict{Allow: matched, Logged: !matched, Mode: p.mode}
	}
}

// firstLine returns the first line of a file, or "" if it cannot be read.
func firstLine(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if i := strings.IndexByte(string(b), '\n'); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// loadAllow reads a domain-per-line allowlist, ignoring blank lines and #
// comments (whole-line or trailing). It returns nil for a missing file, so an
// absent allowlist denies everything under enforce.
func loadAllow(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if e := normEntry(line); e != "" {
			out = append(out, e)
		}
	}
	return out
}
