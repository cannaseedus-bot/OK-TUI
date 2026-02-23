package runtime

import (
	"context"
	"testing"
	"time"
)

// TestMemoryTier tests the memory tier (@mem)
func TestMemoryTier(t *testing.T) {
	mem := NewMemoryTier()

	t.Run("SetAndGet", func(t *testing.T) {
		err := mem.Set("x", 42)
		if err != nil {
			t.Fatalf("Set error: %v", err)
		}

		val, err := mem.Get("x")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}

		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		mem.Set("y", 100)
		err := mem.Delete("y")
		if err != nil {
			t.Fatalf("Delete error: %v", err)
		}

		_, err = mem.Get("y")
		if err == nil {
			t.Error("expected error for deleted variable")
		}
	})

	t.Run("Stack", func(t *testing.T) {
		mem.PushStack(1)
		mem.PushStack(2)
		mem.PushStack(3)

		val, err := mem.PopStack()
		if err != nil || val != 3 {
			t.Errorf("expected 3, got %v (err: %v)", val, err)
		}

		val, err = mem.PopStack()
		if err != nil || val != 2 {
			t.Errorf("expected 2, got %v (err: %v)", val, err)
		}
	})

	t.Run("Registers", func(t *testing.T) {
		err := mem.SetRegister(0, "value1")
		if err != nil {
			t.Fatalf("SetRegister error: %v", err)
		}

		val, err := mem.GetRegister(0)
		if err != nil || val != "value1" {
			t.Errorf("expected 'value1', got %v (err: %v)", val, err)
		}

		// Test bounds
		err = mem.SetRegister(16, "out_of_bounds")
		if err == nil {
			t.Error("expected error for out of bounds register")
		}
	})

	t.Run("HeapSize", func(t *testing.T) {
		// Use a fresh memory tier for this test
		freshMem := NewMemoryTier()
		freshMem.Set("a", 1)
		freshMem.Set("b", 2)
		freshMem.Set("c", 3)

		size := freshMem.GetHeapSize()
		if size != 3 {
			t.Errorf("expected heap size 3, got %d", size)
		}
	})
}

// TestCallTier tests the call tier (@call)
func TestCallTier(t *testing.T) {
	call := NewCallTier()

	// Register a test function
	testFunc := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		x := args["x"].(float64)
		y := args["y"].(float64)
		return x + y, nil
	}

	call.RegisterFunction("add", testFunc)

	t.Run("CallFunction", func(t *testing.T) {
		result, err := call.Call(context.Background(), "add", map[string]interface{}{
			"x": 5.0,
			"y": 3.0,
		})

		if err != nil {
			t.Fatalf("Call error: %v", err)
		}

		if result != 8.0 {
			t.Errorf("expected 8.0, got %v", result)
		}
	})

	t.Run("UndefinedFunction", func(t *testing.T) {
		_, err := call.Call(context.Background(), "undefined", map[string]interface{}{})
		if err == nil {
			t.Error("expected error for undefined function")
		}
	})

	t.Run("CallStack", func(t *testing.T) {
		depth := call.GetCallDepth()
		if depth != 0 {
			t.Errorf("expected call depth 0, got %d", depth)
		}

		stack := call.GetCallStack()
		if len(stack) != 0 {
			t.Errorf("expected empty call stack, got %d frames", len(stack))
		}
	})
}

// TestIPCTier tests the IPC tier (@ipc)
func TestIPCTier(t *testing.T) {
	ipc := NewIPCTier()

	t.Run("CreateChannel", func(t *testing.T) {
		ch, err := ipc.CreateChannel("test_channel")
		if err != nil {
			t.Fatalf("CreateChannel error: %v", err)
		}

		if ch.Name != "test_channel" {
			t.Errorf("expected channel name 'test_channel', got %s", ch.Name)
		}
	})

	t.Run("DuplicateChannel", func(t *testing.T) {
		ipc.CreateChannel("dup")
		_, err := ipc.CreateChannel("dup")
		if err == nil {
			t.Error("expected error for duplicate channel")
		}
	})

	t.Run("SendReceive", func(t *testing.T) {
		ipc.CreateChannel("msgs")

		err := ipc.Send("msgs", "hello")
		if err != nil {
			t.Fatalf("Send error: %v", err)
		}

		msg, err := ipc.Receive("msgs")
		if err != nil {
			t.Fatalf("Receive error: %v", err)
		}

		if msg != "hello" {
			t.Errorf("expected 'hello', got %v", msg)
		}
	})

	t.Run("ChannelSubscription", func(t *testing.T) {
		ipc.CreateChannel("events")
		ch, _ := ipc.GetChannel("events")

		received := make(chan interface{}, 1)
		ch.Subscribe(func(data interface{}) {
			received <- data
		})

		ch.Send("event_data")

		select {
		case data := <-received:
			if data != "event_data" {
				t.Errorf("expected 'event_data', got %v", data)
			}
		case <-time.After(1 * time.Second):
			t.Error("subscription callback not called within timeout")
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		ipc.CreateChannel("cleanup_test")
		ipc.Cleanup()

		// Should not be able to get channel after cleanup
		_, err := ipc.GetChannel("cleanup_test")
		if err == nil {
			t.Error("expected error after cleanup")
		}
	})
}

// TestDatabaseTier tests the database tier (@db)
func TestDatabaseTier(t *testing.T) {
	db := NewDatabaseTier()

	t.Run("Connect", func(t *testing.T) {
		err := db.Connect("mydb", "postgres", "postgresql://localhost/testdb")
		if err != nil {
			t.Fatalf("Connect error: %v", err)
		}
	})

	t.Run("DuplicateConnection", func(t *testing.T) {
		db.Connect("dup", "sqlite", "test.db")
		err := db.Connect("dup", "sqlite", "test.db")
		if err == nil {
			t.Error("expected error for duplicate connection")
		}
	})

	t.Run("Query", func(t *testing.T) {
		db.Connect("test", "sqlite", "test.db")

		results, err := db.Query("test", "SELECT * FROM users")
		if err != nil {
			t.Fatalf("Query error: %v", err)
		}

		// Mock returns empty results
		if results == nil {
			t.Error("expected results, got nil")
		}
	})

	t.Run("Execute", func(t *testing.T) {
		db.Connect("test2", "sqlite", "test.db")

		rowsAffected, err := db.Execute("test2", "INSERT INTO users VALUES (1, 'Alice')")
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		if rowsAffected == 0 {
			t.Error("expected rows affected, got 0")
		}
	})

	t.Run("Disconnect", func(t *testing.T) {
		db.Connect("test3", "sqlite", "test.db")
		err := db.Disconnect("test3")
		if err != nil {
			t.Fatalf("Disconnect error: %v", err)
		}

		// Should fail on disconnected connection
		_, err = db.Query("test3", "SELECT 1")
		if err == nil {
			t.Error("expected error on disconnected connection")
		}
	})
}

// TestAPITier tests the API tier (@api)
func TestAPITier(t *testing.T) {
	api := NewAPITier()

	// Register a test handler
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"status": "ok"}, nil
	}

	t.Run("RegisterRoute", func(t *testing.T) {
		err := api.RegisterRoute("GET", "/hello", handler)
		if err != nil {
			t.Fatalf("RegisterRoute error: %v", err)
		}
	})

	t.Run("DuplicateRoute", func(t *testing.T) {
		api.RegisterRoute("POST", "/data", handler)
		err := api.RegisterRoute("POST", "/data", handler)
		if err == nil {
			t.Error("expected error for duplicate route")
		}
	})

	t.Run("GetRoute", func(t *testing.T) {
		api.RegisterRoute("DELETE", "/item", handler)

		route, err := api.GetRoute("DELETE", "/item")
		if err != nil {
			t.Fatalf("GetRoute error: %v", err)
		}

		if route.Method != "DELETE" || route.Path != "/item" {
			t.Errorf("got unexpected route: %s %s", route.Method, route.Path)
		}
	})

	t.Run("ListRoutes", func(t *testing.T) {
		api.RegisterRoute("GET", "/a", handler)
		api.RegisterRoute("POST", "/b", handler)
		api.RegisterRoute("PUT", "/c", handler)

		routes := api.ListRoutes()
		if len(routes) < 3 {
			t.Errorf("expected at least 3 routes, got %d", len(routes))
		}
	})

	t.Run("RegisterClient", func(t *testing.T) {
		api.RegisterClient("httpClient", &struct{}{})

		client, err := api.GetClient("httpClient")
		if err != nil {
			t.Fatalf("GetClient error: %v", err)
		}

		if client == nil {
			t.Error("expected client, got nil")
		}
	})
}

// TestExecutionContext tests the full execution environment
func TestExecutionContext(t *testing.T) {
	t.Run("ContextCreation", func(t *testing.T) {
		ec := NewExecutionContext()
		defer ec.Cleanup()

		if ec.MemTier == nil {
			t.Error("memory tier not initialized")
		}
		if ec.CallTier == nil {
			t.Error("call tier not initialized")
		}
		if ec.IPCTier == nil {
			t.Error("IPC tier not initialized")
		}
		if ec.DBTier == nil {
			t.Error("database tier not initialized")
		}
		if ec.APIMTier == nil {
			t.Error("API tier not initialized")
		}
	})

	t.Run("GlobalVariables", func(t *testing.T) {
		ec := NewExecutionContext()
		defer ec.Cleanup()

		ec.Variables["test"] = "value"
		if ec.Variables["test"] != "value" {
			t.Error("global variable not set")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		ec := NewExecutionContext()
		defer ec.Cleanup()

		ec.AddError("@mem", "test error", "test context")
		errors := ec.GetErrors()

		if len(errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(errors))
		}

		if errors[0].Message != "test error" {
			t.Errorf("expected 'test error', got %s", errors[0].Message)
		}
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		ec := NewExecutionContext()
		defer ec.Cleanup()

		// Context should timeout after 5 minutes
		select {
		case <-ec.Context.Done():
			t.Error("context cancelled unexpectedly")
		default:
			// Expected
		}
	})
}

// TestTierIntegration tests integration between tiers
func TestTierIntegration(t *testing.T) {
	ec := NewExecutionContext()
	defer ec.Cleanup()

	t.Run("MemAndCall", func(t *testing.T) {
		// Store value in memory
		ec.MemTier.Set("input", 42)

		// Call function that uses memory
		handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			val, _ := ec.MemTier.Get("input")
			return val, nil
		}

		ec.CallTier.RegisterFunction("getInput", handler)
		result, err := ec.CallTier.Call(context.Background(), "getInput", map[string]interface{}{})

		if err != nil {
			t.Fatalf("Call error: %v", err)
		}

		if result != 42 {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("CallAndIPC", func(t *testing.T) {
		ec.IPCTier.CreateChannel("results")

		handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			ec.IPCTier.Send("results", "handler_result")
			return "ok", nil
		}

		ec.CallTier.RegisterFunction("sendResult", handler)
		ec.CallTier.Call(context.Background(), "sendResult", map[string]interface{}{})

		msg, _ := ec.IPCTier.Receive("results")
		if msg != "handler_result" {
			t.Errorf("expected 'handler_result', got %v", msg)
		}
	})
}

// BenchmarkMemoryTier benchmarks memory operations
func BenchmarkMemoryTier(b *testing.B) {
	mem := NewMemoryTier()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.Set("key", i)
		mem.Get("key")
	}
}

// BenchmarkCallTier benchmarks function invocation
func BenchmarkCallTier(b *testing.B) {
	call := NewCallTier()

	fn := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return nil, nil
	}

	call.RegisterFunction("test", fn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		call.Call(context.Background(), "test", map[string]interface{}{})
	}
}
