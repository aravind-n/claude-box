package vhrn

// usageText is ported verbatim from vhrn.sh's usage() so `vhrn help` is unchanged.
const usageText = `vhrn runs Claude Code in a container jailed to the current project, with
default-deny network egress.

Usage:
  vhrn [flags] [claude args...]    run claude inside the box
  vhrn net <subcommand>            manage the egress policy
  vhrn help                        show this help

Flags (must come before claude's own flags):
  --open-net               drop the egress guard for this run (all egress)
  --allow <domain>...      add allowlist domains (comma-separated or repeated)
  --                       stop reading flags; forward the rest to claude

Anything not matched above is forwarded to claude untouched. Use ` + "`--`" + ` to pass a
flag the wrapper would otherwise read:
  vhrn --model opus
  vhrn -- --help     # claude's own help, not this one

net subcommands:
  net status               current mode and allowlist size
  net allow <domain>...    add domains to the allowlist (effective now)
  net denied               domains blocked this session
  net open                 drop the guard (allow everything)
  net guard                re-enable enforcement
  net report               allow everything, but log what would be denied

Environment:
  VHRN_ENGINE        container engine (default: container, then docker)
  VHRN_IMAGE         box image name (default: vhrn-sandbox)
  VHRN_PROXY_IMAGE   proxy image name (default: vhrn-proxy)
  VHRN_PROXY_PORT    proxy port (default: 8080)
`
