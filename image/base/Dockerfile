# vhrn-base: the shared container image — debian + CLI tooling + the tamper-proof egress
# entrypoint, but no agent. Each harness image is `FROM vhrn-base` plus its own CLI,
# so the heavy, slow-changing layers are built and cached once.
FROM debian:trixie-slim

# Base packages: the entrypoint's egress firewall (nftables) and lock cleanup (procps),
# terminal support, and the file/search CLIs an agent reaches for.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    nftables \
    procps \
    ncurses-term \
    less \
    jq \
    ripgrep \
    fd-find \
    zip \
    unzip \
    && ln -s "$(command -v fdfind)" /usr/local/bin/fd \
    && rm -rf /var/lib/apt/lists/*

# GitHub CLI from its own apt repo — arch-aware, so arm64 and x86_64 both work.
RUN mkdir -p -m 755 /etc/apt/keyrings \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Basic C toolchain — cc + libc headers, so native builds work.
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    lld \
    llvm \
    libc6-dev \
    make \
    pkg-config \
    && ln -sf "$(command -v clang)" /usr/local/bin/cc \
    && ln -sf "$(command -v clang++)" /usr/local/bin/c++ \
    && rm -rf /var/lib/apt/lists/*

# Python: the system interpreter plus uv, its package/venv manager.
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-venv \
    && rm -rf /var/lib/apt/lists/* \
    && curl -LsSf https://astral.sh/uv/install.sh \
       | env UV_INSTALL_DIR=/usr/local/bin UV_NO_MODIFY_PATH=1 sh

# mise: extra language runtimes on demand (go, node, …).
RUN curl -fsSL https://mise.run | env MISE_INSTALL_PATH=/usr/local/bin/mise sh

# Non-root user. UID/GID are build args to match the host user on Linux Docker;
# the base image has no uid-1000 user, so nothing needs freeing first.
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd -g "$USER_GID" dev 2>/dev/null || true; \
    useradd -u "$USER_UID" -g "$USER_GID" -ms /bin/bash dev

# dev's user-space bins: the agent launcher (installed by the harness image) and
# mise's tool shims. Set here so every harness image inherits it.
ENV PATH=/home/dev/.local/bin:/home/dev/.local/share/mise/shims:$PATH

# Entrypoint installs the egress firewall then drops to dev; needs procps for the lock check.
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Stay root: the entrypoint sets nftables rules, then drops to the dev user itself.
WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
# Overridden by each harness image and by the command vhrn appends at run time.
CMD ["bash"]
