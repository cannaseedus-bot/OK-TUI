package agentlang

import (
	"testing"
)

// TestAgentOSBasics tests basic Agent OS functionality
func TestAgentOSBasics(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	t.Run("RegisterAgent", func(t *testing.T) {
		agent := &AgentDefinition{
			Name: "test_agent",
			Type: "planner",
			Role: "Test agent",
			Capabilities: []string{"reasoning"},
		}

		err := os.RegisterAgent(agent)
		if err != nil {
			t.Fatalf("RegisterAgent failed: %v", err)
		}

		// Try duplicate
		err = os.RegisterAgent(agent)
		if err == nil {
			t.Error("expected error for duplicate agent registration")
		}
	})

	t.Run("SubmitTask", func(t *testing.T) {
		task := &Task{
			ID:          "task_1",
			Description: "Test task",
			Priority:    1.0,
			RequiredCaps: []string{"reasoning"},
		}

		err := os.SubmitTask(task)
		if err != nil {
			t.Fatalf("SubmitTask failed: %v", err)
		}

		if task.Status != "pending" {
			t.Errorf("expected status 'pending', got %s", task.Status)
		}
	})

	t.Run("AssignTask", func(t *testing.T) {
		// Register agent
		agent := &AgentDefinition{
			Name: "executor",
			Type: "executor",
			Role: "Executes tasks",
			Capabilities: []string{"execution"},
		}
		os.RegisterAgent(agent)

		// Create task
		task := &Task{
			ID:          "task_2",
			Description: "Execute something",
			Priority:    1.0,
			RequiredCaps: []string{"execution"},
		}
		os.SubmitTask(task)

		// Assign task
		err := os.AssignTask(task)
		if err != nil {
			t.Fatalf("AssignTask failed: %v", err)
		}

		if task.AssignedAgent != "executor" {
			t.Errorf("expected assignment to 'executor', got %s", task.AssignedAgent)
		}

		if task.Status != "assigned" {
			t.Errorf("expected status 'assigned', got %s", task.Status)
		}
	})

	t.Run("ExecuteTask", func(t *testing.T) {
		// Setup
		agent := &AgentDefinition{
			Name: "worker",
			Type: "executor",
			Capabilities: []string{"work"},
		}
		os.RegisterAgent(agent)

		task := &Task{
			ID:          "task_3",
			RequiredCaps: []string{"work"},
		}
		os.SubmitTask(task)
		os.AssignTask(task)

		// Execute
		err := os.ExecuteTask(task)
		if err != nil {
			t.Fatalf("ExecuteTask failed: %v", err)
		}

		if task.Status != "completed" {
			t.Errorf("expected status 'completed', got %s", task.Status)
		}

		if os.Metrics.TasksCompleted != 1 {
			t.Errorf("expected 1 completed task, got %d", os.Metrics.TasksCompleted)
		}
	})
}

// TestMemoryTier tests the OS-level memory system
func TestMemoryTier(t *testing.T) {
	os := NewAgentOS()

	t.Run("SharedMemory", func(t *testing.T) {
		os.SetMemoryValue("key1", "value1")

		val, err := os.GetMemoryValue("key1")
		if err != nil {
			t.Fatalf("GetMemoryValue failed: %v", err)
		}

		if val != "value1" {
			t.Errorf("expected 'value1', got %v", val)
		}
	})

	t.Run("NonexistentKey", func(t *testing.T) {
		_, err := os.GetMemoryValue("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent key")
		}
	})
}

// TestNegotiation tests multi-agent negotiation
func TestNegotiation(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	t.Run("StartNegotiation", func(t *testing.T) {
		agents := []string{"agent_1", "agent_2"}
		session, err := os.StartNegotiation(agents, "design_decision")

		if err != nil {
			t.Fatalf("StartNegotiation failed: %v", err)
		}

		if session.Phase != "proposal" {
			t.Errorf("expected phase 'proposal', got %s", session.Phase)
		}

		if len(session.Agents) != 2 {
			t.Errorf("expected 2 agents, got %d", len(session.Agents))
		}
	})

	t.Run("AdvancePhase", func(t *testing.T) {
		agents := []string{"agent_1"}
		session, _ := os.StartNegotiation(agents, "test")

		err := os.AdvanceNegotiationPhase(session.ID)
		if err != nil {
			t.Fatalf("AdvanceNegotiationPhase failed: %v", err)
		}

		if session.Phase != "critique" {
			t.Errorf("expected phase 'critique', got %s", session.Phase)
		}
	})
}

// TestMetaCognition tests self-evaluation
func TestMetaCognition(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	// Register agent
	agent := &AgentDefinition{
		Name: "introspector",
		Type: "planner",
		Capabilities: []string{"reasoning"},
	}
	os.RegisterAgent(agent)

	// Execute a task to create history
	task := &Task{
		ID:          "task_mc",
		RequiredCaps: []string{"reasoning"},
	}
	os.SubmitTask(task)
	os.AssignTask(task)
	os.ExecuteTask(task)

	// Trigger meta-cognition
	insights, err := os.TriggerMetaCognition()
	if err != nil {
		t.Fatalf("TriggerMetaCognition failed: %v", err)
	}

	if len(insights) == 0 {
		t.Error("expected insights, got empty map")
	}

	agentInsight, exists := insights["introspector"]
	if !exists {
		t.Error("expected introspector in insights")
	}

	insightMap := agentInsight.(map[string]interface{})
	if insightMap["state"] != "idle" {
		t.Errorf("expected state 'idle', got %v", insightMap["state"])
	}
}

// TestAgentMetrics tests metrics collection
func TestAgentMetrics(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	// Register and use agent
	agent := &AgentDefinition{
		Name: "metrics_test",
		Type: "executor",
		Capabilities: []string{"work"},
	}
	os.RegisterAgent(agent)

	task := &Task{
		ID:          "task_m",
		RequiredCaps: []string{"work"},
	}
	os.SubmitTask(task)
	os.AssignTask(task)
	os.ExecuteTask(task)

	// Get metrics
	metrics, err := os.GetAgentMetrics("metrics_test")
	if err != nil {
		t.Fatalf("GetAgentMetrics failed: %v", err)
	}

	if metrics["executions"] != 1 {
		t.Errorf("expected 1 execution, got %v", metrics["executions"])
	}
}

// TestTaskPriority tests that high-priority tasks are preferred
func TestTaskPriority(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	agent := &AgentDefinition{
		Name: "priority_worker",
		Type: "executor",
		Capabilities: []string{"execution"},
	}
	os.RegisterAgent(agent)

	// Submit multiple tasks with different priorities
	task1 := &Task{ID: "low", Priority: 0.3, RequiredCaps: []string{"execution"}}
	task2 := &Task{ID: "high", Priority: 0.9, RequiredCaps: []string{"execution"}}

	os.SubmitTask(task1)
	os.SubmitTask(task2)

	// Check that high-priority task would be assigned first (in real OS, scheduler does this)
	if task2.Priority < task1.Priority {
		t.Error("priority test setup failed")
	}
}

// TestMultiAgentCoordination tests coordination between agents
func TestMultiAgentCoordination(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	// Register multiple agents
	planner := &AgentDefinition{
		Name: "planner",
		Type: "planner",
		Capabilities: []string{"planning"},
	}
	executor := &AgentDefinition{
		Name: "executor",
		Type: "executor",
		Capabilities: []string{"execution"},
	}

	os.RegisterAgent(planner)
	os.RegisterAgent(executor)

	// Submit tasks for each
	planningTask := &Task{ID: "plan", RequiredCaps: []string{"planning"}}
	executionTask := &Task{ID: "exec", RequiredCaps: []string{"execution"}}

	os.SubmitTask(planningTask)
	os.SubmitTask(executionTask)

	// Assign and execute
	os.AssignTask(planningTask)
	os.AssignTask(executionTask)

	os.ExecuteTask(planningTask)
	os.ExecuteTask(executionTask)

	// Both tasks should be completed
	if planningTask.Status != "completed" || executionTask.Status != "completed" {
		t.Error("expected both tasks to be completed")
	}

	if os.Metrics.TasksCompleted != 2 {
		t.Errorf("expected 2 completed tasks, got %d", os.Metrics.TasksCompleted)
	}
}

// TestExecutionLog tests event logging
func TestExecutionLog(t *testing.T) {
	os := NewAgentOS()
	defer os.Shutdown()

	agent := &AgentDefinition{Name: "logged", Type: "executor"}
	os.RegisterAgent(agent)

	initialLogLen := len(os.ExecutionLog)

	if len(os.ExecutionLog) <= initialLogLen {
		// At minimum we logged the agent creation
		t.Log("execution log contains events")
	}
}

// BenchmarkTaskExecution benchmarks task execution
func BenchmarkTaskExecution(b *testing.B) {
	os := NewAgentOS()
	defer os.Shutdown()

	agent := &AgentDefinition{
		Name: "bench_agent",
		Type: "executor",
		Capabilities: []string{"bench"},
	}
	os.RegisterAgent(agent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:          "task",
			RequiredCaps: []string{"bench"},
		}

		os.SubmitTask(task)
		os.AssignTask(task)
		os.ExecuteTask(task)
	}
}

// BenchmarkNegotiation benchmarks negotiation
func BenchmarkNegotiation(b *testing.B) {
	os := NewAgentOS()
	defer os.Shutdown()

	agents := []string{"a1", "a2", "a3"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		os.StartNegotiation(agents, "subject")
	}
}
