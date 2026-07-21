package vhrn

import (
	"os"
	"reflect"
	"testing"
)

func TestAllowlistSeedCountAndAtomicAdd(t *testing.T) {
	np := newNetPolicy(t.TempDir())
	if err := np.ensure(); err != nil {
		t.Fatal(err)
	}
	np.seedAllowlistIfAbsent()
	base := np.countDomains()
	if base != 12 {
		t.Fatalf("default domain count = %d, want 12", base)
	}
	// Adds a new domain; ignores duplicates (incl. one already present).
	if err := np.appendMissingAtomic([]string{"docs.rs", "api.anthropic.com", "docs.rs"}); err != nil {
		t.Fatal(err)
	}
	if got := np.countDomains(); got != base+1 {
		t.Errorf("count after add = %d, want %d", got, base+1)
	}
	// Idempotent re-add.
	if err := np.appendMissingAtomic([]string{"docs.rs"}); err != nil {
		t.Fatal(err)
	}
	if got := np.countDomains(); got != base+1 {
		t.Errorf("count after re-add = %d, want %d", got, base+1)
	}
}

func TestResolveMode(t *testing.T) {
	cases := []struct {
		cfg  string
		open bool
		want string
	}{
		{"", false, "enforce"},
		{"enforce", false, "enforce"},
		{"report", false, "report"},
		{"open", false, "open"},
		{"bogus", false, "enforce"},
		{"report", true, "open"}, // --open-net wins over config
		{"", true, "open"},
	}
	for _, c := range cases {
		if got := resolveMode(c.cfg, c.open); got != c.want {
			t.Errorf("resolveMode(%q, %v) = %q, want %q", c.cfg, c.open, got, c.want)
		}
	}
}

func TestDeniedDomains(t *testing.T) {
	np := newNetPolicy(t.TempDir())
	if err := np.ensure(); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(np.denyLog, []byte("t1 evil.com GET\nt2 evil.com GET\nt3 tracker.io POST\n"), 0o644)
	got := np.deniedDomains()
	want := []string{"evil.com", "tracker.io"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("deniedDomains = %v, want %v", got, want)
	}

	empty := newNetPolicy(t.TempDir())
	empty.ensure()
	if got := empty.deniedDomains(); got != nil {
		t.Errorf("empty deniedDomains = %v, want nil", got)
	}
}
