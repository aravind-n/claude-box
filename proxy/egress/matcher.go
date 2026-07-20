package egress

import "strings"

// normEntry canonicalises an allowlist entry: lowercase, no surrounding space,
// no leading "*." or ".", no trailing dot.
func normEntry(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "*.")
	return strings.Trim(s, ".")
}

// normHost canonicalises a request host for comparison.
func normHost(h string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(h)), ".")
}

// hostAllowed reports whether host matches any entry. Entry "example.com"
// matches "example.com" and any subdomain "*.example.com" — and, critically,
// neither "evilexample.com" nor "example.com.attacker.net". This label-anchored
// match is the security-critical core of the proxy.
func hostAllowed(host string, allow []string) bool {
	h := normHost(host)
	if h == "" {
		return false
	}
	for _, e := range allow {
		if h == e || strings.HasSuffix(h, "."+e) {
			return true
		}
	}
	return false
}
