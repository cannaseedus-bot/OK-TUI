package components

import (
	"github.com/gdanh/tcell/v2"
)

// BasePaneData holds common pane data
type BasePaneData struct {
	ID     string
	X      int
	Y      int
	Width  int
	Height int
	Active bool
}

// BasePane provides default implementations for Pane interface
// Embed this in concrete pane types to get base functionality
type BasePane struct {
	BasePaneData
}

// NewBasePane creates a new base pane
func NewBasePane(id string) *BasePane {
	return &BasePane{
		BasePaneData: BasePaneData{
			ID: id,
		},
	}
}

// ID returns the pane identifier
func (p *BasePane) ID() string {
	return p.ID
}

// Bounds returns x, y, width, height
func (p *BasePane) Bounds() (x, y, width, height int) {
	return p.X, p.Y, p.Width, p.Height
}

// SetBounds updates dimensions
func (p *BasePane) SetBounds(x, y, width, height int) {
	p.X = x
	p.Y = y
	p.Width = width
	p.Height = height
}

// SetActive updates active state
func (p *BasePane) SetActive(active bool) {
	p.Active = active
}

// IsActive returns whether pane is active
func (p *BasePane) IsActive() bool {
	return p.Active
}

// DrawBorder draws a box around the pane
func (p *BasePane) DrawBorder(screen tcell.Screen, style tcell.Style) {
	if p.Width < 2 || p.Height < 2 {
		return
	}

	// Top border
	for x := p.X; x < p.X+p.Width; x++ {
		screen.SetContent(x, p.Y, '─', nil, style)
	}

	// Bottom border
	for x := p.X; x < p.X+p.Width; x++ {
		screen.SetContent(x, p.Y+p.Height-1, '─', nil, style)
	}

	// Left border
	for y := p.Y; y < p.Y+p.Height; y++ {
		screen.SetContent(p.X, y, '│', nil, style)
	}

	// Right border
	for y := p.Y; y < p.Y+p.Height; y++ {
		screen.SetContent(p.X+p.Width-1, y, '│', nil, style)
	}

	// Corners
	screen.SetContent(p.X, p.Y, '┌', nil, style)
	screen.SetContent(p.X+p.Width-1, p.Y, '┐', nil, style)
	screen.SetContent(p.X, p.Y+p.Height-1, '└', nil, style)
	screen.SetContent(p.X+p.Width-1, p.Y+p.Height-1, '┘', nil, style)
}

// DrawText draws text in pane with wrapping
func (p *BasePane) DrawText(screen tcell.Screen, x, y int, text string, style tcell.Style) int {
	maxX := p.X + p.Width - 2
	curX := p.X + x
	curY := p.Y + y

	for _, r := range text {
		if r == '\n' {
			curX = p.X + x
			curY++
			if curY >= p.Y+p.Height-1 {
				break
			}
			continue
		}

		if curX >= maxX {
			curX = p.X + x
			curY++
			if curY >= p.Y+p.Height-1 {
				break
			}
		}

		if curY < p.Y+p.Height-1 {
			screen.SetContent(curX, curY, r, nil, style)
			curX++
		}
	}

	return curY - p.Y
}

// GetStyle returns appropriate style based on pane state
func (p *BasePane) GetStyle(base tcell.Style) tcell.Style {
	if p.Active {
		return base.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}
	return base
}

// IsInBounds checks if a coordinate is within the pane
func (p *BasePane) IsInBounds(x, y int) bool {
	return x >= p.X && x < p.X+p.Width && y >= p.Y && y < p.Y+p.Height
}
