package components

import (
	"fmt"
	"time"

	"github.com/gdanh/tcell/v2"
)

// GenerationStep represents a step in code generation
type GenerationStep struct {
	ID    string
	Title string
	Status string // "pending", "running", "completed", "failed"
	Duration time.Duration
}

// TrackerPane displays generation progress and steps
type TrackerPane struct {
	BasePane
	steps        []GenerationStep
	currentStep  int
	scrollOffset int
	startTime    time.Time
}

// NewTrackerPane creates a new tracker pane
func NewTrackerPane() *TrackerPane {
	return &TrackerPane{
		BasePane:     *NewBasePane("tracker"),
		steps:        make([]GenerationStep, 0),
		currentStep:  -1,
		scrollOffset: 0,
		startTime:    time.Now(),
	}
}

// AddStep adds a generation step
func (t *TrackerPane) AddStep(step GenerationStep) {
	t.steps = append(t.steps, step)
}

// UpdateStep updates a step's status
func (t *TrackerPane) UpdateStep(index int, status string, duration time.Duration) {
	if index >= 0 && index < len(t.steps) {
		t.steps[index].Status = status
		t.steps[index].Duration = duration
		if status == "running" {
			t.currentStep = index
		}
	}
}

// Clear clears all steps
func (t *TrackerPane) Clear() {
	t.steps = make([]GenerationStep, 0)
	t.currentStep = -1
	t.startTime = time.Now()
}

// Draw renders the tracker pane
func (t *TrackerPane) Draw(screen tcell.Screen) error {
	if t.Width < 10 || t.Height < 3 {
		return nil
	}

	style := tcell.StyleDefault
	if t.Active {
		style = style.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	}

	// Draw border
	t.DrawBorder(screen, style)

	// Draw header
	headerText := " Generation Progress "
	for x := 1; x < t.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(t.X+x, t.Y+1, rune(headerText[x-1]), nil, style)
		} else {
			screen.SetContent(t.X+x, t.Y+1, ' ', nil, style)
		}
	}

	contentHeight := t.Height - 4

	// Calculate and display progress
	completed := 0
	for _, step := range t.steps {
		if step.Status == "completed" {
			completed++
		}
	}

	// Progress bar
	y := t.Y + 2
	if len(t.steps) > 0 {
		pct := (completed * 100) / len(t.steps)
		barWidth := t.Width - 4
		filledWidth := (barWidth * pct) / 100

		progressText := fmt.Sprintf(" %d/%d ", completed, len(t.steps))
		for x, r := range progressText {
			if t.X+2+x < t.X+t.Width-2 {
				screen.SetContent(t.X+2+x, y, r, nil, style)
			}
		}

		// Draw progress bar
		y++
		for x := 0; x < barWidth; x++ {
			if x < filledWidth {
				screen.SetContent(t.X+2+x, y, '█', nil, style.Foreground(tcell.ColorGreen))
			} else {
				screen.SetContent(t.X+2+x, y, '░', nil, style)
			}
		}
	}

	// Draw steps
	y += 2
	for displayLine := 0; displayLine < contentHeight-3; displayLine++ {
		actualLine := t.scrollOffset + displayLine
		if actualLine >= len(t.steps) || y >= t.Y+t.Height-1 {
			break
		}

		step := t.steps[actualLine]

		// Status icon
		var icon rune
		var stepStyle tcell.Style
		switch step.Status {
		case "completed":
			icon = '✓'
			stepStyle = style.Foreground(tcell.ColorGreen)
		case "running":
			icon = '►'
			stepStyle = style.Foreground(tcell.ColorYellow)
		case "failed":
			icon = '✗'
			stepStyle = style.Foreground(tcell.ColorRed)
		default: // pending
			icon = '○'
			stepStyle = style
		}

		screen.SetContent(t.X+2, y, icon, nil, stepStyle)

		// Step title
		title := " " + step.Title
		if len(title) > t.Width-8 {
			title = title[:t.Width-8] + "..."
		}
		for x, r := range title {
			if t.X+3+x < t.X+t.Width-2 {
				screen.SetContent(t.X+3+x, y, r, nil, stepStyle)
			}
		}

		y++
	}

	return nil
}

// HandleInput processes keyboard input
func (t *TrackerPane) HandleInput(ev tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		if t.scrollOffset > 0 {
			t.scrollOffset--
			return true
		}
	case tcell.KeyDown:
		if t.scrollOffset < len(t.steps)-1 {
			t.scrollOffset++
			return true
		}
	}
	return false
}

// GetState returns pane state for checkpointing
func (t *TrackerPane) GetState() interface{} {
	return map[string]interface{}{
		"stepCount":    len(t.steps),
		"currentStep":  t.currentStep,
		"scrollOffset": t.scrollOffset,
	}
}
