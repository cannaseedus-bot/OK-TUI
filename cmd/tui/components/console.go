package components

import (
	"github.com/gdanh/tcell/v2"
)

// ConsolePane displays command output and shell interaction
type ConsolePane struct {
	BasePane
	buffer       []string // Output lines
	scrollOffset int
	maxLines     int // Keep only last N lines
}

// NewConsolePane creates a new console pane
func NewConsolePane() *ConsolePane {
	return &ConsolePane{
		BasePane:     *NewBasePane("console"),
		buffer:       make([]string, 0),
		scrollOffset: 0,
		maxLines:     1000, // Keep last 1000 lines
	}
}

// Write adds a line to the console buffer
func (c *ConsolePane) Write(line string) {
	c.buffer = append(c.buffer, line)
	// Maintain max buffer size
	if len(c.buffer) > c.maxLines {
		c.buffer = c.buffer[len(c.buffer)-c.maxLines:]
	}
	// Auto-scroll to bottom on new output
	if len(c.buffer) > c.Height-3 {
		c.scrollOffset = len(c.buffer) - (c.Height - 3)
	}
}

// WriteLine adds a line with newline
func (c *ConsolePane) WriteLine(line string) {
	c.Write(line + "\n")
}

// Clear clears the console
func (c *ConsolePane) Clear() {
	c.buffer = make([]string, 0)
	c.scrollOffset = 0
}

// GetContent returns current buffer content
func (c *ConsolePane) GetContent() string {
	if len(c.buffer) == 0 {
		return ""
	}
	// Join buffer lines
	result := ""
	for _, line := range c.buffer {
		result += line
	}
	return result
}

// Draw renders the console pane
func (c *ConsolePane) Draw(screen tcell.Screen) error {
	if c.Width < 10 || c.Height < 3 {
		return nil
	}

	style := tcell.StyleDefault
	if c.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}

	// Draw border
	c.DrawBorder(screen, style)

	// Draw header
	headerText := " Console "
	for x := 1; x < c.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(c.X+x, c.Y+1, rune(headerText[x-1]), nil, style)
		} else {
			screen.SetContent(c.X+x, c.Y+1, ' ', nil, style)
		}
	}

	// Draw content
	contentHeight := c.Height - 4
	startLine := c.scrollOffset
	if startLine < 0 {
		startLine = 0
	}

	for displayLine := 0; displayLine < contentHeight; displayLine++ {
		actualLine := startLine + displayLine
		y := c.Y + 2 + displayLine

		if actualLine >= len(c.buffer) {
			continue
		}

		line := c.buffer[actualLine]
		// Split line if it contains newline
		lines := []rune(line)
		for x := 0; x < len(lines) && x < c.Width-3; x++ {
			r := lines[x]
			if r != '\n' {
				screen.SetContent(c.X+2+x, y, r, nil, style)
			}
		}
	}

	// Draw status line with scroll info
	statusLine := c.Y + c.Height - 1
	scrollInfo := ""
	if len(c.buffer) > contentHeight {
		pct := (c.scrollOffset * 100) / (len(c.buffer) - contentHeight)
		scrollInfo = " " + formatNumber(pct, 3) + "% "
	} else {
		scrollInfo = " All "
	}

	for x, r := range scrollInfo {
		if c.X+1+x < c.X+c.Width-1 {
			screen.SetContent(c.X+1+x, statusLine, r, nil, style)
		}
	}

	return nil
}

// HandleInput processes keyboard input
func (c *ConsolePane) HandleInput(ev tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		if c.scrollOffset > 0 {
			c.scrollOffset--
			return true
		}
	case tcell.KeyDown:
		maxScroll := len(c.buffer) - (c.Height - 3)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if c.scrollOffset < maxScroll {
			c.scrollOffset++
			return true
		}
	case tcell.KeyPgUp:
		c.scrollOffset -= c.Height - 5
		if c.scrollOffset < 0 {
			c.scrollOffset = 0
		}
		return true
	case tcell.KeyPgDn:
		c.scrollOffset += c.Height - 5
		maxScroll := len(c.buffer) - (c.Height - 3)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if c.scrollOffset > maxScroll {
			c.scrollOffset = maxScroll
		}
		return true
	case tcell.KeyHome:
		c.scrollOffset = 0
		return true
	case tcell.KeyEnd:
		c.scrollOffset = len(c.buffer) - (c.Height - 3)
		if c.scrollOffset < 0 {
			c.scrollOffset = 0
		}
		return true
	}
	return false
}

// GetState returns pane state for checkpointing
func (c *ConsolePane) GetState() interface{} {
	return map[string]interface{}{
		"bufferLines": len(c.buffer),
		"scrollOffset": c.scrollOffset,
	}
}
