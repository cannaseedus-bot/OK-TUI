package components

import (
	"fmt"
	"time"

	"github.com/gdanh/tcell/v2"
)

// GenerationStep represents a step in code generation
type GenerationStep struct {
	ID          string
	Number      int
	Title       string
	Description string
	Agent       string            // Which agent is executing
	Tool        string            // Which tool is being used
	Status      string            // "pending", "running", "completed", "failed"
	Duration    time.Duration
	StartTime   time.Time
	EndTime     time.Time
	Output      string
	Error       string
}

// TrackerPane displays generation progress and steps
type TrackerPane struct {
	BasePane
	steps        []GenerationStep
	currentStep  int
	selectedStep int
	scrollOffset int
	startTime    time.Time
	showDetails  bool
	detailIndex  int
}

// NewTrackerPane creates a new tracker pane
func NewTrackerPane() *TrackerPane {
	return &TrackerPane{
		BasePane:     *NewBasePane("tracker"),
		steps:        make([]GenerationStep, 0),
		currentStep:  -1,
		selectedStep: -1,
		scrollOffset: 0,
		startTime:    time.Now(),
		showDetails:  false,
	}
}

// AddStep adds a generation step
func (t *TrackerPane) AddStep(step GenerationStep) {
	step.Number = len(t.steps) + 1
	t.steps = append(t.steps, step)
}

// UpdateStep updates a step's status
func (t *TrackerPane) UpdateStep(index int, status string, duration time.Duration) {
	if index >= 0 && index < len(t.steps) {
		t.steps[index].Status = status
		t.steps[index].Duration = duration
		if status == "running" {
			t.currentStep = index
			t.steps[index].StartTime = time.Now()
		}
		if status == "completed" || status == "failed" {
			t.steps[index].EndTime = time.Now()
		}
	}
}

// SetStepOutput sets output for a step
func (t *TrackerPane) SetStepOutput(index int, output string) {
	if index >= 0 && index < len(t.steps) {
		t.steps[index].Output = output
	}
}

// SetStepError sets error for a step
func (t *TrackerPane) SetStepError(index int, errMsg string) {
	if index >= 0 && index < len(t.steps) {
		t.steps[index].Error = errMsg
	}
}

// Clear clears all steps
func (t *TrackerPane) Clear() {
	t.steps = make([]GenerationStep, 0)
	t.currentStep = -1
	t.selectedStep = -1
	t.scrollOffset = 0
	t.startTime = time.Now()
	t.showDetails = false
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

	headerStyle := style.Foreground(tcell.ColorYellow)

	// Draw border
	t.DrawBorder(screen, style)

	// Draw header
	headerText := " Generation Tracker "
	for x := 1; x < t.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(t.X+x, t.Y+1, rune(headerText[x-1]), nil, headerStyle)
		} else {
			screen.SetContent(t.X+x, t.Y+1, ' ', nil, headerStyle)
		}
	}

	contentHeight := t.Height - 4
	y := t.Y + 2

	// Calculate and display overall progress
	if len(t.steps) > 0 {
		completed := 0
		for _, step := range t.steps {
			if step.Status == "completed" {
				completed++
			}
		}

		pct := (completed * 100) / len(t.steps)
		barWidth := t.Width - 4
		filledWidth := (barWidth * pct) / 100

		progressText := fmt.Sprintf(" %d/%d steps completed (%d%%) ", completed, len(t.steps), pct)
		if len(progressText) > t.Width-4 {
			progressText = fmt.Sprintf(" %d/%d (%d%%) ", completed, len(t.steps), pct)
		}

		for x, r := range progressText {
			if t.X+2+x < t.X+t.Width-2 {
				screen.SetContent(t.X+2+x, y, r, nil, headerStyle)
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
		y++
	}

	// Draw steps
	displayedLines := 0
	for displayLine := 0; displayLine < contentHeight-2 && displayedLines < contentHeight-2; displayLine++ {
		actualLine := t.scrollOffset + displayLine
		if actualLine >= len(t.steps) || y >= t.Y+t.Height-1 {
			break
		}

		step := t.steps[actualLine]

		// Status icon and color
		var icon rune
		var stepStyle tcell.Style
		switch step.Status {
		case "completed":
			icon = '✓'
			stepStyle = style.Foreground(tcell.ColorGreen)
		case "running":
			icon = '►'
			stepStyle = style.Foreground(tcell.ColorCyan)
		case "failed":
			icon = '✗'
			stepStyle = style.Foreground(tcell.ColorRed)
		default: // pending
			icon = '○'
			stepStyle = style.Foreground(tcell.ColorWhite)
		}

		if actualLine == t.selectedStep {
			stepStyle = stepStyle.Reverse(true)
		}

		screen.SetContent(t.X+2, y, icon, nil, stepStyle)

		// Step number and title
		stepNum := fmt.Sprintf("%d", step.Number)
		stepDisplay := fmt.Sprintf(" %s. %s", stepNum, step.Title)

		if len(stepDisplay) > t.Width-8 {
			stepDisplay = stepDisplay[:t.Width-8] + "…"
		}

		for x, r := range stepDisplay {
			if t.X+3+x < t.X+t.Width-2 {
				screen.SetContent(t.X+3+x, y, r, nil, stepStyle)
			}
		}

		// Agent/tool info on same line if space allows
		if step.Agent != "" && t.Width > 40 {
			agentInfo := fmt.Sprintf(" [%s]", step.Agent)
			if step.Tool != "" {
				agentInfo = fmt.Sprintf(" [%s:%s]", step.Agent, step.Tool)
			}
			startX := t.X + t.Width - len(agentInfo) - 2
			if startX > t.X+3+len(stepDisplay) {
				for x, r := range agentInfo {
					if t.X+t.Width-len(agentInfo)-2+x < t.X+t.Width-2 {
						screen.SetContent(t.X+t.Width-len(agentInfo)-2+x, y, r, nil, stepStyle.Foreground(tcell.ColorGray))
					}
				}
			}
		}

		y++
		displayedLines++

		// Show duration/timing on next line for running/completed steps
		if step.Status == "running" || step.Status == "completed" {
			if displayedLines < contentHeight-2 {
				elapsed := step.Duration
				if step.Status == "running" && !step.StartTime.IsZero() {
					elapsed = time.Since(step.StartTime)
				}
				timeStr := fmt.Sprintf("    ⏱ %s", formatDuration(elapsed))
				if len(timeStr) > t.Width-4 {
					timeStr = fmt.Sprintf("    %s", formatDuration(elapsed))
				}

				for x, r := range timeStr {
					if t.X+2+x < t.X+t.Width-2 {
						screen.SetContent(t.X+2+x, y, r, nil, style.Foreground(tcell.ColorGray))
					}
				}
				y++
				displayedLines++
			}
		}
	}

	// Draw scrollbar if needed
	if len(t.steps) > contentHeight {
		indicator := fmt.Sprintf("[%d/%d]", t.scrollOffset+1, len(t.steps))
		for x, r := range indicator {
			if t.X+t.Width-2-len(indicator)+x >= t.X+2 {
				screen.SetContent(t.X+t.Width-2-len(indicator)+x, t.Y+t.Height-2, r, nil, style)
			}
		}
	}

	return nil
}

// formatDuration formats a duration nicely
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// HandleInput processes keyboard input
func (t *TrackerPane) HandleInput(ev tcell.EventKey) bool {
	contentHeight := t.Height - 4
	maxScroll := len(t.steps) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch ev.Key() {
	case tcell.KeyUp:
		if t.selectedStep > 0 {
			t.selectedStep--
			if t.selectedStep < t.scrollOffset {
				t.scrollOffset = t.selectedStep
			}
			return true
		}
	case tcell.KeyDown:
		if t.selectedStep < len(t.steps)-1 {
			t.selectedStep++
			if t.selectedStep >= t.scrollOffset+contentHeight {
				t.scrollOffset = t.selectedStep - contentHeight + 1
			}
			return true
		}
	case tcell.KeyHome:
		t.scrollOffset = 0
		t.selectedStep = 0
		return true
	case tcell.KeyEnd:
		t.selectedStep = len(t.steps) - 1
		if maxScroll > 0 && t.selectedStep > t.scrollOffset+contentHeight-1 {
			t.scrollOffset = maxScroll
		}
		return true
	case tcell.KeyPgUp:
		if t.scrollOffset > 0 {
			t.scrollOffset -= contentHeight / 2
			if t.scrollOffset < 0 {
				t.scrollOffset = 0
			}
			return true
		}
	case tcell.KeyPgDn:
		if t.scrollOffset < maxScroll {
			t.scrollOffset += contentHeight / 2
			if t.scrollOffset > maxScroll {
				t.scrollOffset = maxScroll
			}
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
		"selectedStep": t.selectedStep,
		"scrollOffset": t.scrollOffset,
	}
}
