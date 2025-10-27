Write-Host "⚙️ Setting up Kron..."

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Host "❌ Go not found. Install from https://go.dev/dl/"
  exit 1
}

# Initialize go.mod if missing
if (-not (Test-Path "go.mod")) {
  go mod init github.com/kroniar/kron
}

# Install dependencies
go mod tidy

# Build the binary
Write-Host "🔨 Building binary..."
go build -o kron.exe

# Install path
$InstallPath = "$env:USERPROFILE\.kron"
if (-not (Test-Path $InstallPath)) {
  New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null
}

Move-Item -Force -Path "kron.exe" -Destination "$InstallPath\kron.exe"

# Add to PATH if not already there
if (-not ($env:Path -match [regex]::Escape($InstallPath))) {
  Write-Host "🧩 Adding $InstallPath to PATH..."
  [Environment]::SetEnvironmentVariable("Path", $env:Path + ";" + $InstallPath, "User")
  Write-Host "✅ PATH updated. Restart PowerShell or Command Prompt."
}

Write-Host "✅ Kron installed successfully! Run 'kron' from anywhere 🎉"
