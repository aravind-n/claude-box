package egress

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// --- fakes: the Proxy depends only on interfaces, so its behaviour is testable
// without real policy files, DNS, or sockets. ------------------------------

type fakeChecker struct {
	verdict Verdict
	mode    Mode
}

func (f fakeChecker) Check(string) Verdict { return f.verdict }
func (f fakeChecker) Mode() Mode           { return f.mode }

// fakeDialer ignores the requested address and connects to a fixed target,
// standing in for the SSRF-safe dialer without touching real DNS.
type fakeDialer struct{ target string }

func (f fakeDialer) Dial(ctx context.Context, network, _ string) (net.Conn, error) {
	if f.target == "" {
		return nil, fmt.Errorf("no target")
	}
	var d net.Dialer
	return d.DialContext(ctx, network, f.target)
}

type fakeRecorder struct {
	mu    sync.Mutex
	hosts []string
}

func (f *fakeRecorder) Record(host string, _ Mode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hosts = append(f.hosts, host)
}

func (f *fakeRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.hosts)
}

func TestProxyStatusReportsMode(t *testing.T) {
	px := NewProxy(fakeChecker{mode: ModeOpen}, fakeDialer{}, &fakeRecorder{})
	rr := httptest.NewRecorder()
	px.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/__status", nil))
	if !strings.Contains(rr.Body.String(), `"mode":"open"`) {
		t.Errorf("status body = %q, want mode open", rr.Body.String())
	}
}

func TestProxyHTTPDeniedIsForbiddenAndLogged(t *testing.T) {
	rec := &fakeRecorder{}
	px := NewProxy(
		fakeChecker{verdict: Verdict{Allow: false, Logged: true, Mode: ModeEnforce}},
		fakeDialer{}, rec)
	rr := httptest.NewRecorder()
	px.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://blocked.example/x", nil))
	if rr.Code != http.StatusForbidden {
		t.Errorf("code = %d, want 403", rr.Code)
	}
	if rec.count() != 1 {
		t.Errorf("recorded %d denials, want 1", rec.count())
	}
}

func TestProxyConnectAllowedTunnels(t *testing.T) {
	echo := echoServer(t)
	px := NewProxy(fakeChecker{verdict: Verdict{Allow: true}}, fakeDialer{target: echo}, &fakeRecorder{})
	srv := httptest.NewServer(px)
	defer srv.Close()

	conn := dialProxy(t, srv.URL)
	defer conn.Close()
	fmt.Fprint(conn, "CONNECT anything.com:443 HTTP/1.1\r\nHost: anything.com:443\r\n\r\n")
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil || !strings.Contains(status, "200") {
		t.Fatalf("CONNECT status = %q, err = %v", status, err)
	}
	skipHeaders(br)

	// The tunnel now reaches the echo server through the fake dialer.
	fmt.Fprint(conn, "ping")
	buf := make([]byte, 4)
	if _, err := io.ReadFull(br, buf); err != nil {
		t.Fatalf("tunnel read: %v", err)
	}
	if string(buf) != "ping" {
		t.Errorf("tunnel echo = %q, want ping", string(buf))
	}
}

func TestProxyConnectDeniedIsForbiddenAndLogged(t *testing.T) {
	rec := &fakeRecorder{}
	px := NewProxy(
		fakeChecker{verdict: Verdict{Allow: false, Logged: true, Mode: ModeEnforce}},
		fakeDialer{}, rec)
	srv := httptest.NewServer(px)
	defer srv.Close()

	conn := dialProxy(t, srv.URL)
	defer conn.Close()
	fmt.Fprint(conn, "CONNECT blocked.com:443 HTTP/1.1\r\nHost: blocked.com:443\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || !strings.Contains(status, "403") {
		t.Fatalf("CONNECT status = %q, err = %v", status, err)
	}
	if rec.count() != 1 {
		t.Errorf("recorded %d denials, want 1", rec.count())
	}
}

// --- helpers ---------------------------------------------------------------

func echoServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) { _, _ = io.Copy(c, c); c.Close() }(c)
		}
	}()
	return ln.Addr().String()
}

func dialProxy(t *testing.T, url string) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", strings.TrimPrefix(url, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func skipHeaders(br *bufio.Reader) {
	for {
		line, err := br.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" {
			return
		}
	}
}
