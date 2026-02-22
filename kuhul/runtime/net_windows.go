//go:build windows

package runtime

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PortAvailable checks if a port is available on Windows
func PortAvailable(port int) bool {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// ServiceDiscovery discovers running services on the system
type ServiceDiscovery struct {
	Ollama       string // Ollama API URL
	Orchestrator string // K'UHUL Orchestrator URL
}

// DiscoverServices attempts to find running services on Windows
func DiscoverServices() *ServiceDiscovery {
	discovery := &ServiceDiscovery{
		Ollama:       "http://localhost:11434",
		Orchestrator: "http://localhost:61683",
	}

	// Check if services are available
	if !isServiceAvailable(discovery.Ollama) {
		discovery.Ollama = ""
	}
	if !isServiceAvailable(discovery.Orchestrator) {
		discovery.Orchestrator = ""
	}

	return discovery
}

// isServiceAvailable does a quick health check on a service
func isServiceAvailable(url string) bool {
	// Create a simple HTTP client with timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Try to reach the service
	resp, err := client.Get(url + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 500
}

// GetLocalIPAddress returns the local IP address on Windows
func GetLocalIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// WaitForPort waits for a port to become available on Windows
// Times out after maxWaitSeconds
func WaitForPort(port int, maxWaitSeconds int) bool {
	for i := 0; i < maxWaitSeconds; i++ {
		if PortAvailable(port) {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// GetProcessByPort returns the process ID using a specific port on Windows
// Uses netstat command to find the process
func GetProcessByPort(port int) (int, error) {
	// Use netstat to find process by port
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return -1, fmt.Errorf("netstat command failed: %w", err)
	}

	portStr := fmt.Sprintf(":%d", port)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, portStr) {
			// Parse the PID from netstat output
			fields := strings.Fields(line)
			if len(fields) > 0 {
				pidStr := fields[len(fields)-1]
				pid, err := strconv.Atoi(pidStr)
				if err == nil && pid > 0 {
					return pid, nil
				}
			}
		}
	}

	return -1, fmt.Errorf("no process found on port %d", port)
}

// IsPortInUse checks if a port is in use on Windows
func IsPortInUse(port int) bool {
	_, err := GetProcessByPort(port)
	return err == nil
}

// GetProcessNameByPort returns the process name using a specific port on Windows
func GetProcessNameByPort(port int) (string, error) {
	pid, err := GetProcessByPort(port)
	if err != nil {
		return "", err
	}

	// Use tasklist to get process name by PID
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tasklist command failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("%d", pid)) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("process name not found for PID %d", pid)
}
