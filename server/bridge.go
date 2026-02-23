package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api/xjson"
)

// BridgeConfig holds the configuration for service discovery and proxying
type BridgeConfig struct {
	OllamaURL       string
	OrchestratorURL string
	LastDiscovery   time.Time
	mu              sync.RWMutex
}

var bridgeConfig = &BridgeConfig{
	OllamaURL:       "http://localhost:11434",
	OrchestratorURL: "http://localhost:61683",
}

// DiscoverServices attempts to discover Ollama and Orchestrator services
func DiscoverServices() error {
	bridgeConfig.mu.Lock()
	defer bridgeConfig.mu.Unlock()

	// Try to discover Ollama
	if err := checkService("http://localhost:11434/api/tags", "ollama"); err == nil {
		bridgeConfig.OllamaURL = "http://localhost:11434"
	} else if err := checkService("http://127.0.0.1:11434/api/tags", "ollama"); err == nil {
		bridgeConfig.OllamaURL = "http://127.0.0.1:11434"
	}

	// Try to discover Orchestrator
	if err := checkService("http://localhost:61683/api/health", "orchestrator"); err == nil {
		bridgeConfig.OrchestratorURL = "http://localhost:61683"
	} else if err := checkService("http://127.0.0.1:61683/api/health", "orchestrator"); err == nil {
		bridgeConfig.OrchestratorURL = "http://127.0.0.1:61683"
	}

	bridgeConfig.LastDiscovery = time.Now()
	return nil
}

// checkService checks if a service is available
func checkService(url string, serviceType string) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("service returned status %d", resp.StatusCode)
}

// GetBridgeConfig returns the current bridge configuration
func GetBridgeConfig() *BridgeConfig {
	bridgeConfig.mu.RLock()
	defer bridgeConfig.mu.RUnlock()

	return &BridgeConfig{
		OllamaURL:       bridgeConfig.OllamaURL,
		OrchestratorURL: bridgeConfig.OrchestratorURL,
		LastDiscovery:   bridgeConfig.LastDiscovery,
	}
}

// ServiceDiscoveryResponse represents the service discovery response
type ServiceDiscoveryResponse struct {
	OllamaURL       string    `json:"ollama_url"`
	OrchestratorURL string    `json:"orchestrator_url"`
	LastDiscovery   time.Time `json:"last_discovery"`
	Timestamp       time.Time `json:"timestamp"`
}

// DiscoverServicesHandler handles the service discovery endpoint
func (s *Server) DiscoverServicesHandler(c *gin.Context) {
	// Rediscover services (with caching to avoid constant lookups)
	now := time.Now()
	config := GetBridgeConfig()
	if now.Sub(config.LastDiscovery) > 30*time.Second {
		_ = DiscoverServices()
		config = GetBridgeConfig()
	}

	c.JSON(http.StatusOK, ServiceDiscoveryResponse{
		OllamaURL:       config.OllamaURL,
		OrchestratorURL: config.OrchestratorURL,
		LastDiscovery:   config.LastDiscovery,
		Timestamp:       now,
	})
}

// ProxyRequest proxies a request to a backend service
type ProxyRequest struct {
	Target  string                 `json:"target"` // "ollama" or "orchestrator"
	Path    string                 `json:"path"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

// ProxyResponse represents a proxied response
type ProxyResponse struct {
	Status   int                    `json:"status"`
	Headers  map[string][]string    `json:"headers,omitempty"`
	Body     map[string]interface{} `json:"body,omitempty"`
	Error    string                 `json:"error,omitempty"`
	RawText  string                 `json:"raw_text,omitempty"`
}

// ProxyInferHandler proxies XJSON inference requests to backend services
func (s *Server) ProxyInferHandler(c *gin.Context) {
	var xjsonEnvelope xjson.XJSONEnvelope

	if err := c.ShouldBindJSON(&xjsonEnvelope); err != nil {
		c.JSON(http.StatusBadRequest, xjson.NewErrorResponse(
			"proxy",
			"Invalid XJSON request: "+err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	if !xjsonEnvelope.IsInferRequest() {
		c.JSON(http.StatusBadRequest, xjson.NewErrorResponse(
			"proxy",
			"Request is not an @infer request",
			http.StatusBadRequest,
		))
		return
	}

	req := xjsonEnvelope.Infer
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, xjson.NewErrorResponse(
			"proxy",
			"Validation error: "+err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	config := GetBridgeConfig()

	// Try orchestrator first, then fallback to Ollama
	resp, err := proxyToService(config.OrchestratorURL, req)
	if err != nil {
		// Fallback to Ollama
		resp, err = proxyToService(config.OllamaURL, req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, xjson.NewErrorResponse(
				"proxy",
				"All backend services are unavailable",
				http.StatusServiceUnavailable,
			).WithDetails(err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, resp)
}

// proxyToService sends an @infer request to a backend service
func proxyToService(serviceURL string, req *xjson.InferRequest) (*xjson.CompletionResponse, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Build request body
	envelope := &xjson.XJSONEnvelope{Infer: req}
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request
	endpoint := serviceURL + "/api/infer"
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("service request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("service returned status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Parse response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var respEnvelope xjson.XJSONEnvelope
	if err := json.Unmarshal(respBody, &respEnvelope); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if respEnvelope.IsError() {
		return nil, fmt.Errorf("service error: %s", respEnvelope.Error.Message)
	}

	if !respEnvelope.IsCompletion() {
		return nil, fmt.Errorf("unexpected response format")
	}

	return respEnvelope.Completion, nil
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status       string            `json:"status"` // "healthy", "degraded", "unhealthy"
	Services     map[string]bool    `json:"services"`
	LastCheck    time.Time          `json:"last_check"`
	Message      string            `json:"message,omitempty"`
}

// HealthCheckHandler checks the health of all backend services
func (s *Server) HealthCheckHandler(c *gin.Context) {
	config := GetBridgeConfig()
	services := make(map[string]bool)

	// Check Ollama
	err1 := checkService(config.OllamaURL+"/api/tags", "ollama")
	services["ollama"] = err1 == nil

	// Check Orchestrator
	err2 := checkService(config.OrchestratorURL+"/api/health", "orchestrator")
	services["orchestrator"] = err2 == nil

	// Determine overall health
	status := "healthy"
	message := "All services are available"

	if !services["ollama"] && !services["orchestrator"] {
		status = "unhealthy"
		message = "All backend services are unavailable"
	} else if !services["ollama"] || !services["orchestrator"] {
		status = "degraded"
		if !services["ollama"] {
			message = "Ollama service is unavailable"
		} else {
			message = "Orchestrator service is unavailable"
		}
	}

	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:    status,
		Services:  services,
		LastCheck: time.Now(),
		Message:   message,
	})
}

// RegisterBridgeRoutes registers the bridge endpoints
func (s *Server) RegisterBridgeRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Service discovery
		api.GET("/services/discover", s.DiscoverServicesHandler)

		// Health check
		api.GET("/health", s.HealthCheckHandler)

		// Proxy endpoints
		api.POST("/proxy/infer", s.ProxyInferHandler)
	}
}

// InitBridge initializes the bridge on server startup
func (s *Server) InitBridge() error {
	return DiscoverServices()
}
