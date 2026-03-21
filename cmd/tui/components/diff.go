package components

import (
	"fmt"
	"strings"

	"github.com/gdanh/tcell/v2"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type      string // "added", "removed", "context", "header"
	OldLineNum int
	NewLineNum int
	Content    string
	AgentID    string // Which agent made this change
}

// DiffPane displays real-time code changes in split-pane format
type DiffPane struct {
	BasePane
	oldContent   []string
	newContent   []string
	diffLines    []DiffLine
	scrollOffset int
	horizontalScroll int
}

// NewDiffPane creates a new diff pane
func NewDiffPane() *DiffPane {
	return &DiffPane{
		BasePane:         *NewBasePane("diff"),
		oldContent:       make([]string, 0),
		newContent:       make([]string, 0),
		diffLines:        make([]DiffLine, 0),
		scrollOffset:     0,
		horizontalScroll: 0,
	}
}

// SetContent sets the old and new content for diffing
func (d *DiffPane) SetContent(oldContent, newContent string, agentID string) {
	d.oldContent = strings.Split(oldContent, "\n")
	d.newContent = strings.Split(newContent, "\n")
	d.diffLines = make([]DiffLine, 0)

	// Simple line-by-line diff (can be enhanced with Myers algorithm)
	d.generateDiff(agentID)
	d.scrollOffset = 0
	d.horizontalScroll = 0
}

// generateDiff generates diff lines using simple line-by-line comparison
func (d *DiffPane) generateDiff(agentID string) {
	maxLen := len(d.oldContent)
	if len(d.newContent) > maxLen {
		maxLen = len(d.newContent)
	}

	for i := 0; i < maxLen; i++ {
		oldLine := ""
		newLine := ""
		oldNum := 0
		newNum := 0

		if i < len(d.oldContent) {
			oldLine = d.oldContent[i]
			oldNum = i + 1
		}
		if i < len(d.newContent) {
			newLine = d.newContent[i]
			newNum = i + 1
		}

		if oldLine == newLine {
			// Context line
			if oldLine != "" || newLine != "" {
				d.diffLines = append(d.diffLines, DiffLine{
					Type:       "context",
					OldLineNum: oldNum,
					NewLineNum: newNum,
					Content:    oldLine,
					AgentID:    "",
				})
			}
		} else if oldLine != "" && newLine == "" {
			// Removed line
			d.diffLines = append(d.diffLines, DiffLine{
				Type:       "removed",
				OldLineNum: oldNum,
				NewLineNum: 0,
				Content:    oldLine,
				AgentID:    agentID,
			})
		} else if oldLine == "" && newLine != "" {
			// Added line
			d.diffLines = append(d.diffLines, DiffLine{
				Type:       "added",
				OldLineNum: 0,
				NewLineNum: newNum,
				Content:    newLine,
				AgentID:    agentID,
			})
		} else {
			// Changed line (both exist but different)
			d.diffLines = append(d.diffLines, DiffLine{
				Type:       "removed",
				OldLineNum: oldNum,
				NewLineNum: 0,
				Content:    oldLine,
				AgentID:    agentID,
			})
			d.diffLines = append(d.diffLines, DiffLine{
				Type:       "added",
				OldLineNum: 0,
				NewLineNum: newNum,
				Content:    newLine,
				AgentID:    agentID,
			})
		}
	}
}

// Clear clears the diff
func (d *DiffPane) Clear() {
	d.oldContent = make([]string, 0)
	d.newContent = make([]string, 0)
	d.diffLines = make([]DiffLine, 0)
	d.scrollOffset = 0
	d.horizontalScroll = 0
}

// Draw renders the diff pane
func (d *DiffPane) Draw(screen tcell.Screen) error {
	if d.Width < 20 || d.Height < 3 {
		return nil
	}

	style := tcell.StyleDefault
	if d.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}

	headerStyle := style.Foreground(tcell.ColorYellow)
	addedStyle := style.Foreground(tcell.ColorGreen)
	removedStyle := style.Foreground(tcell.ColorRed)
	contextStyle := style.Foreground(tcell.ColorGray)

	// Draw border
	d.DrawBorder(screen, style)

	// Draw header with file info
	headerText := " Changes "
	for x := 1; x < d.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(d.X+x, d.Y+1, rune(headerText[x-1]), nil, headerStyle)
		} else {
			screen.SetContent(d.X+x, d.Y+1, ' ', nil, headerStyle)
		}
	}

	// Draw column headers for split-pane
	y := d.Y + 2
	colWidth := (d.Width - 4) / 2

	// Old content header
	oldHeader := " Original "
	for x, r := range oldHeader {
		if x < colWidth-1 {
			screen.SetContent(d.X+2+x, y, r, nil, headerStyle)
		}
	}

	// Separator
	screen.SetContent(d.X+2+colWidth, y, '│', nil, headerStyle)

	// New content header
	newHeader := " Modified "
	for x, r := range newHeader {
		if x < colWidth-1 {
			screen.SetContent(d.X+3+colWidth+x, y, r, nil, headerStyle)
		}
	}

	contentHeight := d.Height - 4
	y = d.Y + 3

	if len(d.diffLines) == 0 {
		msg := "No changes"
		for x, r := range msg {
			if d.X+2+x < d.X+d.Width-2 {
				screen.SetContent(d.X+2+x, y, r, nil, style)
			}
		}
		return nil
	}

	// Draw diff lines in split pane format
	for displayLine := 0; displayLine < contentHeight && d.scrollOffset+displayLine < len(d.diffLines); displayLine++ {
		if y >= d.Y+d.Height-1 {
			break
		}

		diffLine := d.diffLines[d.scrollOffset+displayLine]

		// Determine style based on change type
		var lineStyle tcell.Style
		lineChar := " "
		switch diffLine.Type {
		case "added":
			lineStyle = addedStyle
			lineChar = "+"
		case "removed":
			lineStyle = removedStyle
			lineChar = "-"
		default: // context
			lineStyle = contextStyle
			lineChar = " "
		}

		// Draw left side (original)
		content := diffLine.Content
		if len(content) > colWidth-5 {
			content = content[:colWidth-8] + "…"
		}

		// Line number + content for old side
		if diffLine.OldLineNum > 0 {
			lineNumStr := fmt.Sprintf("%3d ", diffLine.OldLineNum)
			for x, r := range lineNumStr {
				if d.X+2+x < d.X+2+colWidth {
					screen.SetContent(d.X+2+x, y, r, nil, lineStyle)
				}
			}
			for x, r := range content {
				if d.X+6+x < d.X+2+colWidth {
					screen.SetContent(d.X+6+x, y, r, nil, lineStyle)
				}
			}
		}

		// Separator
		screen.SetContent(d.X+2+colWidth, y, '│', nil, headerStyle)

		// Draw right side (modified)
		if diffLine.NewLineNum > 0 {
			content := diffLine.Content
			if len(content) > colWidth-5 {
				content = content[:colWidth-8] + "…"
			}

			lineNumStr := fmt.Sprintf("%3d ", diffLine.NewLineNum)
			for x, r := range lineNumStr {
				if d.X+3+colWidth+x < d.X+d.Width-2 {
					screen.SetContent(d.X+3+colWidth+x, y, r, nil, lineStyle)
				}
			}
			for x, r := range content {
				if d.X+7+colWidth+x < d.X+d.Width-2 {
					screen.SetContent(d.X+7+colWidth+x, y, r, nil, lineStyle)
				}
			}
		}

		// Draw change indicator
		screen.SetContent(d.X+2, y, rune(lineChar[0]), nil, lineStyle)

		y++
	}

	// Draw scrollbar
	if len(d.diffLines) > contentHeight {
		indicator := fmt.Sprintf("[%d/%d]", d.scrollOffset+1, len(d.diffLines))
		for x, r := range indicator {
			if d.X+d.Width-2-len(indicator)+x >= d.X+2 {
				screen.SetContent(d.X+d.Width-2-len(indicator)+x, d.Y+d.Height-2, r, nil, style)
			}
		}
	}

	return nil
}

// HandleInput processes keyboard input
func (d *DiffPane) HandleInput(ev tcell.EventKey) bool {
	contentHeight := d.Height - 4
	maxScroll := len(d.diffLines) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch ev.Key() {
	case tcell.KeyUp:
		if d.scrollOffset > 0 {
			d.scrollOffset--
			return true
		}
	case tcell.KeyDown:
		if d.scrollOffset < maxScroll {
			d.scrollOffset++
			return true
		}
	case tcell.KeyHome:
		d.scrollOffset = 0
		return true
	case tcell.KeyEnd:
		d.scrollOffset = maxScroll
		return true
	case tcell.KeyPgUp:
		if d.scrollOffset > 0 {
			d.scrollOffset -= contentHeight
			if d.scrollOffset < 0 {
				d.scrollOffset = 0
			}
			return true
		}
	case tcell.KeyPgDn:
		if d.scrollOffset < maxScroll {
			d.scrollOffset += contentHeight
			if d.scrollOffset > maxScroll {
				d.scrollOffset = maxScroll
			}
			return true
		}
	case tcell.KeyLeft:
		if d.horizontalScroll > 0 {
			d.horizontalScroll--
			return true
		}
	case tcell.KeyRight:
		d.horizontalScroll++
		return true
	}
	return false
}

// GetState returns pane state for checkpointing
func (d *DiffPane) GetState() interface{} {
	return map[string]interface{}{
		"diffLineCount":    len(d.diffLines),
		"scrollOffset":     d.scrollOffset,
		"horizontalScroll": d.horizontalScroll,
	}
}
