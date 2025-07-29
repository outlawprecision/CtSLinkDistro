FROM alpine:3.18

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    wget \
    bash \
    bash-completion \
    openssh-client \
    build-base \
    linux-headers \
    ca-certificates \
    tzdata \
    sudo \
    shadow

# Create vscode user with sudo access
RUN addgroup -g 1000 vscode && \
    adduser -u 1000 -G vscode -s /bin/bash -D vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Go
ENV GO_VERSION=1.21.5
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH

RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Create GOPATH directory and set permissions
RUN mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg && \
    chown -R vscode:vscode $GOPATH

# Install GitHub CLI
RUN apk add --no-cache github-cli

# Install common Go tools
USER vscode
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest

# Set working directory
WORKDIR /workspace

# Switch back to root for any final setup
USER root

# Install additional development tools
RUN apk add --no-cache \
    make \
    vim \
    nano \
    htop \
    tree \
    jq

# Set up shell for vscode user
USER vscode
RUN echo 'export GOROOT=/usr/local/go' >> ~/.bashrc && \
    echo 'export GOPATH=/go' >> ~/.bashrc && \
    echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH' >> ~/.bashrc

# Default command
CMD ["sleep", "infinity"]