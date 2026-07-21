package vhrn

import (
	"fmt"
	"os"
)

// runList shows every known harness and whether `vhrn install` has set it up.
func runList(args []string) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	installed := map[string]string{}
	for _, ih := range readInstalled(home) {
		installed[ih.Name] = ih.Version
	}
	for _, name := range harnessNames() {
		if v, ok := installed[name]; ok {
			fmt.Printf("  %-12s installed (%s)\n", name, v)
		} else {
			fmt.Printf("  %-12s available\n", name)
		}
	}
	return 0
}
