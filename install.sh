#!/usr/bin/env bash
set -euo pipefail

# Install the claude-box wrapper into /usr/local/bin (needs sudo).
sudo install -m 0755 "$(dirname "$0")/claude-box.sh" /usr/local/bin/claude-box
echo "Installed /usr/local/bin/claude-box — run 'claude-box' in any project."
