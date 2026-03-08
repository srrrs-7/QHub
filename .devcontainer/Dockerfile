FROM mcr.microsoft.com/devcontainers/base:2-trixie

# Install dependencies required for tools and firewall
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    make \
    iptables \
    ipset \
    dnsutils \
    jq \
    iproute2 \
    aggregate \
    && rm -rf /var/lib/apt/lists/*

# Install Codex CLI
RUN set -euo pipefail \
    && ARCH="$(uname -m)" \
    && case "${ARCH}" in \
        x86_64|amd64) CODEX_TARGET="x86_64-unknown-linux-musl" ;; \
        aarch64|arm64) CODEX_TARGET="aarch64-unknown-linux-musl" ;; \
        *) echo "Unsupported architecture: ${ARCH}" >&2; exit 1 ;; \
    esac \
    && curl -L -o /tmp/codex.tar.gz "https://github.com/openai/codex/releases/latest/download/codex-${CODEX_TARGET}.tar.gz" \
    && tar -xzf /tmp/codex.tar.gz -C /tmp \
    && chmod +x "/tmp/codex-${CODEX_TARGET}" \
    && mv "/tmp/codex-${CODEX_TARGET}" /usr/local/bin/codex \
    && rm /tmp/codex.tar.gz

# Install Bun as vscode user
USER vscode

WORKDIR /workspace/main

# Install claude code
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install uv
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Install Python 3.12 using uv
RUN /home/vscode/.local/bin/uv python install 3.12

# go migration atlas installation
RUN curl -sSf https://atlasgo.sh | sh

# misspell - spell checker
RUN curl -sfL https://raw.githubusercontent.com/golangci/misspell/HEAD/install-misspell.sh | sh -s -- -b ~/.local/bin

# go orm sqlc installation
RUN curl -L "$(curl -s https://api.github.com/repos/sqlc-dev/sqlc/releases/latest | grep "browser_download_url.*$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | cut -d '"' -f 4)" | tar xz -C ~/.local/bin
