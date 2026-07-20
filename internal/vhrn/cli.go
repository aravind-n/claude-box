// Package vhrn implements the vhrn command: it dispatches subcommands and runs a
// coding agent inside a jailed, egress-guarded container. It is a behavior-
// preserving port of vhrn.sh; comments explain why, not what, and stay terse.
package vhrn

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Run dispatches args and returns the process exit code. It mirrors vhrn.sh's
// top-level control flow: help when it leads, `net ...` for egress policy, else
// the run path with wrapper-owned flags consumed up front.
func Run(args []string) int {
	// Answer help only when it leads, so `vhrn -- --help` and a trailing --help
	// still reach the agent's own help.
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			fmt.Print(usageText)
			return 0
		}
	}

	// `vhrn net ...` mutates the host-side egress policy, then exits. It is the
	// only path to the policy — the box itself cannot reach it.
	if len(args) > 0 && args[0] == "net" {
		return runNet(args[1:])
	}

	f, err := parseRunFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 2
	}
	if err := runAgent(f); err != nil {
		// Propagate the agent's own exit code silently; only wrapper-level
		// failures get a "vhrn:" message.
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	return 0
}
