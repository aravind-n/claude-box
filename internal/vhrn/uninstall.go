package vhrn

import (
	"fmt"
	"os"
	"os/exec"
)

// runUninstall drops a harness from the installed registry and regenerates the
// shell aliases so its alias disappears. With --image it also deletes the harness
// image (the shared base and proxy are left in place for other harnesses).
func runUninstall(args []string) int {
	var name string
	rmImage := false
	for _, a := range args {
		if a == "--image" {
			rmImage = true
		} else if name == "" {
			name = a
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: vhrn uninstall <harness> [--image]")
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	// Capture the version before dropping the entry, so --image deletes the exact
	// ref that was installed (a versioned registry ref, or the bare local name).
	version, wasInstalled := installedVersion(home, name)

	if err := removeInstalled(home, name); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	if err := syncAliases(home); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: warning: could not update shell aliases: %v\n", err)
	}

	alias := name
	if h, ok := lookupHarness(name); ok {
		alias = h.Alias
		if rmImage && !wasInstalled {
			fmt.Fprintf(os.Stderr, "vhrn: %q was not installed; no image to remove\n", name)
		} else if rmImage {
			if engine, err := detectEngine(); err == nil {
				img := harnessImageRef(h, version)
				fmt.Fprintf(os.Stderr, "vhrn: removing image %s...\n", img)
				if err := removeImage(engine, img); err != nil {
					fmt.Fprintf(os.Stderr, "vhrn: warning: could not remove image %s: %v\n", img, err)
				}
			}
		}
	} else if rmImage {
		fmt.Fprintf(os.Stderr, "vhrn: warning: unknown harness %q; cannot remove its image\n", name)
	}

	fmt.Printf("Uninstalled %s. Restart your shell to drop the `%s` alias.\n", name, alias)
	return 0
}

// removeImageArgv is the engine-specific image-delete command (Docker and Apple
// container differ: `image rm` vs `image delete`).
func removeImageArgv(engine, image string) []string {
	if engine == "docker" {
		return []string{"image", "rm", image}
	}
	return []string{"image", "delete", image}
}

func removeImage(engine, image string) error {
	cmd := exec.Command(engine, removeImageArgv(engine, image)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
