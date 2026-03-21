package components

import (
	"strings"

	"github.com/gdanh/tcell/v2"
)

// EditorPane displays a file with line numbers and syntax highlighting
type EditorPane struct {
	BasePane
	content      []string // Lines of file content
	filename     string
	scrollOffset int // How many lines scrolled down
	cursorX      int
	cursorY      int
}

// NewEditorPane creates a new editor pane
func NewEditorPane() *EditorPane {
	return &EditorPane{
		BasePane:     *NewBasePane("editor"),
		content:      make([]string, 0),
		scrollOffset: 0,
		cursorX:      0,
		cursorY:      0,
	}
}

// SetContent sets the file content to display
func (e *EditorPane) SetContent(filename string, content string) {
	e.filename = filename
	e.content = strings.Split(content, "\n")
	e.scrollOffset = 0
	e.cursorX = 0
	e.cursorY = 0
}

// GetFilename returns the current filename
func (e *EditorPane) GetFilename() string {
	return e.filename
}

// Clear clears the content
func (e *EditorPane) Clear() {
	e.content = make([]string, 0)
	e.filename = ""
	e.scrollOffset = 0
	e.cursorX = 0
	e.cursorY = 0
}

// Draw renders the editor pane
func (e *EditorPane) Draw(screen tcell.Screen) error {
	if e.Width < 10 || e.Height < 3 {
		return nil
	}

	// Draw border
	style := tcell.StyleDefault
	if e.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}
	e.DrawBorder(screen, style)

	// Draw filename header
	headerText := " " + e.filename + " "
	if len(headerText) > e.Width-2 {
		headerText = " " + e.filename[len(e.filename)-e.Width+5:] + " "
	}
	for x := 1; x < e.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(e.X+x, e.Y+1, rune(headerText[x-1]), nil, style)
		} else {
			screen.SetContent(e.X+x, e.Y+1, ' ', nil, style)
		}
	}

	// Calculate available space
	contentHeight := e.Height - 4 // Border, header, status line
	lineNumWidth := calculateLineNumberWidth(len(e.content))
	contentWidth := e.Width - lineNumWidth - 3

	// Draw content with line numbers
	for displayLine := 0; displayLine < contentHeight; displayLine++ {
		actualLine := e.scrollOffset + displayLine
		y := e.Y + 2 + displayLine

		if actualLine >= len(e.content) {
			// Draw empty line indicator
			screen.SetContent(e.X+1, y, '~', nil, style)
			continue
		}

		// Draw line number
		lineNumStr := formatLineNumber(actualLine+1, lineNumWidth)
		for i, r := range lineNumStr {
			screen.SetContent(e.X+1+i, y, r, nil, style)
		}

		// Draw line content
		line := e.content[actualLine]
		for x, r := range line {
			screenX := e.X + lineNumWidth + 2 + x
			if screenX >= e.X+e.Width-1 {
				break
			}
			// Simple syntax highlighting - highlight certain keywords
			textStyle := style
			if isKeyword(string(r)) {
				textStyle = style.Foreground(tcell.ColorYellow)
			}
			screen.SetContent(screenX, y, r, nil, textStyle)
		}
	}

	// Draw status line
	statusLine := e.Y + e.Height - 1
	statusText := " Ln " + formatNumber(e.cursorY+e.scrollOffset+1, 4) +
		" Col " + formatNumber(e.cursorX+1, 4) +
		" | Lines: " + formatNumber(len(e.content), 5)

	for x, r := range statusText {
		if e.X+1+x < e.X+e.Width-1 {
			screen.SetContent(e.X+1+x, statusLine, r, nil, style)
		}
	}

	// Clear rest of status line
	for x := e.X + 1 + len(statusText); x < e.X+e.Width-1; x++ {
		screen.SetContent(x, statusLine, ' ', nil, style)
	}

	return nil
}

// HandleInput processes keyboard input
func (e *EditorPane) HandleInput(ev tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		if e.cursorY > 0 {
			e.cursorY--
		} else if e.scrollOffset > 0 {
			e.scrollOffset--
		}
		return true
	case tcell.KeyDown:
		if e.cursorY+e.scrollOffset < len(e.content)-1 {
			if e.cursorY < e.Height-5 {
				e.cursorY++
			} else if e.scrollOffset+e.Height-4 < len(e.content) {
				e.scrollOffset++
			}
		}
		return true
	case tcell.KeyLeft:
		if e.cursorX > 0 {
			e.cursorX--
		}
		return true
	case tcell.KeyRight:
		line := e.content[e.cursorY+e.scrollOffset]
		if e.cursorX < len(line) {
			e.cursorX++
		}
		return true
	case tcell.KeyPgUp:
		e.scrollOffset -= e.Height - 5
		if e.scrollOffset < 0 {
			e.scrollOffset = 0
		}
		return true
	case tcell.KeyPgDn:
		e.scrollOffset += e.Height - 5
		if e.scrollOffset >= len(e.content) {
			e.scrollOffset = len(e.content) - 1
		}
		return true
	case tcell.KeyHome:
		e.cursorX = 0
		return true
	case tcell.KeyEnd:
		if e.cursorY+e.scrollOffset < len(e.content) {
			e.cursorX = len(e.content[e.cursorY+e.scrollOffset])
		}
		return true
	}
	return false
}

// GetState returns pane state for checkpointing
func (e *EditorPane) GetState() interface{} {
	return map[string]interface{}{
		"filename":      e.filename,
		"scrollOffset":  e.scrollOffset,
		"cursorX":       e.cursorX,
		"cursorY":       e.cursorY,
		"contentLines":  len(e.content),
	}
}

// Helper functions

func calculateLineNumberWidth(lineCount int) int {
	if lineCount == 0 {
		return 1
	}
	width := 1
	count := lineCount
	for count >= 10 {
		width++
		count /= 10
	}
	return width
}

func formatLineNumber(num, width int) string {
	numStr := formatNumber(num, width)
	return numStr + " │ "
}

func formatNumber(num, width int) string {
	str := strings.Repeat(" ", width) + string(rune('0'+rune(num%10)))
	num /= 10
	for i := len(str) - 2; num > 0 && i >= 0; i-- {
		str = str[:i] + string(rune('0'+rune(num%10))) + str[i+1:]
		num /= 10
	}
	if len(str) > width {
		str = str[len(str)-width:]
	}
	return str
}

func isKeyword(s string) bool {
	keywords := map[string]bool{
		"f": true, "t": true, "i": true, "e": true, "d": true, "l": true, "s": true,
	}
	return keywords[s]
}
