#!/usr/bin/env bash
set -euo pipefail

BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

echo -e "${BOLD}LibKill Installer${RESET}"
echo "Supply-chain compromise scanner and cleaner"
echo ""

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo -e "${RED}Unsupported architecture: $ARCH${RESET}"; exit 1 ;;
esac

case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *) echo -e "${RED}Unsupported OS: $OS${RESET}"; exit 1 ;;
esac

VERSION="${1:-latest}"
INSTALL_DIR="${HOME}/.local/bin"
BINARY="libkill-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then BINARY="${BINARY}.exe"; fi

DOWNLOAD_URL="https://github.com/firfircelik/libkill/releases/${VERSION}/download/${BINARY}"

mkdir -p "$INSTALL_DIR"

echo "Downloading libkill-${OS}-${ARCH}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "${INSTALL_DIR}/libkill"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$DOWNLOAD_URL" -O "${INSTALL_DIR}/libkill"
else
  echo -e "${RED}curl or wget required${RESET}"
  exit 1
fi

chmod +x "${INSTALL_DIR}/libkill"
echo -e "${GREEN}Installed: ${INSTALL_DIR}/libkill${RESET}"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "Add to your PATH:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo ""
  echo "Or add this line to your ~/.zshrc or ~/.bashrc"
fi

echo ""
echo "Run 'libkill scan' to scan your system."
echo "Run 'libkill install' to set up auto-scanning daemon."
