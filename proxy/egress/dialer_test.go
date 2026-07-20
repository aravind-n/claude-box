package egress

import (
	"net"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"2606:4700:4700::1111", true},
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false}, // cloud metadata endpoint
		{"100.64.0.1", false},      // CGNAT
		{"::1", false},
		{"fc00::1", false}, // IPv6 ULA
		{"0.0.0.0", false},
	}
	for _, c := range cases {
		if got := isPublicIP(net.ParseIP(c.ip)); got != c.want {
			t.Errorf("isPublicIP(%s) = %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestFirstPublicIP(t *testing.T) {
	pub := net.ParseIP("8.8.8.8")
	pub2 := net.ParseIP("1.1.1.1")
	priv := net.ParseIP("10.0.0.1")
	meta := net.ParseIP("169.254.169.254")

	if got, err := firstPublicIP([]net.IP{pub, pub2}); err != nil || !got.Equal(pub) {
		t.Errorf("all-public: got %v, %v; want %v, nil", got, err, pub)
	}
	// A single private address must fail the whole set (DNS-rebinding guard).
	if _, err := firstPublicIP([]net.IP{pub, priv}); err == nil {
		t.Error("public+private accepted; want rejection")
	}
	if _, err := firstPublicIP([]net.IP{meta}); err == nil {
		t.Error("metadata address accepted; want rejection")
	}
	if _, err := firstPublicIP(nil); err == nil {
		t.Error("empty set accepted; want error")
	}
}
