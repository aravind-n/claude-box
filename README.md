# claude-box

Run Claude Code inside a container jailed to the current project directory. This
ensures claude can't break out of the project directory allowing you to freely
run features like `--dangerously-skip-permissions`

The current project is bind-mounted at its real path; a sandbox copy of `~/.claude` is synced in on every run; session history is written straight back to your host's `~/.claude/projects/<key>`, so in-box and native history stay unified.

## Requirements

- [Apple Container](https://github.com/apple/container) or Docker (auto-detected, `container` first)
- Claude Code, plus `gh` on the host if you want GitHub auth forwarded

## Install

```sh
make            # build the claude-sandbox image
./install.sh    # install the wrapper to /usr/local/bin (needs sudo)
```

## Usage

Run in any project directory. Arguments are passed straight through to `claude`:

```sh
claude-box
claude-box --model opus
```

## Make targets

| Target | Description |
| --- | --- |
| `make` / `make build` | Build the image (default) |
| `make rebuild` | Rebuild with no cache |
| `make clean` | Remove the image |
| `make ENGINE=docker …` | Force Docker instead of `container` |

## Notes

- `gh` auth is forwarded as an env token (`$GH_TOKEN`/`$GITHUB_TOKEN`, else `gh auth token`), so git-over-HTTPS works inside the box. SSH remotes stay unauthenticated.
- Edits to `~/.claude-sandbox` are wiped each run — change `~/.claude` instead.
- The "trust this folder?" prompt may reappear each run; This is a known issue. Open the folder once in native `claude` to persist trust. See `CLAUDE.md` for details.
