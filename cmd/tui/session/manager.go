package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cannaseedus-bot/Ollama-K/cmd/tui"
)

// Session represents a TUI session with state, generation context, and chat history
type Session struct {
	ID              string                    `json:"id"`
	CreatedAt       time.Time                 `json:"created_at"`
	LastUpdated     time.Time                 `json:"last_updated"`
	Generation      *GenerationSession        `json:"generation,omitempty"`
	UIState         *tui.TUIState             `json:"ui_state,omitempty"`
	Messages        []Message                 `json:"messages,omitempty"`
	AgentDecisions  []tui.AgentDecision       `json:"agent_decisions,omitempty"`
	FileChanges     []tui.FileChange          `json:"file_changes,omitempty"`
}

// Message represents a chat message in session
type Message struct {
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"` // "user" or "agent"
	Content   string    `json:"content"`
}

// GenerationSession tracks a code generation workflow
type GenerationSession struct {
	ID           string             `json:"id"`
	StartTime    time.Time          `json:"start_time"`
	Status       string             `json:"status"` // "pending", "running", "completed", "error"
	CurrentStep  int                `json:"current_step"`
	Steps        []*GenerationStep  `json:"steps,omitempty"`
	ProjectRoot  string             `json:"project_root"`
}

// GenerationStep represents a single step in generation
type GenerationStep struct {
	ID          string        `json:"id"`
	Number      int           `json:"number"`
	Description string        `json:"description"`
	Agent       string        `json:"agent"`
	Tool        string        `json:"tool,omitempty"`
	Status      string        `json:"status"` // "pending", "running", "completed", "error"
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Output      string        `json:"output,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// SessionManager manages session lifecycle and persistence
type SessionManager struct {
	sessions      map[string]*Session
	currentID     string
	storePath     string
	mu            sync.RWMutex
	checkpointMgr *CheckpointManager
}

// NewSessionManager creates a new session manager
func NewSessionManager(storePath string) (*SessionManager, error) {
	// Ensure store directory exists
	if err := os.MkdirAll(storePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session store directory: %w", err)
	}

	sm := &SessionManager{
		sessions:  make(map[string]*Session),
		storePath: storePath,
	}

	// Initialize checkpoint manager
	checkpointMgr, err := NewCheckpointManager(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize checkpoint manager: %w", err)
	}
	sm.checkpointMgr = checkpointMgr

	// Load existing sessions
	if err := sm.loadSessions(); err != nil {
		return nil, fmt.Errorf("failed to load sessions: %w", err)
	}

	// Load current session
	sm.loadCurrentSessionID()

	return sm, nil
}

// loadSessions loads all sessions from disk
func (sm *SessionManager) loadSessions() error {
	files, err := ioutil.ReadDir(sm.storePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		sessionID := file.Name()[:len(file.Name())-5] // Remove .json
		session, err := sm.loadSession(sessionID)
		if err != nil {
			// Log error but continue loading other sessions
			fmt.Printf("Error loading session %s: %v\n", sessionID, err)
			continue
		}

		sm.sessions[sessionID] = session
	}

	return nil
}

// loadSession loads a single session from disk
func (sm *SessionManager) loadSession(sessionID string) (*Session, error) {
	sessionPath := filepath.Join(sm.storePath, sessionID+".json")
	data, err := ioutil.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// loadCurrentSessionID loads the current session ID from file
func (sm *SessionManager) loadCurrentSessionID() {
	currentPath := filepath.Join(sm.storePath, ".current")
	data, err := ioutil.ReadFile(currentPath)
	if err == nil {
		sm.currentID = string(data)
	}
}

// saveCurrentSessionID saves the current session ID to file
func (sm *SessionManager) saveCurrentSessionID() error {
	currentPath := filepath.Join(sm.storePath, ".current")
	return ioutil.WriteFile(currentPath, []byte(sm.currentID), 0644)
}

// NewSession creates a new session
func (sm *SessionManager) NewSession(sessionID string, projectRoot string) *Session {
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		CreatedAt: now,
		LastUpdated: now,
		Generation: &GenerationSession{
			ID:          sessionID,
			StartTime:   now,
			Status:      "pending",
			ProjectRoot: projectRoot,
			Steps:       make([]*GenerationStep, 0),
		},
		UIState:        &tui.TUIState{},
		Messages:       make([]Message, 0),
		AgentDecisions: make([]tui.AgentDecision, 0),
		FileChanges:    make([]tui.FileChange, 0),
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[sessionID] = session
	sm.currentID = sessionID

	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// GetCurrentSession returns the current session
func (sm *SessionManager) GetCurrentSession() (*Session, error) {
	if sm.currentID == "" {
		return nil, fmt.Errorf("no current session")
	}

	return sm.GetSession(sm.currentID)
}

// SetCurrentSession sets the current active session
func (sm *SessionManager) SetCurrentSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	sm.currentID = sessionID
	return sm.saveCurrentSessionID()
}

// SaveSession saves a session to disk
func (sm *SessionManager) SaveSession(session *Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session.LastUpdated = time.Now()
	sessionPath := filepath.Join(sm.storePath, session.ID+".json")

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(sessionPath, data, 0644); err != nil {
		return err
	}

	sm.sessions[session.ID] = session
	return nil
}

// ListSessions returns all session IDs
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		sessions = append(sessions, id)
	}

	return sessions
}

// DeleteSession deletes a session
func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionPath := filepath.Join(sm.storePath, sessionID+".json")
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(sm.sessions, sessionID)

	if sm.currentID == sessionID {
		sm.currentID = ""
		if len(sm.sessions) > 0 {
			for id := range sm.sessions {
				sm.currentID = id
				break
			}
		}
	}

	return nil
}

// GetCheckpointManager returns the checkpoint manager
func (sm *SessionManager) GetCheckpointManager() *CheckpointManager {
	return sm.checkpointMgr
}
