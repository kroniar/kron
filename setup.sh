#!/usr/bin/env bash
set -e

echo "⚙️  Setting up Kron..."

# Check Go installation
if ! command -v go &> /dev/null; then
  echo "❌ Go not found. Please install Go first: https://go.dev/dl/"
  exit 1
fi

# Initialize Go module if missing
if [ ! -f "go.mod" ]; then
  go mod init github.com/kroniar/kron
fi

# Install dependencies
go mod tidy

# Build the binary
echo "🔨 Building binary..."
go build -o kron

# Choose global install path
INSTALL_DIR="$HOME/.local/bin"

# Create if not exists
mkdir -p "$INSTALL_DIR"

# Move binary
mv kron "$INSTALL_DIR/"

# Ensure path is set
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "🧩 Adding $INSTALL_DIR to PATH..."
  SHELL_RC="$HOME/.bashrc"
  [ -n "$ZSH_VERSION" ] && SHELL_RC="$HOME/.zshrc"
  echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
  echo "✅ Added to PATH. Restart your terminal or run:"
  echo "   source $SHELL_RC"
fi

echo "✅ Kron installed successfully!"
echo "You can now run 'kron' from anywhere 🎉"
