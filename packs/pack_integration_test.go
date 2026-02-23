package packs

import (
	"testing"
	"time"

	"github.com/ollama/ollama/kuhul/runtime"
)

// TestLamOPackOllamaIntegration tests real Ollama API integration
// Skip this test if Ollama is not running
func TestLamOPackOllamaIntegration(t *testing.T) {
	// This test requires Ollama to be running on localhost:11434
	// Run with: go test ./packs -v -run TestLamOPackOllamaIntegration

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	state := runtime.NewRuntimeState()
	pack := NewLamOPack()

	// Initialize pack
	if err := pack.Init(state); err != nil {
		t.Fatalf("failed to initialize pack: %v", err)
	}

	t.Run("GenerateWithInfer", func(t *testing.T) {
		// Skip if Ollama is not available
		if pack.client == nil {
			t.Skip("Ollama client not available")
		}

		handlers := pack.Handlers()
		inferHandler, ok := handlers["lam_o.infer"]
		if !ok {
			t.Fatal("infer handler not found")
		}

		// Try with a small model if available
		ctx := &runtime.Context{
			Body: map[string]interface{}{
				"model":       "llama2",
				"prompt":      "What is 2+2?",
				"temperature": 0.5,
			},
		}

		// Set a timeout for the actual API call
		done := make(chan interface{})
		go func() {
			result, _ := inferHandler(ctx)
			done <- result
		}()

		// Wait with timeout
		select {
		case result := <-done:
			if result == nil {
				t.Error("handler returned nil result")
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("expected map result, got %T", result)
				return
			}

			if !resultMap["ok"].(bool) {
				t.Errorf("handler failed: %v", resultMap)
				return
			}

			if _, hasResponse := resultMap["response"]; !hasResponse {
				t.Error("response field missing from result")
			}

		case <-time.After(60 * time.Second):
			t.Skip("Ollama did not respond in time - might not be running")
		}
	})

	t.Run("ListModels", func(t *testing.T) {
		if pack.client == nil {
			t.Skip("Ollama client not available")
		}

		handlers := pack.Handlers()
		listHandler, ok := handlers["lam_o.list_models"]
		if !ok {
			t.Fatal("list_models handler not found")
		}

		ctx := &runtime.Context{
			Body: map[string]interface{}{},
		}

		result, err := listHandler(ctx)
		if err != nil {
			t.Errorf("handler execution error = %v", err)
			return
		}

		if result == nil {
			t.Error("handler returned nil result")
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("expected map result, got %T", result)
			return
		}

		// Check if error response or success
		if resultMap["ok"] != nil {
			if !resultMap["ok"].(bool) {
				// Ollama might not be running, skip
				t.Skip("Ollama not available: " + resultMap["error"].(string))
			}
		}
	})

	t.Run("EmbedWithOllama", func(t *testing.T) {
		if pack.client == nil {
			t.Skip("Ollama client not available")
		}

		handlers := pack.Handlers()
		embedHandler, ok := handlers["lam_o.embed"]
		if !ok {
			t.Fatal("embed handler not found")
		}

		ctx := &runtime.Context{
			Body: map[string]interface{}{
				"text":  "hello world",
				"model": "nomic-embed-text",
			},
		}

		done := make(chan interface{})
		go func() {
			result, _ := embedHandler(ctx)
			done <- result
		}()

		select {
		case result := <-done:
			if result == nil {
				t.Skip("Ollama not available")
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("expected map result, got %T", result)
				return
			}

			if resultMap["ok"] != nil && !resultMap["ok"].(bool) {
				t.Skip("Ollama not available: " + resultMap["error"].(string))
			}

		case <-time.After(30 * time.Second):
			t.Skip("Ollama did not respond in time")
		}
	})

	t.Run("ShowModel", func(t *testing.T) {
		if pack.client == nil {
			t.Skip("Ollama client not available")
		}

		handlers := pack.Handlers()
		showHandler, ok := handlers["lam_o.show_model"]
		if !ok {
			t.Fatal("show_model handler not found")
		}

		ctx := &runtime.Context{
			Body: map[string]interface{}{
				"model": "llama2",
			},
		}

		result, err := showHandler(ctx)
		if err != nil {
			t.Errorf("handler execution error = %v", err)
			return
		}

		if result == nil {
			t.Error("handler returned nil result")
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("expected map result, got %T", result)
			return
		}

		// If ok is false, Ollama is not available
		if resultMap["ok"] != nil && !resultMap["ok"].(bool) {
			t.Skip("Ollama not available")
		}
	})
}

// TestOllamaAPIClientCreation verifies API client initialization
func TestOllamaAPIClientCreation(t *testing.T) {
	pack := NewLamOPack()

	if pack.client == nil {
		t.Error("API client should be initialized in NewLamOPack")
	}

	if pack.baseURL == nil {
		t.Error("base URL should be initialized")
	}

	if pack.endpoint != "http://localhost:11434" {
		t.Errorf("expected default endpoint, got %s", pack.endpoint)
	}
}

// TestOllamaAPIErrorHandling verifies error handling when Ollama is unavailable
func TestOllamaAPIErrorHandling(t *testing.T) {
	state := runtime.NewRuntimeState()
	pack := NewLamOPack()

	// Reinitialize with unreachable endpoint
	state.Variables.Set("@lam_o_endpoint", "http://localhost:19999")
	err := pack.Init(state)
	if err != nil {
		t.Errorf("init error (should not fail immediately): %v", err)
	}

	handlers := pack.Handlers()
	inferHandler, ok := handlers["lam_o.infer"]
	if !ok {
		t.Fatal("infer handler not found")
	}

	ctx := &runtime.Context{
		Body: map[string]interface{}{
			"model":  "llama2",
			"prompt": "test",
		},
	}

	// Create a channel to receive the result with timeout
	done := make(chan interface{})
	go func() {
		result, _ := inferHandler(ctx)
		done <- result
	}()

	select {
	case result := <-done:
		if result == nil {
			t.Error("handler returned nil result")
			return
		}

		// Handler can return either map or error response object
		switch v := result.(type) {
		case map[string]interface{}:
			// Should have an error response (503 Service Unavailable or similar)
			if v["ok"] == nil || v["ok"].(bool) {
				t.Log("Note: If Ollama is running on the test machine, this test needs adjustment")
			}
		default:
			// Error response object - this is acceptable
			t.Logf("Got error response of type %T", v)
		}

	case <-time.After(5 * time.Second):
		// Timeout is acceptable for unreachable endpoint
		t.Log("Request timed out as expected for unreachable endpoint")
	}
}

// BenchmarkOllamaPackHandlers benchmarks pack handler performance
// Only run if Ollama is available
func BenchmarkOllamaPackHandlers(b *testing.B) {
	state := runtime.NewRuntimeState()
	pack := NewLamOPack()

	if err := pack.Init(state); err != nil {
		b.Fatalf("failed to initialize pack: %v", err)
	}

	handlers := pack.Handlers()
	listHandler := handlers["lam_o.list_models"]

	b.ResetTimer()

	// Benchmark list models (fastest operation)
	for i := 0; i < b.N; i++ {
		ctx := &runtime.Context{Body: map[string]interface{}{}}
		_, _ = listHandler(ctx)
	}
}
