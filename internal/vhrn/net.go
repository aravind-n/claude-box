package vhrn

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// netPolicy locates the host-side egress policy files. They are mounted only into
// the proxy, never the box, so an in-box process can never widen its own egress.
type netPolicy struct {
	dir       string // <cache>/net
	allowlist string
	modeFile  string
	denyLog   string
}

func newNetPolicy(cache string) netPolicy {
	dir := filepath.Join(cache, "net")
	return netPolicy{
		dir:       dir,
		allowlist: filepath.Join(dir, "allowlist"),
		modeFile:  filepath.Join(dir, "mode"),
		denyLog:   filepath.Join(dir, "denied.log"),
	}
}

// ensure creates the policy dir world-writable so the proxy container (a different
// uid) can append to denied.log.
func (p netPolicy) ensure() error {
	if err := os.MkdirAll(p.dir, 0o777); err != nil {
		return err
	}
	os.Chmod(p.dir, 0o777)
	return nil
}

const defaultAllowlist = `# vhrn egress allowlist — one domain per line, matching the domain and its
# subdomains. Edit freely, or run ` + "`vhrn net allow <domain>`" + ` while a box runs.
api.anthropic.com
claude.ai
platform.claude.com
statsig.anthropic.com
sentry.io
github.com
githubusercontent.com
registry.npmjs.org
pypi.org
files.pythonhosted.org
astral.sh
mise.jdx.dev
`

// seedAllowlistIfAbsent writes the default allowlist on first run; it never
// clobbers later edits.
func (p netPolicy) seedAllowlistIfAbsent() {
	if fileExists(p.allowlist) {
		return
	}
	os.WriteFile(p.allowlist, []byte(defaultAllowlist), 0o644)
}

// lines returns the current allowlist file contents, one entry per line.
func (p netPolicy) lines() []string {
	f, err := os.Open(p.allowlist)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		out = append(out, s.Text())
	}
	return out
}

// countDomains counts non-comment, non-blank allowlist entries (grep -cvE
// '^[[:space:]]*(#|$)').
func (p netPolicy) countDomains() int {
	n := 0
	for _, line := range p.lines() {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		n++
	}
	return n
}

// appendMissing appends domains not already present (exact line match), mirroring
// the run path's --allow handling (grep -qxF || printf >>). Non-atomic, as in the
// wrapper's run path.
func (p netPolicy) appendMissing(domains []string) {
	if len(domains) == 0 {
		return
	}
	set := map[string]bool{}
	for _, l := range p.lines() {
		set[l] = true
	}
	f, err := os.OpenFile(p.allowlist, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	for _, d := range domains {
		if !set[d] {
			fmt.Fprintln(f, d)
			set[d] = true
		}
	}
}

// appendMissingAtomic is `net allow`: it writes the updated allowlist to a same-dir
// temp file and renames it into place, so the proxy (reading concurrently) never
// sees a torn file.
func (p netPolicy) appendMissingAtomic(domains []string) error {
	var buf bytes.Buffer
	set := map[string]bool{}
	if data, err := os.ReadFile(p.allowlist); err == nil {
		buf.Write(data)
		for _, l := range strings.Split(string(data), "\n") {
			set[l] = true
		}
		if len(data) > 0 && data[len(data)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}
	for _, d := range domains {
		if !set[d] {
			buf.WriteString(d + "\n")
			set[d] = true
		}
	}
	tmp, err := os.CreateTemp(p.dir, "allowlist.*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	tmp.Close()
	os.Chmod(name, 0o666)
	return os.Rename(name, p.allowlist) // atomic on the same fs; proxy re-reads
}

func (p netPolicy) writeMode(mode string) {
	os.WriteFile(p.modeFile, []byte(mode+"\n"), 0o644)
}

func (p netPolicy) truncateDenyLog() {
	os.WriteFile(p.denyLog, []byte{}, 0o666)
	os.Chmod(p.denyLog, 0o666)
}

// deniedDomains returns the unique, sorted set of domains from the deny log's
// second field (awk '{print $2}' | sort -u).
func (p netPolicy) deniedDomains() []string {
	data, err := os.ReadFile(p.denyLog)
	if err != nil || len(data) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if d := fields[1]; !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

// runNet handles `vhrn net <subcommand>`: it mutates the host-side egress policy
// the running box reads. This is the only path to that policy — the box has none.
func runNet(args []string) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
		return 1
	}
	np := newNetPolicy(vhrnCache(home))
	os.MkdirAll(np.dir, 0o755)

	cmd := "status"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	switch cmd {
	case "status":
		mode := "enforce"
		if b, err := os.ReadFile(np.modeFile); err == nil {
			mode = strings.TrimSpace(string(b))
		}
		fmt.Printf("mode:    %s\n", mode)
		fmt.Printf("allowed: %d domain(s) (%s)\n", np.countDomains(), np.allowlist)
	case "denied":
		domains := np.deniedDomains()
		if len(domains) == 0 {
			fmt.Println("no denials recorded this session")
			return 0
		}
		for _, d := range domains {
			fmt.Println(d)
		}
	case "allow":
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "usage: vhrn net allow <domain>...")
			return 2
		}
		if err := np.appendMissingAtomic(args); err != nil {
			fmt.Fprintf(os.Stderr, "vhrn: %v\n", err)
			return 1
		}
		fmt.Printf("allowed: %s\n", strings.Join(args, " "))
	case "open":
		np.writeMode("open")
		fmt.Println("egress guard OFF (open) — all public hosts allowed")
	case "guard":
		np.writeMode("enforce")
		fmt.Println("egress guard ON (enforce) — allowlist enforced")
	case "report":
		np.writeMode("report")
		fmt.Println("egress guard REPORT — all allowed, denials logged")
	default:
		fmt.Fprintln(os.Stderr, "usage: vhrn net {status|denied|allow <domain>...|open|guard|report}")
		return 2
	}
	return 0
}
