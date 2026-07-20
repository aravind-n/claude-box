package vhrn

import "testing"

func TestDetectEngineExplicitOverride(t *testing.T) {
	// An explicit engine that exists on PATH is honored. `ls` stands in for a real
	// engine binary so the test is deterministic across machines.
	t.Setenv("VHRN_ENGINE", "ls")
	got, err := detectEngine()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ls" {
		t.Errorf("engine = %q, want %q", got, "ls")
	}
}

func TestDetectEngineExplicitMissing(t *testing.T) {
	t.Setenv("VHRN_ENGINE", "vhrn-no-such-engine-xyz")
	if _, err := detectEngine(); err == nil {
		t.Fatalf("expected error for a missing explicit engine")
	}
}
