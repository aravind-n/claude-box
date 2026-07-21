package vhrn

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// proxy is a running egress-proxy sidecar. The box's in-container firewall pins all
// egress to it; policy files live host-side and are mounted only into this sidecar.
type proxy struct {
	engine string
	name   string
}

// startProxy launches the detached proxy sidecar and resolves its IP (engines
// differ; retry until it has one).
func startProxy(engine, image string, np netPolicy, port string) (*proxy, string, error) {
	name := fmt.Sprintf("vhrn-proxy-%d", os.Getpid())
	cmd := exec.Command(engine, "run", "-d", "--rm", "--name", name,
		"--volume", np.dir+":/etc/vhrn",
		"--env", "VHRN_ALLOWLIST=/etc/vhrn/allowlist",
		"--env", "VHRN_MODE_FILE=/etc/vhrn/mode",
		"--env", "VHRN_DENY_LOG=/etc/vhrn/denied.log",
		"--env", "VHRN_PROXY_LISTEN=:"+port,
		image,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("proxy failed to start (is the %q image built?)", image)
	}
	p := &proxy{engine: engine, name: name}

	var ip string
	for i := 0; i < 30; i++ {
		if ip = p.inspectIP(); ip != "" {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if ip == "" {
		p.stop()
		return nil, "", fmt.Errorf("proxy failed to start (is the %q image built?)", image)
	}
	return p, ip, nil
}

func (p *proxy) stop() {
	if p == nil {
		return
	}
	exec.Command(p.engine, "stop", p.name).Run()
}

func (p *proxy) inspectIP() string {
	if p.engine == "docker" {
		out, err := exec.Command("docker", "inspect", "-f",
			"{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", p.name).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	out, err := exec.Command("container", "inspect", p.name).Output()
	if err != nil {
		return ""
	}
	return firstIPv4(string(out))
}

var ipv4Re = regexp.MustCompile(`([0-9]{1,3}\.){3}[0-9]{1,3}`)

// firstIPv4 mirrors `grep -m1 ipv4Address | grep -oE <dotted quad>`: the first
// dotted quad on the first line mentioning ipv4Address. Apple's inspect JSON
// escapes the CIDR slash (192.168.64.73\/24), so we match only the quad.
func firstIPv4(inspectOutput string) string {
	for _, line := range strings.Split(inspectOutput, "\n") {
		if strings.Contains(line, "ipv4Address") {
			return ipv4Re.FindString(line)
		}
	}
	return ""
}

// stopOnSignal keeps the sidecar from leaking if vhrn is signaled. SIGINT is left
// to the interactive child (the agent) — the parent stays alive to wait and clean
// up on exit; SIGTERM tears down the sidecar and exits.
func stopOnSignal(p *proxy) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range ch {
			if sig == syscall.SIGTERM {
				p.stop()
				os.Exit(1)
			}
		}
	}()
}
