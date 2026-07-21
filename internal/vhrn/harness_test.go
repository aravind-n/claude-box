package vhrn

import "testing"

func TestLookupHarness(t *testing.T) {
	h, ok := lookupHarness("claude")
	if !ok {
		t.Fatal("claude should be a known harness")
	}
	if h.Image != "vhrn-claude" || h.Command != "claude" || h.Alias != "claude" {
		t.Errorf("claude spec mismatch: %+v", h)
	}
	if h.ConfigDirEnv != "CLAUDE_CONFIG_DIR" || h.StateDir != ".claude" {
		t.Errorf("claude persistence spec mismatch: %+v", h)
	}
	if len(h.Credentials) == 0 {
		t.Error("claude should bootstrap at least one credentials file")
	}
	if _, ok := lookupHarness("nope"); ok {
		t.Error("unknown harness should not resolve")
	}
}

func TestHarnessNamesSorted(t *testing.T) {
	names := harnessNames()
	if len(names) == 0 {
		t.Fatal("expected at least one harness")
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("harnessNames not sorted: %v", names)
		}
	}
}
