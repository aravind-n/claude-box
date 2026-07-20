package egress

import (
	"context"
	"fmt"
	"net"
	"time"
)

// SafeDialer dials only public unicast addresses. It resolves the destination
// and refuses to connect if any resolved address is non-public, closing off
// SSRF to loopback, the container host, and cloud metadata endpoints even for
// an allowlisted name; it then dials the validated IP directly, so a rebinding
// re-resolve cannot slip a private address in behind the check.
type SafeDialer struct {
	Timeout time.Duration
}

// Dial satisfies the Dialer interface and http.Transport.DialContext.
func (d SafeDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	dst, err := firstPublicIP(ips)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", host, err)
	}
	dialer := net.Dialer{Timeout: d.Timeout}
	return dialer.DialContext(ctx, network, net.JoinHostPort(dst.String(), port))
}

// firstPublicIP returns the address to dial from a resolved set. It requires
// every address to be public — a single private one fails the whole lookup, so
// DNS rebinding cannot pair a public record with a private one — and returns the
// first. This is the SSRF gate; see isPublicIP.
func firstPublicIP(ips []net.IP) (net.IP, error) {
	var dst net.IP
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return nil, fmt.Errorf("refusing non-public address %s", ip)
		}
		if dst == nil {
			dst = ip
		}
	}
	if dst == nil {
		return nil, fmt.Errorf("no address")
	}
	return dst, nil
}

// isPublicIP reports whether ip is a globally routable unicast address.
func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}
	// Go's IsPrivate covers RFC1918 and IPv6 ULA but not CGNAT 100.64.0.0/10.
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return false
	}
	return true
}
