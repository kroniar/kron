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
go build -o kron main.go

# Install path
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"
mv kron "$INSTALL_DIR/"

# Add to PATH if missing
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "🧩 Adding $INSTALL_DIR to PATH..."
  SHELL_RC="$HOME/.bashrc"
  [ -n "$ZSH_VERSION" ] && SHELL_RC="$HOME/.zshrc"
  echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
  echo "✅ Added to PATH. Restart your terminal or run:"
  echo "   source $SHELL_RC"
fi

# Copy default setup config
CONFIG_DIR="$HOME/.kron"
mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/setup.yaml" ]; then
  echo "📦 Copying default setup configuration..."
  cp configs/kron/setup.yaml "$CONFIG_DIR/setup.yaml"
else
  echo "✅ Existing ~/.kron/setup.yaml found, skipping copy."
fi

echo "✅ Kron installed successfully!"
echo "You can now run 'kron' from anywhere 🎉"
