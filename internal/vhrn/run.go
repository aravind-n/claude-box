package vhrn

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// boxConfig is the resolved host-side state for one run: paths, engine/image, and
// the extra --volume/--env args assembled during preparation.
type boxConfig struct {
	engine      string
	image       string
	project     string // physical cwd (pwd -P)
	key         string // history key: [^A-Za-z0-9] -> '-'
	cache       string // ~/.cache/vhrn
	sandbox     string // <cache>/sandbox      -> /home/dev/.claude
	sandboxJSON string // <cache>/sandbox.json -> /home/dev/.claude.json
	realClaude  string // ~/.claude
	history     string // ~/.claude/projects/<key>

	gitMount []string
	ghEnv    []string
	termEnv  []string
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]`)

// historyKey reproduces Claude's projects/<key> encoding so in-box history
// unifies with native history (sed 's/[^A-Za-z0-9]/-/g').
func historyKey(project string) string {
	return nonAlnum.ReplaceAllString(project, "-")
}

// prepareBox performs all host-side preparation: resolve paths and engine, sync a
// sandbox copy of ~/.claude, and assemble the gitconfig/gh/terminal run args.
func prepareBox() (*boxConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	project, err := filepath.EvalSymlinks(wd) // pwd -P: physical path
	if err != nil {
		return nil, err
	}
	engine, err := detectEngine()
	if err != nil {
		return nil, err
	}

	cache := vhrnCache(home)
	realClaude := filepath.Join(home, ".claude")

	cfg := &boxConfig{
		engine:      engine,
		image:       envOr("VHRN_IMAGE", "vhrn-sandbox"),
		project:     project,
		key:         historyKey(project),
		cache:       cache,
		sandbox:     filepath.Join(cache, "sandbox"),
		sandboxJSON: filepath.Join(cache, "sandbox.json"),
		realClaude:  realClaude,
	}
	cfg.history = filepath.Join(realClaude, "projects", cfg.key)

	if err := os.MkdirAll(cfg.sandbox, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.history, 0o755); err != nil {
		return nil, err
	}

	// Copy globals in, dereferencing symlinks so symlinked skills come across.
	for _, d := range []string{"skills", "commands", "agents"} {
		syncClaudeSubdir(realClaude, cfg.sandbox, d)
	}
	for _, fn := range []string{"settings.json", "statusline.sh"} {
		copyFileInto(realClaude, cfg.sandbox, fn)
	}
	copyClaudeJSON(realClaude+".json", cfg.sandboxJSON)

	cfg.gitMount = gitConfigMount(home, cache)
	cfg.ghEnv = ghTokenEnv()
	cfg.termEnv = terminalEnv()
	return cfg, nil
}

func runAgent(f runFlags) error {
	cfg, err := prepareBox()
	if err != nil {
		return err
	}
	return startBox(cfg, f)
}

// startBox seeds the egress policy, starts the proxy sidecar, then runs the jailed
// box with all egress pinned to the proxy. The box run inherits the terminal so the
// agent is interactive; its exit status is returned verbatim.
func startBox(cfg *boxConfig, f runFlags) error {
	np := newNetPolicy(cfg.cache)
	port := envOr("VHRN_PROXY_PORT", "8080")
	mode := "enforce"
	if f.openNet {
		mode = "open"
	}

	if err := np.ensure(); err != nil {
		return err
	}
	np.seedAllowlistIfAbsent()
	np.appendMissing(f.extraAllow) // session --allow additions persist, like `net allow`
	np.writeMode(mode)
	np.truncateDenyLog()

	if err := writeBoxGuide(cfg.realClaude, cfg.sandbox, f.openNet); err != nil {
		fmt.Fprintf(os.Stderr, "vhrn: warning: could not write box CLAUDE.md: %v\n", err)
	}

	// Apple container needs its system service up; Docker manages its own daemon.
	if cfg.engine == "container" {
		exec.Command("container", "system", "start").Run()
	}

	p, ip, err := startProxy(cfg.engine, np, port)
	if err != nil {
		return err
	}
	defer p.stop()
	stopOnSignal(p)

	proxyURL := fmt.Sprintf("http://%s:%s", ip, port)
	if mode == "open" {
		fmt.Fprintln(os.Stderr, "vhrn: network guard OFF (open) — all public egress allowed this session.")
		if len(cfg.ghEnv) > 0 {
			fmt.Fprintln(os.Stderr, "vhrn: a GitHub token is present in the box with the guard off.")
		}
	}

	// NET_ADMIN lets the entrypoint install the egress firewall (dropped before dev runs).
	args := []string{
		"run", "-it", "--rm",
		"--cap-add", "CAP_NET_ADMIN",
		"--env", "VHRN_SANDBOX=1",
		"--env", "VHRN_NET=" + mode,
		"--env", "VHRN_PROXY_IP=" + ip,
		"--env", "VHRN_PROXY_PORT=" + port,
		"--env", "HTTP_PROXY=" + proxyURL,
		"--env", "HTTPS_PROXY=" + proxyURL,
		"--env", "http_proxy=" + proxyURL,
		"--env", "https_proxy=" + proxyURL,
		"--volume", cfg.project + ":" + cfg.project,
		"--workdir", cfg.project,
		"--volume", cfg.sandbox + ":/home/dev/.claude",
		"--volume", cfg.sandboxJSON + ":/home/dev/.claude.json",
		"--volume", cfg.history + ":/home/dev/.claude/projects/" + cfg.key,
	}
	args = append(args, cfg.gitMount...)
	args = append(args, cfg.termEnv...)
	args = append(args, cfg.ghEnv...)
	args = append(args, cfg.image, "claude")
	args = append(args, f.rest...)

	cmd := exec.Command(cfg.engine, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// vhrnCache is the XDG cache root for vhrn (${XDG_CACHE_HOME:-~/.cache}/vhrn).
func vhrnCache(home string) string {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		cacheHome = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheHome, "vhrn")
}
