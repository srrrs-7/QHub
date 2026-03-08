FROM mcr.microsoft.com/devcontainers/base:2-trixie

# Install dependencies required for Bun and uv
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    make \
    && rm -rf /var/lib/apt/lists/*

# Install Bun as vscode user
USER vscode

WORKDIR /workspace/main

# Install claude code
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install uv
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Install Python 3.12 using uv
RUN /home/vscode/.local/bin/uv python install 3.12

# Install spec-kit (Spec-Driven Development toolkit)
RUN /home/vscode/.local/bin/uv tool install specify-cli --from git+https://github.com/github/spec-kit.git

# go migration atlas installation

RUN curl -sSf https://atlasgo.sh | sh

# go orm sqlc installation
RUN curl -L "$(curl -s https://api.github.com/repos/sqlc-dev/sqlc/releases/latest | grep "browser_download_url.*$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | cut -d '"' -f 4)" | tar xz -C ~/.local/bin

# golangci-lint installation
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/.local/bin
