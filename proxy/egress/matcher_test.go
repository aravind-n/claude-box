package egress

import "testing"

func TestHostAllowed(t *testing.T) {
	allow := []string{"github.com", "api.anthropic.com", "githubusercontent.com"}
	cases := []struct {
		host string
		want bool
	}{
		{"github.com", true},
		{"api.github.com", true},
		{"raw.githubusercontent.com", true},
		{"api.anthropic.com", true},
		{"API.Anthropic.COM", true}, // case-insensitive
		{"github.com.", true},       // trailing FQDN dot
		// security-critical rejections
		{"github.com.attacker.net", false}, // suffix-append attack
		{"evilgithub.com", false},          // prefix-glue attack
		{"notgithub.com", false},
		{"anthropic.com", false},      // apex not implied by api.anthropic.com
		{"xapi.anthropic.com", false}, // label-boundary, not a real subdomain
		{"attacker.net", false},
		{"", false},
	}
	for _, c := range cases {
		if got := hostAllowed(c.host, allow); got != c.want {
			t.Errorf("hostAllowed(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}

func TestNormEntry(t *testing.T) {
	cases := map[string]string{
		"  GitHub.com ": "github.com",
		"*.example.com": "example.com",
		".example.com":  "example.com",
		"example.com.":  "example.com",
		"":              "",
	}
	for in, want := range cases {
		if got := normEntry(in); got != want {
			t.Errorf("normEntry(%q) = %q, want %q", in, got, want)
		}
	}
}
