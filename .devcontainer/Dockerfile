FROM mcr.microsoft.com/devcontainers/go:1-1.24

ARG GO_VERSION=1.26.1
RUN ARCH="$(dpkg --print-architecture)" \
    && rm -rf /usr/local/go \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
    | tar -C /usr/local -xz
