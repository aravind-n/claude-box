package egress

import "testing"

func TestParseMode(t *testing.T) {
	cases := map[string]Mode{
		"enforce": ModeEnforce,
		"open":    ModeOpen,
		"report":  ModeReport,
		"OPEN\n":  ModeOpen,
		"":        ModeEnforce, // fail closed
		"garbage": ModeEnforce,
	}
	for in, want := range cases {
		if got := parseMode(in); got != want {
			t.Errorf("parseMode(%q) = %v, want %v", in, got, want)
		}
	}
}
