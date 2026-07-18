#!/usr/bin/env bash
set -euo pipefail

# Run Claude Code in dangerous mode, jailed to the current project.

REAL_CLAUDE="$HOME/.claude"
SANDBOX="$HOME/.claude-sandbox"
SANDBOX_JSON="$HOME/.claude-sandbox.json"
PROJECT="$(pwd -P)"

# Reproduce Claude's projects/<key> encoding so history unifies with native.
KEY="$(printf '%s' "$PROJECT" | sed 's/[^A-Za-z0-9]/-/g')"
HISTORY="$REAL_CLAUDE/projects/$KEY"

mkdir -p "$SANDBOX" "$HISTORY"

# Copy globals in, dereferencing symlinks so symlinked skills come across.
copy_dir() {
  [ -d "$REAL_CLAUDE/$1" ] || return 0
  if command -v rsync >/dev/null 2>&1; then
    rsync -aL --delete "$REAL_CLAUDE/$1/" "$SANDBOX/$1/" \
      || echo "claude-box: warning: some '$1' entries were skipped (broken symlink?)" >&2
  else
    rm -rf "${SANDBOX:?}/$1"
    cp -RL "$REAL_CLAUDE/$1" "$SANDBOX/$1" \
      || echo "claude-box: warning: some '$1' entries were skipped (broken symlink?)" >&2
  fi
}
copy_file() {
  [ -f "$REAL_CLAUDE/$1" ] || return 0
  cp -L "$REAL_CLAUDE/$1" "$SANDBOX/$1" 2>/dev/null \
    || echo "claude-box: warning: could not copy '$1'" >&2
}
for d in skills commands agents; do copy_dir "$d"; done
for f in settings.json CLAUDE.md;  do copy_file "$f"; done

# Copy the config file
if [ -f "$REAL_CLAUDE.json" ]; then
  cp -L "$REAL_CLAUDE.json" "$SANDBOX_JSON" 2>/dev/null \
    || echo "claude-box: warning: could not copy .claude.json" >&2
fi
[ -s "$SANDBOX_JSON" ] || printf '{}\n' > "$SANDBOX_JSON"

container system start >/dev/null 2>&1 || true

exec container run -it --rm \
  --volume "$PROJECT:$PROJECT" \
  --workdir "$PROJECT" \
  --volume "$SANDBOX:/home/dev/.claude" \
  --volume "$SANDBOX_JSON:/home/dev/.claude.json" \
  --volume "$HISTORY:/home/dev/.claude/projects/$KEY" \
  --env TERM="${TERM:-xterm-256color}" \
  --env COLORTERM=truecolor \
  --env FORCE_COLOR=3 \
  claude-sandbox \
  claude "$@"
