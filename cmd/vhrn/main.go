// Command vhrn runs a coding agent in a container jailed to the current project,
// with default-deny network egress. This is the Go port of the original vhrn.sh
// wrapper; behavior is intended to match it exactly.
package main

import (
	"os"

	"vhrn/internal/vhrn"
)

func main() {
	os.Exit(vhrn.Run(os.Args[1:]))
}
