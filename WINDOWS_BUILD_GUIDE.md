# Windows Build Guide - Ollama-K

**Complete instructions for building and testing Ollama-K on Windows 10/11**

---

## Quick Start

```powershell
# Clone repository
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
git checkout claude/ollama-windows-port-solnj

# Build Windows binary
.\build-windows.ps1

# Run tests
go test -v ./kuhul/runtime/...

# Run server
.\ollama.exe serve
```

---

## System Requirements

### Minimum
- **OS**: Windows 10 (Build 19041+) or Windows 11
- **Processor**: x86-64 with SSE4.2
- **RAM**: 4 GB
- **Disk**: 500 MB free
- **Network**: Optional (for Ollama/Orchestrator discovery)

### Recommended
- **OS**: Windows 11
- **Processor**: Intel/AMD 6th gen or newer
- **RAM**: 8+ GB
- **Disk**: 2 GB free
- **GPU**: NVIDIA CUDA Compute Capability 3.5+ (for Phase 3)

### Development
- **Go**: 1.24.7 or newer
- **Git**: Latest version
- **PowerShell**: 7.0+
- **CMake**: 3.13+ (for Phase 3 - MLIR/LLVM)
- **LLVM**: 14+ (for Phase 3 - optional)

---

## Installation

### 1. Install Go

**Using Windows Package Manager** (Recommended)
```powershell
# Check if Go is installed
go version

# Install via scoop
scoop install go

# Or via chocolatey
choco install golang

# Or download from https://golang.org/dl
```

**Verify Installation**
```powershell
go version
go env GOPATH
go env GOROOT
```

### 2. Install Git

```powershell
# Via scoop
scoop install git

# Via chocolatey
choco install git

# Via Windows Package Manager
winget install Git.Git
```

### 3. Clone Repository

```powershell
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
git checkout claude/ollama-windows-port-solnj

# Verify you're on the correct branch
git branch --show-current
# Output: claude/ollama-windows-port-solnj
```

### 4. Download Dependencies

```powershell
go mod download
go mod verify
```

---

## Building

### Local Build

**Using PowerShell Build Script** (Recommended)
```powershell
# From repository root
.\build-windows.ps1

# Output: ./ollama.exe
```

**Using Direct go build**
```powershell
# Build for Windows
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -v -o ollama.exe ./cmd/ollama

# Verify binary
dir ollama.exe
.\ollama.exe version
```

**Build Options**
```powershell
# Debug build with symbols
go build -v -gcflags="all=-N -l" -o ollama-debug.exe ./cmd/ollama

# Release build with optimizations
go build -v -ldflags="-s -w" -o ollama.exe ./cmd/ollama

# With custom version
go build -v -ldflags="-X main.Version=3.0.0" -o ollama.exe ./cmd/ollama
```

### Cross-Platform Build (from Linux/Mac)

```bash
# Build Windows binary on Linux
GOOS=windows GOARCH=amd64 go build -v -o ollama.exe ./cmd/ollama

# Build Windows binary on macOS
GOOS=windows GOARCH=amd64 go build -v -o ollama.exe ./cmd/ollama
```

---

## Testing

### Run All Tests

```powershell
# Run all tests with verbose output
go test -v ./...

# Run tests with timeout
go test -v -timeout 60s ./...

# Run tests with race detector
go test -v -race ./...
```

### Platform Abstraction Layer Tests

```powershell
# Run all platform abstraction tests
go test -v ./kuhul/runtime/platform_abstraction_test.go

# Run specific test
go test -v -run TestPathHandling ./kuhul/runtime/...

# Run Windows-specific tests
go test -v -run "Windows" ./kuhul/runtime/...

# Test categories
go test -v -run "TestPath" ./kuhul/runtime/...      # Path handling
go test -v -run "TestProcess" ./kuhul/runtime/...   # Process execution
go test -v -run "TestNetwork" ./kuhul/runtime/...   # Network/ports
go test -v -run "TestEnvironment" ./kuhul/runtime/... # Environment vars
```

### HTTP Bridge Tests

```powershell
# Run bridge tests
go test -v ./server/bridge_test.go

# Run specific bridge test
go test -v -run TestBridgeConfig ./server/...

# Run with coverage
go test -v -cover ./server/...
```

### Test Coverage Report

```powershell
# Generate coverage report
go test -coverprofile=coverage.out ./kuhul/runtime/...
go tool cover -html=coverage.out -o coverage.html

# View in default browser
Start-Process coverage.html
```

### Integration Tests

```powershell
# Run end-to-end tests
go test -v -run TestEndToEnd ./kuhul/...

# Run with long timeout
go test -v -timeout 120s ./kuhul/...
```

---

## Running the Server

### Basic Execution

```powershell
# Run server with default settings
.\ollama.exe serve

# Server will:
# 1. Start on http://localhost:7860
# 2. Auto-discover Ollama (11434) and Orchestrator (61683)
# 3. Expose PWA interface
# 4. Provide HTTP bridge endpoints
```

### With Custom Configuration

```powershell
# Set environment variables
$env:OLLAMA_HOST="127.0.0.1:7860"
$env:OLLAMA_NUM_PARALLEL=4
$env:OLLAMA_NUM_THREADS=8
$env:OLLAMA_DEBUG=1

.\ollama.exe serve

# Or use command line arguments
.\ollama.exe serve --host 0.0.0.0 --port 8080
```

### Service Discovery Test

```powershell
# In another PowerShell window
# Check service discovery
curl http://localhost:7860/api/services/discover | ConvertFrom-Json | Format-List

# Expected output:
# ollama_url       : http://localhost:11434
# orchestrator_url : http://localhost:61683
# last_discovery   : 2026-02-23T...
# timestamp        : 2026-02-23T...
```

### Health Check

```powershell
# Check server health
curl http://localhost:7860/api/health | ConvertFrom-Json | Format-List

# Expected output:
# status    : healthy
# services  : @{ollama=True; orchestrator=False}
# last_check: 2026-02-23T...
```

---

## Windows-Specific Features

### Path Handling

```powershell
# Test path operations
go test -v -run "TestPath" ./kuhul/runtime/platform_abstraction_test.go

# Paths work with both separators:
# C:\Users\Admin\file.txt (Windows style)
# C:/Users/Admin/file.txt (Unix style - auto-converted)

# UNC paths supported:
# \\server\share\file.txt
```

### Environment Variables

```powershell
# K'UHUL can access Windows environment variables
# sys.env_get("USERPROFILE")   → C:\Users\YourName
# sys.env_get("APPDATA")        → C:\Users\YourName\AppData\Roaming
# sys.env_get("LOCALAPPDATA")   → C:\Users\YourName\AppData\Local
# sys.env_get("TEMP")           → C:\Users\YourName\AppData\Local\Temp
```

### Process Execution

```powershell
# K'UHUL can execute Windows processes
# sys.proc_run("notepad.exe", @("file.txt"))
# sys.proc_run("powershell.exe", @("-Command", "Get-Process"))

# Automatic .exe extension handling:
# sys.proc_run("cmd", @("/c", "echo", "test"))  → calls cmd.exe
```

### Registry Access

```powershell
# K'UHUL can read/write Windows Registry
# sys.registry_get("HKCU", "Software\MyApp", "Setting")
# sys.registry_set("HKCU", "Software\MyApp", "Setting", "Value")
# sys.registry_list("HKLM", "Software\Microsoft")
```

### Port Management

```powershell
# K'UHUL can check port availability
# sys.port_available(8080)           → true/false
# sys.get_process_by_port(11434)     → PID
# sys.get_process_name_by_port(3000) → "node.exe"
```

---

## Troubleshooting

### Build Issues

**Problem**: `go: command not found`
```powershell
# Solution: Install Go from https://golang.org/dl
# Or: scoop install go
$env:Path -split ";" | Select-String "Go"
```

**Problem**: `Module not found`
```powershell
# Solution: Run go mod download
go mod download
go mod verify
```

**Problem**: `Compilation errors with platform_abstraction_test.go`
```powershell
# Solution: File was auto-formatted. Regenerate:
go fmt ./kuhul/runtime/platform_abstraction_test.go
go test -v ./kuhul/runtime/...
```

### Runtime Issues

**Problem**: `Port 7860 already in use`
```powershell
# Solution 1: Find and kill process using port
Get-Process | Where-Object { $_.Handles -gt 500 } | Stop-Process -Force

# Solution 2: Use different port
$env:OLLAMA_PORT="7861"
.\ollama.exe serve

# Solution 3: Check what's using the port
netstat -ano | Select-String ":7860"
taskkill /PID <PID> /F
```

**Problem**: `Service discovery fails`
```powershell
# Check if Ollama is running
curl http://localhost:11434/api/tags 2>$null

# Check if Orchestrator is running
curl http://localhost:61683/api/health 2>$null

# Both failing is OK - server continues without them
```

**Problem**: `Permission denied accessing Registry`
```powershell
# Solution: Run PowerShell as Administrator
# Or: Registry access works without admin for user keys (HKCU)
```

### Performance Issues

**Problem**: `Slow startup`
```powershell
# Check disk I/O
Get-Process | Where-Object { $_.ProcessName -eq "ollama" } | Get-Process
Measure-Command { .\ollama.exe serve }

# Check memory
Get-Process ollama | Select-Object WorkingSet, VirtualMemorySize
```

**Problem**: `High CPU usage`
```powershell
# Check what's consuming CPU
Get-Process | Sort-Object CPU -Descending | Select-Object -First 10

# Reduce parallel tasks
$env:OLLAMA_NUM_PARALLEL=1
.\ollama.exe serve
```

---

## CI/CD Integration

### GitHub Actions

The repository includes automatic Windows builds via `.github/workflows/windows-build.yml`:

```powershell
# Trigger on push to windows-port branch
git push origin claude/ollama-windows-port-solnj

# GitHub Actions will:
# 1. Set up Windows environment
# 2. Build ollama.exe
# 3. Run all tests
# 4. Generate coverage reports
# 5. Upload artifacts
```

**View Build Status**
- Go to: https://github.com/cannaseedus-bot/Ollama-K/actions
- Filter by: `windows-build`
- Check: Build logs, test results, artifacts

### Local Pre-Commit Hooks

```powershell
# Create pre-commit hook
$hookContent = @'
#!/bin/sh
# Run tests before commit
go test -v ./kuhul/runtime/... || exit 1
go fmt ./...
'@

$hookContent | Out-File -FilePath ".git/hooks/pre-commit" -Encoding UTF8 -NoNewline

# Make executable (optional on Windows)
```

---

## Development Workflow

### 1. Create Feature Branch

```powershell
git checkout claude/ollama-windows-port-solnj
git pull origin claude/ollama-windows-port-solnj
git checkout -b feature/your-feature-name
```

### 2. Make Changes

```powershell
# Edit files
code .

# Run tests frequently
go test -v ./kuhul/runtime/...

# Check formatting
go fmt ./...
```

### 3. Commit Changes

```powershell
git add .
git commit -m "Feature: Description of changes"

# Follow commit conventions:
# Feature: New functionality
# Fix: Bug fix
# Refactor: Code reorganization
# Docs: Documentation updates
# Test: Test additions
# Ci: CI/CD updates
```

### 4. Push and Create PR

```powershell
git push origin feature/your-feature-name

# Go to GitHub and create Pull Request:
# Base: claude/ollama-windows-port-solnj
# Compare: feature/your-feature-name
```

### 5. Merge to Main Branch

```powershell
# After review and CI/CD passes
git checkout main
git pull origin main
git merge --no-ff feature/your-feature-name
git push origin main
```

---

## Advanced Topics

### Phase 3: MLIR/LLVM Setup (Future)

When implementing Phase 3, you'll need:

```powershell
# Install LLVM (via scoop)
scoop install llvm

# Or download from https://releases.llvm.org/

# Set environment variables
$env:LLVM_DIR="C:\Path\To\LLVM"
$env:MLIR_SRC="C:\Path\To\MLIR"

# Build with MLIR support
CGO_ENABLED=1 `
  CGO_CXXFLAGS="-I$env:LLVM_DIR\include" `
  CGO_LDFLAGS="-L$env:LLVM_DIR\lib -lMLIR" `
  go build -tags mlir -o ollama-mlir.exe ./cmd/ollama
```

### Performance Profiling

```powershell
# CPU profiling
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./kuhul/runtime/...
go tool pprof cpu.prof

# Memory analysis
go tool pprof mem.prof
```

### Debugging

```powershell
# Build debug binary
go build -gcflags="all=-N -l" -o ollama-debug.exe ./cmd/ollama

# Use debugger (requires delve)
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/ollama
```

---

## Additional Resources

### Documentation
- [WINDOWS_PORT_STATUS.md](WINDOWS_PORT_STATUS.md) - Detailed status report
- [WINDOWS_PORT_ROADMAP.md](WINDOWS_PORT_ROADMAP.md) - Project roadmap
- [MLIR_LLVM_PHASE3_PLAN.md](MLIR_LLVM_PHASE3_PLAN.md) - Phase 3 specification
- [KUHUL_CLI_ARCHITECTURE.md](KUHUL_CLI_ARCHITECTURE.md) - CLI TUI design

### External Links
- [Go Installation Guide](https://golang.org/doc/install)
- [Go on Windows](https://github.com/golang/go/wiki/WindowsInstallation)
- [Git for Windows](https://git-scm.com/download/win)
- [PowerShell 7](https://github.com/PowerShell/PowerShell)

### Community
- GitHub Issues: Report bugs and feature requests
- Discussions: Ask questions and share ideas
- Pull Requests: Contribute improvements

---

## Support

For issues or questions:

1. **Check existing documentation** - Many common issues are documented above
2. **Review test logs** - `go test -v` provides detailed information
3. **Check GitHub Issues** - Search for similar problems
4. **Create new issue** - Include:
   - Windows version (10/11, build number)
   - Go version
   - Error messages
   - Steps to reproduce

---

## Contributing

We welcome Windows-specific improvements! Areas for contribution:

- [ ] GPU/CUDA support (Phase 3)
- [ ] Windows installer (.msi)
- [ ] Performance optimizations
- [ ] Additional test coverage
- [ ] Documentation improvements

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

**Last Updated**: February 23, 2026
**Branch**: `claude/ollama-windows-port-solnj`
**Status**: ✅ Phase 1-2 Complete, Phase 3 Planned
