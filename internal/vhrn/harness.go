package vhrn

import "sort"

// Harness describes one coding agent vhrn can run in the box. It is the single
// source of truth a subcommand, install, run, and persistence all read from, so
// adding an agent (codex, aider, ...) is a spec here plus a thin `FROM vhrn-base`
// Dockerfile — not a fork of the CLI.
type Harness struct {
	Name    string // registry key and subcommand, e.g. "claude"
	Image   string // box image built for it, e.g. "vhrn-claude"
	Command string // in-box argv[0], e.g. "claude"
	Alias   string // shell alias installed for it (`alias <Alias>='vhrn <Name>'`)

	// Default egress domains unioned into the host allowlist at install time.
	AllowDomains []string

	// Persistence — the three home-dir buckets (see state.go):
	//   - box-owned state:  StateDir mounts the persistent state/<Name>/ over it,
	//     with ConfigDirEnv pointing the agent's config dir there so its config
	//     JSON + refreshed credentials persist box→box and are rename-safe.
	//   - forwarded creds:  Credentials are copied from the host HostConfig dir
	//     into the state store once, only when the store is empty (bootstrap).
	//   - synced config:    SyncDirs/SyncFiles are the disposable host→box copies
	//     layered back on top as nested mounts each run.
	StateDir     string   // box-home-relative persistent dir, e.g. ".claude"
	ConfigDirEnv string   // env var pointing the agent's config dir at StateDir
	HostConfig   string   // host-home-relative dir to sync/bootstrap FROM, e.g. ".claude"
	SyncDirs     []string // disposable synced subdirs, e.g. skills/commands/agents
	SyncFiles    []string // disposable synced files, e.g. settings.json/statusline.sh
	Credentials  []string // StateDir-relative bootstrap-only files, e.g. .credentials.json
	ConfigJSON   string   // StateDir-relative config file holding login/onboarding/trust (claude: .claude.json)
	SeedTrust    bool     // pre-seed onboarding + per-project trust into ConfigJSON — the sandbox is the trust boundary
}

// harnesses is the built-in registry. Only claude exists today; the struct shape
// is what a codex/aider spec would fill in.
var harnesses = map[string]Harness{
	"claude": {
		Name:    "claude",
		Image:   "vhrn-claude",
		Command: "claude",
		Alias:   "claude",
		AllowDomains: []string{
			"api.anthropic.com",
			"claude.ai",
			"platform.claude.com",
			"statsig.anthropic.com",
			"sentry.io",
		},
		StateDir:     ".claude",
		ConfigDirEnv: "CLAUDE_CONFIG_DIR",
		HostConfig:   ".claude",
		SyncDirs:     []string{"skills", "commands", "agents"},
		SyncFiles:    []string{"settings.json", "statusline.sh"},
		Credentials:  []string{".credentials.json"},
		ConfigJSON:   ".claude.json",
		SeedTrust:    true,
	},
}

// lookupHarness returns the spec for name and whether it is a known harness.
func lookupHarness(name string) (Harness, bool) {
	h, ok := harnesses[name]
	return h, ok
}

// harnessNames returns the known harness names, sorted for stable output.
func harnessNames() []string {
	names := make([]string, 0, len(harnesses))
	for n := range harnesses {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
