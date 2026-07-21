package vhrn

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfigNoFilesYieldsDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	cfg, err := loadConfig(home, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, defaultConfig()) {
		t.Errorf("no config files should yield defaults, got %+v", cfg)
	}
}

func TestLoadConfigPrecedence(t *testing.T) {
	home := t.TempDir()
	cfgHome := filepath.Join(home, "cfg")
	t.Setenv("XDG_CONFIG_HOME", cfgHome)
	os.MkdirAll(filepath.Join(cfgHome, "vhrn"), 0o755)
	os.WriteFile(filepath.Join(cfgHome, "vhrn", "config.toml"), []byte(`
[toolchains]
tools = ["go@1.26"]
[net]
mode = "report"
allow = ["global.example"]
`), 0o644)

	project := t.TempDir()
	os.WriteFile(filepath.Join(project, ".vhrn.toml"), []byte(`
[net]
allow = ["project.example"]
`), 0o644)

	cfg, err := loadConfig(home, project)
	if err != nil {
		t.Fatal(err)
	}
	// project overrides global for a field it sets
	if !reflect.DeepEqual(cfg.Net.Allow, []string{"project.example"}) {
		t.Errorf("net.allow = %v, want [project.example]", cfg.Net.Allow)
	}
	// set only in global -> inherited
	if cfg.Net.Mode != "report" {
		t.Errorf("net.mode = %q, want report", cfg.Net.Mode)
	}
	if !reflect.DeepEqual(cfg.Toolchains.Tools, []string{"go@1.26"}) {
		t.Errorf("tools = %v, want [go@1.26]", cfg.Toolchains.Tools)
	}
	// set nowhere -> default
	if !reflect.DeepEqual(cfg.Run.BlockedDirs, []string{"~", "/"}) {
		t.Errorf("blocked_dirs = %v, want default", cfg.Run.BlockedDirs)
	}
}

func TestLoadConfigMalformedIsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	project := t.TempDir()
	os.WriteFile(filepath.Join(project, ".vhrn.toml"), []byte("this is = not valid = toml"), 0o644)
	if _, err := loadConfig(home, project); err == nil {
		t.Error("malformed config should be an error")
	}
}

func TestCheckBlockedDir(t *testing.T) {
	// Resolve symlinks so the test mirrors prepareBox, which passes a physical cwd
	// (on macOS t.TempDir() lives under the /var -> /private/var symlink).
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	blocked := []string{"~", "/"}

	// Exact home and exact / are refused.
	if checkBlockedDir(home, home, blocked) == nil {
		t.Error("cwd == $HOME should be blocked")
	}
	if checkBlockedDir("/", home, []string{"/"}) == nil {
		t.Error("cwd == / should be blocked")
	}
	// A subdirectory of home is allowed — exact-match, not subtree.
	sub := filepath.Join(home, "projects", "x")
	os.MkdirAll(sub, 0o755)
	if err := checkBlockedDir(sub, home, blocked); err != nil {
		t.Errorf("a project under $HOME must run: %v", err)
	}
	// No blocked dirs -> nothing refused.
	if err := checkBlockedDir(home, home, nil); err != nil {
		t.Errorf("empty blocked list should allow anything: %v", err)
	}
}
