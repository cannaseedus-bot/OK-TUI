# Windows Testing & Validation Guide

**For**: Ollama-K Windows Port (Phase 1-2)
**Branch**: `claude/ollama-windows-port-solnj`
**Status**: Ready for comprehensive Windows validation

---

## Pre-Testing Checklist

### System Requirements
- [ ] Windows 10 (Build 19041+) or Windows 11
- [ ] Go 1.24.7 or newer installed
- [ ] Git installed and in PATH
- [ ] PowerShell 7.0+ (for build script)
- [ ] 4+ GB RAM available
- [ ] 500+ MB disk space free
- [ ] Network connectivity (for service discovery tests)

### Dependencies
- [ ] Go installed: `go version`
- [ ] Git configured: `git config --global user.name "Your Name"`
- [ ] Repository cloned: `git clone https://github.com/cannaseedus-bot/Ollama-K.git`
- [ ] Correct branch: `git checkout claude/ollama-windows-port-solnj`

---

## Phase 1: Build Validation

### 1.1 Repository Setup
```powershell
# Clone repository
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K

# Checkout Windows port branch
git checkout claude/ollama-windows-port-solnj

# Verify branch
git branch --show-current
# Expected: claude/ollama-windows-port-solnj
```

**Validation**:
- [ ] Repository cloned successfully
- [ ] Branch is `claude/ollama-windows-port-solnj`
- [ ] No uncommitted changes: `git status` (clean)

### 1.2 Dependency Download
```powershell
# Download all Go modules
go mod download

# Verify checksums
go mod verify
```

**Validation**:
- [ ] No network errors
- [ ] All modules downloaded successfully
- [ ] Checksums verified: `ok` output

### 1.3 Build Release Binary
```powershell
# Method 1: Using build script (recommended)
.\build-windows.ps1

# Method 2: Manual build
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -v -o ollama.exe ./cmd/ollama
```

**Validation**:
- [ ] `ollama.exe` created (check: `dir ollama.exe`)
- [ ] File size reasonable (expected: 30-50 MB)
- [ ] No build errors
- [ ] Build time recorded

### 1.4 Build Debug Binary
```powershell
# Build debug version for troubleshooting
.\build-windows.ps1 -BuildType Debug
```

**Validation**:
- [ ] `ollama-debug.exe` created
- [ ] File size larger than release (due to symbols)
- [ ] Symbols retained for debugging

---

## Phase 2: Binary Verification

### 2.1 Binary Properties
```powershell
# Check binary metadata
dir ollama.exe | Format-Table FullName, Length, LastWriteTime

# Get file version (if embedded)
(Get-Item ollama.exe).VersionInfo | Format-List
```

**Validation**:
- [ ] File exists
- [ ] File size reasonable
- [ ] Timestamp correct

### 2.2 Binary Execution
```powershell
# Test version command
.\ollama.exe version

# Expected output: Version information
```

**Validation**:
- [ ] Binary executes without errors
- [ ] Version information displayed
- [ ] No crash on startup
- [ ] Exit code 0

### 2.3 Help Output
```powershell
# Test help command
.\ollama.exe --help

# Test serve help
.\ollama.exe serve --help
```

**Validation**:
- [ ] Help displays correctly
- [ ] All commands listed
- [ ] No encoding issues
- [ ] Usage instructions clear

---

## Phase 3: Unit Tests

### 3.1 Platform Abstraction Tests
```powershell
# Run path handling tests
go test -v -run "TestPath" ./kuhul/runtime/...

# Run process execution tests
go test -v -run "TestProcess" ./kuhul/runtime/...

# Run network tests
go test -v -run "TestNetwork" ./kuhul/runtime/...

# Run environment tests
go test -v -run "TestEnvironment" ./kuhul/runtime/...

# Run all platform abstraction tests
go test -v ./kuhul/runtime/platform_abstraction_test.go
```

**Validation**:
- [ ] All path tests pass
- [ ] All process tests pass
- [ ] All network tests pass
- [ ] All environment tests pass
- [ ] Total: 100% passing
- [ ] No race conditions detected

### 3.2 Bridge Tests
```powershell
# Run HTTP bridge tests
go test -v ./server/bridge_test.go

# Run with coverage
go test -v -cover ./server/...
```

**Validation**:
- [ ] All bridge tests pass
- [ ] Service discovery logic tested
- [ ] Health check tested
- [ ] Proxy endpoints tested
- [ ] Configuration tested

### 3.3 Full Test Suite
```powershell
# Run all tests with timeout
go test -v -timeout 60s ./...

# Run all tests with race detector
go test -v -race ./...
```

**Validation**:
- [ ] All tests passing
- [ ] No race conditions
- [ ] No deadlocks
- [ ] Execution time acceptable (< 60s)

---

## Phase 4: Windows-Specific Features

### 4.1 Path Handling

```powershell
# Test various path formats
$paths = @(
    "C:\Users\$env:USERNAME\file.txt",
    "C:/Users/$env:USERNAME/file.txt",
    "\\server\share\file.txt",
    ".\relative\path.txt",
    "..\parent\path.txt"
)

foreach ($path in $paths) {
    Write-Host "Testing: $path"
    # Should all normalize correctly
}
```

**Validation**:
- [ ] Backslash paths work
- [ ] Forward slash paths work
- [ ] UNC paths supported
- [ ] Relative paths work
- [ ] Path joining correct
- [ ] Path separators normalized

### 4.2 Environment Variables

```powershell
# Test Windows environment variables
$env:TEST_VAR = "test_value"

# These should all be accessible from ollama:
# $env:USERPROFILE    → C:\Users\YourName
# $env:APPDATA        → C:\Users\YourName\AppData\Roaming
# $env:LOCALAPPDATA   → C:\Users\YourName\AppData\Local
# $env:TEMP           → C:\Users\YourName\AppData\Local\Temp
# $env:USERNAME       → YourName
# $env:COMPUTERNAME   → ComputerName
```

**Validation**:
- [ ] HOME directory correct
- [ ] TEMP directory accessible
- [ ] APP data paths exist
- [ ] Custom variables readable
- [ ] Variable expansion works

### 4.3 Process Execution

```powershell
# Test command execution
# ollama should be able to run:
# - cmd.exe commands
# - powershell scripts
# - external executables
# - shell builtins with proper quoting

# Manual verification:
# sys.proc_run("notepad", @())              → opens notepad
# sys.proc_run("cmd", @("/c", "echo", "test")) → should output "test"
```

**Validation**:
- [ ] Commands with .exe extension work
- [ ] Commands without extension work (auto .exe added)
- [ ] Command line arguments parsed correctly
- [ ] Quote handling correct
- [ ] Process exit codes returned
- [ ] STDOUT/STDERR captured

### 4.4 Port Management

```powershell
# Test port availability checking
# Find an available port
$testPort = 65432

# Check if port is available
netstat -ano | Select-String ":$testPort"
# Should show empty (port available)

# Port-to-process lookup
netstat -ano | findstr ":11434"
# Should show Ollama process (if running)
```

**Validation**:
- [ ] Port availability detection works
- [ ] Port-to-process mapping works
- [ ] Service discovery finds Ollama (11434)
- [ ] Service discovery finds Orchestrator (61683)
- [ ] Health checks work correctly

### 4.5 Registry Access

```powershell
# Test Registry access (if running as admin)
# Can verify through manual testing that:
# - HKCU registry keys readable
# - HKLM registry keys readable (if admin)
# - Registry values retrieved correctly
# - Registry type conversion works

# Check if running as admin
$isAdmin = ([Security.Principal.WindowsPrincipal] `
    [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
Write-Host "Running as Admin: $isAdmin"
```

**Validation**:
- [ ] HKCU read access works
- [ ] HKCU write access works (if admin)
- [ ] HKLM read access works
- [ ] Registry value types correct
- [ ] Multi-value strings handled
- [ ] Registry enumeration works

---

## Phase 5: Server Runtime

### 5.1 Start Server
```powershell
# Terminal 1: Start server
.\ollama.exe serve

# Expected output:
# - Service discovery messages
# - Health check messages
# - HTTP server started on :7860
# - No errors in startup
```

**Validation**:
- [ ] Server starts without errors
- [ ] No crash on startup
- [ ] HTTP listener active
- [ ] Service discovery running
- [ ] Health checks executing
- [ ] Log output clear and informative

### 5.2 Health Checks
```powershell
# Terminal 2: Check server health
$healthCheck = curl http://localhost:7860/api/health -UseBasicParsing | ConvertFrom-Json
Write-Host "Health Status:" $healthCheck.status
Write-Host "Ollama Available:" $healthCheck.services.ollama
Write-Host "Orchestrator Available:" $healthCheck.services.orchestrator
```

**Validation**:
- [ ] Health endpoint responds
- [ ] Status is "healthy"
- [ ] Service status reported
- [ ] Response time < 1s

### 5.3 Service Discovery
```powershell
# Check service discovery
$discovery = curl http://localhost:7860/api/services/discover -UseBasicParsing | ConvertFrom-Json
Write-Host "Ollama URL:" $discovery.ollama_url
Write-Host "Orchestrator URL:" $discovery.orchestrator_url
Write-Host "Timestamp:" $discovery.timestamp
```

**Validation**:
- [ ] Discovery endpoint responds
- [ ] Returns service URLs
- [ ] Timestamp valid
- [ ] Service detection working
- [ ] Fallback handled (if services down)

### 5.4 Web Interface
```powershell
# Open web interface in browser
Start-Process "http://localhost:7860"
```

**Validation**:
- [ ] Web page loads
- [ ] No 404 errors
- [ ] UI renders correctly
- [ ] Interactive controls work
- [ ] No JavaScript errors
- [ ] Responsive on mobile

### 5.5 Proxy Endpoints
```powershell
# Test proxy endpoints (requires Ollama or Orchestrator running)
# If available: POST /api/proxy/infer

# Example:
$body = @{
    model = "llama2"
    prompt = "Hello, world!"
    stream = $false
} | ConvertTo-Json

curl -Method POST `
     -Uri "http://localhost:7860/api/proxy/infer" `
     -ContentType "application/json" `
     -Body $body -UseBasicParsing
```

**Validation**:
- [ ] Endpoints respond
- [ ] Proxy forwarding works
- [ ] Error handling correct
- [ ] Timeouts handled
- [ ] Fallback chains work

---

## Phase 6: Performance Testing

### 6.1 Memory Usage
```powershell
# Monitor memory while server running
Get-Process ollama | Select-Object WorkingSet, VirtualMemorySize, Handles

# Acceptable ranges:
# - WorkingSet: 50-200 MB
# - VirtualMemorySize: 100-500 MB
# - Handles: 100-500
```

**Validation**:
- [ ] Memory usage reasonable
- [ ] No memory leaks (stable after startup)
- [ ] Handle count stable
- [ ] Virtual memory reasonable

### 6.2 CPU Usage
```powershell
# Monitor CPU while idle
Get-Process ollama | Select-Object CPU, ThreadCount

# Acceptable:
# - Idle CPU: < 5%
# - Thread count: < 50
```

**Validation**:
- [ ] Low CPU usage at idle
- [ ] Reasonable thread count
- [ ] No spinning loops
- [ ] Responsive to requests

### 6.3 Startup Time
```powershell
# Measure startup time
$start = Get-Date
.\ollama.exe serve | Read-Host
$end = Get-Date
$duration = $end - $start
Write-Host "Startup time: $($duration.TotalSeconds) seconds"

# Acceptable: < 5 seconds
```

**Validation**:
- [ ] Startup time reasonable (< 5s)
- [ ] Fast initialization
- [ ] No unnecessary delays

### 6.4 Response Time
```powershell
# Measure API response time
Measure-Command {
    curl http://localhost:7860/api/health -UseBasicParsing
} | Select-Object TotalMilliseconds

# Acceptable: < 100 ms
```

**Validation**:
- [ ] Health check responds < 100ms
- [ ] Discovery endpoint responsive
- [ ] No latency issues
- [ ] Concurrent requests handled

---

## Phase 7: Stress & Edge Cases

### 7.1 Concurrent Requests
```powershell
# Test with concurrent requests
1..10 | ForEach-Object {
    $jobParams = @{
        ScriptBlock = {
            curl http://localhost:7860/api/health -UseBasicParsing
        }
    }
    Start-Job @jobParams
} | Wait-Job | Receive-Job

# Should all complete without errors
```

**Validation**:
- [ ] Handles concurrent requests
- [ ] No race conditions
- [ ] Thread-safe operations
- [ ] No dropped requests

### 7.2 Long Running Server
```powershell
# Run server for extended period (30+ minutes)
# Monitor for:
# - Memory leaks
# - Handle leaks
# - CPU creep
# - Responsiveness degradation

# Keep checking health:
while ($true) {
    $health = curl http://localhost:7860/api/health -UseBasicParsing 2>$null
    Write-Host "$(Get-Date): $($health.StatusCode)"
    Start-Sleep -Seconds 60
}
```

**Validation**:
- [ ] Server stays responsive
- [ ] Memory usage stable
- [ ] Handles stable
- [ ] No degradation over time

### 7.3 Port Already in Use
```powershell
# Test when port 7860 is occupied
# Run something else on port 7860
$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://+:7860/")
$listener.Start()

# Try to start ollama.exe serve
# Should gracefully handle error

$listener.Stop()
```

**Validation**:
- [ ] Error message clear
- [ ] Suggests alternative port
- [ ] No crash
- [ ] Helpful error guidance

### 7.4 Service Down Scenarios
```powershell
# Test when Ollama is not running (port 11434 closed)
# Test when Orchestrator is not running (port 61683 closed)
# Server should:
# - Still start successfully
# - Report services as unavailable
# - Continue operating with fallback

.\ollama.exe serve
# Check /api/health → services.ollama: false
```

**Validation**:
- [ ] Server handles missing services
- [ ] Health endpoint accurate
- [ ] Continues with reduced functionality
- [ ] Can recover when services restart

---

## Phase 8: Cross-Platform Compatibility

### 8.1 Build on Multiple Systems
```powershell
# Test building from different Windows versions
# Windows 10 (latest)
# Windows 11
# Windows Server 2022 (if available)
```

**Validation**:
- [ ] Builds on Windows 10
- [ ] Builds on Windows 11
- [ ] No platform-specific errors
- [ ] Binary compatible across versions

### 8.2 Run on Different Hardware
```powershell
# Test on:
# - Intel processors
# - AMD processors
# - Laptop
# - Desktop
# - Virtual machine
```

**Validation**:
- [ ] Runs on Intel
- [ ] Runs on AMD
- [ ] Runs on laptops
- [ ] Runs on VMs
- [ ] No architecture-specific issues

---

## Defect Tracking

### Format for Issues Found

When you encounter issues, document them with:

```
**Issue Title**: [Descriptive title]

**Severity**:
- Critical (blocks functionality)
- High (degrades functionality)
- Medium (workaround available)
- Low (cosmetic/minor)

**Steps to Reproduce**:
1. Step 1
2. Step 2
3. ...

**Expected Result**:
[What should happen]

**Actual Result**:
[What actually happened]

**Environment**:
- Windows Version: [10/11 + build number]
- Go Version: [output of `go version`]
- Processor: [Intel/AMD model]
- RAM: [GB available]

**Workaround**:
[If applicable]

**Screenshots/Logs**:
[Error output, logs, screenshots]
```

---

## Testing Sign-Off

When complete, verify all checkboxes:

### Critical Path ✅
- [ ] Build successful
- [ ] Binary executes
- [ ] Tests pass 100%
- [ ] Server starts
- [ ] Health check works

### Recommended ✅
- [ ] All Windows features tested
- [ ] Performance acceptable
- [ ] Long-running test passed
- [ ] Stress test passed
- [ ] Web interface works

### Optional ✅
- [ ] Multi-platform tested
- [ ] Edge cases handled
- [ ] Documentation reviewed
- [ ] Examples tested
- [ ] Troubleshooting guide used

---

## Sign-Off

```
Tester Name: _________________________
Date: _____________________
Windows Version: _____________________
Go Version: _____________________
Build Version: _____________________

All critical tests passed: [ ] Yes [ ] No

Issues Found: [ ] Yes [ ] No
If yes, count: _____

Ready for Release: [ ] Yes [ ] No

Notes:
_____________________________________________
_____________________________________________
_____________________________________________
```

---

## Next Steps After Testing

1. **All tests pass**: Proceed to release preparation
2. **Issues found**: Create GitHub issues with details
3. **Requests for changes**: Document feature requests
4. **Performance concerns**: Note for optimization

---

**Last Updated**: February 23, 2026
**Version**: 1.0
**Branch**: claude/ollama-windows-port-solnj
