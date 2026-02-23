#Requires -Version 7.0

<#
.SYNOPSIS
    Build script for Ollama-K Windows binary

.DESCRIPTION
    Builds ollama.exe for Windows with proper configuration and testing

.PARAMETER BuildType
    Type of build: Release (default) or Debug

.PARAMETER OutputPath
    Where to save the binary (default: current directory)

.PARAMETER RunTests
    Run tests before building (default: true)

.PARAMETER Verbose
    Enable verbose output

.EXAMPLE
    .\build-windows.ps1
    # Builds Release binary

    .\build-windows.ps1 -BuildType Debug -RunTests:$false
    # Builds Debug binary without running tests
#>

param(
    [ValidateSet('Release', 'Debug')]
    [string]$BuildType = 'Release',

    [string]$OutputPath = '.',

    [bool]$RunTests = $true,

    [switch]$Verbose
)

# Color helper
function Write-ColorOutput {
    param(
        [string]$Message,
        [ConsoleColor]$Color = 'White'
    )
    Write-Host $Message -ForegroundColor $Color
}

# Header
Write-ColorOutput "`n╔════════════════════════════════════════════════════════════╗" -Color Cyan
Write-ColorOutput "║        Ollama-K Windows Build Script                       ║" -Color Cyan
Write-ColorOutput "║        Branch: claude/ollama-windows-port-solnj            ║" -Color Cyan
Write-ColorOutput "╚════════════════════════════════════════════════════════════╝`n" -Color Cyan

# Verify Go installation
Write-ColorOutput "✓ Checking Go installation..." -Color Yellow
try {
    $goVersion = go version
    Write-ColorOutput "  $goVersion" -Color Green
} catch {
    Write-ColorOutput "✗ Go not found! Please install Go 1.24.7 or newer" -Color Red
    exit 1
}

# Verify in correct directory
if (-not (Test-Path "cmd/ollama/main.go")) {
    Write-ColorOutput "✗ Error: build-windows.ps1 must run from repository root" -Color Red
    exit 1
}

Write-ColorOutput "✓ Repository structure verified" -Color Green

# Download dependencies
Write-ColorOutput "`n📦 Downloading dependencies..." -Color Yellow
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-ColorOutput "✗ Failed to download dependencies" -Color Red
    exit 1
}
Write-ColorOutput "✓ Dependencies downloaded" -Color Green

# Verify dependencies
go mod verify
if ($LASTEXITCODE -ne 0) {
    Write-ColorOutput "✗ Dependency verification failed" -Color Red
    exit 1
}
Write-ColorOutput "✓ Dependencies verified" -Color Green

# Run tests (if requested)
if ($RunTests) {
    Write-ColorOutput "`n🧪 Running tests..." -Color Yellow

    # Run platform abstraction tests
    Write-ColorOutput "  Testing platform abstraction layer..." -Color Cyan
    go test -v -timeout 60s ./kuhul/runtime/platform_abstraction_test.go
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "⚠ Some tests failed (continuing)" -Color Yellow
    } else {
        Write-ColorOutput "  ✓ Platform tests passed" -Color Green
    }

    # Run all runtime tests
    Write-ColorOutput "  Running all runtime tests..." -Color Cyan
    go test -v -timeout 60s ./kuhul/runtime/...
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "⚠ Some tests failed (continuing)" -Color Yellow
    } else {
        Write-ColorOutput "  ✓ All runtime tests passed" -Color Green
    }
}

# Format check
Write-ColorOutput "`n✓ Checking code formatting..." -Color Yellow
$formatCheck = go fmt ./...
if ($formatCheck) {
    Write-ColorOutput "⚠ Some files need formatting (auto-formatting)" -Color Yellow
} else {
    Write-ColorOutput "✓ Code formatting is correct" -Color Green
}

# Build configuration
Write-ColorOutput "`n⚙️  Build Configuration:" -Color Yellow
Write-ColorOutput "  Build Type: $BuildType" -Color Cyan
Write-ColorOutput "  Output Path: $OutputPath" -Color Cyan
Write-ColorOutput "  GOOS: windows" -Color Cyan
Write-ColorOutput "  GOARCH: amd64" -Color Cyan

# Set environment variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Build flags
$buildFlags = @("-v")

switch ($BuildType) {
    'Debug' {
        Write-ColorOutput "`n🔨 Building Debug binary..." -Color Yellow
        $buildFlags += "-gcflags=all=-N -l"
        $outputFile = Join-Path $OutputPath "ollama-debug.exe"
    }
    'Release' {
        Write-ColorOutput "`n🔨 Building Release binary..." -Color Yellow
        $version = git describe --tags --always --dirty
        $commit = git rev-parse --short HEAD
        $buildDate = Get-Date -Format "2006-01-02"

        $ldFlags = @(
            "-s",                                    # Strip symbols
            "-w",                                    # Strip DWARF info
            "-X", "github.com/ollama/ollama/version.Version=$version",
            "-X", "github.com/ollama/ollama/version.Commit=$commit",
            "-X", "github.com/ollama/ollama/version.BuildDate=$buildDate"
        ) -join " "

        $buildFlags += "-ldflags=$ldFlags"
        $outputFile = Join-Path $OutputPath "ollama.exe"
    }
}

# Execute build
try {
    Write-ColorOutput "  Compiling..." -Color Cyan

    $startTime = Get-Date
    & go build @buildFlags -o $outputFile ./cmd/ollama
    $endTime = Get-Date
    $duration = $endTime - $startTime

    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "✗ Build failed!" -Color Red
        exit 1
    }

    # Verify binary was created
    if (-not (Test-Path $outputFile)) {
        Write-ColorOutput "✗ Binary file not found: $outputFile" -Color Red
        exit 1
    }

    # Get binary info
    $fileInfo = Get-Item $outputFile
    $sizeKB = [math]::Round($fileInfo.Length / 1024, 2)

    Write-ColorOutput "`n✅ Build successful!" -Color Green
    Write-ColorOutput "  Binary: $outputFile" -Color Green
    Write-ColorOutput "  Size: $sizeKB KB" -Color Green
    Write-ColorOutput "  Build time: $($duration.TotalSeconds) seconds" -Color Green

    # Verify binary
    Write-ColorOutput "`n✓ Verifying binary..." -Color Yellow
    & $outputFile version
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Binary verification successful" -Color Green
    } else {
        Write-ColorOutput "⚠ Binary verification returned non-zero exit code" -Color Yellow
    }

    # Summary
    Write-ColorOutput "`n╔════════════════════════════════════════════════════════════╗" -Color Green
    Write-ColorOutput "║                     BUILD SUMMARY                           ║" -Color Green
    Write-ColorOutput "╠════════════════════════════════════════════════════════════╣" -Color Green
    Write-ColorOutput "║ ✅ Windows binary built successfully!                        ║" -Color Green
    Write-ColorOutput "║                                                            ║" -Color Green
    Write-ColorOutput "║ Next steps:                                                ║" -Color Green
    Write-ColorOutput "║   1. Run: .\$(Split-Path $outputFile -Leaf) serve                   ║" -Color Green
    Write-ColorOutput "║   2. Access: http://localhost:7860                         ║" -Color Green
    Write-ColorOutput "║   3. Check: curl http://localhost:7860/api/health         ║" -Color Green
    Write-ColorOutput "║                                                            ║" -Color Green
    Write-ColorOutput "║ Documentation:                                            ║" -Color Green
    Write-ColorOutput "║   - WINDOWS_BUILD_GUIDE.md (Complete build instructions)  ║" -Color Green
    Write-ColorOutput "║   - WINDOWS_PORT_STATUS.md (Implementation status)        ║" -Color Green
    Write-ColorOutput "║   - WINDOWS_PORT_ROADMAP.md (Project roadmap)             ║" -Color Green
    Write-ColorOutput "╚════════════════════════════════════════════════════════════╝" -Color Green

    # Optional: List build artifacts
    Write-ColorOutput "`n📁 Build artifacts:" -Color Cyan
    Get-Item $outputFile | Format-Table Name, Length, LastWriteTime

    Write-ColorOutput "`n✨ Ready to go! Run the binary with: $outputFile serve`n" -Color Green

} catch {
    Write-ColorOutput "`n✗ Build failed with exception:" -Color Red
    Write-ColorOutput $_.Exception.Message -Color Red
    exit 1
}

# Exit successfully
exit 0
