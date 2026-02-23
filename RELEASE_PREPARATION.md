# v1.0 Windows Release Preparation Guide

**Project**: Ollama-K Windows Port (Phase 1-2)
**Target Release Date**: [To be determined after testing]
**Release Branch**: `claude/ollama-windows-port-solnj`

---

## Pre-Release Checklist

### Code Quality
- [ ] All tests passing (100%)
- [ ] No TODO/FIXME comments left in code
- [ ] Code formatted with `go fmt`
- [ ] Code vetted with `go vet`
- [ ] Dependencies verified with `go mod verify`
- [ ] No race conditions detected
- [ ] No compiler warnings
- [ ] No test warnings

### Windows Validation
- [ ] Tested on Windows 10 (latest build)
- [ ] Tested on Windows 11
- [ ] All platform abstraction features verified
- [ ] Web interface fully functional
- [ ] Service discovery working
- [ ] Performance acceptable
- [ ] Long-running stability confirmed
- [ ] Stress tests passed

### Documentation
- [ ] Windows Build Guide complete
- [ ] Windows Testing Guide complete
- [ ] Changelog generated
- [ ] Release notes written
- [ ] API documentation updated
- [ ] Known limitations documented
- [ ] Installation guide ready
- [ ] Troubleshooting guide complete

### Release Artifacts
- [ ] Binary built (`ollama.exe`)
- [ ] Debug symbols available
- [ ] Checksums calculated
- [ ] License headers included
- [ ] Version strings set
- [ ] Build metadata included

---

## Version Management

### Version Numbering

Following semantic versioning: `MAJOR.MINOR.PATCH`

**Current**: 1.0.0 (Windows Port Release)

```
1 = Major version (Windows port complete)
0 = Minor version (features)
0 = Patch version (bug fixes)
```

### Version String Update

1. **Update in code**:
```bash
# Edit version file
vim cmd/ollama/main.go

# Set:
const Version = "1.0.0-windows"
const BuildDate = "2026-02-23"
const Commit = "$(git rev-parse --short HEAD)"
```

2. **Create Git tag**:
```bash
git tag -a v1.0.0-windows -m "Windows Port Release v1.0.0"
git push origin v1.0.0-windows
```

3. **Update in build script**:
```bash
# build-windows.ps1 should embed:
-X main.Version=1.0.0-windows
```

---

## Release Notes Template

Create file: `RELEASE_NOTES_v1.0.0.md`

```markdown
# Ollama-K v1.0.0 - Windows Release

**Release Date**: [Date]
**Download**: [GitHub Releases URL]
**Branch**: claude/ollama-windows-port-solnj

## 🎉 What's New

### Major Features
- ✅ Full Windows 10/11 support
- ✅ Platform abstraction layer
- ✅ HTTP bridge server (port 7860)
- ✅ Service discovery (Ollama + Orchestrator)
- ✅ K'UHUL language integration
- ✅ 40+ system builtin functions

### Windows-Specific Capabilities
- ✅ Windows path handling (backslash/UNC)
- ✅ Process execution with .exe support
- ✅ Windows Registry access
- ✅ Port management and service lookup
- ✅ Environment variable handling
- ✅ Native Windows process management

### Server Features
- ✅ Progressive Web App (PWA) interface
- ✅ Health checking
- ✅ Auto-discovery of Ollama/Orchestrator
- ✅ XJSON inference proxy
- ✅ Thread-safe configuration
- ✅ Performance optimized

## 🔧 Installation

### System Requirements
- Windows 10 (Build 19041+) or Windows 11
- Go 1.24.7+ (for building from source)
- 4 GB RAM minimum (8 GB recommended)
- 500 MB disk space

### Quick Start

```powershell
# Download release
Invoke-WebRequest -Uri "https://github.com/.../ollama.exe" -OutFile ollama.exe

# Run server
.\ollama.exe serve

# Open web interface
Start-Process "http://localhost:7860"
```

### From Source

```powershell
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
git checkout claude/ollama-windows-port-solnj
.\build-windows.ps1
.\ollama.exe serve
```

## 📋 Documentation

- [Windows Build Guide](WINDOWS_BUILD_GUIDE.md) - Complete build instructions
- [Windows Testing Guide](WINDOWS_TESTING_GUIDE.md) - Validation procedures
- [Windows Port Status](WINDOWS_PORT_STATUS.md) - Implementation details
- [K'UHUL Architecture](KUHUL_CLI_ARCHITECTURE.md) - TUI design

## 🐛 Known Limitations

### Phase 1-2 (Current)
- No GPU acceleration (Phase 3)
- No MLIR/LLVM compiler (Phase 3)
- Web-only interface (CLI TUI coming)
- Service discovery requires Ollama/Orchestrator on same network

### Planned for Later
- Phase 3: MLIR/LLVM compiler layer
- Phase 4: GPU/CUDA acceleration
- Phase 5: Windows Installer (.msi)
- Phase 6: CLI with TUI interface

## 🔄 Upgrading

From previous version:

```powershell
# Backup current binary
Copy-Item ollama.exe ollama-backup.exe

# Download new version
Invoke-WebRequest -Uri "..." -OutFile ollama-new.exe

# Replace binary
Remove-Item ollama.exe
Rename-Item ollama-new.exe ollama.exe

# Restart server
.\ollama.exe serve
```

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/cannaseedus-bot/Ollama-K/issues)
- **Discussions**: [GitHub Discussions](https://github.com/cannaseedus-bot/Ollama-K/discussions)
- **Documentation**: [WINDOWS_BUILD_GUIDE.md](WINDOWS_BUILD_GUIDE.md)

## 📜 License

[See LICENSE file]

## ✅ Testing & Validation

This release has been tested on:
- Windows 10 (Build [number])
- Windows 11 (Build [number])
- Intel and AMD processors
- Multiple network configurations

All platform abstraction features verified:
- ✅ Path handling
- ✅ Process execution
- ✅ Network operations
- ✅ Registry access
- ✅ Service discovery
- ✅ Health monitoring

## 🙏 Special Thanks

Thanks to the contributors and testers who made this Windows release possible!

---

**Release Created**: [Date]
**Release Manager**: [Name]
**Status**: ✅ Ready for Distribution
```

---

## Changelog Generation

Create file: `CHANGELOG_v1.0.0.md`

```markdown
# Changelog - v1.0.0

## [1.0.0] - Windows Release - 2026-02-23

### Added - Phase 1: Platform Abstraction Layer

#### Path & File System (1.1)
- Added Windows path normalization (backslash/forward slash)
- Added UNC path support (\\server\share)
- Added drive letter extraction
- Added absolute vs relative path detection
- Added environment variable expansion
- Added cross-platform path joining

#### Process Execution (1.2)
- Added command execution with automatic .exe extension
- Added shell command support via cmd.exe /c
- Added process enumeration via PowerShell
- Added process termination via taskkill
- Added command line parsing with proper quote handling
- Added shell environment inheritance

#### Network & Port Management (1.3)
- Added port availability checking
- Added Ollama service discovery (default 11434)
- Added Orchestrator service discovery (default 61683)
- Added port-to-process mapping via netstat
- Added health checking with configurable timeout
- Added service endpoint detection

#### Windows Registry Access (1.4)
- Added registry value read/write/delete operations
- Added hive enumeration (HKCR, HKCU, HKLM, HKU, HKCC)
- Added registry subkey enumeration
- Added type conversion (string, DWORD, multi-string)
- Added Windows-only implementation with Unix stubs

### Added - Phase 2: PWA Bridge & Server

#### HTTP Bridge (2.1)
- Added HTTP server on port 7860
- Added service discovery endpoint (GET /api/services/discover)
- Added health check endpoint (GET /api/health)
- Added XJSON inference proxy (POST /api/proxy/infer)
- Added orchestrator-first fallback logic
- Added thread-safe configuration management

#### PWA Interface (2.2)
- Added Progressive Web App interface
- Added real-time service status display
- Added health monitoring dashboard
- Added responsive design for mobile
- Added offline capability

### Added - K'UHUL Language Integration

#### 40+ System Builtins
- sys.path_normalize() - Cross-platform path normalization
- sys.path_join() - Platform-aware path joining
- sys.proc_run() - Execute external processes
- sys.proc_list() - Enumerate running processes
- sys.port_available() - Check port availability
- sys.registry_get() - Read Windows registry
- sys.registry_set() - Write to Windows registry
- ... and 33 more platform abstraction functions

### Added - Build Infrastructure

#### GitHub Actions CI/CD
- Added automatic Windows build pipeline
- Added test execution on Windows runner
- Added code formatting validation
- Added test coverage reporting
- Added artifact uploading
- Added build status notifications

#### PowerShell Build Script
- Added one-command build automation
- Added Release and Debug build modes
- Added pre-build validation
- Added automatic testing
- Added binary verification

### Added - Documentation

- Added WINDOWS_BUILD_GUIDE.md (650+ lines)
- Added WINDOWS_TESTING_GUIDE.md (400+ lines)
- Added RELEASE_PREPARATION.md (this file)
- Added K'UHUL CLI Architecture documentation
- Added Windows Port Roadmap
- Added Windows Port Status Report

### Changed

- Refactored common functions into shared libraries
- Updated build system for Windows support
- Enhanced error messages for Windows-specific issues
- Improved logging for service discovery
- Updated test utilities for cross-platform compatibility

### Fixed

- Fixed path separator handling for Windows
- Fixed process execution with special characters
- Fixed environment variable resolution
- Fixed registry access permissions
- Fixed concurrent service discovery calls

### Deprecated

- None in v1.0.0

### Removed

- None in v1.0.0

### Security

- Added input validation for registry operations
- Added path traversal prevention
- Added command injection prevention
- Added safe process execution
- Added environment variable filtering

### Performance

- Optimized path normalization (cached results)
- Optimized process enumeration (parallel lookups)
- Optimized service discovery (timeout handling)
- Optimized memory usage (reduced allocations)
- Optimized startup time (lazy initialization)

### Testing

- Added 381 lines of platform abstraction tests
- Added 317 lines of HTTP bridge tests
- Added Windows-specific feature tests
- Added performance benchmarks
- Added stress tests
- Added edge case handling

---

## Release Artifacts

### Binary Files

```
ollama.exe                  [Main Windows executable]
ollama-debug.exe           [Debug version with symbols]
ollama-1.0.0-windows.zip   [Release package]
```

### Checksums

```
SHA256 (ollama.exe): [generate with: certutil -hashfile ollama.exe SHA256]
SHA256 (ollama-debug.exe): [same process]
```

### Archive Contents

```
ollama-1.0.0-windows.zip/
├── ollama.exe
├── README.md
├── WINDOWS_BUILD_GUIDE.md
├── WINDOWS_TESTING_GUIDE.md
├── CHANGELOG.md
├── LICENSE
└── CHECKSUMS.txt
```

---

## Distribution Channels

### 1. GitHub Releases

```bash
# Create release on GitHub
# Upload files:
# - ollama.exe
# - ollama-debug.exe
# - RELEASE_NOTES_v1.0.0.md
# - CHANGELOG.md
# - checksums.txt
```

### 2. Direct Download

```
https://github.com/cannaseedus-bot/Ollama-K/releases/download/v1.0.0-windows/ollama.exe
```

### 3. Package Managers

Plan for future:
- Windows Package Manager (winget)
- Chocolatey
- Scoop
- Microsoft Store

---

## Post-Release Activities

### Immediate (Day 1)
- [ ] Publish GitHub Release
- [ ] Announce on GitHub Discussions
- [ ] Announce on social media
- [ ] Monitor for issues
- [ ] Respond to initial feedback

### Short Term (Week 1)
- [ ] Collect user feedback
- [ ] Track download statistics
- [ ] Monitor for bugs
- [ ] Prepare patch if needed
- [ ] Document common issues

### Medium Term (Weeks 2-4)
- [ ] Plan Phase 3 implementation
- [ ] Create Windows installer
- [ ] Prepare v1.0.1 (if bugs found)
- [ ] Begin GPU support research

---

## Release Validation Checklist

Before publishing:

### Code Quality ✅
- [ ] `go fmt ./...` - No formatting issues
- [ ] `go vet ./...` - No vet warnings
- [ ] `go test ./...` - All tests pass
- [ ] `go test -race ./...` - No race conditions

### Windows Testing ✅
- [ ] Windows 10 validation complete
- [ ] Windows 11 validation complete
- [ ] All platform features verified
- [ ] Performance benchmarks met
- [ ] Stability testing (30+ min) passed
- [ ] Stress testing passed
- [ ] Service discovery working
- [ ] Web interface fully functional

### Documentation ✅
- [ ] README updated
- [ ] Changelog complete
- [ ] Release notes written
- [ ] Build guide reviewed
- [ ] Testing guide reviewed
- [ ] Known limitations documented
- [ ] Upgrade instructions clear

### Artifacts ✅
- [ ] Binary built (`ollama.exe`)
- [ ] Debug binary built (`ollama-debug.exe`)
- [ ] Checksums calculated
- [ ] Zip file created
- [ ] GitHub release prepared
- [ ] All files present

---

## Release Timeline

```
Day -5: Testing complete
Day -3: Documentation final review
Day -1: Final validation
Day 0:  Publish release
Day +1: Monitor for issues
Day +7: Collect feedback
Week 2: Plan Phase 3
```

---

## Sign-Off

```
Release Manager: ________________________
Date: _________________________
Version: 1.0.0-windows
Status: ✅ Ready for Release

Final Approval:
[ ] All items complete
[ ] No blockers
[ ] Ready to publish
```

---

**Last Updated**: February 23, 2026
**Version**: 1.0 (Release Preparation)
**Status**: Ready for Windows Testing Phase
