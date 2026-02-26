package tui

import (
	"github.com/gdanh/tcell/v2"
)

// Renderer handles all screen rendering
type Renderer struct {
	screen tcell.Screen
	width  int
	height int
}

// NewRenderer creates a new renderer
func NewRenderer(screen tcell.Screen) *Renderer {
	width, height := screen.Size()
	return &Renderer{
		screen: screen,
		width:  width,
		height: height,
	}
}

// GetSize returns current screen size
func (r *Renderer) GetSize() (width, height int) {
	return r.screen.Size()
}

// UpdateSize updates renderer size
func (r *Renderer) UpdateSize() {
	r.width, r.height = r.screen.Size()
}

// Clear clears the entire screen
func (r *Renderer) Clear() {
	r.screen.Clear()
}

// Render performs full screen render
func (r *Renderer) Render(manager *TUIManager) error {
	// Clear screen
	r.screen.Clear()

	// Update pane layouts
	r.layoutPanes(manager)

	// Draw each pane
	for _, pane := range manager.panes {
		if pane != nil {
			pane.Draw(r.screen)
		}
	}

	// Draw status bar
	r.drawStatusBar(manager)

	// Show cursor
	r.screen.ShowCursor(10, r.height-1)

	// Update display
	r.screen.Sync()

	return nil
}

// layoutPanes arranges panes based on layout mode
func (r *Renderer) layoutPanes(manager *TUIManager) {
	if manager.state.Layout == "4-panel" {
		r.layout4Panel(manager)
	} else {
		r.layout4Panel(manager) // Default to 4-panel
	}
}

// layout4Panel arranges 4 panes in 2x2 grid
func (r *Renderer) layout4Panel(manager *TUIManager) {
	if r.width < 40 || r.height < 15 {
		return // Too small to layout
	}

	midX := r.width / 2
	midY := r.height / 2

	// Editor (top-left)
	if editor := manager.panes["editor"]; editor != nil {
		editor.SetBounds(0, 0, midX, midY)
	}

	// Console (bottom-left)
	if console := manager.panes["console"]; console != nil {
		consoleHeight := r.height - midY - 1
		console.SetBounds(0, midY, midX, consoleHeight)
	}

	// Project (top-right)
	if project := manager.panes["project"]; project != nil {
		editorWidth := r.width - midX
		project.SetBounds(midX, 0, editorWidth, midY)
	}

	// Tracker (bottom-right)
	if tracker := manager.panes["tracker"]; tracker != nil {
		trackerWidth := r.width - midX
		trackerHeight := r.height - midY - 1
		tracker.SetBounds(midX, midY, trackerWidth, trackerHeight)
	}
}

// drawStatusBar draws the status/help line at bottom
func (r *Renderer) drawStatusBar(manager *TUIManager) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	statusText := "Tab: Switch | Ctrl+1-4: Pane | h: Help | q: Quit"

	y := r.height - 1
	for x, r := range statusText {
		if x < r.width {
			r.screen.SetContent(x, y, r, nil, statusStyle)
		}
	}

	// Clear rest of line
	for x := len(statusText); x < r.width; x++ {
		r.screen.SetContent(x, y, ' ', nil, statusStyle)
	}
}

// DrawText draws text at a position
func (r *Renderer) DrawText(x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		if x+i < r.width && y < r.height {
			r.screen.SetContent(x+i, y, ch, nil, style)
		}
	}
}

// FillRect fills a rectangle with a character
func (r *Renderer) FillRect(x, y, width, height int, ch rune, style tcell.Style) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if x+dx < r.width && y+dy < r.height {
				r.screen.SetContent(x+dx, y+dy, ch, nil, style)
			}
		}
	}
}

// Sync forces screen update
func (r *Renderer) Sync() {
	r.screen.Sync()
}
