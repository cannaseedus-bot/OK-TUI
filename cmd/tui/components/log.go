package components

import (
	"fmt"
	"time"

	"github.com/gdanh/tcell/v2"
)

// LogEntry represents an agent decision log entry
type LogEntry struct {
	Timestamp   time.Time
	AgentID     string
	Action      string // "created_tool", "created_subagent", "modified_file"
	Target      string
	Reasoning   string
	Confidence  float64
}

// LogPane displays agent decisions and autonomy tracking
type LogPane struct {
	BasePane
	entries      []LogEntry
	scrollOffset int
	selectedIdx  int
	expandedIdx  int // Which entry is expanded
}

// NewLogPane creates a new log pane
func NewLogPane() *LogPane {
	return &LogPane{
		BasePane:    *NewBasePane("log"),
		entries:     make([]LogEntry, 0),
		scrollOffset: 0,
		selectedIdx: -1,
		expandedIdx: -1,
	}
}

// AddEntry adds a log entry
func (l *LogPane) AddEntry(entry LogEntry) {
	l.entries = append(l.entries, entry)
	// Auto-scroll to show latest
	if len(l.entries) > (l.Height - 4) {
		l.scrollOffset = len(l.entries) - (l.Height - 4)
	}
}

// Clear clears all entries
func (l *LogPane) Clear() {
	l.entries = make([]LogEntry, 0)
	l.scrollOffset = 0
	l.selectedIdx = -1
	l.expandedIdx = -1
}

// Draw renders the log pane
func (l *LogPane) Draw(screen tcell.Screen) error {
	if l.Width < 10 || l.Height < 3 {
		return nil
	}

	style := tcell.StyleDefault
	if l.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}

	headerStyle := style.Foreground(tcell.ColorYellow)
	actionStyle := style.Foreground(tcell.ColorCyan)

	// Draw border
	l.DrawBorder(screen, style)

	// Draw header
	headerText := " Agent Decisions "
	for x := 1; x < l.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(l.X+x, l.Y+1, rune(headerText[x-1]), nil, headerStyle)
		} else {
			screen.SetContent(l.X+x, l.Y+1, ' ', nil, headerStyle)
		}
	}

	contentHeight := l.Height - 4
	y := l.Y + 2

	if len(l.entries) == 0 {
		msg := "No decisions yet"
		for x, r := range msg {
			if l.X+2+x < l.X+l.Width-2 {
				screen.SetContent(l.X+2+x, y, r, nil, style)
			}
		}
		return nil
	}

	// Draw entries
	displayLine := 0
	for idx := l.scrollOffset; idx < len(l.entries) && displayLine < contentHeight; idx++ {
		if y >= l.Y+l.Height-1 {
			break
		}

		entry := l.entries[idx]
		isSelected := (idx == l.selectedIdx)
		isExpanded := (idx == l.expandedIdx)

		entryStyle := actionStyle
		if isSelected {
			entryStyle = entryStyle.Reverse(true)
		}

		// Main entry line
		timeStr := entry.Timestamp.Format("15:04:05")
		actionIcon := "→"
		switch entry.Action {
		case "created_tool":
			actionIcon = "🔧"
		case "created_subagent":
			actionIcon = "👤"
		case "modified_file":
			actionIcon = "📝"
		}

		display := fmt.Sprintf("%s %s [%s] %s: %s", timeStr, actionIcon, entry.AgentID, entry.Action, entry.Target)
		if len(display) > l.Width-4 {
			display = display[:l.Width-7] + "…"
		}

		for x, r := range display {
			if l.X+2+x < l.X+l.Width-2 {
				screen.SetContent(l.X+2+x, y, r, nil, entryStyle)
			}
		}

		y++
		displayLine++

		// Show reasoning if expanded and space available
		if isExpanded && len(entry.Reasoning) > 0 && displayLine < contentHeight {
			reasoningText := fmt.Sprintf("  → %s", entry.Reasoning)
			if len(reasoningText) > l.Width-4 {
				reasoningText = reasoningText[:l.Width-7] + "…"
			}

			for x, r := range reasoningText {
				if l.X+2+x < l.X+l.Width-2 {
					screen.SetContent(l.X+2+x, y, r, nil, style.Foreground(tcell.ColorGray))
				}
			}
			y++
			displayLine++

			// Show confidence if available
			if entry.Confidence > 0 && displayLine < contentHeight {
				confStr := fmt.Sprintf("  Confidence: %.0f%%", entry.Confidence*100)
				for x, r := range confStr {
					if l.X+2+x < l.X+l.Width-2 {
						screen.SetContent(l.X+2+x, y, r, nil, style.Foreground(tcell.ColorGray))
					}
				}
				y++
				displayLine++
			}
		}
	}

	// Draw scrollbar indicator
	if len(l.entries) > contentHeight {
		indicator := fmt.Sprintf("[%d/%d]", l.scrollOffset+1, len(l.entries))
		for x, r := range indicator {
			if l.X+l.Width-2-len(indicator)+x >= l.X+2 {
				screen.SetContent(l.X+l.Width-2-len(indicator)+x, l.Y+l.Height-2, r, nil, style)
			}
		}
	}

	return nil
}

// HandleInput processes keyboard input
func (l *LogPane) HandleInput(ev tcell.EventKey) bool {
	contentHeight := l.Height - 4
	maxScroll := len(l.entries) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch ev.Key() {
	case tcell.KeyUp:
		if l.selectedIdx > 0 {
			l.selectedIdx--
			if l.selectedIdx < l.scrollOffset {
				l.scrollOffset = l.selectedIdx
			}
			return true
		}
	case tcell.KeyDown:
		if l.selectedIdx < len(l.entries)-1 {
			l.selectedIdx++
			if l.selectedIdx >= l.scrollOffset+contentHeight {
				l.scrollOffset = l.selectedIdx - contentHeight + 1
			}
			return true
		}
	case tcell.KeyHome:
		l.scrollOffset = 0
		l.selectedIdx = 0
		return true
	case tcell.KeyEnd:
		l.selectedIdx = len(l.entries) - 1
		if maxScroll > 0 && l.selectedIdx > l.scrollOffset+contentHeight-1 {
			l.scrollOffset = maxScroll
		}
		return true
	case tcell.KeyEnter:
		if l.selectedIdx >= 0 {
			if l.expandedIdx == l.selectedIdx {
				l.expandedIdx = -1
			} else {
				l.expandedIdx = l.selectedIdx
			}
			return true
		}
	case tcell.KeyPgUp:
		if l.scrollOffset > 0 {
			l.scrollOffset -= contentHeight / 2
			if l.scrollOffset < 0 {
				l.scrollOffset = 0
			}
			return true
		}
	case tcell.KeyPgDn:
		if l.scrollOffset < maxScroll {
			l.scrollOffset += contentHeight / 2
			if l.scrollOffset > maxScroll {
				l.scrollOffset = maxScroll
			}
			return true
		}
	}
	return false
}

// GetState returns pane state for checkpointing
func (l *LogPane) GetState() interface{} {
	return map[string]interface{}{
		"entryCount":   len(l.entries),
		"scrollOffset": l.scrollOffset,
		"selectedIdx":  l.selectedIdx,
		"expandedIdx":  l.expandedIdx,
	}
}
