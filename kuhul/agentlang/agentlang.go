// Package agentlang extends K'UHUL with agent definitions, meta-cognition,
// and knowledge graph operations for building autonomous multi-agent systems
package agentlang

import (
	"fmt"
)

// AgentDefinition represents a K'UHUL agent specification
type AgentDefinition struct {
	Name          string
	Type          string // "planner", "executor", "analyst"
	Language      string // "kuhul", "typescript", "python"
	Role          string
	Goals         []Goal
	BehaviorTree  *BehaviorTree
	RewardSignals map[string]float64
	Capabilities  []string
	Memory        *AgentMemory
}

// Goal represents an agent goal with priorities and constraints
type Goal struct {
	Name       string
	Priority   float64 // 0.0 to 1.0
	Constraint string
	Deadline   string // interval notation
	Metrics    []string
}

// BehaviorTree represents the agent's decision tree
type BehaviorTree struct {
	Root  *BehaviorNode
	Nodes map[string]*BehaviorNode
}

// BehaviorNode represents a single node in the behavior tree
type BehaviorNode struct {
	ID       string
	Type     string // "selector", "sequence", "condition", "action"
	Children []string
	Action   string
	Condition string
	Weight   float64
}

// AgentMemory represents memory modules for an agent
type AgentMemory struct {
	ShortTerm *MemoryModule
	LongTerm  *MemoryModule
	Episodic  *MemoryModule
}

// MemoryModule represents a memory storage and retrieval system
type MemoryModule struct {
	Capacity             int
	DecayRate            float64
	ConsolidationStrategy string
	Format               string
	Entries              map[string]interface{}
}

// MetaCognition represents the agent's ability to think about thinking
type MetaCognition struct {
	IntrospectionLoops map[string]*IntrospectionLoop
	ExplanationFormat  string
	ConfidenceScores   map[string]float64
}

// IntrospectionLoop represents a self-evaluation cycle
type IntrospectionLoop struct {
	Name        string
	Frequency   string
	Monitors    []string
	Adaptations []string
}

// KnowledgeGraph represents semantic knowledge structure
type KnowledgeGraph struct {
	Nodes      map[string]*KnowledgeNode
	Edges      map[string]*KnowledgeEdge
	Queries    []GraphQuery
	InferenceRules []string
}

// KnowledgeNode represents a concept in the knowledge graph
type KnowledgeNode struct {
	ID         string
	Type       string // "concept", "procedure", "relationship"
	Properties map[string]interface{}
	Confidence float64
}

// KnowledgeEdge represents relationships in the knowledge graph
type KnowledgeEdge struct {
	From       string
	To         string
	RelationType string // "causal", "hierarchical", "associative"
	Strength   float64
}

// GraphQuery represents a query against the knowledge graph
type GraphQuery struct {
	Pattern    string
	Variables  []string
	Results    []map[string]interface{}
}

// TemporalReasoning represents time-aware planning
type TemporalReasoning struct {
	Representation   string // "interval_temporal_logic"
	Horizons         map[string]*PlanningHorizon
	TemporalPatterns []TemporalPattern
}

// PlanningHorizon represents a time-based planning scope
type PlanningHorizon struct {
	Name            string
	Duration        string
	ReplanFrequency string
	Tasks           []string
}

// TemporalPattern represents detected time-based patterns
type TemporalPattern struct {
	Type       string // "periodicity", "trend", "anomaly"
	Confidence float64
	Data       []float64
}

// AgentNegotiation represents multi-agent consensus protocols
type AgentNegotiation struct {
	Style  string // "consensus", "voting", "hierarchical"
	Phases []NegotiationPhase
	Rules  []string
}

// NegotiationPhase represents a stage in negotiation
type NegotiationPhase struct {
	Name        string
	Description string
	Actions     []string
}

// EmergentGoal represents goals that emerge from agent behavior
type EmergentGoal struct {
	Name          string
	DetectedAt    string
	Evidence      []string
	Confidence    float64
	Promotion     string // "explicit_goal", "sub_agent"
	Resources     map[string]float64
}

// ResourceOrchestration represents allocation and management
type ResourceOrchestration struct {
	ComputationalBudget map[string]int
	LoadBalancing       *LoadBalancingStrategy
	ContentionResolution []string
}

// LoadBalancingStrategy represents how resources are distributed
type LoadBalancingStrategy struct {
	Strategy              string
	RedistributionTriggers []string
	Protocols            []string
}

// ValueSystem represents the agent's core values
type ValueSystem struct {
	CoreValues        []Value
	ConflictResolution string
	OverrideRules      []string
}

// Value represents a core value with weight and measurement
type Value struct {
	Name        string
	Weight      float64
	Measurement string
}

// ErrorHandling represents fault tolerance and recovery
type ErrorHandling struct {
	FaultTolerance    map[string]string // level -> strategy
	RecoveryStrategies []string
	LearningFromFailures bool
	FailureDatabase   string
}

// AgentCompiler compiles K'UHUL agent definitions
type AgentCompiler struct {
	Definitions map[string]*AgentDefinition
	KnowledgeGraph *KnowledgeGraph
	MetaCognition *MetaCognition
	TemporalReasoning *TemporalReasoning
	ErrorHandling *ErrorHandling
}

// NewAgentCompiler creates a new agent compiler
func NewAgentCompiler() *AgentCompiler {
	return &AgentCompiler{
		Definitions: make(map[string]*AgentDefinition),
		KnowledgeGraph: &KnowledgeGraph{
			Nodes: make(map[string]*KnowledgeNode),
			Edges: make(map[string]*KnowledgeEdge),
		},
		MetaCognition: &MetaCognition{
			IntrospectionLoops: make(map[string]*IntrospectionLoop),
			ConfidenceScores: make(map[string]float64),
		},
		TemporalReasoning: &TemporalReasoning{
			Horizons: make(map[string]*PlanningHorizon),
		},
		ErrorHandling: &ErrorHandling{
			FaultTolerance: make(map[string]string),
			RecoveryStrategies: []string{},
		},
	}
}

// RegisterAgent registers an agent definition
func (ac *AgentCompiler) RegisterAgent(def *AgentDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	if _, exists := ac.Definitions[def.Name]; exists {
		return fmt.Errorf("agent %q already registered", def.Name)
	}

	ac.Definitions[def.Name] = def
	return nil
}

// GetAgent retrieves an agent definition
func (ac *AgentCompiler) GetAgent(name string) (*AgentDefinition, error) {
	def, exists := ac.Definitions[name]
	if !exists {
		return nil, fmt.Errorf("agent %q not found", name)
	}
	return def, nil
}

// ListAgents returns all registered agents
func (ac *AgentCompiler) ListAgents() []*AgentDefinition {
	agents := make([]*AgentDefinition, 0)
	for _, def := range ac.Definitions {
		agents = append(agents, def)
	}
	return agents
}

// AddKnowledgeNode adds a node to the knowledge graph
func (ac *AgentCompiler) AddKnowledgeNode(node *KnowledgeNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if _, exists := ac.KnowledgeGraph.Nodes[node.ID]; exists {
		return fmt.Errorf("node %q already exists", node.ID)
	}

	ac.KnowledgeGraph.Nodes[node.ID] = node
	return nil
}

// AddKnowledgeEdge adds an edge to the knowledge graph
func (ac *AgentCompiler) AddKnowledgeEdge(edge *KnowledgeEdge) error {
	if edge.From == "" || edge.To == "" {
		return fmt.Errorf("edge endpoints cannot be empty")
	}

	key := fmt.Sprintf("%s->%s", edge.From, edge.To)

	if _, exists := ac.KnowledgeGraph.Edges[key]; exists {
		return fmt.Errorf("edge %q already exists", key)
	}

	ac.KnowledgeGraph.Edges[key] = edge
	return nil
}

// QueryKnowledgeGraph executes a query
func (ac *AgentCompiler) QueryKnowledgeGraph(query *GraphQuery) ([]map[string]interface{}, error) {
	// Simple pattern matching - can be extended
	results := make([]map[string]interface{}, 0)

	for nodeID, node := range ac.KnowledgeGraph.Nodes {
		match := false
		for _, prop := range node.Properties {
			if prop == query.Pattern {
				match = true
				break
			}
		}

		if match {
			results = append(results, map[string]interface{}{
				"node": nodeID,
				"type": node.Type,
			})
		}
	}

	query.Results = results
	return results, nil
}

// GetMetaCognitionInsight performs introspection
func (ac *AgentCompiler) GetMetaCognitionInsight(loopName string) (map[string]interface{}, error) {
	loop, exists := ac.MetaCognition.IntrospectionLoops[loopName]
	if !exists {
		return nil, fmt.Errorf("introspection loop %q not found", loopName)
	}

	insight := map[string]interface{}{
		"loop": loop.Name,
		"monitors": loop.Monitors,
		"timestamp": "now",
		"adaptations_applied": loop.Adaptations,
	}

	return insight, nil
}

// SetTemporalHorizon sets a planning horizon
func (ac *AgentCompiler) SetTemporalHorizon(horizon *PlanningHorizon) error {
	if horizon.Name == "" {
		return fmt.Errorf("horizon name cannot be empty")
	}

	ac.TemporalReasoning.Horizons[horizon.Name] = horizon
	return nil
}

// GetTemporalHorizon retrieves a planning horizon
func (ac *AgentCompiler) GetTemporalHorizon(name string) (*PlanningHorizon, error) {
	horizon, exists := ac.TemporalReasoning.Horizons[name]
	if !exists {
		return nil, fmt.Errorf("horizon %q not found", name)
	}
	return horizon, nil
}

// SetErrorRecoveryStrategy sets a fault tolerance strategy
func (ac *AgentCompiler) SetErrorRecoveryStrategy(level, strategy string) {
	ac.ErrorHandling.FaultTolerance[level] = strategy
}

// GetErrorRecoveryStrategy retrieves a fault tolerance strategy
func (ac *AgentCompiler) GetErrorRecoveryStrategy(level string) (string, error) {
	strategy, exists := ac.ErrorHandling.FaultTolerance[level]
	if !exists {
		return "", fmt.Errorf("strategy for level %q not found", level)
	}
	return strategy, nil
}
