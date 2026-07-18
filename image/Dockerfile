FROM node:lts-slim

# Install essential CLI tools, build dependencies, and procps (required for pgrep lock checks)
RUN apt-get update && apt-get install -y \
    git \
    curl \
    jq \
    less \
    build-essential \
    sudo \
    procps \
    ncurses-term \
    && rm -rf /var/lib/apt/lists/*

# Non-root user with passwordless sudo. UID/GID are build args to match the host user on
# Linux Docker; free the base image's `node` user (uid 1000) first so `dev` can take it.
ARG USER_UID=1000
ARG USER_GID=1000
RUN userdel -r node 2>/dev/null || true; \
    groupdel node 2>/dev/null || true; \
    getent group $USER_GID >/dev/null 2>&1 || groupadd -g $USER_GID dev; \
    useradd -u $USER_UID -g $USER_GID -ms /bin/bash dev && \
    echo "dev ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Claude Code globally
RUN npm install -g @anthropic-ai/claude-code

# Copy and configure the self-healing lock management script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER dev
WORKDIR /workspace

# Ensure the lock check runs on every boot while defaulting to launching 'claude'
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]

