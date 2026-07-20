package vhrn

import (
	"os"
	"path/filepath"
	"strings"
)

// writeBoxGuide rebuilds the sandbox CLAUDE.md fresh each run: the host global
// CLAUDE.md (if any) followed by a guard-aware section that tracks the net mode, so
// it never accumulates across runs.
func writeBoxGuide(realClaude, sandbox string, openNet bool) error {
	var b strings.Builder
	if data, err := os.ReadFile(filepath.Join(realClaude, "CLAUDE.md")); err == nil {
		b.Write(data)
	}
	b.WriteString(boxGuideHeader)
	if openNet {
		b.WriteString(boxGuideOpen)
	} else {
		b.WriteString(boxGuideGuard)
	}
	return os.WriteFile(filepath.Join(sandbox, "CLAUDE.md"), []byte(b.String()), 0o644)
}

const boxGuideHeader = "\n# vhrn environment\n\n" +
	"You are running inside vhrn: a container jailed to this project with a\n" +
	"network egress guard. Adapt as follows:\n\n" +
	"- **No sudo, no apt.** Install tools in user space: `mise use -g <tool>` for\n" +
	"  runtimes (node, go, python, ...), `uv tool install <pkg>` for Python CLIs, and\n" +
	"  `npm i -g <pkg>` after `mise use -g node` for npm CLIs.\n"

const boxGuideOpen = "- **Network egress is unrestricted this session** (the guard is off via `--open-net`).\n"

const boxGuideGuard = "- **Network egress is allowlisted (default-deny).** A blocked request fails with\n" +
	"  an error naming the domain. You cannot change the allowlist from inside the\n" +
	"  box; tell the user the exact host(s) and ask them to run\n" +
	"  `vhrn net allow <host>` on the host, then retry — no restart is needed.\n"
