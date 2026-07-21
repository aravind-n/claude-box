package vhrn

import (
	"strings"
	"testing"
)

func TestToolchainTagStable(t *testing.T) {
	a := toolchainTag("vhrn-claude", []string{"go@1.26", "node@22"})
	b := toolchainTag("vhrn-claude", []string{"node@22", " go@1.26 ", "node@22"}) // reorder + whitespace + dup
	if a != b {
		t.Errorf("tag must be order/whitespace/dup independent: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "vhrn-claude-tc-") {
		t.Errorf("unexpected tag %q", a)
	}
	if toolchainTag("vhrn-claude", []string{"go@1.26"}) == a {
		t.Error("different tool sets should produce different tags")
	}
}

func TestToolchainDockerfile(t *testing.T) {
	df := toolchainDockerfile("vhrn-claude", []string{"node@22", "go@1.26"})
	if !strings.Contains(df, "FROM vhrn-claude") {
		t.Errorf("missing FROM:\n%s", df)
	}
	if !strings.Contains(df, "mise use -g go@1.26 node@22") { // normalized order
		t.Errorf("tools not provisioned in sorted order:\n%s", df)
	}
	if !strings.Contains(df, "USER dev") || !strings.Contains(df, "USER root") {
		t.Errorf("should provision as dev then return to root:\n%s", df)
	}
}

func TestEnsureToolchainImageNoTools(t *testing.T) {
	// No tools must pass the harness image through untouched, without touching the
	// engine.
	img, err := ensureToolchainImage("container", "ghcr.io/x/vhrn-claude:v1", "vhrn-claude", nil)
	if err != nil || img != "ghcr.io/x/vhrn-claude:v1" {
		t.Errorf("no tools should pass fromImage through: %q, %v", img, err)
	}
}
