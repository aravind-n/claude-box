package vhrn

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the merged vhrn configuration. Precedence is project .vhrn.toml over
// global config.toml over built-in defaults (CLI flags win over all of it, applied
// in the run path). Unset fields (nil slice / empty string) let a lower-precedence
// source show through, so each file need only specify what it changes.
type Config struct {
	Run        RunConfig        `toml:"run"`
	Toolchains ToolchainsConfig `toml:"toolchains"`
	Net        NetConfig        `toml:"net"`
}

// RunConfig guards where a box may launch. blocked_dirs are refused as an exact
// resolved cwd (not a subtree), so ordinary projects under $HOME still run while
// jailing all of $HOME or / is prevented.
type RunConfig struct {
	BlockedDirs []string `toml:"blocked_dirs"`
}

// ToolchainsConfig lists tools provisioned into the box beyond the base image,
// e.g. "go@1.26", "node@22".
type ToolchainsConfig struct {
	Tools []string `toml:"tools"`
}

// NetConfig folds into the egress policy: extra allowlist domains and the guard
// mode (enforce | report | open).
type NetConfig struct {
	Allow []string `toml:"allow"`
	Mode  string   `toml:"mode"`
}

// defaultConfig is the lowest-precedence layer.
func defaultConfig() Config {
	return Config{
		Run: RunConfig{BlockedDirs: []string{"~", "/"}},
		Net: NetConfig{Mode: "enforce"},
	}
}

// mergeConfig overlays over onto base: a field in over wins only when it is set,
// so an unspecified key falls through to the lower-precedence layer.
func mergeConfig(base, over Config) Config {
	out := base
	if over.Run.BlockedDirs != nil {
		out.Run.BlockedDirs = over.Run.BlockedDirs
	}
	if over.Toolchains.Tools != nil {
		out.Toolchains.Tools = over.Toolchains.Tools
	}
	if over.Net.Allow != nil {
		out.Net.Allow = over.Net.Allow
	}
	if over.Net.Mode != "" {
		out.Net.Mode = over.Net.Mode
	}
	return out
}

// loadConfig merges defaults, the global config, and the project config in
// precedence order. Missing files are not an error; a malformed one is.
func loadConfig(home, project string) (Config, error) {
	cfg := defaultConfig()
	for _, path := range []string{
		filepath.Join(vhrnConfigDir(home), "config.toml"),
		filepath.Join(project, ".vhrn.toml"),
	} {
		c, err := readConfigFile(path)
		if err != nil {
			return cfg, err
		}
		if c != nil {
			cfg = mergeConfig(cfg, *c)
		}
	}
	return cfg, nil
}

// checkBlockedDir refuses to launch when the resolved cwd exactly matches a
// blocked dir. The match is exact, not subtree: subtree-blocking ~ would refuse
// every project under $HOME, so exact-match is what actually prevents jailing all
// of $HOME or / (or any other exact dir the user lists) while leaving ordinary
// projects runnable.
func checkBlockedDir(project, home string, blocked []string) error {
	for _, b := range blocked {
		if resolveDir(b, home) == project {
			return fmt.Errorf("refusing to run in %s (blocked_dirs); cd into a project subdirectory", project)
		}
	}
	return nil
}

// resolveDir expands a leading ~ then resolves symlinks so a blocked entry can be
// compared against the physical cwd (which prepareBox has already resolved).
func resolveDir(p, home string) string {
	switch {
	case p == "~":
		p = home
	case strings.HasPrefix(p, "~/"):
		p = filepath.Join(home, p[2:])
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// readConfigFile parses one TOML config file; a missing file yields (nil, nil).
func readConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &c, nil
}
