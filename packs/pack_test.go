package packs

import (
	"testing"

	"github.com/ollama/ollama/kuhul/runtime"
)

// TestPackInterface verifies all packs implement the Pack interface
func TestPackInterface(t *testing.T) {
	packs := []Pack{
		NewLamOPack(),
		NewSCXQ2Pack(),
		NewASXRAMPack(),
		NewMX2LMPack(),
	}

	for _, pack := range packs {
		t.Run(pack.Name(), func(t *testing.T) {
			// Test Name() implementation
			if name := pack.Name(); name == "" {
				t.Errorf("pack Name() returned empty string")
			}

			// Test Version() implementation
			if version := pack.Version(); version == "" {
				t.Errorf("pack Version() returned empty string")
			}

			// Test Description() implementation
			if desc := pack.Description(); desc == "" {
				t.Errorf("pack Description() returned empty string")
			}

			// Test Handlers() returns a map
			handlers := pack.Handlers()
			if handlers == nil {
				t.Errorf("pack Handlers() returned nil")
			}

			// Test Vectors() returns a map
			vectors := pack.Vectors()
			if vectors == nil {
				t.Errorf("pack Vectors() returned nil")
			}

			// Test Variables() returns a map
			variables := pack.Variables()
			if variables == nil {
				t.Errorf("pack Variables() returned nil")
			}
		})
	}
}

// TestPackRegistration verifies pack registration works
func TestPackRegistration(t *testing.T) {
	tests := []struct {
		name     string
		pack     Pack
		shouldExist bool
	}{
		{"pack_lam_o", NewLamOPack(), true},
		{"pack_scxq2", NewSCXQ2Pack(), true},
		{"pack_asx_ram", NewASXRAMPack(), true},
		{"pack_mx2lm", NewMX2LMPack(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pack, exists := Get(tt.pack.Name())
			if tt.shouldExist {
				if !exists {
					t.Errorf("expected pack %s to be registered, but it wasn't", tt.name)
				}
				if pack == nil {
					t.Errorf("retrieved pack %s is nil", tt.name)
				} else if pack.Name() != tt.pack.Name() {
					t.Errorf("expected pack name %s, got %s", tt.pack.Name(), pack.Name())
				}
			}
		})
	}
}

// TestAllPacks verifies All() returns all registered packs
func TestAllPacks(t *testing.T) {
	packs := All()
	if len(packs) == 0 {
		t.Error("expected at least 1 pack to be registered, got 0")
		return
	}

	packNames := make(map[string]bool)
	for _, pack := range packs {
		packNames[pack.Name()] = true
	}

	expectedPacks := []string{"pack_lam_o", "pack_scxq2", "pack_asx_ram", "pack_mx2lm"}
	for _, expected := range expectedPacks {
		if !packNames[expected] {
			t.Errorf("expected pack %s in All() results, but it wasn't found", expected)
		}
	}
}

// TestLamOPackHandlers verifies lam.o pack handlers
func TestLamOPackHandlers(t *testing.T) {
	pack := NewLamOPack()
	handlers := pack.Handlers()

	requiredHandlers := []string{
		"lam_o.infer",
		"lam_o.chat",
		"lam_o.generate",
		"lam_o.embed",
		"lam_o.list_models",
		"lam_o.show_model",
	}

	for _, handler := range requiredHandlers {
		if _, ok := handlers[handler]; !ok {
			t.Errorf("expected handler %s not found in lam.o pack", handler)
		}
	}
}

// TestLamOPackVariables verifies lam.o pack variables
func TestLamOPackVariables(t *testing.T) {
	pack := NewLamOPack()
	vars := pack.Variables()

	if endpoint, ok := vars["@lam_o_endpoint"]; !ok || endpoint == "" {
		t.Error("@lam_o_endpoint not found or empty in variables")
	}

	if defaultModel, ok := vars["@lam_o_default_model"]; !ok || defaultModel == "" {
		t.Error("@lam_o_default_model not found or empty in variables")
	}
}

// TestSCXQ2PackHandlers verifies scxq2 pack handlers
func TestSCXQ2PackHandlers(t *testing.T) {
	pack := NewSCXQ2Pack()
	handlers := pack.Handlers()

	requiredHandlers := []string{
		"scxq2.fingerprint",
		"scxq2.verify",
		"scxq2.compress",
		"scxq2.decompress",
	}

	for _, handler := range requiredHandlers {
		if _, ok := handlers[handler]; !ok {
			t.Errorf("expected handler %s not found in scxq2 pack", handler)
		}
	}
}

// TestSCXQ2PackVectors verifies scxq2 pack vectors
func TestSCXQ2PackVectors(t *testing.T) {
	pack := NewSCXQ2Pack()
	vectors := pack.Vectors()

	if _, ok := vectors["@fingerprint"]; !ok {
		t.Error("@fingerprint vector not found in scxq2 pack")
	}
}

// TestASXRAMPackHandlers verifies asx_ram pack handlers
func TestASXRAMPackHandlers(t *testing.T) {
	pack := NewASXRAMPack()
	handlers := pack.Handlers()

	requiredHandlers := []string{
		"asx_ram.get",
		"asx_ram.set",
		"asx_ram.delete",
		"asx_ram.list",
		"asx_ram.clear",
	}

	for _, handler := range requiredHandlers {
		if _, ok := handlers[handler]; !ok {
			t.Errorf("expected handler %s not found in asx_ram pack", handler)
		}
	}
}

// TestMX2LMPackHandlers verifies mx2lm pack handlers
func TestMX2LMPackHandlers(t *testing.T) {
	pack := NewMX2LMPack()
	handlers := pack.Handlers()

	requiredHandlers := []string{
		"mx2lm.route",
		"mx2lm.pipeline",
		"mx2lm.broadcast",
		"mx2lm.status",
	}

	for _, handler := range requiredHandlers {
		if _, ok := handlers[handler]; !ok {
			t.Errorf("expected handler %s not found in mx2lm pack", handler)
		}
	}
}

// TestPackInit verifies pack initialization
func TestPackInit(t *testing.T) {
	packs := []Pack{
		NewLamOPack(),
		NewSCXQ2Pack(),
		NewASXRAMPack(),
		NewMX2LMPack(),
	}

	state := runtime.NewRuntimeState()

	for _, pack := range packs {
		t.Run(pack.Name(), func(t *testing.T) {
			err := pack.Init(state)
			if err != nil {
				t.Errorf("pack.Init() error = %v", err)
			}
		})
	}
}

// TestPackHandlerExecution verifies handlers can be executed
func TestPackHandlerExecution(t *testing.T) {
	state := runtime.NewRuntimeState()
	pack := NewSCXQ2Pack()

	// Initialize pack
	if err := pack.Init(state); err != nil {
		t.Fatalf("failed to initialize pack: %v", err)
	}

	handlers := pack.Handlers()
	if _, ok := handlers["scxq2.fingerprint"]; !ok {
		t.Fatal("fingerprint handler not found")
	}

	// Create a context
	ctx := &runtime.Context{
		Body: map[string]interface{}{
			"data": "test data",
		},
	}

	// Execute handler
	result, err := handlers["scxq2.fingerprint"](ctx)
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
		t.Errorf("expected result to be map[string]interface{}, got %T", result)
		return
	}

	if ok, isOk := resultMap["ok"].(bool); !isOk || !ok {
		t.Error("handler result 'ok' field should be true")
	}

	if _, hasFingerprint := resultMap["fingerprint"]; !hasFingerprint {
		t.Error("handler result should have 'fingerprint' field")
	}
}

// TestLamOPackInferHandler tests the infer handler
func TestLamOPackInferHandler(t *testing.T) {
	state := runtime.NewRuntimeState()
	pack := NewLamOPack()

	if err := pack.Init(state); err != nil {
		t.Fatalf("failed to initialize pack: %v", err)
	}

	handlers := pack.Handlers()
	inferHandler, ok := handlers["lam_o.infer"]
	if !ok {
		t.Fatal("infer handler not found")
	}

	ctx := &runtime.Context{
		Body: map[string]interface{}{
			"model":  "llama2",
			"prompt": "What is 2+2?",
		},
	}

	result, err := inferHandler(ctx)
	if err != nil {
		t.Errorf("handler execution error = %v", err)
		return
	}

	if result == nil {
		t.Error("handler returned nil result")
	}
}

// TestASXRAMPackMemoryOperations tests ASXRAM pack operations
func TestASXRAMPackMemoryOperations(t *testing.T) {
	state := runtime.NewRuntimeState()
	pack := NewASXRAMPack()

	if err := pack.Init(state); err != nil {
		t.Fatalf("failed to initialize pack: %v", err)
	}

	handlers := pack.Handlers()

	// Test SET
	setHandler := handlers["asx_ram.set"]
	setCtx := &runtime.Context{
		Body: map[string]interface{}{
			"key":   "test_key",
			"value": "test_value",
		},
	}

	result, _ := setHandler(setCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); !hasOk || !okVal {
			t.Error("set operation should return ok: true")
		}
	}

	// Test GET
	getHandler := handlers["asx_ram.get"]
	getCtx := &runtime.Context{
		Body: map[string]interface{}{
			"key": "test_key",
		},
	}

	result, _ = getHandler(getCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); !hasOk || !okVal {
			t.Error("get operation should return ok: true")
		}
		if value, hasValue := resultMap["value"]; hasValue && value != "test_value" {
			t.Errorf("get operation should return the stored value, got %v", value)
		}
	}

	// Test DELETE
	deleteHandler := handlers["asx_ram.delete"]
	delCtx := &runtime.Context{
		Body: map[string]interface{}{
			"key": "test_key",
		},
	}

	result, _ = deleteHandler(delCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); !hasOk || !okVal {
			t.Error("delete operation should return ok: true")
		}
	}

	// Test GET after DELETE (should return not ok)
	result, _ = getHandler(getCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); hasOk && okVal {
			t.Error("get after delete should return ok: false")
		}
	}
}

// TestMX2LMPackOrchestration tests mx2lm pack orchestration
func TestMX2LMPackOrchestration(t *testing.T) {
	state := runtime.NewRuntimeState()
	pack := NewMX2LMPack()

	if err := pack.Init(state); err != nil {
		t.Fatalf("failed to initialize pack: %v", err)
	}

	handlers := pack.Handlers()

	// Test STATUS
	statusHandler := handlers["mx2lm.status"]
	statusCtx := &runtime.Context{Body: map[string]interface{}{}}

	result, _ := statusHandler(statusCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); !hasOk || !okVal {
			t.Error("status should return ok: true")
		}
		if mode, hasMode := resultMap["mode"].(string); !hasMode || mode != "orchestrator" {
			t.Error("status should return mode: orchestrator")
		}
	}

	// Test ROUTE
	routeHandler := handlers["mx2lm.route"]
	routeCtx := &runtime.Context{
		Body: map[string]interface{}{
			"target":  "llama2",
			"action":  "infer",
			"payload": map[string]interface{}{"prompt": "test"},
		},
	}

	result, _ = routeHandler(routeCtx)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if okVal, hasOk := resultMap["ok"].(bool); !hasOk || !okVal {
			t.Error("route should return ok: true")
		}
		if routed, hasRouted := resultMap["routed"].(bool); !hasRouted || !routed {
			t.Error("route should return routed: true")
		}
	}
}

// TestPackDiscovery tests pack discovery functionality
func TestPackDiscovery(t *testing.T) {
	allPacks := All()

	tests := []struct {
		name string
		desc string
	}{
		{"pack_lam_o", "Llama/Ollama"},
		{"pack_scxq2", "SCXQ2"},
		{"pack_asx_ram", "ASX-RAM"},
		{"pack_mx2lm", "MX2LM"},
	}

	for _, tt := range tests {
		found := false
		for _, pack := range allPacks {
			if pack.Name() == tt.name {
				found = true
				if pack.Description() == "" {
					t.Errorf("pack %s has empty description", tt.name)
				}
				break
			}
		}
		if !found {
			t.Errorf("pack %s not found in discovery", tt.name)
		}
	}
}
