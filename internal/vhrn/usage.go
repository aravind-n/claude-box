package vhrn

// usageText is the subcommand-first help. Bare `vhrn` prints it; a harness is run
// explicitly as `vhrn <harness> ...`.
const usageText = `vhrn runs coding agents in a container jailed to the current project, with
default-deny network egress.

Usage:
  vhrn install <harness>                  build images, seed egress, add a shell alias
  vhrn uninstall <harness>                remove the alias/registry entry (--image drops the image)
  vhrn <harness> [flags] [-- ] [args...]  run a harness in the box
  vhrn list                               show known and installed harnesses
  vhrn net <subcommand>                   manage the egress policy
  vhrn help                               show this help

Harnesses:
  claude                   Claude Code

Run flags (after the harness name, before the agent's own flags):
  --open-net               drop the egress guard for this run (all egress)
  --allow <domain>...      add allowlist domains (comma-separated or repeated)
  --                       stop reading flags; forward the rest to the agent

After ` + "`vhrn install claude`" + ` a shell alias lets you run ` + "`claude`" + ` directly; ` + "`command claude`" + `
or ` + "`\\claude`" + ` still reaches the real binary. Examples:
  vhrn claude --model opus         # forwards --model opus to claude
  vhrn claude --open-net           # drop the guard for this session
  vhrn claude -- --help            # the agent's own help, not this one

net subcommands:
  net status               current mode and allowlist size
  net allow <domain>...    add domains to the allowlist (effective now)
  net denied               domains blocked this session
  net open                 drop the guard (allow everything)
  net guard                re-enable enforcement
  net report               allow everything, but log what would be denied

Environment:
  VHRN_ENGINE        container engine (default: container, then docker)
  VHRN_IMAGE         box image name (default: per-harness, e.g. vhrn-claude)
  VHRN_PROXY_IMAGE   proxy image name (default: vhrn-proxy)
  VHRN_PROXY_PORT    proxy port (default: 8080)
`
