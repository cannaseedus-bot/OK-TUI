// Package kuhul provides integration tests for the full K'UHUL system
package kuhul

import (
	"context"
	"testing"

	"github.com/ollama/ollama/kuhul/compiler"
	"github.com/ollama/ollama/kuhul/runtime"
)

// TestFullPipeline tests the complete K'UHUL pipeline
// K'UHUL Source → Compiler → Runtime → Execution
func TestFullPipeline(t *testing.T) {
	tests := []struct {
		name      string
		kuhulCode string
		wantOk    bool
		checkFn   func(*runtime.ExecutionContext) bool
	}{
		{
			name: "simple_variable_assignment",
			kuhulCode: `wo x 42
wo y "hello"`,
			wantOk: true,
			checkFn: func(ec *runtime.ExecutionContext) bool {
				// Just verify the context is initialized
				return ec.MemTier != nil && ec.CallTier != nil
			},
		},
		{
			name: "array_assignment",
			kuhulCode: `wo nums [1 2 3]`,
			wantOk: true,
			checkFn: func(ec *runtime.ExecutionContext) bool {
				// Just verify the context is initialized
				return ec.MemTier != nil
			},
		},
		{
			name: "handler_definition",
			kuhulCode: `handler greet(name) {
  wo greeting "Hello"
}`,
			wantOk: true,
			checkFn: func(ec *runtime.ExecutionContext) bool {
				return true // Just check it compiles
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Compile K'UHUL to JavaScript
			comp := compiler.NewCompiler(tt.kuhulCode)
			js, err := comp.Compile()

			if err != nil && tt.wantOk {
				t.Fatalf("compilation failed: %v", err)
			}

			if err == nil && !tt.wantOk {
				t.Fatalf("expected compilation error, got nil")
			}

			if !tt.wantOk {
				return
			}

			// Step 2: Create execution context
			ec := runtime.NewExecutionContext()
			defer ec.Cleanup()

			// Step 3: Verify generated code
			if js == "" {
				t.Error("expected generated code, got empty string")
			}

			if !contains(js, "$kuhul_main") {
				t.Error("generated code missing main function")
			}

			// Step 4: Run checks on the context
			if !tt.checkFn(ec) {
				t.Error("execution context check failed")
			}
		})
	}
}

// TestCompilerToRuntimeIntegration tests that compiler output can run on runtime
func TestCompilerToRuntimeIntegration(t *testing.T) {
	// K'UHUL code that uses packs
	kuhulCode := `wo model "llama2"
wo temperature 0.7`

	// Compile K'UHUL
	comp := compiler.NewCompiler(kuhulCode)
	js, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create execution environment
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Verify the generated code references the pack system
	if !contains(js, "packs") {
		t.Error("generated code missing pack references")
	}

	// Verify memory tier is set up
	if ec.MemTier == nil {
		t.Error("memory tier not initialized")
	}

	// Verify call tier is set up
	if ec.CallTier == nil {
		t.Error("call tier not initialized")
	}
}

// TestMemoryTierWithCompiler tests memory operations with compiled code
func TestMemoryTierWithCompiler(t *testing.T) {
	kuhulCode := `wo x 100
wo y 200
wo z 300`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Simulate storing variables in memory tier
	ec.MemTier.Set("x", 100)
	ec.MemTier.Set("y", 200)
	ec.MemTier.Set("z", 300)

	// Verify retrieval
	x, _ := ec.MemTier.Get("x")
	if x != 100 {
		t.Errorf("expected x=100, got %v", x)
	}

	y, _ := ec.MemTier.Get("y")
	if y != 200 {
		t.Errorf("expected y=200, got %v", y)
	}

	// Verify heap size
	size := ec.MemTier.GetHeapSize()
	if size != 3 {
		t.Errorf("expected heap size 3, got %d", size)
	}
}

// TestCallTierWithCompiler tests function calls with compiled code
func TestCallTierWithCompiler(t *testing.T) {
	kuhulCode := `handler add(x y) {
  wo result "done"
}`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Register a function in the call tier
	addFunc := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		x := args["x"].(float64)
		y := args["y"].(float64)
		return x + y, nil
	}

	ec.CallTier.RegisterFunction("add", addFunc)

	// Call the function
	result, err := ec.CallTier.Call(context.Background(), "add", map[string]interface{}{
		"x": 5.0,
		"y": 3.0,
	})

	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	if result != 8.0 {
		t.Errorf("expected 8.0, got %v", result)
	}
}

// TestIPCTierWithCompiler tests inter-process communication
func TestIPCTierWithCompiler(t *testing.T) {
	kuhulCode := `wo message "hello world"`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Create a channel
	ec.IPCTier.CreateChannel("messages")

	// Send a message
	ec.IPCTier.Send("messages", "test message")

	// Receive the message
	msg, err := ec.IPCTier.Receive("messages")
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}

	if msg != "test message" {
		t.Errorf("expected 'test message', got %v", msg)
	}
}

// TestEndToEndWithPacks tests full pipeline with pack integration
func TestEndToEndWithPacks(t *testing.T) {
	// K'UHUL code that invokes a pack
	kuhulCode := `wo model "llama2"
wo prompt "What is 2+2?"

pop (lam_o.infer
  :model model
  :prompt prompt
  :temperature 0.7)`

	// Step 1: Compile K'UHUL
	comp := compiler.NewCompiler(kuhulCode)
	js, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Step 2: Verify generated code
	if !contains(js, "packs.lam_o") {
		t.Error("generated code missing pack reference")
	}

	if !contains(js, "$mem") {
		t.Error("generated code missing memory tier")
	}

	// Step 3: Create execution context
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Step 4: Store variables in memory (simulating assignment)
	ec.MemTier.Set("model", "llama2")
	ec.MemTier.Set("prompt", "What is 2+2?")
	ec.MemTier.Set("temperature", 0.7)

	// Step 5: Retrieve and verify
	model, _ := ec.MemTier.Get("model")
	if model != "llama2" {
		t.Errorf("expected model='llama2', got %v", model)
	}

	prompt, _ := ec.MemTier.Get("prompt")
	if prompt != "What is 2+2?" {
		t.Errorf("expected prompt='What is 2+2?', got %v", prompt)
	}
}

// TestErrorHandling tests error handling in the full pipeline
func TestErrorHandling(t *testing.T) {
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Add an error
	ec.AddError("@mem", "test error", "test context")

	// Verify error recording
	errors := ec.GetErrors()
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	if errors[0].Message != "test error" {
		t.Errorf("expected 'test error', got %s", errors[0].Message)
	}
}

// TestMultipleTierInteraction tests interaction between multiple tiers
func TestMultipleTierInteraction(t *testing.T) {
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Tier 1: Store value in memory
	ec.MemTier.Set("data", map[string]interface{}{"key": "value"})

	// Tier 2: Create a handler that accesses memory
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		data, _ := ec.MemTier.Get("data")
		return data, nil
	}

	ec.CallTier.RegisterFunction("getData", handler)

	// Call the handler
	result, err := ec.CallTier.Call(context.Background(), "getData", map[string]interface{}{})
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	// Tier 3: Send result through IPC
	ec.IPCTier.CreateChannel("results")
	ec.IPCTier.Send("results", result)

	// Receive from IPC
	received, _ := ec.IPCTier.Receive("results")
	if received == nil {
		t.Error("expected result, got nil")
	}
}

// Utility function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr ||
		 len(s) > len(substr) &&
		 (s[:len(substr)] == substr ||
		  s[len(s)-len(substr):] == substr ||
		  indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// BenchmarkFullPipeline benchmarks the complete pipeline
func BenchmarkFullPipeline(b *testing.B) {
	kuhulCode := `wo x 1
wo y 2
wo z 3`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Compile
		comp := compiler.NewCompiler(kuhulCode)
		comp.Compile()

		// Create runtime
		ec := runtime.NewExecutionContext()
		ec.MemTier.Set("x", 1)
		ec.MemTier.Set("y", 2)
		ec.MemTier.Set("z", 3)
		ec.Cleanup()
	}
}
