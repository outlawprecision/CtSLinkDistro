# Minimal approach: Build basic container, install tools after container starts
FROM golang:1.21.5-alpine3.18

# Fix SSL certificates
RUN apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

# Install system dependencies (no Go tool downloads during build)
RUN apk add --no-cache \
    git \
    curl \
    wget \
    bash \
    bash-completion \
    openssh-client \
    build-base \
    linux-headers \
    tzdata \
    sudo \
    shadow \
    make \
    vim \
    nano \
    htop \
    tree \
    jq \
    github-cli

# Create vscode user with sudo access
RUN addgroup -g 1000 vscode && \
    adduser -u 1000 -G vscode -s /bin/bash -D vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set Go environment variables
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_DIR=/etc/ssl/certs

# Create GOPATH directory and set permissions
RUN mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg && \
    chown -R vscode:vscode $GOPATH

# Set working directory
WORKDIR /workspace

# Set up shell for vscode user (no Go tool installation during build)
USER vscode
RUN echo 'export GOPATH=/go' >> ~/.bashrc && \
    echo 'export PATH=$GOPATH/bin:$PATH' >> ~/.bashrc && \
    echo 'export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt' >> ~/.bashrc && \
    echo 'export SSL_CERT_DIR=/etc/ssl/certs' >> ~/.bashrc

# Create a script to install Go tools after container starts
RUN echo '#!/bin/bash' > /home/vscode/install-go-tools.sh && \
    echo 'echo "Installing Go development tools..."' >> /home/vscode/install-go-tools.sh && \
    echo 'go install golang.org/x/tools/cmd/goimports@latest' >> /home/vscode/install-go-tools.sh && \
    echo 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest' >> /home/vscode/install-go-tools.sh && \
    echo 'go install golang.org/x/tools/gopls@latest' >> /home/vscode/install-go-tools.sh && \
    echo 'go install github.com/go-delve/delve/cmd/dlv@latest' >> /home/vscode/install-go-tools.sh && \
    echo 'go install honnef.co/go/tools/cmd/staticcheck@latest' >> /home/vscode/install-go-tools.sh && \
    echo 'echo "Go tools installation complete!"' >> /home/vscode/install-go-tools.sh && \
    chmod +x /home/vscode/install-go-tools.sh

# Default command
CMD ["sleep", "infinity"]