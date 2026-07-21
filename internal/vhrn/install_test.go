package vhrn

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildArgv(t *testing.T) {
	got := buildArgv("vhrn-claude", "/tmp/ctx/Dockerfile", "/tmp/ctx", []string{"--build-arg", "BASE=vhrn-base"})
	want := []string{"build", "--tag", "vhrn-claude", "--file", "/tmp/ctx/Dockerfile", "--build-arg", "BASE=vhrn-base", "/tmp/ctx"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildArgv = %v, want %v", got, want)
	}
}

func TestPullArgv(t *testing.T) {
	// Both engines pull via `<engine> image pull` — Apple container has no
	// top-level `pull` subcommand.
	got := pullArgv("ghcr.io/aravind-n/vhrn-claude:v0.2.0")
	want := []string{"image", "pull", "ghcr.io/aravind-n/vhrn-claude:v0.2.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pullArgv = %v, want %v", got, want)
	}
}

func TestBuildTempDirUnderHomeCache(t *testing.T) {
	// The build context must land under the home tree, not the system temp: Apple
	// container's build can't read a context under macOS's /var/folders.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", "")
	dir, err := buildTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if want := filepath.Join(home, ".cache", "vhrn", "build"); !strings.HasPrefix(dir, want) {
		t.Errorf("build dir %q must be under %q, not the system temp", dir, want)
	}
}
