package vhrn

import "testing"

func TestFirstIPv4(t *testing.T) {
	// Apple container inspect escapes the CIDR slash; only the dotted quad matters.
	apple := `{
  "networks": [
    { "ipv4Address": "192.168.64.73\/24", "gateway": "192.168.64.1" }
  ]
}`
	if got := firstIPv4(apple); got != "192.168.64.73" {
		t.Errorf("firstIPv4(apple) = %q, want %q", got, "192.168.64.73")
	}
	if got := firstIPv4("no address here\nsecond line"); got != "" {
		t.Errorf("firstIPv4(none) = %q, want empty", got)
	}
}
