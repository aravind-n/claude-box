package vhrn

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitConfigMount copies the host ~/.gitconfig into the cache (dereferencing
// symlinks) and mounts it at /home/dev/.gitconfig so in-box commits use the user's
// identity. A disposable copy — re-synced each run. Returns nil when absent.
func gitConfigMount(home, cache string) []string {
	src := filepath.Join(home, ".gitconfig")
	dst := filepath.Join(cache, "gitconfig")
	if !fileExists(src) {
		os.Remove(dst)
		return nil
	}
	if err := copyFile(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: warning: could not copy .gitconfig\n")
		return nil
	}
	return []string{"--volume", dst + ":/home/dev/.gitconfig"}
}

// ghTokenEnv resolves a GitHub token — explicit env wins, else `gh auth token`
// (the only route that works with macOS Keychain storage, where no file holds it)
// — and passes it as GH_TOKEN. Returns nil when the host has no gh login.
func ghTokenEnv() []string {
	tok := os.Getenv("GH_TOKEN")
	if tok == "" {
		tok = os.Getenv("GITHUB_TOKEN")
	}
	if tok == "" && lookPath("gh") {
		if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
			tok = strings.TrimSpace(string(out))
		}
	}
	if tok == "" {
		return nil
	}
	return []string{"--env", "GH_TOKEN=" + tok}
}

// terminalEnv forwards the terminal identity verbatim: Claude branches per-terminal
// rendering on these, so they are never forced or invented. TERM falls back to
// xterm-256color; the rest cross only when set.
func terminalEnv() []string {
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}
	env := []string{"--env", "TERM=" + term}
	for _, v := range []string{"COLORTERM", "TERM_PROGRAM", "TERM_PROGRAM_VERSION"} {
		if val := os.Getenv(v); val != "" {
			env = append(env, "--env", v+"="+val)
		}
	}
	return env
}
