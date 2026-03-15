#!/bin/bash

set -e

# LLM Detector Installation Script
# Usage: curl -sSL https://your-domain.com/install.sh | bash

REPO="llm-detector"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux|darwin)
            PLATFORM="${OS}-${ARCH}"
            ;;
        *)
            echo "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    echo "$PLATFORM"
}

# Download and install
download() {
    PLATFORM=$(detect_platform)
    
    if [ "$VERSION" = "latest" ]; then
        # Get latest release URL
        DOWNLOAD_URL="https://github.com/yourusername/llm-detector/releases/latest/download/llm-detector-${PLATFORM}.tar.gz"
    else
        DOWNLOAD_URL="https://github.com/yourusername/llm-detector/releases/download/${VERSION}/llm-detector-${PLATFORM}.tar.gz"
    fi
    
    echo "Downloading LLM Detector for ${PLATFORM}..."
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    # Download
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o llm-detector.tar.gz
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O llm-detector.tar.gz
    else
        echo "Error: curl or wget required"
        exit 1
    fi
    
    # Extract
    echo "Extracting..."
    tar -xzf llm-detector.tar.gz
    
    # Install binary
    echo "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        cp llm-detector/llm-detector "$INSTALL_DIR/"
        mkdir -p /etc/llm-detector
        cp -r llm-detector/fingerprints /etc/llm-detector/
    else
        sudo cp llm-detector/llm-detector "$INSTALL_DIR/"
        sudo mkdir -p /etc/llm-detector
        sudo cp -r llm-detector/fingerprints /etc/llm-detector/
    fi
    
    # Cleanup
    cd -
    rm -rf "$TMP_DIR"
    
    echo "✓ LLM Detector installed successfully!"
    echo "  Binary: ${INSTALL_DIR}/llm-detector"
    echo "  Data: /etc/llm-detector/fingerprints"
    echo ""
    echo "Usage: llm-detector -t localhost:11434"
}

# Main
main() {
    echo "LLM Detector Installer"
    echo "======================"
    echo ""
    
    # Check for required tools
    if ! command -v tar &> /dev/null; then
        echo "Error: tar is required"
        exit 1
    fi
    
    download
}

main "$@"
