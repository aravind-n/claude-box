package vhrn

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runInstall pulls a harness's image and the matching-version proxy from the
// registry, unions its egress domains into the allowlist, records the
// harness+version in the installed registry, and writes shell aliases. --local uses
// images already built by `make` instead of pulling (for development/offline).
func runInstall(args []string) int {
	var arg string
	local := false
	for _, a := range args {
		if a == "--local" {
			local = true
		} else if arg == "" {
			arg = a
		}
	}
	if arg == "" {
		fmt.Fprintln(os.Stderr, "usage: vhrn install <harness>[@version] [--local]")
		return 2
	}
	name, version := parseHarnessArg(arg)
	if local {
		version = localVersion
	}
	h, ok := lookupHarness(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "vhrn: unknown harness %q (known: %s)\n", name, strings.Join(harnessNames(), ", "))
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	engine, err := detectEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}

	if err := provisionImages(engine, h, version); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}

	// Union base defaults + this harness's domains into the host allowlist,
	// append-if-missing so later user edits are respected.
	np := newNetPolicy(vhrnCache(home))
	if err := np.ensure(); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	np.seedAllowlistIfAbsent()
	np.appendMissing(h.AllowDomains)

	if err := addInstalled(home, name, version); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	if err := syncAliases(home); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: warning: could not update shell aliases: %v\n", err)
	}

	fmt.Printf("Installed %s (%s). Restart your shell (or source your rc file) to use `%s`.\n", name, version, h.Alias)
	return 0
}

// provisionImages makes the harness and its matching proxy available at the chosen
// version: pull both from the registry, or (for --local) verify the make-built
// images exist. The box and proxy are always the same version — a matched set.
func provisionImages(engine string, h Harness, version string) error {
	if engine == "container" {
		exec.Command("container", "system", "start").Run() // Apple engine needs its service up
	}
	harnessImg := harnessImageRef(h, version)
	proxyImg := proxyImageRef(version)

	if version == localVersion {
		for _, img := range []string{harnessImg, proxyImg} {
			if !imageExists(engine, img) {
				return fmt.Errorf("local image %q not found — run `make build` first", img)
			}
		}
		return nil
	}
	for _, img := range []string{proxyImg, harnessImg} {
		fmt.Fprintf(os.Stderr, "vhrn: pulling %s...\n", img)
		if err := pullImage(engine, img); err != nil {
			return fmt.Errorf("pulling %s: %w", img, err)
		}
	}
	return nil
}

// pullArgv is the engine image-pull command. Both Docker and Apple container use
// `<engine> image pull` — Apple container has no top-level `pull` subcommand.
func pullArgv(image string) []string { return []string{"image", "pull", image} }

// pullImage pulls an image with the engine, streaming progress to stderr.
func pullImage(engine, image string) error {
	cmd := exec.Command(engine, pullArgv(image)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// --- local image build, used only for the toolchain derived image (see
// toolchain.go); user-facing images are pulled, not built. ---

// buildTempDir creates a build-context temp dir under the vhrn cache. It must live
// in the home tree, not the system temp: Apple container's build cannot read a
// context under macOS's /var/folders and silently drops files from it.
func buildTempDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(vhrnCache(home), "build")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(root, "ctx-*")
}

// buildArgv assembles the engine build command line (pure, for testing).
func buildArgv(image, dockerfile, context string, extra []string) []string {
	args := []string{"build", "--tag", image, "--file", dockerfile}
	args = append(args, extra...)
	return append(args, context)
}

// buildImage runs the engine build, streaming output so the user sees progress.
func buildImage(engine, image, dockerfile, context string, extra []string) error {
	cmd := exec.Command(engine, buildArgv(image, dockerfile, context, extra)...)
	cmd.Stdout = os.Stderr // keep our stdout clean; build chatter goes to stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
