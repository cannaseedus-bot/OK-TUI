package tui

import (
	"time"
)

// TUIState represents the complete state of the TUI
type TUIState struct {
	Layout          string                 // "4-panel", "5-panel"
	ActivePane      string                 // ID of currently active pane
	ScrollPositions map[string]int         // Scroll position per pane
	SelectionState  map[string]interface{} // Selection state per pane
	Cursor          CursorState            // Current cursor position

	// Generation context
	CurrentFile     string        // File currently being edited
	ProjectRoot     string        // Root directory of project
	SessionID       string        // Current session ID
	GenerationStep  int           // Current step in generation
	FileChanges     []FileChange  // Recent file changes
	AgentDecisions  []AgentAction // Recent agent decisions
	Messages        []ChatMessage // Chat history

	// Timestamp for checkpoint
	LastUpdate time.Time
}

// CursorState tracks cursor position
type CursorState struct {
	X int
	Y int
}

// FileChange represents a file modification event
type FileChange struct {
	Path      string    `json:"path"`
	Type      string    `json:"type"`      // "created", "modified", "deleted"
	Timestamp time.Time `json:"timestamp"`
	AgentID   string    `json:"agent_id"`
	Tool      string    `json:"tool"`
	OldSize   int       `json:"old_size"`
	NewSize   int       `json:"new_size"`
	Lines     int       `json:"lines"` // Number of lines changed
}

// AgentAction represents an agent decision or tool creation
type AgentAction struct {
	Timestamp time.Time              `json:"timestamp"`
	AgentID   string                 `json:"agent_id"`
	Action    string                 `json:"action"`      // "created_tool", "created_agent", etc.
	Target    string                 `json:"target"`      // What was created
	Reasoning string                 `json:"reasoning"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Role      string    `json:"role"`      // "user", "assistant", "agent"
	Content   string    `json:"content"`
	AgentID   string    `json:"agent_id,omitempty"` // If from agent
}

// Pane is the interface all UI panes must implement
type Pane interface {
	// ID returns unique identifier for pane
	ID() string

	// Bounds returns x, y, width, height
	Bounds() (x, y, width, height int)

	// SetBounds updates the pane dimensions
	SetBounds(x, y, width, height int)

	// Draw renders the pane to the screen
	Draw(buffer *ScreenBuffer) error

	// HandleInput processes keyboard/mouse input, returns true if handled
	HandleInput(ev interface{}) bool

	// SetActive updates active state (for styling)
	SetActive(active bool)

	// Clear resets pane content
	Clear()

	// GetState returns pane-specific state for checkpointing
	GetState() interface{}
}

// PaneConfig holds layout information for a pane
type PaneConfig struct {
	ID       string
	Position string // "top-left", "top-right", "bottom-left", "bottom-right", "bottom-full"
	MinWidth int
	MinHeight int
	Flex     bool // True if pane should expand to fill available space
}

// ScreenBuffer represents terminal screen for rendering
type ScreenBuffer struct {
	Width  int
	Height int
	// Could be tcell.Screen or similar
	screen interface{} // Store actual screen implementation
}

// GenerationSession tracks ongoing code generation
type GenerationSession struct {
	ID           string
	StartTime    time.Time
	CurrentStep  int
	TotalSteps   int
	Status       string // "running", "paused", "completed", "failed"
	FileChanges  []FileChange
	LastActivity time.Time
}

// NewTUIState creates a new TUI state with defaults
func NewTUIState(projectRoot string) *TUIState {
	return &TUIState{
		Layout:          "4-panel",
		ActivePane:      "editor",
		ScrollPositions: make(map[string]int),
		SelectionState:  make(map[string]interface{}),
		ProjectRoot:     projectRoot,
		FileChanges:     make([]FileChange, 0),
		AgentDecisions:  make([]AgentAction, 0),
		Messages:        make([]ChatMessage, 0),
		LastUpdate:      time.Now(),
	}
}

// RecordFileChange adds a file change to the state
func (s *TUIState) RecordFileChange(change FileChange) {
	s.FileChanges = append(s.FileChanges, change)
	// Keep only last 100 changes
	if len(s.FileChanges) > 100 {
		s.FileChanges = s.FileChanges[len(s.FileChanges)-100:]
	}
	s.LastUpdate = time.Now()
}

// RecordAgentAction adds an agent action to the state
func (s *TUIState) RecordAgentAction(action AgentAction) {
	s.AgentDecisions = append(s.AgentDecisions, action)
	// Keep only last 100 decisions
	if len(s.AgentDecisions) > 100 {
		s.AgentDecisions = s.AgentDecisions[len(s.AgentDecisions)-100:]
	}
	s.LastUpdate = time.Now()
}

// AddMessage adds a chat message to history
func (s *TUIState) AddMessage(msg ChatMessage) {
	s.Messages = append(s.Messages, msg)
	s.LastUpdate = time.Now()
}
