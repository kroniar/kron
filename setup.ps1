Write-Host "⚙️ Setting up Kron..."

# Check Go installation
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Host "❌ Go not found. Install from https://go.dev/dl/"
  exit 1
}

# Initialize Go module if missing
if (-not (Test-Path "go.mod")) {
  go mod init github.com/kroniar/kron
}

# Install dependencies
go mod tidy

# Build binary
Write-Host "🔨 Building binary..."
go build -o kron.exe main.go

# Install path
$InstallPath = "$env:USERPROFILE\.local\bin"
if (-not (Test-Path $InstallPath)) {
  New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null
}

Move-Item -Force -Path "kron.exe" -Destination "$InstallPath\kron.exe"

# Add to PATH if missing
if (-not ($env:Path -match [regex]::Escape($InstallPath))) {
  Write-Host "🧩 Adding $InstallPath to PATH..."
  [Environment]::SetEnvironmentVariable("Path", $env:Path + ";" + $InstallPath, "User")
  Write-Host "✅ PATH updated. Restart PowerShell or Command Prompt."
}

# Copy setup config
$ConfigDir = "$env:USERPROFILE\.kron"
if (-not (Test-Path $ConfigDir)) {
  New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
}

if (-not (Test-Path "$ConfigDir\setup.yaml")) {
  Write-Host "📦 Copying default setup configuration..."
  Copy-Item "configs\setup\setup.yaml" "$ConfigDir\setup.yaml"
} else {
  Write-Host "✅ Existing setup.yaml found, skipping copy."
}

Write-Host "✅ Kron installed successfully! Run 'kron' from anywhere 🎉"
