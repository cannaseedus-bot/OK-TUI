# Windows Test & Validation Checklist

**Version**: 1.0.0-rc1
**Branch**: `claude/ollama-windows-port-solnj`
**Date**: February 23, 2026
**Status**: Ready for Windows 10/11 Testing

---

## Quick Start Testing (5 minutes)

### Prerequisites
- [ ] Windows 10 (Build 19041+) or Windows 11
- [ ] Go 1.24.7+ installed
- [ ] Git installed
- [ ] PowerShell 7.0+ (or 5.1+)
- [ ] Administrator access (optional, for Registry tests)

### Build & Run
```powershell
# 1. Clone repository
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
git checkout claude/ollama-windows-port-solnj

# 2. Build binary
.\build-windows.ps1

# 3. Run server (in one PowerShell window)
.\ollama.exe serve

# 4. Test in another PowerShell window
curl http://localhost:7860/api/health | ConvertFrom-Json | Format-List
```

---

## Phase 1: Build Validation

### 1.1 Build Environment
- [ ] Windows 10 Build 19041+
- [ ] Windows 11 (any version)
- [ ] Go version matches (1.24.7+)
- [ ] Git installed and configured
- [ ] PowerShell 7.0+ available

**Validation Commands**:
```powershell
# Check Windows version
[System.Environment]::OSVersion.VersionString

# Check Go version
go version

# Check Git version
git --version

# Check PowerShell version
$PSVersionTable.PSVersion
```

### 1.2 Build Success
- [ ] `.\build-windows.ps1` completes without errors
- [ ] `ollama.exe` binary created (< 50 MB expected)
- [ ] Binary is executable (can run `.\ollama.exe version`)
- [ ] No compilation warnings or errors

**Expected Output**:
```
✅ Build successful!
  Binary: .\ollama.exe
  Size: ~40 MB
  Build time: ~15 seconds
```

### 1.3 Debug Build
- [ ] `.\build-windows.ps1 -BuildType Debug` completes
- [ ] Debug binary created (> 100 MB expected)
- [ ] Can be debugged with delve if installed

**Test**:
```powershell
.\build-windows.ps1 -BuildType Debug
.\ollama-debug.exe version
```

### 1.4 Cross-Build Validation
- [ ] Can cross-compile on Linux: `GOOS=windows GOARCH=amd64 go build`
- [ ] Can cross-compile on macOS: `GOOS=windows GOARCH=amd64 go build`
- [ ] Resulting binary works on Windows

---

## Phase 2: Unit Tests

### 2.1 Platform Abstraction Tests
```powershell
go test -v -timeout 30s ./kuhul/runtime/platform_abstraction_test.go
```

**Tests to Pass**:
- [ ] TestPathHandling
- [ ] TestEnvironmentVariables
- [ ] TestFileOperations
- [ ] TestProcessExecution
- [ ] TestCommandExists
- [ ] TestNetworkPortManagement
- [ ] TestServiceDiscovery
- [ ] TestBuiltinIntegration
- [ ] TestCrossPlatformConsistency

**Expected**: All tests pass (0.05s typical)

### 2.2 HTTP Bridge Tests
```powershell
go test -v -timeout 30s ./server/bridge_test.go
```

**Tests to Pass**:
- [ ] TestBridgeConfiguration
- [ ] TestServiceDiscoveryEndpoint
- [ ] TestHealthCheckEndpoint
- [ ] TestInferenceProxy
- [ ] TestFallbackLogic

**Expected**: All tests pass (0.03s typical)

### 2.3 Runtime Tests
```powershell
go test -v -timeout 60s ./kuhul/runtime/...
```

**Expected**: All tests pass (0.05s total)

### 2.4 Full Test Suite
```powershell
go test -v -timeout 120s ./...
```

**Expected**: 100% pass rate

---

## Phase 3: Feature Tests

### 3.1 Path Handling Tests

**Test 1: Windows Path Normalization**
```powershell
# Create test script
@'
package main
import (
    "fmt"
    "github.com/ollama/ollama/kuhul/runtime"
)
func main() {
    tests := []string{
        `C:\Users\Admin\file.txt`,
        `C:/Users/Admin/file.txt`,
        `\\server\share\file.txt`,
        `..\relative\path`,
    }
    for _, path := range tests {
        normalized := runtime.NormalizePath(path)
        fmt.Printf("%s → %s\n", path, normalized)
    }
}
'@ > test_path.go

go run test_path.go
```

**Expected Output**:
```
C:\Users\Admin\file.txt → C:\Users\Admin\file.txt
C:/Users/Admin/file.txt → C:\Users\Admin\file.txt
\\server\share\file.txt → \\server\share\file.txt
..\relative\path → ..\relative\path
```

**Test 2: Environment Variables**
```powershell
# Test home directory detection
go run -run TestEnvironmentVariables ./kuhul/runtime/...

# Verify output shows:
# USERPROFILE: C:\Users\YourName
# APPDATA: C:\Users\YourName\AppData\Roaming
# LOCALAPPDATA: C:\Users\YourName\AppData\Local
# TEMP: C:\Users\YourName\AppData\Local\Temp
```

### 3.2 Process Execution Tests

**Test 1: Simple Command Execution**
```powershell
# Test notepad launch
go run test_proc.go  # Runs sys.proc_run("notepad.exe")

# Verify: Notepad opens
```

**Test 2: PowerShell Command**
```powershell
# Test PowerShell execution
go run test_ps.go  # Runs Get-Process command

# Verify: Process list returned
```

**Test 3: Command with Arguments**
```powershell
# Test echo command
go run test_echo.go  # Runs cmd.exe /c echo test

# Verify: Output contains "test"
```

### 3.3 Network & Port Tests

**Test 1: Port Availability**
```powershell
go test -v -run TestPortAvailability ./kuhul/runtime/...

# Tests:
# - Port 65432 available
# - Port 80 (in use or not available)
# - Random ephemeral port
```

**Test 2: Service Discovery**
```powershell
# Start services on different ports:
# Ollama: 11434 (if available)
# Orchestrator: 61683 (if available)

go test -v -run TestServiceDiscovery ./kuhul/runtime/...

# Verify: Services detected correctly
```

**Test 3: Port-to-Process Mapping**
```powershell
go test -v -run TestPortToProcess ./kuhul/runtime/...

# Verify: Can identify process using port
```

### 3.4 Registry Tests (Admin Required)

**Test 1: Registry Read (No Admin)**
```powershell
# Read user registry (HKCU) - works without admin
go test -v -run TestRegistryRead ./kuhul/runtime/...

# Verify: Can read user settings
```

**Test 2: Registry Write (Admin)**
```powershell
# Run as Administrator
Start-Process powershell -Verb RunAs -ArgumentList "cd $PWD; go test -v -run TestRegistryWrite ./kuhul/runtime/..."

# Verify: Can write registry values
```

**Test 3: Registry Enumeration**
```powershell
go test -v -run TestRegistryEnum ./kuhul/runtime/...

# Verify: Can list registry hives and keys
```

---

## Phase 4: Server Functionality Tests

### 4.1 Start Server
```powershell
# Terminal 1: Start server
.\ollama.exe serve

# Expected output:
# [INFO] Ollama-K bridge starting...
# [INFO] Listening on http://localhost:7860
```

### 4.2 Health Check
```powershell
# Terminal 2: Check health
curl http://localhost:7860/api/health | ConvertFrom-Json | Format-List

# Expected output:
# status    : healthy
# services  : @{ollama=; orchestrator=}
# last_check: 2026-02-23T...
```

### 4.3 Service Discovery
```powershell
curl http://localhost:7860/api/services/discover | ConvertFrom-Json | Format-List

# Expected output:
# ollama_url       : http://localhost:11434
# orchestrator_url : http://localhost:61683 (or null)
# last_discovery   : 2026-02-23T...
# timestamp        : 2026-02-23T...
```

### 4.4 Inference Proxy (if Ollama running)
```powershell
# If Ollama is running on 11434:
$body = @{
    model = "llama2"
    xjson = "..."
} | ConvertTo-Json

curl -X POST http://localhost:7860/api/proxy/infer `
  -ContentType "application/json" `
  -Body $body | ConvertFrom-Json
```

### 4.5 Custom Port Configuration
```powershell
# Test custom port
$env:OLLAMA_PORT = "8080"
.\ollama.exe serve

# In another terminal:
curl http://localhost:8080/api/health

# Verify: Server runs on 8080
```

### 4.6 Multiple Instances
```powershell
# Terminal 1: First instance on 7860
.\ollama.exe serve

# Terminal 2: Second instance on 7861
$env:OLLAMA_PORT = "7861"
.\ollama.exe serve

# Terminal 3: Test both
curl http://localhost:7860/api/health
curl http://localhost:7861/api/health

# Verify: Both instances respond
```

---

## Phase 5: Windows-Specific Features

### 5.1 Path Handling (Windows)
- [ ] Backslash paths work correctly
- [ ] Forward slash paths auto-convert
- [ ] UNC paths (`\\server\share`) work
- [ ] Relative paths resolve correctly
- [ ] Environment variables expand (`%USERPROFILE%`)

**Test Script**:
```powershell
$paths = @(
    "C:\Users\Admin\file.txt",
    "C:/Users/Admin/file.txt",
    "\\server\share\file.txt",
    "%USERPROFILE%\AppData"
)

foreach ($path in $paths) {
    $normalized = & go run path_test.go $path
    Write-Host "$path → $normalized"
}
```

### 5.2 Process Execution (Windows)
- [ ] .exe extension handling (cmd vs cmd.exe)
- [ ] PowerShell command execution
- [ ] Process enumeration works
- [ ] Process termination works
- [ ] Command line parsing with quotes

### 5.3 Environment Variables
- [ ] USERPROFILE returns home directory
- [ ] APPDATA returns roaming app data
- [ ] LOCALAPPDATA returns local app data
- [ ] TEMP returns temp directory
- [ ] PATH resolution works

### 5.4 Registry Access (with Admin)
- [ ] HKCU registry read
- [ ] HKCU registry write
- [ ] HKLM registry read (if accessible)
- [ ] Registry hive enumeration
- [ ] Type conversion (string, DWORD)

### 5.5 Port Management
- [ ] Port availability checking
- [ ] Ephemeral port detection
- [ ] Port-to-process lookup
- [ ] Multiple port handling

---

## Phase 6: Performance Testing

### 6.1 Startup Time
```powershell
$startTime = Get-Date
.\ollama.exe serve &
$serverPid = $?
Start-Sleep -Seconds 2

$startupTime = (Get-Date) - $startTime
Write-Host "Startup time: $($startupTime.TotalMilliseconds)ms"

# Expected: < 500ms
```

### 6.2 Memory Usage
```powershell
$process = Get-Process ollama
$memMB = $process.WorkingSet / 1MB
Write-Host "Memory usage: ${memMB}MB"

# Expected: 50-200MB at rest
```

### 6.3 Request Latency
```powershell
$times = @()
for ($i = 0; $i -lt 100; $i++) {
    $start = Get-Date
    curl http://localhost:7860/api/health | Out-Null
    $elapsed = (Get-Date) - $start
    $times += $elapsed.TotalMilliseconds
}

$avg = ($times | Measure-Object -Average).Average
$max = ($times | Measure-Object -Maximum).Maximum
Write-Host "Avg latency: ${avg}ms, Max: ${max}ms"

# Expected: < 50ms average
```

### 6.4 Concurrent Requests
```powershell
# Run 10 concurrent requests
$jobs = @()
for ($i = 0; $i -lt 10; $i++) {
    $jobs += Start-Job {
        curl http://localhost:7860/api/health
    }
}

$results = $jobs | Wait-Job | Receive-Job
$successCount = ($results | Measure-Object).Count

Write-Host "Success: $successCount / 10"
# Expected: All 10 succeed
```

---

## Phase 7: Error Handling & Edge Cases

### 7.1 Port Already in Use
```powershell
# Start server twice on same port
.\ollama.exe serve &
.\ollama.exe serve

# Expected: Second instance fails gracefully
# Error message: "Address already in use" or similar
```

### 7.2 Missing Dependencies
```powershell
# Move Go installation temporarily
# Try to run build
.\build-windows.ps1

# Expected: Clear error message about missing Go
```

### 7.3 Invalid Configuration
```powershell
# Set invalid port
$env:OLLAMA_PORT = "99999"
.\ollama.exe serve

# Expected: Error about invalid port
```

### 7.4 Service Unavailable
```powershell
# Stop Ollama if running
# Start bridge server
.\ollama.exe serve

# Check health
curl http://localhost:7860/api/health

# Expected: Shows "ollama": false (service unavailable)
```

### 7.5 Path Traversal Protection
```powershell
# Test path traversal
$body = @{
    path = "../../etc/passwd"
} | ConvertTo-Json

curl -X POST http://localhost:7860/api/files `
  -ContentType "application/json" `
  -Body $body

# Expected: 403 Forbidden or similar
```

---

## Phase 8: Cleanup & Validation

### 8.1 Stop Server
```powershell
# Find ollama process
Get-Process ollama

# Terminate gracefully (Ctrl+C)
# Or force kill if needed:
Stop-Process -Name ollama -Force

# Verify stopped
Get-Process ollama -ErrorAction SilentlyContinue
# Should return nothing
```

### 8.2 Cleanup Temporary Files
```powershell
# Remove build artifacts
Remove-Item ollama.exe
Remove-Item ollama-debug.exe
Remove-Item test_*.go
Remove-Item coverage.* -ErrorAction SilentlyContinue
```

### 8.3 Verify No Corruption
```powershell
# Check git status (should be clean)
git status

# Verify repository integrity
git fsck --full
```

---

## Test Results Template

### System Information
```
OS: Windows [10/11]
Build: [Build number]
Go Version: [Version]
Architecture: x86-64
Processor: [Model]
RAM: [Amount]
Disk: [Free Space]
```

### Build Results
```
✅ Release Build: [PASS/FAIL]
✅ Debug Build: [PASS/FAIL]
✅ Cross-Build: [PASS/FAIL]
```

### Unit Test Results
```
✅ Platform Abstraction: [PASS/FAIL]
✅ HTTP Bridge: [PASS/FAIL]
✅ Runtime Tests: [PASS/FAIL]
✅ Full Suite: [PASS/FAIL]
```

### Feature Tests
```
✅ Path Handling: [PASS/FAIL]
✅ Process Execution: [PASS/FAIL]
✅ Network/Ports: [PASS/FAIL]
✅ Registry Access: [PASS/FAIL]
```

### Server Tests
```
✅ Start Server: [PASS/FAIL]
✅ Health Check: [PASS/FAIL]
✅ Service Discovery: [PASS/FAIL]
✅ Inference Proxy: [PASS/FAIL]
```

### Performance
```
✅ Startup Time: [___]ms (target: < 500ms)
✅ Memory Usage: [___]MB (target: 50-200MB)
✅ Request Latency: [___]ms (target: < 50ms)
✅ Concurrent Requests: [___]/10 (target: 10/10)
```

---

## Known Issues & Workarounds

### Issue: Port Already in Use
**Symptoms**: `Address already in use` error
**Solution**:
```powershell
netstat -ano | Select-String ":7860"
taskkill /PID <PID> /F
```

### Issue: Registry Access Denied
**Symptoms**: `Access denied` when reading HKLM
**Solution**: Run PowerShell as Administrator or use HKCU instead

### Issue: Path with Spaces
**Symptoms**: Command execution fails with paths containing spaces
**Solution**: Paths are auto-quoted; if issue persists, use forward slashes

### Issue: Slow Startup
**Symptoms**: Takes > 1000ms to start
**Solution**: Check disk I/O, close other applications

---

## Sign-Off Checklist

### Testing Completed
- [ ] Build validation passed
- [ ] Unit tests passed
- [ ] Feature tests passed
- [ ] Server functionality verified
- [ ] Performance acceptable
- [ ] Error handling verified
- [ ] Windows-specific features working
- [ ] No regressions detected

### Quality Gates
- [ ] Code follows Go conventions
- [ ] All tests pass
- [ ] No race conditions
- [ ] Memory usage acceptable
- [ ] Performance meets targets
- [ ] Documentation complete
- [ ] Ready for release

### Sign-Off
- **Tester**: _________________
- **Date**: _________________
- **Windows Version**: _________________
- **Notes**: _________________

---

## Next Steps

1. **Validation**: Execute all tests on Windows 10/11
2. **Bug Fixes**: Report and fix any issues
3. **Release**: Proceed to v1.0.0 release
4. **Users**: Deploy to production

---

**Branch**: `claude/ollama-windows-port-solnj`
**Version**: 1.0.0-rc1
**Ready for**: Windows 10/11 Validation
