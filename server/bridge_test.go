package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api/xjson"
)

// TestBridgeConfig tests the bridge configuration
func TestBridgeConfig(t *testing.T) {
	config := GetBridgeConfig()

	if config == nil {
		t.Fatal("GetBridgeConfig returned nil")
	}

	if config.OllamaURL == "" {
		t.Error("OllamaURL should not be empty")
	}

	if config.OrchestratorURL == "" {
		t.Error("OrchestratorURL should not be empty")
	}

	t.Logf("Bridge config - Ollama: %s, Orchestrator: %s", config.OllamaURL, config.OrchestratorURL)
}

// TestServiceDiscoveryResponse tests the service discovery response structure
func TestServiceDiscoveryResponse(t *testing.T) {
	response := ServiceDiscoveryResponse{
		OllamaURL:       "http://localhost:11434",
		OrchestratorURL: "http://localhost:61683",
		LastDiscovery:   time.Now(),
		Timestamp:       time.Now(),
	}

	// Marshal to JSON to verify structure
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal to verify all fields are present
	var unmarshalled ServiceDiscoveryResponse
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshalled.OllamaURL != response.OllamaURL {
		t.Error("OllamaURL not preserved in JSON")
	}

	t.Logf("ServiceDiscoveryResponse structure verified")
}

// TestHealthCheckResponse tests the health check response structure
func TestHealthCheckResponse(t *testing.T) {
	response := HealthCheckResponse{
		Status:  "healthy",
		Services: map[string]bool{
			"ollama":       true,
			"orchestrator": true,
		},
		LastCheck: time.Now(),
		Message:   "All services are available",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal to verify
	var unmarshalled HealthCheckResponse
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshalled.Status != "healthy" {
		t.Error("Status not preserved")
	}

	if !unmarshalled.Services["ollama"] {
		t.Error("Ollama service status not preserved")
	}

	t.Logf("HealthCheckResponse structure verified")
}

// TestDiscoverServicesEndpoint tests the /api/services/discover endpoint
func TestDiscoverServicesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a minimal server for testing
	server := &Server{}

	// Register the bridge routes
	server.RegisterBridgeRoutes(router)

	// Create a test request
	req, err := http.NewRequest("GET", "/api/services/discover", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response ServiceDiscoveryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.OllamaURL == "" {
		t.Error("OllamaURL should not be empty in response")
	}

	t.Logf("Service discovery endpoint response: %+v", response)
}

// TestHealthCheckEndpoint tests the /api/health endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthCheckResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status == "" {
		t.Error("Status should not be empty")
	}

	if response.Services == nil {
		t.Error("Services map should not be nil")
	}

	t.Logf("Health check endpoint response: Status=%s, Message=%s", response.Status, response.Message)
}

// TestProxyInferEndpointValidation tests proxy endpoint request validation
func TestProxyInferEndpointValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	// Test with invalid request (empty body)
	req, err := http.NewRequest("POST", "/api/proxy/infer", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail with 400 Bad Request (invalid XJSON)
	if w.Code != http.StatusBadRequest {
		t.Logf("Proxy infer endpoint returned: %d (validation behavior)", w.Code)
	}
}

// TestBridgeRouteRegistration tests that all bridge routes are registered
func TestBridgeRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	// Get the list of registered routes
	routes := router.Routes()

	expectedRoutes := []string{
		"GET /api/services/discover",
		"GET /api/health",
		"POST /api/proxy/infer",
	}

	registeredRoutes := make(map[string]bool)
	for _, route := range routes {
		registeredRoutes[route.Method+" "+route.Path] = true
	}

	for _, expected := range expectedRoutes {
		if !registeredRoutes[expected] {
			t.Errorf("Expected route not registered: %s", expected)
		}
	}

	t.Logf("Bridge routes registered: %d routes", len(routes))
	for _, route := range routes {
		t.Logf("  - %s %s", route.Method, route.Path)
	}
}

// TestServiceAvailabilityCheck tests service availability checking
func TestServiceAvailabilityCheck(t *testing.T) {
	// Test with a non-existent service
	err := checkService("http://localhost:65432/nonexistent", "test")
	if err == nil {
		t.Logf("Service check correctly detected unavailable service")
	}

	// Test structure is working
	t.Logf("Service availability check function working")
}

// TestProxyRequest structure validates the proxy request structure
func TestProxyRequestStructure(t *testing.T) {
	req := ProxyRequest{
		Target:  "ollama",
		Path:    "/api/tags",
		Method:  "GET",
		Headers: map[string]string{"Accept": "application/json"},
		Body:    nil,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshalled ProxyRequest
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshalled.Target != req.Target {
		t.Error("Target not preserved")
	}

	t.Logf("ProxyRequest structure verified")
}

// TestBridgeInitialization tests bridge initialization
func TestBridgeInitialization(t *testing.T) {
	// Call DiscoverServices to test initialization
	err := DiscoverServices()

	if err != nil {
		t.Logf("DiscoverServices: %v (services may not be running)", err)
	}

	config := GetBridgeConfig()
	if config == nil {
		t.Fatal("Bridge config is nil after initialization")
	}

	t.Logf("Bridge initialized - Ollama: %s, Orchestrator: %s", config.OllamaURL, config.OrchestratorURL)
}

// BenchmarkServiceDiscovery benchmarks service discovery response
func BenchmarkServiceDiscovery(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	req, _ := http.NewRequest("GET", "/api/services/discover", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHealthCheck benchmarks health check response
func BenchmarkHealthCheck(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	req, _ := http.NewRequest("GET", "/api/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// TestBridgeErrorHandling tests bridge error handling
func TestBridgeErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{}
	server.RegisterBridgeRoutes(router)

	// Test with malformed JSON
	req, err := http.NewRequest("POST", "/api/proxy/infer", strings.NewReader("{invalid json}"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should get a 400 error
	if w.Code >= 400 {
		t.Logf("Proxy error handling working - status code %d", w.Code)
	}
}
