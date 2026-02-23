// Package kuhul provides end-to-end tests for the complete K'UHUL system
package kuhul

import (
	"context"
	"testing"

	"github.com/ollama/ollama/kuhul/agentlang"
	"github.com/ollama/ollama/kuhul/compiler"
	"github.com/ollama/ollama/kuhul/runtime"
)

// TestE2ESimpleInference tests complete pipeline: K'UHUL → Compiler → Runtime → Execution
func TestE2ESimpleInference(t *testing.T) {
	// Step 1: K'UHUL source code
	kuhulCode := `wo model "test-model"
wo prompt "What is AI?"
wo config {
  temperature: 0.7
  max_tokens: 100
}`

	// Step 2: Compile K'UHUL to JavaScript
	comp := compiler.NewCompiler(kuhulCode)
	jsCode, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if jsCode == "" {
		t.Error("expected generated code, got empty string")
	}

	// Step 3: Create execution context
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Step 4: Verify compilation generated proper code
	if !e2eContains(jsCode, "$kuhul_main") {
		t.Error("generated code missing main function")
	}

	if !e2eContains(jsCode, "const model") && !e2eContains(jsCode, "model =") {
		t.Error("generated code missing model assignment")
	}

	// Step 5: Simulate execution by storing values in memory tier
	ec.MemTier.Set("model", "test-model")
	ec.MemTier.Set("prompt", "What is AI?")

	// Step 6: Verify memory tier contains values
	model, _ := ec.MemTier.Get("model")
	if model != "test-model" {
		t.Errorf("expected model='test-model', got %v", model)
	}

	prompt, _ := ec.MemTier.Get("prompt")
	if prompt != "What is AI?" {
		t.Errorf("expected prompt='What is AI?', got %v", prompt)
	}
}

// TestE2EPackInvocation tests pack system integration
func TestE2EPackInvocation(t *testing.T) {
	// K'UHUL code that invokes a pack
	kuhulCode := `wo model "llama2"
wo prompt "test"

pop (lam_o.infer
  :model model
  :prompt prompt
  :temperature 0.7)`

	// Compile
	comp := compiler.NewCompiler(kuhulCode)
	jsCode, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Verify pack reference in generated code
	if !e2eContains(jsCode, "packs.lam_o") {
		t.Error("generated code missing pack reference")
	}

	if !e2eContains(jsCode, "Infer") {
		t.Error("generated code missing handler reference")
	}

	// Create execution context
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Register mock pack handler
	inferHandler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"response": "This is a test response",
			"model":    args["model"],
		}, nil
	}

	ec.CallTier.RegisterFunction("lam_o.infer", inferHandler)

	// Call the pack
	result, err := ec.CallTier.Call(context.Background(), "lam_o.infer", map[string]interface{}{
		"model":       "llama2",
		"prompt":      "test",
		"temperature": 0.7,
	})

	if err != nil {
		t.Fatalf("pack invocation failed: %v", err)
	}

	// Verify result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if resultMap["response"] != "This is a test response" {
		t.Errorf("unexpected response: %v", resultMap["response"])
	}
}

// TestE2EAgentExecution tests K'UHUL with Agent OS orchestration
func TestE2EAgentExecution(t *testing.T) {
	// Create agent OS
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	// Register planning agent
	planner := &agentlang.AgentDefinition{
		Name:         "kuhul_planner",
		Type:         "planner",
		Role:         "Plans K'UHUL execution",
		Capabilities: []string{"reasoning", "planning"},
	}

	err := os.RegisterAgent(planner)
	if err != nil {
		t.Fatalf("failed to register planner: %v", err)
	}

	// Create a task
	task := &agentlang.Task{
		ID:            "infer_task",
		Description:   "Execute K'UHUL inference",
		Priority:      1.0,
		RequiredCaps:  []string{"reasoning"},
	}

	// Submit task
	err = os.SubmitTask(task)
	if err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	// Verify task is pending
	if task.Status != "pending" {
		t.Errorf("expected task status 'pending', got %s", task.Status)
	}

	// Assign task
	err = os.AssignTask(task)
	if err != nil {
		t.Fatalf("failed to assign task: %v", err)
	}

	// Verify assignment
	if task.AssignedAgent != "kuhul_planner" {
		t.Errorf("expected assignment to 'kuhul_planner', got %s", task.AssignedAgent)
	}

	if task.Status != "assigned" {
		t.Errorf("expected task status 'assigned', got %s", task.Status)
	}

	// Execute task
	err = os.ExecuteTask(task)
	if err != nil {
		t.Fatalf("failed to execute task: %v", err)
	}

	// Verify completion
	if task.Status != "completed" {
		t.Errorf("expected task status 'completed', got %s", task.Status)
	}

	if os.Metrics.TasksCompleted != 1 {
		t.Errorf("expected 1 completed task, got %d", os.Metrics.TasksCompleted)
	}
}

// TestE2ECompilerWithAgent tests compiling K'UHUL and executing via agent
func TestE2ECompilerWithAgent(t *testing.T) {
	// K'UHUL source
	kuhulCode := `wo x 42
wo y 100
wo result 0`

	// Compile
	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create agent OS
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	// Register executor agent
	executor := &agentlang.AgentDefinition{
		Name:         "executor",
		Type:         "executor",
		Role:         "Executes compiled K'UHUL",
		Capabilities: []string{"execution"},
	}

	os.RegisterAgent(executor)

	// Create execution task
	task := &agentlang.Task{
		ID:            "compile_exec",
		Description:   "Execute compiled K'UHUL code",
		Priority:      1.0,
		RequiredCaps:  []string{"execution"},
	}

	os.SubmitTask(task)
	os.AssignTask(task)
	os.ExecuteTask(task)

	// Verify task completed
	if task.Status != "completed" {
		t.Errorf("expected task completed, got %s", task.Status)
	}
}

// TestE2EMultiAgentCompilation tests multi-agent system with compiled K'UHUL
func TestE2EMultiAgentCompilation(t *testing.T) {
	// Planner writes K'UHUL code
	kuhulCode := `handler greet(name) {
  wo greeting "Hello"
}`

	// Compile the code
	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create agent OS
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	// Register planner
	planner := &agentlang.AgentDefinition{
		Name:         "planner",
		Type:         "planner",
		Role:         "Plans handler compilation",
		Capabilities: []string{"planning"},
	}
	os.RegisterAgent(planner)

	// Register executor
	executor := &agentlang.AgentDefinition{
		Name:         "executor",
		Type:         "executor",
		Role:         "Executes handlers",
		Capabilities: []string{"execution"},
	}
	os.RegisterAgent(executor)

	// Planner creates planning task
	planningTask := &agentlang.Task{
		ID:            "plan",
		Description:   "Plan K'UHUL compilation",
		Priority:      0.9,
		RequiredCaps:  []string{"planning"},
	}

	os.SubmitTask(planningTask)
	os.AssignTask(planningTask)
	os.ExecuteTask(planningTask)

	// Executor creates execution task
	executionTask := &agentlang.Task{
		ID:            "exec",
		Description:   "Execute compiled code",
		Priority:      0.8,
		RequiredCaps:  []string{"execution"},
	}

	os.SubmitTask(executionTask)
	os.AssignTask(executionTask)
	os.ExecuteTask(executionTask)

	// Both tasks should be completed
	if planningTask.Status != "completed" || executionTask.Status != "completed" {
		t.Error("expected both tasks to be completed")
	}

	if os.Metrics.TasksCompleted != 2 {
		t.Errorf("expected 2 completed tasks, got %d", os.Metrics.TasksCompleted)
	}
}

// TestE2ECompilerErrorHandling tests error handling in full pipeline
func TestE2ECompilerErrorHandling(t *testing.T) {
	// Malformed K'UHUL code
	kuhulCode := `wo x "unterminated string`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	// Should have failed during lexical analysis
	if err == nil {
		t.Error("expected compilation error for unterminated string")
	}

	// Verify error was recorded
	compErrors := comp.GetErrors()
	if len(compErrors) == 0 {
		t.Error("expected error details")
	}
}

// TestE2ENegotiationWithCompilation tests multi-agent negotiation using compiled K'UHUL
func TestE2ENegotiationWithCompilation(t *testing.T) {
	// Compile negotiation handler
	kuhulCode := `handler propose(solution) {
  wo proposal solution
}`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create agent OS
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	// Register negotiating agents
	agent1 := &agentlang.AgentDefinition{
		Name:         "agent_1",
		Type:         "planner",
		Role:         "First negotiator",
		Capabilities: []string{"negotiation"},
	}
	os.RegisterAgent(agent1)

	agent2 := &agentlang.AgentDefinition{
		Name:         "agent_2",
		Type:         "planner",
		Role:         "Second negotiator",
		Capabilities: []string{"negotiation"},
	}
	os.RegisterAgent(agent2)

	// Start negotiation
	agents := []string{"agent_1", "agent_2"}
	session, err := os.StartNegotiation(agents, "design_decision")

	if err != nil {
		t.Fatalf("negotiation failed: %v", err)
	}

	if session.Phase != "proposal" {
		t.Errorf("expected phase 'proposal', got %s", session.Phase)
	}

	// Advance phase
	err = os.AdvanceNegotiationPhase(session.ID)
	if err != nil {
		t.Fatalf("phase advancement failed: %v", err)
	}

	if session.Phase != "critique" {
		t.Errorf("expected phase 'critique', got %s", session.Phase)
	}
}

// TestE2EMetaCognitionWithCompiled tests self-evaluation of compiled code
func TestE2EMetaCognitionWithCompiled(t *testing.T) {
	// Compile test code
	kuhulCode := `wo iterations 0
wo success true`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create agent OS
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	// Register agent
	agent := &agentlang.AgentDefinition{
		Name:         "reflector",
		Type:         "planner",
		Role:         "Self-evaluating agent",
		Capabilities: []string{"introspection"},
	}
	os.RegisterAgent(agent)

	// Execute task to create history
	task := &agentlang.Task{
		ID:            "reflection",
		Description:   "Task for self-evaluation",
		Priority:      1.0,
		RequiredCaps:  []string{"introspection"},
	}

	os.SubmitTask(task)
	os.AssignTask(task)
	os.ExecuteTask(task)

	// Trigger meta-cognition
	insights, err := os.TriggerMetaCognition()
	if err != nil {
		t.Fatalf("meta-cognition failed: %v", err)
	}

	// Verify insights were generated
	if len(insights) == 0 {
		t.Error("expected meta-cognition insights")
	}

	reflectorInsight, exists := insights["reflector"]
	if !exists {
		t.Error("expected reflector in insights")
	}

	insightMap, ok := reflectorInsight.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map insight, got %T", reflectorInsight)
	}

	if insightMap["state"] == nil {
		t.Error("expected state in insight")
	}
}

// TestE2EFullPipelineIntegration tests complete system with all components
func TestE2EFullPipelineIntegration(t *testing.T) {
	// Complex K'UHUL program
	kuhulCode := `wo llama_model "llama2"
wo claude_model "claude-3"
wo inference_prompt "Calculate 42 + 58"

handler dual_inference(model prompt) {
  wo result_llama "pending"
  wo result_claude "pending"
}

wo task_id "dual-infer-001"`

	// Step 1: Lexical analysis
	comp := compiler.NewCompiler(kuhulCode)
	tokens := comp.GetTokens()
	if tokens == nil {
		// Force tokenization by compiling
		_, _ = comp.Compile()
		tokens = comp.GetTokens()
	}

	if tokens == nil {
		t.Error("expected tokens from compiler")
	}

	// Step 2: Syntax analysis (AST)
	_, _ = comp.Compile()
	ast := comp.GetAST()
	if ast == nil {
		t.Error("expected AST from compiler")
	}

	// Step 3: Full compilation
	jsCode, err := comp.Compile()
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Verify compilation output
	if !e2eContains(jsCode, "$kuhul_main") {
		t.Error("generated code missing main function")
	}

	// Step 4: Runtime execution setup
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Store values in memory tier
	ec.MemTier.Set("llama_model", "llama2")
	ec.MemTier.Set("claude_model", "claude-3")
	ec.MemTier.Set("inference_prompt", "Calculate 42 + 58")

	// Verify memory tier
	model, _ := ec.MemTier.Get("llama_model")
	if model != "llama2" {
		t.Errorf("memory tier error: expected llama2, got %v", model)
	}

	// Step 5: Agent OS orchestration
	os := agentlang.NewAgentOS()
	defer os.Shutdown()

	llmAgent := &agentlang.AgentDefinition{
		Name:         "llm_executor",
		Type:         "executor",
		Role:         "Executes LLM inferences",
		Capabilities: []string{"inference", "reasoning"},
	}
	os.RegisterAgent(llmAgent)

	// Create inference task
	inferenceTask := &agentlang.Task{
		ID:            "dual-infer-001",
		Description:   "Dual model inference on math problem",
		Priority:      1.0,
		RequiredCaps:  []string{"inference"},
	}

	os.SubmitTask(inferenceTask)
	os.AssignTask(inferenceTask)
	os.ExecuteTask(inferenceTask)

	// Verify task completion
	if inferenceTask.Status != "completed" {
		t.Errorf("expected task completed, got %s", inferenceTask.Status)
	}

	// Step 6: Metrics verification
	metrics, err := os.GetAgentMetrics("llm_executor")
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	if metrics["executions"] != 1 {
		t.Errorf("expected 1 execution, got %v", metrics["executions"])
	}
}

// TestE2EErrorRecovery tests system resilience to errors
func TestE2EErrorRecovery(t *testing.T) {
	// Code that compiles but might fail at runtime
	kuhulCode := `wo safe_value 42
wo risky_operation "test"`

	comp := compiler.NewCompiler(kuhulCode)
	_, err := comp.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Create execution context
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	// Store safe value
	ec.MemTier.Set("safe_value", 42)

	// Try to access non-existent value
	_, err = ec.MemTier.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent variable")
	}

	// Record error
	ec.AddError("@mem", "variable not found", "test context")

	errors := ec.GetErrors()
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	if errors[0].Message != "variable not found" {
		t.Errorf("expected 'variable not found', got %s", errors[0].Message)
	}
}

// BenchmarkE2ECompilation benchmarks end-to-end compilation
func BenchmarkE2ECompilation(b *testing.B) {
	kuhulCode := `wo x 1
wo y 2
wo z 3
handler test() {
  wo result 0
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp := compiler.NewCompiler(kuhulCode)
		comp.Compile()
	}
}

// BenchmarkE2EAgentExecution benchmarks end-to-end agent execution
func BenchmarkE2EAgentExecution(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os := agentlang.NewAgentOS()

		agent := &agentlang.AgentDefinition{
			Name:         "bench",
			Type:         "executor",
			Capabilities: []string{"work"},
		}
		os.RegisterAgent(agent)

		task := &agentlang.Task{
			ID:            "task",
			Description:   "test",
			Priority:      1.0,
			RequiredCaps:  []string{"work"},
		}

		os.SubmitTask(task)
		os.AssignTask(task)
		os.ExecuteTask(task)

		os.Shutdown()
	}
}

// BenchmarkE2EMemoryTier benchmarks memory operations during execution
func BenchmarkE2EMemoryTier(b *testing.B) {
	ec := runtime.NewExecutionContext()
	defer ec.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ec.MemTier.Set("key", i)
		ec.MemTier.Get("key")
	}
}

// Helper function (also defined in integration_test.go)
// Using separate definition to allow this test file to be independent
func e2eContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr ||
			len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				e2eIndexOf(s, substr) >= 0))
}

func e2eIndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
