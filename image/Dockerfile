FROM debian:trixie-slim

# Baked-in CLI tooling. Installed at build time as root, so the box ships fully
# equipped and never needs sudo at runtime. nftables is for the entrypoint's
# egress firewall; ca-certificates for the https installers below.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    curl \
    jq \
    less \
    procps \
    ncurses-term \
    python3 \
    python3-venv \
    ripgrep \
    fd-find \
    zip \
    unzip \
    nftables \
    && ln -s "$(command -v fdfind)" /usr/local/bin/fd \
    && rm -rf /var/lib/apt/lists/*

# GitHub CLI from its official apt repo — arch-aware, so arm64 (Apple container) and x86_64 both work.
RUN mkdir -p -m 755 /etc/apt/keyrings \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# uv (Python) and mise (runtime manager) installed system-wide so the
# unprivileged dev user gets them on PATH — the sudo-free path to more tools.
RUN curl -LsSf https://astral.sh/uv/install.sh \
    | env UV_INSTALL_DIR=/usr/local/bin UV_NO_MODIFY_PATH=1 sh
RUN curl -fsSL https://mise.run | env MISE_INSTALL_PATH=/usr/local/bin/mise sh

# Non-root user. UID/GID are build args to match the host user on Linux Docker;
# the base image has no uid-1000 user, so nothing needs freeing first.
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd -g "$USER_GID" dev 2>/dev/null || true; \
    useradd -u "$USER_UID" -g "$USER_GID" -ms /bin/bash dev

# Claude Code as the self-contained native binary (no Node runtime). Installed as
# dev into ~/.local, which the runtime bind-mount over ~/.claude does not shadow.
USER dev
RUN HOME=/home/dev curl -fsSL https://claude.ai/install.sh | bash
USER root

# dev's user-space bins: the claude launcher and mise's tool shims.
ENV PATH=/home/dev/.local/bin:/home/dev/.local/share/mise/shims:$PATH

# Entrypoint installs the egress firewall then drops to dev; needs procps for the lock check.
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Stay root: the entrypoint sets nftables rules, then drops to the dev user itself.
WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
