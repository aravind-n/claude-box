package vhrn

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// installedHarness is a registry entry: a harness name and the image version it was
// installed at (a tag like "v0.2.0" or "latest", or "local" for a make-built image).
type installedHarness struct {
	Name    string
	Version string
}

// vhrnConfigDir is the XDG config root for vhrn (${XDG_CONFIG_HOME:-~/.config}/vhrn).
// It holds host-owned, user-facing state: the installed registry and config.toml.
func vhrnConfigDir(home string) string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "vhrn")
}

// installedRegistryPath is the "name version" list of installed harnesses. The
// shell aliases are regenerated from it and the run path resolves the image ref
// from it, so it is the source of truth for what `vhrn install` has set up.
func installedRegistryPath(home string) string {
	return filepath.Join(vhrnConfigDir(home), "installed")
}

// readInstalled returns installed harnesses sorted by name, de-duplicated by name.
// Lines are "name version"; a bare "name" defaults to version "latest".
func readInstalled(home string) []installedHarness {
	f, err := os.Open(installedRegistryPath(home))
	if err != nil {
		return nil
	}
	defer f.Close()
	byName := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		t := strings.TrimSpace(s.Text())
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		fields := strings.Fields(t)
		version := "latest"
		if len(fields) > 1 {
			version = fields[1]
		}
		byName[fields[0]] = version
	}
	out := make([]installedHarness, 0, len(byName))
	for n, v := range byName {
		out = append(out, installedHarness{Name: n, Version: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// installedVersion returns the version a harness is installed at, and whether it is
// installed at all.
func installedVersion(home, name string) (string, bool) {
	for _, h := range readInstalled(home) {
		if h.Name == name {
			return h.Version, true
		}
	}
	return "", false
}

func isInstalled(home, name string) bool {
	_, ok := installedVersion(home, name)
	return ok
}

// writeInstalled writes the registry atomically (same-dir temp + rename), sorted and
// de-duplicated by name, one "name version" per line.
func writeInstalled(home string, hs []installedHarness) error {
	dir := vhrnConfigDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	sorted := append([]installedHarness(nil), hs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	var buf bytes.Buffer
	buf.WriteString("# vhrn installed harnesses — managed by `vhrn install`/`uninstall`.\n")
	seen := map[string]bool{}
	for _, h := range sorted {
		if h.Name == "" || seen[h.Name] {
			continue
		}
		seen[h.Name] = true
		version := h.Version
		if version == "" {
			version = "latest"
		}
		fmt.Fprintf(&buf, "%s %s\n", h.Name, version)
	}
	tmp, err := os.CreateTemp(dir, "installed.*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	tmp.Close()
	return os.Rename(name, installedRegistryPath(home))
}

// addInstalled records a harness at a version, updating the version if it is already
// present.
func addInstalled(home, name, version string) error {
	hs := readInstalled(home)
	for i := range hs {
		if hs[i].Name == name {
			hs[i].Version = version
			return writeInstalled(home, hs)
		}
	}
	return writeInstalled(home, append(hs, installedHarness{Name: name, Version: version}))
}

func removeInstalled(home, name string) error {
	var kept []installedHarness
	for _, h := range readInstalled(home) {
		if h.Name != name {
			kept = append(kept, h)
		}
	}
	return writeInstalled(home, kept)
}
