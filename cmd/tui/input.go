package tui

import (
	"github.com/gdanh/tcell/v2"
)

// InputHandler manages keyboard and mouse input
type InputHandler struct {
	manager *TUIManager
}

// NewInputHandler creates a new input handler
func NewInputHandler(manager *TUIManager) *InputHandler {
	return &InputHandler{
		manager: manager,
	}
}

// HandleEvent processes terminal events
func (ih *InputHandler) HandleEvent(ev tcell.Event) bool {
	switch event := ev.(type) {
	case *tcell.EventKey:
		return ih.handleKey(event)
	case *tcell.EventResize:
		// TUIManager handles resize
		return true
	}
	return false
}

// handleKey processes keyboard input
func (ih *InputHandler) handleKey(ev *tcell.EventKey) bool {
	// Global shortcuts
	switch ev.Key() {
	case tcell.KeyEscape:
		return false // Signal to exit
	case tcell.KeyCtrlC:
		return false // Signal to exit
	}

	switch ev.Rune() {
	case 'q':
		if ev.Modifiers()&tcell.ModCtrl == 0 {
			return false // q to quit
		}
	case 'h':
		if ev.Modifiers()&tcell.ModCtrl == 0 {
			ih.manager.ShowHelp()
			return true
		}
	case '?':
		ih.manager.ShowHelp()
		return true
	}

	// Pane switching with Tab
	if ev.Key() == tcell.KeyTab {
		ih.manager.NextPane()
		return true
	}

	// Shift+Tab for previous pane
	if ev.Key() == tcell.KeyBacktab {
		ih.manager.PreviousPane()
		return true
	}

	// Arrow keys to switch panes
	if ev.Modifiers()&tcell.ModAlt != 0 {
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyLeft:
			ih.manager.PreviousPane()
			return true
		case tcell.KeyDown, tcell.KeyRight:
			ih.manager.NextPane()
			return true
		}
	}

	// Number keys to switch to specific pane
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Rune() {
		case '1':
			ih.manager.SetActivePane("editor")
			return true
		case '2':
			ih.manager.SetActivePane("console")
			return true
		case '3':
			ih.manager.SetActivePane("project")
			return true
		case '4':
			ih.manager.SetActivePane("tracker")
			return true
		}
	}

	// Delegate to active pane
	if ih.manager.activePane != nil {
		return ih.manager.activePane.HandleInput(ev)
	}

	return false
}
