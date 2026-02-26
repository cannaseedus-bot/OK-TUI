package tui

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/gdanh/tcell/v2"
	"github.com/ollama/ollama/cmd/tui/components"
)

// TUIManager is the main TUI coordinator
type TUIManager struct {
	screen       tcell.Screen
	renderer     *Renderer
	inputHandler *InputHandler
	state        *TUIState

	panes      map[string]interface{} // All available panes
	activePane interface{}             // Currently active pane
	panOrder   []string               // Order of panes for Tab cycling

	projectRoot string
	sessionID   string

	// Shutdown signal
	done    chan struct{}
	mutex   sync.RWMutex
	running bool

	// Event channels
	resizeChan chan struct{}
}

// NewTUIManager creates a new TUI manager
func NewTUIManager(projectRoot string) (*TUIManager, error) {
	// Initialize tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	manager := &TUIManager{
		screen:      screen,
		renderer:    NewRenderer(screen),
		state:       NewTUIState(projectRoot),
		panes:       make(map[string]interface{}),
		panOrder:    []string{"editor", "console", "project", "tracker"},
		projectRoot: projectRoot,
		done:        make(chan struct{}),
		resizeChan:  make(chan struct{}, 1),
		running:     true,
	}

	manager.inputHandler = NewInputHandler(manager)

	// Create panes
	manager.panes["editor"] = components.NewEditorPane()
	manager.panes["console"] = components.NewConsolePane()
	manager.panes["project"] = components.NewProjectPane()
	manager.panes["tracker"] = components.NewTrackerPane()

	// Set initial active pane
	manager.SetActivePane("editor")

	// Set screen defaults
	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))

	return manager, nil
}

// SetActivePane sets the active pane by ID
func (m *TUIManager) SetActivePane(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Deactivate current pane
	if m.activePane != nil {
		if p, ok := m.activePane.(interface{ SetActive(bool) }); ok {
			p.SetActive(false)
		}
	}

	// Activate new pane
	if pane, ok := m.panes[id]; ok {
		if p, ok := pane.(interface{ SetActive(bool) }); ok {
			p.SetActive(true)
		}
		m.activePane = pane
		m.state.ActivePane = id
	}
}

// NextPane switches to the next pane in order
func (m *TUIManager) NextPane() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	currentID := m.state.ActivePane
	for i, id := range m.panOrder {
		if id == currentID {
			nextIdx := (i + 1) % len(m.panOrder)
			m.SetActivePane(m.panOrder[nextIdx])
			return
		}
	}
	// Default to first pane
	m.SetActivePane(m.panOrder[0])
}

// PreviousPane switches to the previous pane in order
func (m *TUIManager) PreviousPane() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	currentID := m.state.ActivePane
	for i, id := range m.panOrder {
		if id == currentID {
			prevIdx := (i - 1 + len(m.panOrder)) % len(m.panOrder)
			m.SetActivePane(m.panOrder[prevIdx])
			return
		}
	}
	// Default to last pane
	m.SetActivePane(m.panOrder[len(m.panOrder)-1])
}

// ShowHelp displays help information
func (m *TUIManager) ShowHelp() {
	console := m.panes["console"].(*components.ConsolePane)
	if console != nil {
		console.Clear()
		console.WriteLine("╔════════════════════════════════════════╗")
		console.WriteLine("║       KUHUL TUI - Keyboard Help        ║")
		console.WriteLine("╠════════════════════════════════════════╣")
		console.WriteLine("║ Tab / Shift+Tab    - Switch panes       ║")
		console.WriteLine("║ Ctrl+1/2/3/4      - Jump to pane      ║")
		console.WriteLine("║ Alt+Up/Down/Left/Right - Cycle panes  ║")
		console.WriteLine("║ Arrow Keys         - Navigate in pane  ║")
		console.WriteLine("║ Page Up/Down       - Scroll faster     ║")
		console.WriteLine("║ Home/End           - Jump to start/end ║")
		console.WriteLine("║ h / ?              - Show this help    ║")
		console.WriteLine("║ q                  - Quit              ║")
		console.WriteLine("╚════════════════════════════════════════╝")
	}
}

// LoadFile loads a file into the editor pane
func (m *TUIManager) LoadFile(filepath string) error {
	// TODO: Read file content
	// For now just set placeholder
	editor := m.panes["editor"].(*components.EditorPane)
	if editor != nil {
		editor.SetContent(filepath, "// File: "+filepath+"\n// Content would be loaded here")
		m.state.CurrentFile = filepath
	}
	return nil
}

// AddConsoleOutput adds a line to console output
func (m *TUIManager) AddConsoleOutput(line string) {
	console := m.panes["console"].(*components.ConsolePane)
	if console != nil {
		console.Write(line)
	}
}

// AddFileChange records a file change event
func (m *TUIManager) AddFileChange(change FileChange) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.RecordFileChange(change)

	// Update project pane
	project := m.panes["project"].(*components.ProjectPane)
	if project != nil {
		files := make([]string, 0)
		for _, fc := range m.state.FileChanges {
			files = append(files, filepath.Base(fc.Path))
		}
		project.SetFiles(files)
	}

	// Update console
	m.AddConsoleOutput(fmt.Sprintf("Modified: %s", change.Path))
}

// AddAgentDecision records an agent decision
func (m *TUIManager) AddAgentDecision(action AgentAction) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.RecordAgentAction(action)
	m.AddConsoleOutput(fmt.Sprintf("Agent %s: %s %s", action.AgentID, action.Action, action.Target))
}

// Run starts the TUI event loop
func (m *TUIManager) Run() error {
	// Initial render
	m.renderer.UpdateSize()
	m.renderer.Render(m)

	// Event loop
	for m.running {
		// Handle events with timeout
		m.screen.SetTimeout(100 * time.Millisecond)

		select {
		case <-m.done:
			return nil
		case <-m.resizeChan:
			m.renderer.UpdateSize()
			m.renderer.Render(m)
		default:
			// Process one event
			ev := m.screen.PollEvent()
			if ev == nil {
				// Timeout - just refresh
				m.renderer.Render(m)
				continue
			}

			// Handle event
			shouldExit := false
			if keyEvent, ok := ev.(*tcell.EventKey); ok {
				if keyEvent.Rune() == 'q' && keyEvent.Modifiers() == 0 {
					shouldExit = true
				}
				if keyEvent.Key() == tcell.KeyEscape {
					shouldExit = true
				}
			}

			if shouldExit {
				m.running = false
				break
			}

			// Handle resize
			if _, ok := ev.(*tcell.EventResize); ok {
				m.renderer.UpdateSize()
			}

			// Input handler
			m.inputHandler.HandleEvent(ev)

			// Render
			m.renderer.Render(m)
		}
	}

	return nil
}

// Close gracefully shuts down the TUI
func (m *TUIManager) Close() error {
	m.running = false
	close(m.done)

	if m.screen != nil {
		m.screen.Fini()
	}

	return nil
}
