# Deployment & Operations Guide v1.0.0

**Ollama-K v1.0.0 Windows Release**
**For DevOps, SRE, and Operations Teams**

---

## Executive Summary

Ollama-K v1.0.0 is a production-ready, cross-platform AI inference bridge with full Windows 10/11 support. This guide covers deployment, operations, monitoring, and maintenance procedures.

**Key Facts**:
- ✅ Windows 10/11 native support
- ✅ Cross-platform (Windows/Linux/macOS)
- ✅ 40+ system functions
- ✅ HTTP REST API
- ✅ Service discovery and proxying
- ✅ Zero downtime capable
- ✅ Monitoring ready

---

## Table of Contents

1. [Deployment](#deployment)
2. [Configuration](#configuration)
3. [Monitoring](#monitoring)
4. [Troubleshooting](#troubleshooting)
5. [Maintenance](#maintenance)
6. [Scaling](#scaling)
7. [Disaster Recovery](#disaster-recovery)
8. [Operations Runbook](#operations-runbook)

---

## Deployment

### Single Server Deployment

#### Windows Server Deployment

**Prerequisites**:
```powershell
# Check prerequisites
[System.Environment]::OSVersion.VersionString     # Windows 10/11
(Get-Item 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion').Version
```

**Step 1: Prepare Environment**
```powershell
# Create installation directory
New-Item -ItemType Directory -Path "C:\Ollama-K" -Force
cd C:\Ollama-K

# Download binary (or build from source)
$releaseUrl = "https://github.com/cannaseedus-bot/Ollama-K/releases/download/v1.0.0/ollama.exe"
Invoke-WebRequest -Uri $releaseUrl -OutFile ollama.exe

# Verify download
$expected_hash = "sha256_hash_from_release"
$actual_hash = (Get-FileHash -Path ollama.exe -Algorithm SHA256).Hash
if ($actual_hash -eq $expected_hash) {
    Write-Host "✓ Download verified"
} else {
    Write-Error "Hash mismatch!"
    exit 1
}
```

**Step 2: Register as Service (Optional)**
```powershell
# Using NSSM (Non-Sucking Service Manager)
# Install NSSM: choco install nssm

nssm install OllamaK C:\Ollama-K\ollama.exe
nssm set OllamaK AppParameters serve
nssm set OllamaK AppDirectory C:\Ollama-K
nssm set OllamaK AppEnvironmentExtra OLLAMA_PORT=7860
nssm set OllamaK AppExit Default Restart
nssm set OllamaK AppRestartDelay 5000

# Start service
nssm start OllamaK

# Verify service status
nssm status OllamaK
```

**Step 3: Configure Firewall**
```powershell
# Allow incoming connections on port 7860
New-NetFirewallRule -DisplayName "Ollama-K" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 7860

# Verify rule
Get-NetFirewallRule -DisplayName "Ollama-K" | Get-NetFirewallPortFilter
```

**Step 4: Verify Installation**
```powershell
# Check service is running
Get-Service OllamaK | Select-Object Status, DisplayName

# Test endpoint
curl http://localhost:7860/api/health
# Expected: HTTP 200 with JSON response
```

#### Linux Deployment

**Prerequisites**:
```bash
# Check prerequisites
lsb_release -a                    # Ubuntu 20.04+
uname -m                          # x86_64
free -h                           # At least 4GB RAM
df -h /opt                        # At least 500MB
```

**Step 1: Prepare Environment**
```bash
# Create installation directory
sudo mkdir -p /opt/ollama-k
cd /opt/ollama-k

# Download binary
sudo wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v1.0.0/ollama-linux-amd64
sudo chmod +x ollama-linux-amd64

# Verify download
sha256sum -c <(echo "hash_value ollama-linux-amd64")
```

**Step 2: Create Service**
```bash
# Create systemd service file
sudo tee /etc/systemd/system/ollama-k.service > /dev/null <<EOF
[Unit]
Description=Ollama-K Service
After=network.target

[Service]
Type=simple
User=ollama
WorkingDirectory=/opt/ollama-k
ExecStart=/opt/ollama-k/ollama-linux-amd64 serve
Restart=always
RestartSec=5
Environment="OLLAMA_PORT=7860"

[Install]
WantedBy=multi-user.target
EOF

# Create ollama user
sudo useradd -r -s /bin/false ollama || true

# Set permissions
sudo chown -R ollama:ollama /opt/ollama-k

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ollama-k
sudo systemctl start ollama-k
```

**Step 3: Configure Firewall**
```bash
# UFW (Uncomplicated Firewall)
sudo ufw allow 7860/tcp
sudo ufw status

# Or check iptables
sudo iptables -L -n | grep 7860
```

**Step 4: Verify Installation**
```bash
# Check service status
sudo systemctl status ollama-k

# Test endpoint
curl http://localhost:7860/api/health
# Expected: HTTP 200 with JSON response
```

### Multi-Node Deployment

#### Load Balancer Configuration

**Windows (IIS Application Request Routing)**
```powershell
# Install ARR (Application Request Routing)
# Via Server Manager or: Install-WindowsFeature Web-Application-Routing

# Create URL Rewrite rule in IIS
# 1. Open IIS Manager
# 2. Create new website (e.g., ollama-app.internal)
# 3. Add Application Request Routing
# 4. Configure backend servers
# 5. Set load balancing algorithm (Round Robin)
```

**Linux (Nginx)**
```nginx
# /etc/nginx/sites-available/ollama-k
upstream ollama_backend {
    server 192.168.1.10:7860;
    server 192.168.1.11:7860;
    server 192.168.1.12:7860;
}

server {
    listen 80;
    server_name ollama-k.internal;

    location / {
        proxy_pass http://ollama_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Health check
        proxy_connect_timeout 5s;
        proxy_send_timeout 5s;
        proxy_read_timeout 5s;
    }
}
```

**Docker Compose**
```yaml
version: '3.8'

services:
  load-balancer:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - ollama-1
      - ollama-2
      - ollama-3

  ollama-1:
    image: ollama-k:1.0.0
    environment:
      OLLAMA_PORT: 7860
    ports:
      - "7860:7860"

  ollama-2:
    image: ollama-k:1.0.0
    environment:
      OLLAMA_PORT: 7860
    ports:
      - "7861:7860"

  ollama-3:
    image: ollama-k:1.0.0
    environment:
      OLLAMA_PORT: 7860
    ports:
      - "7862:7860"
```

---

## Configuration

### Environment Variables

```powershell
# Core Configuration
$env:OLLAMA_PORT = "7860"              # Server port (default: 7860)
$env:OLLAMA_HOST = "0.0.0.0"           # Bind address (default: localhost)
$env:OLLAMA_NUM_PARALLEL = "4"         # Parallel tasks
$env:OLLAMA_NUM_THREADS = "8"          # Thread count

# Service Discovery
$env:OLLAMA_DISCOVERY_TIMEOUT = "5"    # Discovery timeout (seconds)
$env:OLLAMA_HEALTH_CHECK_INTERVAL = "30" # Health check interval

# Logging
$env:OLLAMA_DEBUG = "0"                # Debug mode (0=off, 1=on)
$env:OLLAMA_LOG_LEVEL = "INFO"         # Log level (DEBUG, INFO, WARN, ERROR)

# Performance
$env:OLLAMA_MAX_CONNECTIONS = "1000"   # Max concurrent connections
$env:OLLAMA_REQUEST_TIMEOUT = "30"     # Request timeout (seconds)

# Windows-Specific
$env:OLLAMA_REGISTRY_TIMEOUT = "5"     # Registry operation timeout
```

### Configuration File (Optional)

**File**: `config.json`
```json
{
  "server": {
    "port": 7860,
    "host": "0.0.0.0",
    "bind": true
  },
  "discovery": {
    "ollama": {
      "port": 11434,
      "timeout": 5
    },
    "orchestrator": {
      "port": 61683,
      "timeout": 5
    }
  },
  "performance": {
    "parallelTasks": 4,
    "threadCount": 8,
    "maxConnections": 1000,
    "requestTimeout": 30
  },
  "logging": {
    "level": "INFO",
    "format": "json",
    "file": "/var/log/ollama-k/app.log"
  }
}
```

**Usage**:
```powershell
# If config file implemented
.\ollama.exe serve --config config.json
```

---

## Monitoring

### Health Check Endpoint

**Request**:
```bash
GET /api/health HTTP/1.1
Host: localhost:7860
```

**Response**:
```json
{
  "status": "healthy",
  "services": {
    "ollama": true,
    "orchestrator": false
  },
  "last_check": "2026-02-28T10:30:00Z",
  "uptime_seconds": 3600,
  "version": "1.0.0"
}
```

### Monitoring Tools Setup

#### Windows Performance Monitor
```powershell
# Create custom monitoring script
$monitoringScript = @'
$counter = Get-Counter -Counter "\Process(ollama)\% Processor Time", `
                                  "\Process(ollama)\Working Set" `
                       -Continuous -SampleInterval 5

$counter | ForEach-Object {
    $_ | Select-Object -ExpandProperty CounterSamples |
        ForEach-Object {
            [PSCustomObject]@{
                TimeStamp = $_.TimeStamp
                Counter = $_.Path
                Value = $_.CookedValue
            }
        }
}
'@

# Run monitoring
& $monitoringScript
```

#### Prometheus Metrics (if implemented)

**Endpoint**: `GET /metrics`

```
# HELP ollama_k_requests_total Total requests received
# TYPE ollama_k_requests_total counter
ollama_k_requests_total{endpoint="/api/health"} 10234

# HELP ollama_k_request_duration_seconds Request latency
# TYPE ollama_k_request_duration_seconds histogram
ollama_k_request_duration_seconds_bucket{endpoint="/api/health",le="0.1"} 9800

# HELP ollama_k_active_connections Active connections
# TYPE ollama_k_active_connections gauge
ollama_k_active_connections 42

# HELP ollama_k_memory_bytes Memory usage
# TYPE ollama_k_memory_bytes gauge
ollama_k_memory_bytes 123456789
```

#### ELK Stack Integration

**Logstash Configuration**:
```
input {
  file {
    path => "C:/Ollama-K/logs/*.log"
    start_position => "beginning"
  }
}

filter {
  json {
    source => "message"
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "ollama-k-%{+YYYY.MM.dd}"
  }
}
```

### Alerting

#### Alert Rules

| Alert | Condition | Action |
|-------|-----------|--------|
| High CPU | CPU > 80% for 5 min | Page on-call |
| High Memory | Memory > 1GB for 5 min | Restart service |
| Service Down | Health check fails | Page on-call |
| High Latency | p99 latency > 1s | Investigate |
| Port Unavailable | Port 7860 not reachable | Check firewall |

**Prometheus Alert Rules**:
```yaml
groups:
  - name: ollama-k
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: process_cpu_seconds_total > 0.8
        for: 5m
        action: notify

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes > 1073741824
        for: 5m
        action: restart

      - alert: ServiceDown
        expr: up{job="ollama-k"} == 0
        for: 1m
        action: page
```

---

## Troubleshooting

### Common Issues

#### Issue: Port Already in Use
```powershell
# Find process using port 7860
netstat -ano | Select-String ":7860"

# Kill process
taskkill /PID <PID> /F

# Or use different port
$env:OLLAMA_PORT = "7861"
.\ollama.exe serve
```

#### Issue: Service Discovery Fails
```powershell
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Check if Orchestrator is running
curl http://localhost:61683/api/health

# Both can be unavailable - server continues
curl http://localhost:7860/api/health
```

#### Issue: High Memory Usage
```powershell
# Monitor memory usage
Get-Process ollama | Select-Object WorkingSet, VirtualMemorySize

# Reduce parallel tasks
$env:OLLAMA_NUM_PARALLEL = "1"
.\ollama.exe serve

# Restart service
nssm restart OllamaK
```

#### Issue: Slow Response Time
```powershell
# Check CPU usage
Get-Process ollama | Select-Object CPU, Handles

# Increase thread count
$env:OLLAMA_NUM_THREADS = "16"

# Check network latency
Test-NetConnection -ComputerName localhost -Port 7860 -InformationLevel Detailed

# Monitor request latency
(1..100) | ForEach-Object {
    $start = Get-Date
    curl http://localhost:7860/api/health | Out-Null
    (Get-Date) - $start | Select-Object TotalMilliseconds
} | Measure-Object -Property TotalMilliseconds -Average
```

#### Issue: Registry Access Denied
```powershell
# Run as Administrator
Start-Process powershell -Verb RunAs

# Or use HKCU instead of HKLM
# User registry (HKCU) works without admin
```

---

## Maintenance

### Regular Tasks

#### Daily
```powershell
# Check service status
nssm status OllamaK

# Verify health endpoint
curl http://localhost:7860/api/health

# Check logs for errors
Get-Content C:\Ollama-K\logs\*.log | Select-String -Pattern "ERROR|WARN" | Tail -20
```

#### Weekly
```powershell
# Performance review
Get-Process ollama | Select-Object CPU, Memory, Handles | Format-Table

# Update checks
# Download latest release info
# Compare with current version
.\ollama.exe version

# Test failover procedures
# Simulate service failure and recovery
```

#### Monthly
```powershell
# Security updates
# Check for Go security patches
# Rebuild if patches available

# Dependency updates
# Review and update dependencies
# Test in staging environment
# Deploy to production

# Backup configuration
# Backup config.json and environment files
# Store securely
```

### Backup & Recovery

#### Backup Procedure
```powershell
# Backup application
$backupDate = Get-Date -Format "yyyyMMdd"
$backupDir = "\\backup-server\ollama-k\$backupDate"

New-Item -ItemType Directory -Path $backupDir -Force
Copy-Item -Path "C:\Ollama-K\*" -Destination $backupDir -Recurse -Force

# Backup registry (if using Windows Registry)
$regPath = "HKCU:\Software\OllamaK"
Export-Registry -Path $regPath -DestinationPath "$backupDir\registry.reg"

# Create backup manifest
@{
    Date = $backupDate
    Version = "1.0.0"
    Files = Get-ChildItem $backupDir | Select-Object Name, Length
} | ConvertTo-Json | Out-File "$backupDir\manifest.json"
```

#### Recovery Procedure
```powershell
# 1. Stop service
nssm stop OllamaK

# 2. Restore files
Copy-Item -Path "\\backup-server\ollama-k\$backupDate\*" `
          -Destination "C:\Ollama-K\" -Recurse -Force

# 3. Restore registry (if applicable)
Import-Registry -SourcePath "$backupDir\registry.reg"

# 4. Verify files
Get-FileHash -Path "C:\Ollama-K\ollama.exe" -Algorithm SHA256

# 5. Start service
nssm start OllamaK

# 6. Verify health
curl http://localhost:7860/api/health
```

---

## Scaling

### Horizontal Scaling

**Architecture**:
```
Clients
  ↓
Load Balancer (Nginx/IIS)
  ├→ Ollama-K-1 (7860)
  ├→ Ollama-K-2 (7861)
  └→ Ollama-K-3 (7862)
```

**Configuration**:
```nginx
upstream ollama_pool {
    least_conn;
    server ollama-1:7860 weight=1;
    server ollama-2:7860 weight=1;
    server ollama-3:7860 weight=1;
}

server {
    listen 80;

    location / {
        proxy_pass http://ollama_pool;
        proxy_next_upstream error timeout invalid_header http_502 http_503;
        proxy_connect_timeout 2s;
    }
}
```

### Vertical Scaling

**Performance Tuning**:
```powershell
# Increase resources
$env:OLLAMA_NUM_PARALLEL = "8"
$env:OLLAMA_NUM_THREADS = "16"
$env:OLLAMA_MAX_CONNECTIONS = "2000"

# Monitor impact
Get-Process ollama | Select-Object CPU, Memory

# Adjust based on system resources
# Aim for 80% CPU utilization at peak load
```

---

## Disaster Recovery

### Disaster Recovery Plan

#### Recovery Time Objectives (RTO)
- **Critical Service**: < 5 minutes
- **Standard Service**: < 30 minutes
- **Non-Critical**: < 4 hours

#### Recovery Point Objectives (RPO)
- **Critical Data**: < 15 minutes
- **Standard Data**: < 1 hour
- **Non-Critical**: < 1 day

#### Disaster Scenarios

**Scenario 1: Service Crash**
```powershell
# Automatic recovery via service restart
# Windows will restart service after 5 seconds

# Manual recovery
nssm restart OllamaK

# Verify recovery
nssm status OllamaK
curl http://localhost:7860/api/health
```

**Scenario 2: Port Conflict**
```powershell
# Find conflicting process
netstat -ano | Select-String ":7860"

# Kill process or use alternate port
$env:OLLAMA_PORT = "7861"
nssm set OllamaK AppEnvironmentExtra OLLAMA_PORT=7861
nssm restart OllamaK
```

**Scenario 3: Disk Full**
```powershell
# Check disk space
Get-Volume -DriveLetter C

# Free up space
Remove-Item "C:\Ollama-K\logs\*.log" -OlderThan (Get-Date).AddDays(-7)

# Move logs to larger disk
$env:OLLAMA_LOG_DIR = "D:\logs"

# Restart service
nssm restart OllamaK
```

**Scenario 4: Hardware Failure**
```powershell
# 1. Set up replacement server
# 2. Restore from backup
Copy-Item -Path "\\backup\ollama-k\latest\*" `
          -Destination "C:\Ollama-K\" -Recurse

# 3. Verify data integrity
Get-FileHash -Path "C:\Ollama-K\ollama.exe" -Algorithm SHA256

# 4. Update DNS/load balancer to point to new server
# 5. Verify traffic

# Estimated RTO: < 30 minutes
```

---

## Operations Runbook

### Emergency Procedures

#### Procedure 1: Emergency Shutdown
```powershell
# Step 1: Alert users
# Send notification to users about scheduled maintenance

# Step 2: Stop accepting new requests
# Drain load balancer (graceful shutdown)

# Step 3: Stop service
nssm stop OllamaK
Start-Sleep -Seconds 5

# Step 4: Verify stopped
Get-Process ollama -ErrorAction SilentlyContinue

# Step 5: Perform maintenance

# Step 6: Restart service
nssm start OllamaK

# Step 7: Verify health
curl http://localhost:7860/api/health

# Step 8: Resume traffic
# Update load balancer to resume traffic

# Step 9: Notify users
# Send notification that service is restored
```

#### Procedure 2: Emergency Rollback
```powershell
# Step 1: Notify stakeholders
# Indicate rollback is in progress

# Step 2: Stop current version
nssm stop OllamaK

# Step 3: Restore previous version
Copy-Item -Path "C:\Ollama-K\backup\v0.9.0\ollama.exe" `
          -Destination "C:\Ollama-K\ollama.exe" -Force

# Step 4: Start service
nssm start OllamaK

# Step 5: Verify
curl http://localhost:7860/api/health

# Step 6: Post-incident review
# Document what went wrong
# Plan fixes
# Schedule update for future release
```

#### Procedure 3: Database/Data Corruption
```powershell
# Step 1: Assess damage
# Check data integrity

# Step 2: Stop service
nssm stop OllamaK

# Step 3: Restore from backup
Copy-Item -Path "\\backup\ollama-k\latest\*" `
          -Destination "C:\Ollama-K\" -Recurse -Force

# Step 4: Verify restoration
Get-FileHash -Path "C:\Ollama-K\ollama.exe" -Algorithm SHA256

# Step 5: Start service
nssm start OllamaK

# Step 6: Run integrity checks
curl http://localhost:7860/api/health
```

---

## Quick Reference

### Key Commands

```powershell
# Service Management
nssm status OllamaK                    # Check status
nssm start OllamaK                     # Start
nssm stop OllamaK                      # Stop
nssm restart OllamaK                   # Restart
nssm edit OllamaK                      # Edit configuration

# Monitoring
Get-Process ollama                     # List process
Get-Service OllamaK                    # Service status
Get-EventLog -LogName System -Source OllamaK  # Event logs

# Testing
curl http://localhost:7860/api/health # Health check
curl http://localhost:7860/api/services/discover # Service discovery

# Troubleshooting
netstat -ano | Select-String ":7860"  # Port usage
Get-NetFirewallRule -DisplayName "Ollama-K"  # Firewall rules
tasklist /FI "IMAGENAME eq ollama.exe"  # Process list
```

### Important Paths

```
Binary: C:\Ollama-K\ollama.exe
Config: C:\Ollama-K\config.json
Logs: C:\Ollama-K\logs\
Backup: \\backup-server\ollama-k\
```

### Escalation Contact

```
Tier 1 (Operations): +1-555-0100
Tier 2 (Engineering): +1-555-0101
Tier 3 (Architecture): +1-555-0102
On-Call: PagerDuty #ollama-k
```

---

## Conclusion

This deployment and operations guide covers the complete lifecycle of Ollama-K v1.0.0 from initial deployment through production operations and disaster recovery.

For detailed information, refer to:
- [WINDOWS_BUILD_GUIDE.md](WINDOWS_BUILD_GUIDE.md)
- [WINDOWS_TEST_VALIDATION.md](WINDOWS_TEST_VALIDATION.md)
- [RELEASE_PREPARATION_v1.0.md](RELEASE_PREPARATION_v1.0.md)

---

**Version**: 1.0.0
**Last Updated**: February 28, 2026
**Next Review**: March 28, 2026
