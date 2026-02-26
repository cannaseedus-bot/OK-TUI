package components

import (
	"github.com/gdanh/tcell/v2"
)

// ProjectPane displays the project structure and file tree
type ProjectPane struct {
	BasePane
	files        []string // File list
	scrollOffset int
}

// NewProjectPane creates a new project pane
func NewProjectPane() *ProjectPane {
	return &ProjectPane{
		BasePane:     *NewBasePane("project"),
		files:        make([]string, 0),
		scrollOffset: 0,
	}
}

// SetFiles updates the file list
func (p *ProjectPane) SetFiles(files []string) {
	p.files = files
	p.scrollOffset = 0
}

// Clear clears the file list
func (p *ProjectPane) Clear() {
	p.files = make([]string, 0)
	p.scrollOffset = 0
}

// Draw renders the project pane
func (p *ProjectPane) Draw(screen tcell.Screen) error {
	if p.Width < 10 || p.Height < 3 {
		return nil
	}

	style := tcell.StyleDefault
	if p.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}

	// Draw border
	p.DrawBorder(screen, style)

	// Draw header
	headerText := " Project "
	for x := 1; x < p.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(p.X+x, p.Y+1, rune(headerText[x-1]), nil, style)
		} else {
			screen.SetContent(p.X+x, p.Y+1, ' ', nil, style)
		}
	}

	// Draw content (placeholder - showing "No files")
	contentHeight := p.Height - 4
	if len(p.files) == 0 {
		// Placeholder text
		msg := "No files yet"
		for x, r := range msg {
			if p.X+2+x < p.X+p.Width-2 {
				screen.SetContent(p.X+2+x, p.Y+3, r, nil, style)
			}
		}
	} else {
		// Show files
		for displayLine := 0; displayLine < contentHeight; displayLine++ {
			actualLine := p.scrollOffset + displayLine
			y := p.Y + 2 + displayLine

			if actualLine >= len(p.files) {
				break
			}

			file := p.files[actualLine]
			if len(file) > p.Width-3 {
				file = "..." + file[len(file)-(p.Width-6):]
			}

			for x, r := range file {
				if p.X+2+x < p.X+p.Width-2 {
					screen.SetContent(p.X+2+x, y, r, nil, style)
				}
			}
		}
	}

	return nil
}

// HandleInput processes keyboard input
func (p *ProjectPane) HandleInput(ev tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		if p.scrollOffset > 0 {
			p.scrollOffset--
			return true
		}
	case tcell.KeyDown:
		if p.scrollOffset < len(p.files)-1 {
			p.scrollOffset++
			return true
		}
	}
	return false
}

// GetState returns pane state for checkpointing
func (p *ProjectPane) GetState() interface{} {
	return map[string]interface{}{
		"fileCount":    len(p.files),
		"scrollOffset": p.scrollOffset,
	}
}
