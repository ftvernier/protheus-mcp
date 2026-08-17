$ErrorActionPreference = "Stop"

Write-Host "Downloading Go dependencies..."
go mod download

Write-Host "Running tests..."
go test ./...

New-Item -ItemType Directory -Force -Path dist | Out-Null

Write-Host "Building Windows AMD64 binary..."
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags="-s -w" -o dist/protheus-mcp-windows-amd64.exe ./cmd/protheus-mcp

Write-Host "Done: dist/protheus-mcp-windows-amd64.exe"
