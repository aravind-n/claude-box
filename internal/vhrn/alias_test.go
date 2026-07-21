package vhrn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAliasLine(t *testing.T) {
	if got := aliasLine("zsh", "claude", "vhrn claude"); got != "alias claude='vhrn claude'" {
		t.Errorf("zsh alias = %q", got)
	}
	if got := aliasLine("fish", "claude", "vhrn claude"); got != "alias claude 'vhrn claude'" {
		t.Errorf("fish alias = %q", got)
	}
}

func TestAliasBlock(t *testing.T) {
	if aliasBlock(nil, "zsh") != "" {
		t.Error("empty harness set should yield no block")
	}
	b := aliasBlock([]Harness{{Name: "claude", Alias: "claude"}}, "bash")
	if !strings.Contains(b, aliasStart) || !strings.Contains(b, aliasEnd) {
		t.Error("block missing markers")
	}
	if !strings.Contains(b, "alias claude='vhrn claude'") {
		t.Errorf("block missing alias line:\n%s", b)
	}
}

func TestWriteAliasBlockRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	os.WriteFile(path, []byte("export FOO=1\n"), 0o644)

	block := aliasBlock([]Harness{{Name: "claude", Alias: "claude"}}, "zsh")
	if err := writeAliasBlock(path, block); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(after), "export FOO=1\n") {
		t.Errorf("surrounding content not preserved:\n%s", after)
	}
	if !strings.Contains(string(after), "alias claude='vhrn claude'") {
		t.Error("alias not written")
	}

	// Regenerating with the same block must not duplicate it.
	if err := writeAliasBlock(path, block); err != nil {
		t.Fatal(err)
	}
	regen, _ := os.ReadFile(path)
	if strings.Count(string(regen), aliasStart) != 1 {
		t.Errorf("block duplicated on regenerate:\n%s", regen)
	}

	// Empty block removes it and restores the original content exactly.
	if err := writeAliasBlock(path, ""); err != nil {
		t.Fatal(err)
	}
	cleared, _ := os.ReadFile(path)
	if string(cleared) != "export FOO=1\n" {
		t.Errorf("block not cleanly removed: %q", cleared)
	}
}

func TestWriteAliasBlockNoSpuriousFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	if err := writeAliasBlock(path, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("removing a block from an absent file should not create it")
	}
}

func TestSyncAliases(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	t.Setenv("SHELL", "/bin/zsh")

	bashrc := filepath.Join(home, ".bashrc")
	os.WriteFile(bashrc, []byte("# bash\n"), 0o644) // exists -> managed
	zshrc := filepath.Join(home, ".zshrc")          // current shell -> created
	fishrc := filepath.Join(home, ".config", "fish", "config.fish")

	if err := addInstalled(home, "claude", "latest"); err != nil {
		t.Fatal(err)
	}
	if err := syncAliases(home); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{bashrc, zshrc} {
		data, err := os.ReadFile(p)
		if err != nil || !strings.Contains(string(data), "alias claude=") {
			t.Errorf("%s should carry the alias (err=%v)", p, err)
		}
	}
	if _, err := os.Stat(fishrc); !os.IsNotExist(err) {
		t.Error("fish rc is neither existing nor the current shell; should be left alone")
	}

	// Uninstalling clears the blocks.
	if err := removeInstalled(home, "claude"); err != nil {
		t.Fatal(err)
	}
	if err := syncAliases(home); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(zshrc); strings.Contains(string(data), aliasStart) {
		t.Error("alias block should be gone after uninstall")
	}
}
