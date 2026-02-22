//go:build !windows

package runtime

import (
	"net"
	"net/http"
	"time"
)

// PortAvailable checks if a port is available on the system
func PortAvailable(port int) bool {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", string(rune(port))))
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

// DiscoverServices attempts to find running services
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

// GetLocalIPAddress returns the local IP address
func GetLocalIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// WaitForPort waits for a port to become available
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
