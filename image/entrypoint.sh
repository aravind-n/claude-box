#!/usr/bin/env bash
set -euo pipefail

# Remove a stale git index lock in the project (the workdir) left by a crash.
LOCK_FILE="$PWD/.git/index.lock"
if [ -f "$LOCK_FILE" ]; then
  if ! pgrep -x git >/dev/null 2>&1; then
    echo "[claude-box] Removing orphaned $LOCK_FILE from a previous crash." >&2
    rm -f "$LOCK_FILE"
  else
    echo "[claude-box] Warning: a git process is holding the index lock." >&2
  fi
fi

exec "$@"
