package vhrn

import "testing"

func TestParseHarnessArg(t *testing.T) {
	cases := []struct{ in, name, version string }{
		{"claude", "claude", "latest"},
		{"claude@v0.2.0", "claude", "v0.2.0"},
		{"claude@sha-abc123", "claude", "sha-abc123"},
		{"claude@", "claude", "latest"}, // trailing @ is latest, not empty
	}
	for _, c := range cases {
		n, v := parseHarnessArg(c.in)
		if n != c.name || v != c.version {
			t.Errorf("parseHarnessArg(%q) = %q,%q want %q,%q", c.in, n, v, c.name, c.version)
		}
	}
}

func TestImageRefs(t *testing.T) {
	t.Setenv("VHRN_REGISTRY", "")
	h := Harness{Name: "claude", Image: "vhrn-claude"}

	if got := harnessImageRef(h, "v0.2.0"); got != "ghcr.io/aravind-n/vhrn-claude:v0.2.0" {
		t.Errorf("harnessImageRef = %q", got)
	}
	if got := proxyImageRef("v0.2.0"); got != "ghcr.io/aravind-n/vhrn-proxy:v0.2.0" {
		t.Errorf("proxyImageRef = %q", got)
	}
	// A local install uses the bare, make-built image names.
	if got := harnessImageRef(h, localVersion); got != "vhrn-claude" {
		t.Errorf("local harnessImageRef = %q", got)
	}
	if got := proxyImageRef(localVersion); got != "vhrn-proxy" {
		t.Errorf("local proxyImageRef = %q", got)
	}
	// VHRN_REGISTRY overrides the registry.
	t.Setenv("VHRN_REGISTRY", "example.com/team")
	if got := harnessImageRef(h, "latest"); got != "example.com/team/vhrn-claude:latest" {
		t.Errorf("override harnessImageRef = %q", got)
	}
}
