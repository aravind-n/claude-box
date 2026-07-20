// Command berm-proxy is the egress guard for berm: an HTTP CONNECT
// (and plain-HTTP) forward proxy that permits outbound connections only to an
// allowlisted set of domains. The guard logic lives in the egress package; this
// entrypoint only reads configuration from the environment and wires it up.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"berm/proxy/egress"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	allowPath := env("BERM_ALLOWLIST", "/etc/berm/allowlist")
	modePath := env("BERM_MODE_FILE", "/etc/berm/mode")
	listen := env("BERM_PROXY_LISTEN", ":8080")

	policy := egress.NewPolicy(allowPath, modePath)
	dialer := egress.SafeDialer{Timeout: 10 * time.Second}
	denyLog := egress.NewDenyLog(env("BERM_DENY_LOG", ""))
	proxy := egress.NewProxy(policy, dialer, denyLog)

	log.Printf("berm egress proxy on %s (allowlist=%s mode=%s)", listen, allowPath, modePath)
	srv := &http.Server{
		Addr:              listen,
		Handler:           proxy,
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
