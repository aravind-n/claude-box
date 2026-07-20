// Command claude-box-proxy is the egress guard for claude-box: an HTTP CONNECT
// (and plain-HTTP) forward proxy that permits outbound connections only to an
// allowlisted set of domains. The guard logic lives in the egress package; this
// entrypoint only reads configuration from the environment and wires it up.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"claude-box/proxy/egress"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	allowPath := env("CLAUDE_BOX_ALLOWLIST", "/etc/claude-box/allowlist")
	modePath := env("CLAUDE_BOX_MODE_FILE", "/etc/claude-box/mode")
	listen := env("CLAUDE_BOX_PROXY_LISTEN", ":8080")

	policy := egress.NewPolicy(allowPath, modePath)
	dialer := egress.SafeDialer{Timeout: 10 * time.Second}
	denyLog := egress.NewDenyLog(env("CLAUDE_BOX_DENY_LOG", ""))
	proxy := egress.NewProxy(policy, dialer, denyLog)

	log.Printf("claude-box egress proxy on %s (allowlist=%s mode=%s)", listen, allowPath, modePath)
	srv := &http.Server{
		Addr:              listen,
		Handler:           proxy,
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
