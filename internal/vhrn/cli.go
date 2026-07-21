// Package vhrn implements the vhrn command: a subcommand-first CLI that installs,
// manages, and runs coding agents ("harnesses") inside a jailed, egress-guarded
// container. Comments explain why, not what, and stay terse.
package vhrn

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Run dispatches a subcommand and returns the process exit code. vhrn is
// subcommand-first: bare `vhrn` prints help, `vhrn <harness> ...` runs an agent in
// the box, and install/uninstall/list/net manage the environment around it.
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Print(usageText)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Print(usageText)
		return 0
	case "net":
		return runNet(args[1:])
	case "install":
		return runInstall(args[1:])
	case "uninstall":
		return runUninstall(args[1:])
	case "list":
		return runList(args[1:])
	}

	// A known harness name runs that agent; the wrapper's own flags come right
	// after it, then everything else forwards to the agent verbatim.
	if h, ok := lookupHarness(args[0]); ok {
		f, err := parseRunFlags(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
			return 2
		}
		if err := runHarness(h, f); err != nil {
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

	fmt.Fprintf(os.Stderr, "vhrn: unknown command %q — run 'vhrn help'\n", args[0])
	return 2
}
