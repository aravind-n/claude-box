package vhrn

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// normalizeTools trims, drops empties, de-duplicates, and sorts the tool list so
// the content hash is stable regardless of order or incidental whitespace.
func normalizeTools(tools []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tools {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// toolchainTag is the content-addressed image tag for a tool set,
// <base>-tc-<hash12> (base is the clean local image name, e.g. vhrn-claude — not
// the pulled registry ref, which carries a colon and can't be a tag prefix). The
// same tools always map to the same tag, so the derived image is built once.
func toolchainTag(base string, tools []string) string {
	sum := sha256.Sum256([]byte(strings.Join(normalizeTools(tools), "\n")))
	return fmt.Sprintf("%s-tc-%s", base, hex.EncodeToString(sum[:])[:12])
}

// toolchainDockerfile derives an image FROM the harness image that provisions the
// tools with mise, as the unprivileged dev user (mise installs into its home).
func toolchainDockerfile(baseImage string, tools []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "FROM %s\n", baseImage)
	b.WriteString("USER dev\n")
	fmt.Fprintf(&b, "RUN mise use -g %s\n", strings.Join(normalizeTools(tools), " "))
	b.WriteString("USER root\n")
	return b.String()
}

// imageExists reports whether the engine already has the image locally.
func imageExists(engine, image string) bool {
	return exec.Command(engine, "image", "inspect", image).Run() == nil
}

// ensureToolchainImage returns the image to run: fromImage unchanged when no tools
// are declared, else a content-addressed derived image (FROM fromImage, tagged from
// the clean tagBase), built once and cached by its tag.
func ensureToolchainImage(engine, fromImage, tagBase string, tools []string) (string, error) {
	norm := normalizeTools(tools)
	if len(norm) == 0 {
		return fromImage, nil
	}
	tag := toolchainTag(tagBase, norm)
	if imageExists(engine, tag) {
		return tag, nil
	}
	tmp, err := buildTempDir()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	dockerfile := filepath.Join(tmp, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte(toolchainDockerfile(fromImage, norm)), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "vhrn: provisioning toolchain (%s) into %s...\n", strings.Join(norm, ", "), tag)
	if err := buildImage(engine, tag, dockerfile, tmp, nil); err != nil {
		return "", err
	}
	return tag, nil
}
