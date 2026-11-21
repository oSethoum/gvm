#!/usr/bin/env sh
set -e

REPO="oSethoum/gvm"
INSTALL_DIR="${GVM_INSTALL_DIR:-$HOME/.gvm/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  i386|i686) ARCH="386" ;;
esac

case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

LATEST="$(curl -sSL https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
if [ -z "$LATEST" ]; then
  echo "Could not determine latest release"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/gvm_${LATEST#v}_${OS}_${ARCH}.tar.gz"
echo "Downloading $URL"
mkdir -p "$INSTALL_DIR"
curl -sSL "$URL" | tar -xz -C "$INSTALL_DIR"

echo "Installed to $INSTALL_DIR/gvm"

# Ensure gvm bin dir is on PATH for this shell
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) PATH="${PATH}:${INSTALL_DIR}"; export PATH ;;
esac

# Ensure GOBIN is on PATH for this shell (when set)
if [ -n "${GOBIN:-}" ]; then
  case ":${PATH}:" in
    *":${GOBIN}:"*) ;;
    *) PATH="${PATH}:${GOBIN}"; export PATH ;;
  esac
fi

# Persist to shell rc
RC="$HOME/.bashrc"
case "$SHELL" in
  *zsh) RC="$HOME/.zshrc" ;;
esac
if ! grep -q "$INSTALL_DIR" "$RC" 2>/dev/null; then
  echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$RC"
fi
if [ -n "${GOBIN:-}" ] && ! grep -qF "${GOBIN}" "$RC" 2>/dev/null; then
  echo "export PATH=\"\$PATH:${GOBIN}\"" >> "$RC"
fi
if ! grep -q 'GOBIN' "$RC" 2>/dev/null; then
  echo 'export PATH="$PATH:$GOBIN"' >> "$RC"
fi

# Emit eval-able lines for current shell
echo "export PATH=\"\$PATH:$INSTALL_DIR\""
if [ -n "${GOBIN:-}" ]; then
  echo "export PATH=\"\$PATH:${GOBIN}\""
fi
echo 'export PATH="$PATH:$GOBIN"'
