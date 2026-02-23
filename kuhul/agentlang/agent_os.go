// Package agentlang/agent_os provides the Agent OS orchestration layer
// that integrates K'UHUL agents with the 5-tier runtime
package agentlang

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentOS represents the complete agent operating system
type AgentOS struct {
	// Core components
	Agents            map[string]*RunningAgent
	KnowledgeGraph    *KnowledgeGraph
	MetaCognition     *MetaCognition
	TemporalReasoner  *TemporalReasoner
	NegotiationEngine *NegotiationEngine
	ValueSystem       *ValueSystem

	// State management
	mu                sync.RWMutex
	TaskQueue         *TaskQueue
	MemoryTier        *MemoryTier
	ExecutionLog      []ExecutionEvent
	GlobalContext     context.Context
	Cancel            context.CancelFunc

	// Metrics
	Metrics *OSMetrics
}

// RunningAgent represents an executing agent instance
type RunningAgent struct {
	Definition   *AgentDefinition
	State        string // "idle", "active", "blocked", "error"
	Memory       *AgentMemory
	CurrentTask  *Task
	History      []*ExecutionRecord
	Confidence   float64
	LastUpdated  time.Time
	mu            sync.RWMutex
}

// Task represents a unit of work
type Task struct {
	ID            string
	Description   string
	Priority      float64
	Deadline      time.Time
	RequiredCaps  []string
	AssignedAgent string
	Status        string // "pending", "assigned", "executing", "completed", "failed"
	Result        interface{}
	Error         error
}

// TaskQueue manages pending work
type TaskQueue struct {
	Tasks map[string]*Task
	Queue []*Task
	mu    sync.Mutex
}

// ExecutionRecord tracks an agent's action
type ExecutionRecord struct {
	Timestamp   time.Time
	Action      string
	Input       interface{}
	Output      interface{}
	Error       error
	DurationMs  int64
	Reward      float64
}

// ExecutionEvent represents an event in the system
type ExecutionEvent struct {
	Timestamp time.Time
	EventType string // "agent_created", "task_completed", "negotiation", "error"
	Agent     string
	Details   map[string]interface{}
}

// TemporalReasoner handles time-aware planning
type TemporalReasoner struct {
	Horizons        map[string]*PlanningHorizon
	ScheduledTasks  []*ScheduledTask
	PredictionModel map[string]float64
}

// ScheduledTask represents a task scheduled for a specific time
type ScheduledTask struct {
	Task      *Task
	ScheduledTime time.Time
	Horizon   string
}

// NegotiationEngine handles multi-agent consensus
type NegotiationEngine struct {
	Protocol      string
	Proposals     map[string]*Proposal
	ActiveSession *NegotiationSession
}

// Proposal represents a suggested solution
type Proposal struct {
	ID          string
	AgentID     string
	Content     map[string]interface{}
	Reasoning   string
	Confidence  float64
	Critiques   []*Critique
	Revisions   int
}

// Critique represents feedback on a proposal
type Critique struct {
	AgentID  string
	Feedback string
	Score    float64
}

// NegotiationSession represents an ongoing negotiation
type NegotiationSession struct {
	ID       string
	Agents   []string
	Phase    string // "proposal", "critique", "revision", "agreement"
	Proposals []*Proposal
	Agreement map[string]interface{}
}

// MemoryTier represents OS-level memory
type MemoryTier struct {
	Shared       map[string]interface{}
	AgentPrivate map[string]map[string]interface{}
	mu           sync.RWMutex
}

// OSMetrics tracks system performance
type OSMetrics struct {
	TasksCompleted   int64
	TasksFailed      int64
	AverageLatency   float64
	AverageReward    float64
	NegotiationCount int64
	AgentUtilization map[string]float64
}

// NewAgentOS creates a new Agent OS instance
func NewAgentOS() *AgentOS {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)

	return &AgentOS{
		Agents:           make(map[string]*RunningAgent),
		KnowledgeGraph:   &KnowledgeGraph{Nodes: make(map[string]*KnowledgeNode), Edges: make(map[string]*KnowledgeEdge)},
		MetaCognition:    &MetaCognition{IntrospectionLoops: make(map[string]*IntrospectionLoop)},
		TemporalReasoner: &TemporalReasoner{Horizons: make(map[string]*PlanningHorizon)},
		NegotiationEngine: &NegotiationEngine{Proposals: make(map[string]*Proposal)},
		ValueSystem:      &ValueSystem{},
		TaskQueue:        &TaskQueue{Tasks: make(map[string]*Task), Queue: make([]*Task, 0)},
		MemoryTier:       &MemoryTier{Shared: make(map[string]interface{}), AgentPrivate: make(map[string]map[string]interface{})},
		ExecutionLog:     make([]ExecutionEvent, 0),
		GlobalContext:    ctx,
		Cancel:           cancel,
		Metrics:          &OSMetrics{AgentUtilization: make(map[string]float64)},
	}
}

// RegisterAgent registers an agent with the OS
func (os *AgentOS) RegisterAgent(def *AgentDefinition) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	if _, exists := os.Agents[def.Name]; exists {
		return fmt.Errorf("agent %q already registered", def.Name)
	}

	runningAgent := &RunningAgent{
		Definition:  def,
		State:       "idle",
		Memory:      def.Memory,
		History:     make([]*ExecutionRecord, 0),
		Confidence:  1.0,
		LastUpdated: time.Now(),
	}

	os.Agents[def.Name] = runningAgent
	os.logEvent("agent_created", def.Name, map[string]interface{}{"agent": def.Name, "type": def.Type})

	return nil
}

// SubmitTask submits a task to the task queue
func (os *AgentOS) SubmitTask(task *Task) error {
	os.TaskQueue.mu.Lock()
	defer os.TaskQueue.mu.Unlock()

	if _, exists := os.TaskQueue.Tasks[task.ID]; exists {
		return fmt.Errorf("task %q already exists", task.ID)
	}

	task.Status = "pending"
	os.TaskQueue.Tasks[task.ID] = task
	os.TaskQueue.Queue = append(os.TaskQueue.Queue, task)

	return nil
}

// AssignTask finds an appropriate agent and assigns the task
func (os *AgentOS) AssignTask(task *Task) error {
	os.mu.RLock()
	defer os.mu.RUnlock()

	// Find agent with required capabilities
	var targetAgent *RunningAgent

	for _, agent := range os.Agents {
		if agent.State == "idle" && hasCapabilities(agent.Definition.Capabilities, task.RequiredCaps) {
			if targetAgent == nil || agent.Confidence > targetAgent.Confidence {
				targetAgent = agent
			}
		}
	}

	if targetAgent == nil {
		return fmt.Errorf("no available agent for task %s", task.ID)
	}

	targetAgent.mu.Lock()
	targetAgent.CurrentTask = task
	targetAgent.State = "active"
	targetAgent.mu.Unlock()

	task.AssignedAgent = targetAgent.Definition.Name
	task.Status = "assigned"

	return nil
}

// ExecuteTask executes a task (synchronous for now)
func (os *AgentOS) ExecuteTask(task *Task) error {
	// Execute the task (placeholder)
	startTime := time.Now()

	// Simulate work
	time.Sleep(100 * time.Millisecond)

	// Record execution
	agent, err := os.getAgent(task.AssignedAgent)
	if err != nil {
		return err
	}

	record := &ExecutionRecord{
		Timestamp:  time.Now(),
		Action:     fmt.Sprintf("execute_%s", task.ID),
		Input:      task,
		Output:     "task_completed",
		Error:      nil,
		DurationMs: int64(time.Since(startTime).Milliseconds()),
		Reward:     1.0, // Would be computed from reward signals
	}

	agent.mu.Lock()
	agent.History = append(agent.History, record)
	agent.CurrentTask = nil
	agent.State = "idle"
	agent.mu.Unlock()

	task.Status = "completed"
	os.Metrics.TasksCompleted++

	os.logEvent("task_completed", agent.Definition.Name, map[string]interface{}{"task": task.ID})

	return nil
}

// StartNegotiation initiates a multi-agent negotiation
func (os *AgentOS) StartNegotiation(agents []string, subject string) (*NegotiationSession, error) {
	session := &NegotiationSession{
		ID:        fmt.Sprintf("neg_%d", time.Now().Unix()),
		Agents:    agents,
		Phase:     "proposal",
		Proposals: make([]*Proposal, 0),
		Agreement: make(map[string]interface{}),
	}

	os.NegotiationEngine.ActiveSession = session
	os.Metrics.NegotiationCount++

	os.logEvent("negotiation_started", "", map[string]interface{}{"agents": agents, "subject": subject})

	return session, nil
}

// AdvanceNegotiationPhase moves to the next phase
func (os *AgentOS) AdvanceNegotiationPhase(sessionID string) error {
	session := os.NegotiationEngine.ActiveSession
	if session == nil || session.ID != sessionID {
		return fmt.Errorf("session %s not found", sessionID)
	}

	phases := []string{"proposal", "critique", "revision", "agreement"}
	currentIdx := indexOf(phases, session.Phase)

	if currentIdx < len(phases)-1 {
		session.Phase = phases[currentIdx+1]
	}

	os.logEvent("negotiation_phase_advanced", "", map[string]interface{}{"phase": session.Phase})

	return nil
}

// TriggerMetaCognition runs self-evaluation
func (os *AgentOS) TriggerMetaCognition() (map[string]interface{}, error) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	insights := make(map[string]interface{})

	for agentName, agent := range os.Agents {
		agent.mu.RLock()

		agentInsight := map[string]interface{}{
			"state":       agent.State,
			"confidence":  agent.Confidence,
			"history_len": len(agent.History),
			"last_updated": agent.LastUpdated,
		}

		if len(agent.History) > 0 {
			lastRecord := agent.History[len(agent.History)-1]
			agentInsight["last_reward"] = lastRecord.Reward
			agentInsight["last_duration_ms"] = lastRecord.DurationMs
		}

		insights[agentName] = agentInsight
		agent.mu.RUnlock()
	}

	os.logEvent("meta_cognition_triggered", "", insights)

	return insights, nil
}

// GetMemoryValue retrieves a value from shared memory
func (os *AgentOS) GetMemoryValue(key string) (interface{}, error) {
	os.MemoryTier.mu.RLock()
	defer os.MemoryTier.mu.RUnlock()

	val, exists := os.MemoryTier.Shared[key]
	if !exists {
		return nil, fmt.Errorf("key %q not found in shared memory", key)
	}

	return val, nil
}

// SetMemoryValue stores a value in shared memory
func (os *AgentOS) SetMemoryValue(key string, value interface{}) {
	os.MemoryTier.mu.Lock()
	defer os.MemoryTier.mu.Unlock()

	os.MemoryTier.Shared[key] = value
}

// GetAgentMetrics returns metrics for a specific agent
func (os *AgentOS) GetAgentMetrics(agentName string) (map[string]interface{}, error) {
	agent, err := os.getAgent(agentName)
	if err != nil {
		return nil, err
	}

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	totalReward := 0.0
	for _, record := range agent.History {
		totalReward += record.Reward
	}

	avgReward := 0.0
	if len(agent.History) > 0 {
		avgReward = totalReward / float64(len(agent.History))
	}

	return map[string]interface{}{
		"name":          agent.Definition.Name,
		"state":         agent.State,
		"confidence":    agent.Confidence,
		"executions":    len(agent.History),
		"avg_reward":    avgReward,
		"total_reward":  totalReward,
		"last_updated":  agent.LastUpdated,
	}, nil
}

// Shutdown gracefully shuts down the OS
func (os *AgentOS) Shutdown() error {
	os.Cancel()
	os.mu.Lock()
	defer os.mu.Unlock()

	os.logEvent("shutdown", "", map[string]interface{}{
		"tasks_completed": os.Metrics.TasksCompleted,
		"tasks_failed":    os.Metrics.TasksFailed,
	})

	return nil
}

// Helper functions

func (os *AgentOS) getAgent(name string) (*RunningAgent, error) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	agent, exists := os.Agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %q not found", name)
	}

	return agent, nil
}

func (os *AgentOS) logEvent(eventType, agent string, details map[string]interface{}) {
	event := ExecutionEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		Agent:     agent,
		Details:   details,
	}

	os.ExecutionLog = append(os.ExecutionLog, event)
}

func hasCapabilities(agentCaps, requiredCaps []string) bool {
	capMap := make(map[string]bool)
	for _, cap := range agentCaps {
		capMap[cap] = true
	}

	for _, req := range requiredCaps {
		if !capMap[req] {
			return false
		}
	}

	return true
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
