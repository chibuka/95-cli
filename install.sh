#!/bin/bash
set -e

# 95 CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/chibuka/95-cli/main/install.sh | bash


REPO="chibuka/95-cli"
BINARY_NAME="95"
INSTALL_DIR="$HOME/.local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    local os arch

    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux*)
            OS="unknown-linux-gnu"
            ;;
        darwin*)
            OS="apple-darwin"
            ;;
        *)
            echo -e "${RED}error: unsupported operating system: $os${NC}"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            ARCH="x86_64"
            ;;
        aarch64|arm64)
            ARCH="aarch64"
            ;;
        *)
            echo -e "${RED}error: unsupported architecture: $arch${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${ARCH}-${OS}"
}

# Get latest release version
get_latest_version() {
    echo -e "${CYAN}→ fetching latest release...${NC}"

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        echo -e "${RED}error: failed to fetch latest version${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ latest version: $VERSION${NC}"
}

install_binary() {
    local download_url="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${PLATFORM}"
    local tmp_file="/tmp/${BINARY_NAME}"

    echo -e "${CYAN}→ downloading $BINARY_NAME for $PLATFORM...${NC}"

    if ! curl -fsSL "$download_url" -o "$tmp_file"; then
        echo -e "${RED}error: failed to download binary${NC}"
        echo -e "${YELLOW}url: $download_url${NC}"
        exit 1
    fi

    # Make binary executable
    chmod +x "$tmp_file"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Move binary to install directory
    mv "$tmp_file" "$INSTALL_DIR/$BINARY_NAME"

    echo -e "${GREEN}✓ installed to $INSTALL_DIR/$BINARY_NAME${NC}"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo -e "${YELLOW}warning: $INSTALL_DIR is not in your PATH${NC}"
        echo ""
        echo "add this line to your shell config file (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo -e "${CYAN}  export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        echo ""
    fi
}

print_success() {
    echo ""
    echo -e "${GREEN}installation complete!${NC}"
    echo ""
    echo "run your first command:"
    echo -e "${CYAN}  $BINARY_NAME${NC}"
    echo ""
}

main() {
    echo ""
    echo -e "${YELLOW} █████╗ ███████╗${NC}"
    echo -e "${YELLOW}██╔══██╗██╔════╝${NC}"
    echo -e "${YELLOW}╚██████║███████╗${NC}"
    echo -e "${YELLOW} ╚═══██║╚════██║${NC}"
    echo -e "${YELLOW} █████╔╝███████║${NC}"
    echo -e "${YELLOW} ╚════╝ ╚══════╝${NC}"
    echo ""
    echo "ninefive cli installer"
    echo ""

    detect_platform
    get_latest_version
    install_binary
    check_path
    print_success
}

main
