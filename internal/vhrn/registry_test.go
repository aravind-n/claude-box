package vhrn

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInstalledRegistry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))

	if got := readInstalled(home); got != nil {
		t.Fatalf("fresh registry should be empty, got %v", got)
	}
	if isInstalled(home, "claude") {
		t.Fatal("claude should not be installed yet")
	}

	if err := addInstalled(home, "claude", "v0.2.0"); err != nil {
		t.Fatal(err)
	}
	if err := addInstalled(home, "codex", "latest"); err != nil {
		t.Fatal(err)
	}
	// Re-adding a harness updates its version in place.
	if err := addInstalled(home, "claude", "v0.3.0"); err != nil {
		t.Fatal(err)
	}

	want := []installedHarness{{"claude", "v0.3.0"}, {"codex", "latest"}}
	if got := readInstalled(home); !reflect.DeepEqual(got, want) {
		t.Fatalf("readInstalled = %v, want %v", got, want)
	}
	if v, ok := installedVersion(home, "claude"); !ok || v != "v0.3.0" {
		t.Errorf("installedVersion(claude) = %q,%v want v0.3.0,true", v, ok)
	}

	if err := removeInstalled(home, "claude"); err != nil {
		t.Fatal(err)
	}
	if got, want := readInstalled(home), []installedHarness{{"codex", "latest"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("after remove = %v, want %v", got, want)
	}
}

func TestReadInstalledBareNameDefaultsLatest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	os.MkdirAll(vhrnConfigDir(home), 0o755)
	os.WriteFile(installedRegistryPath(home), []byte("claude\n"), 0o644)
	if v, ok := installedVersion(home, "claude"); !ok || v != "latest" {
		t.Errorf("bare name should default to latest, got %q,%v", v, ok)
	}
}
