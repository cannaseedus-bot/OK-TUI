package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdanh/tcell/v2"
)

// FileNode represents a file in the project tree
type FileNode struct {
	Path         string
	Name         string
	IsDir        bool
	Modified     bool      // If file has been modified
	Dependencies []string  // Files this file depends on
	Size         int64
	Children     []*FileNode
	Expanded     bool
}

// ProjectPane displays the project structure and file tree
type ProjectPane struct {
	BasePane
	root          *FileNode
	displayLines  []*FileNode // Flattened view for rendering
	scrollOffset  int
	selectedIndex int
	modifiedFiles map[string]bool
}

// NewProjectPane creates a new project pane
func NewProjectPane() *ProjectPane {
	return &ProjectPane{
		BasePane:      *NewBasePane("project"),
		displayLines:  make([]*FileNode, 0),
		scrollOffset:  0,
		modifiedFiles: make(map[string]bool),
	}
}

// SetProjectRoot scans and builds the project tree
func (p *ProjectPane) SetProjectRoot(rootPath string) error {
	if info, err := os.Stat(rootPath); err != nil || !info.IsDir() {
		return fmt.Errorf("invalid project root: %s", rootPath)
	}

	p.root = &FileNode{
		Path:        rootPath,
		Name:        filepath.Base(rootPath),
		IsDir:       true,
		Children:    make([]*FileNode, 0),
		Expanded:    true,
	}

	// Scan directory recursively
	if err := p.scanDirectory(p.root, rootPath); err != nil {
		return err
	}

	p.rebuildDisplayLines()
	return nil
}

// scanDirectory recursively scans a directory
func (p *ProjectPane) scanDirectory(parent *FileNode, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil // Silently skip unreadable directories
	}

	for _, entry := range entries {
		// Skip hidden files and common directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.Name() == "node_modules" || entry.Name() == "vendor" || entry.Name() == "target" {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		info, _ := entry.Info()

		node := &FileNode{
			Path:     fullPath,
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Children: make([]*FileNode, 0),
			Size:     info.Size(),
		}

		parent.Children = append(parent.Children, node)

		// Recursively scan subdirectories
		if entry.IsDir() {
			p.scanDirectory(node, fullPath)
		}
	}

	return nil
}

// MarkFileModified marks a file as modified
func (p *ProjectPane) MarkFileModified(filePath string) {
	p.modifiedFiles[filePath] = true

	// Mark all parent directories
	dir := filepath.Dir(filePath)
	for dir != p.root.Path && dir != "/" {
		p.modifiedFiles[dir] = true
		dir = filepath.Dir(dir)
	}

	p.rebuildDisplayLines()
}

// SetFiles updates with a list of modified files
func (p *ProjectPane) SetFiles(files []string) {
	p.modifiedFiles = make(map[string]bool)
	for _, file := range files {
		p.MarkFileModified(file)
	}
}

// rebuildDisplayLines flattens the tree for rendering
func (p *ProjectPane) rebuildDisplayLines() {
	p.displayLines = make([]*FileNode, 0)
	if p.root != nil {
		p.flattenTree(p.root, 0)
	}
}

// flattenTree recursively flattens the tree
func (p *ProjectPane) flattenTree(node *FileNode, depth int) {
	// Skip root node in display
	if depth > 0 {
		p.displayLines = append(p.displayLines, node)
	}

	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			p.flattenTree(child, depth+1)
		}
	}
}

// Clear clears the file list
func (p *ProjectPane) Clear() {
	p.root = nil
	p.displayLines = make([]*FileNode, 0)
	p.modifiedFiles = make(map[string]bool)
	p.scrollOffset = 0
	p.selectedIndex = 0
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

	headerStyle := style.Foreground(tcell.ColorYellow)
	modifiedStyle := style.Foreground(tcell.ColorGreen)

	// Draw border
	p.DrawBorder(screen, style)

	// Draw header
	headerText := " Project Tree "
	for x := 1; x < p.Width-1; x++ {
		if x < len(headerText)+1 {
			screen.SetContent(p.X+x, p.Y+1, rune(headerText[x-1]), nil, headerStyle)
		} else {
			screen.SetContent(p.X+x, p.Y+1, ' ', nil, headerStyle)
		}
	}

	// Draw content
	contentHeight := p.Height - 4
	if len(p.displayLines) == 0 {
		// Placeholder text
		msg := "Scan project to build tree"
		for x, r := range msg {
			if p.X+2+x < p.X+p.Width-2 {
				screen.SetContent(p.X+2+x, p.Y+3, r, nil, style)
			}
		}
	} else {
		// Show file tree
		for displayLine := 0; displayLine < contentHeight; displayLine++ {
			actualLine := p.scrollOffset + displayLine
			y := p.Y + 2 + displayLine

			if actualLine >= len(p.displayLines) {
				break
			}

			node := p.displayLines[actualLine]

			// Calculate depth and indentation
			depth := p.calculateDepth(node)
			indent := strings.Repeat("  ", depth)

			// Format node display
			prefix := "  "
			if node.IsDir {
				if node.Expanded {
					prefix = "▼ "
				} else {
					prefix = "▶ "
				}
			} else {
				prefix = "  "
			}

			// Check if modified
			nodeStyle := style
			if p.modifiedFiles[node.Path] {
				nodeStyle = modifiedStyle
			}
			if actualLine == p.selectedIndex {
				nodeStyle = nodeStyle.Reverse(true)
			}

			display := indent + prefix + node.Name
			if len(display) > p.Width-3 {
				display = "..." + display[len(display)-(p.Width-6):]
			}

			for x, r := range display {
				if p.X+2+x < p.X+p.Width-2 {
					screen.SetContent(p.X+2+x, y, r, nil, nodeStyle)
				}
			}
		}
	}

	// Draw scrollbar indicator
	if len(p.displayLines) > contentHeight {
		indicator := fmt.Sprintf("[%d/%d]", p.scrollOffset+1, len(p.displayLines))
		for x, r := range indicator {
			if p.X+p.Width-2-len(indicator)+x >= p.X+2 {
				screen.SetContent(p.X+p.Width-2-len(indicator)+x, p.Y+p.Height-2, r, nil, style)
			}
		}
	}

	return nil
}

// calculateDepth calculates the depth of a node in the tree
func (p *ProjectPane) calculateDepth(node *FileNode) int {
	depth := 0
	current := node.Path
	for current != p.root.Path && current != "/" {
		depth++
		current = filepath.Dir(current)
	}
	return depth
}

// HandleInput processes keyboard input
func (p *ProjectPane) HandleInput(ev tcell.EventKey) bool {
	contentHeight := p.Height - 4
	maxScroll := len(p.displayLines) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch ev.Key() {
	case tcell.KeyUp:
		if p.selectedIndex > 0 {
			p.selectedIndex--
			if p.selectedIndex < p.scrollOffset {
				p.scrollOffset = p.selectedIndex
			}
			return true
		}
	case tcell.KeyDown:
		if p.selectedIndex < len(p.displayLines)-1 {
			p.selectedIndex++
			if p.selectedIndex >= p.scrollOffset+contentHeight {
				p.scrollOffset = p.selectedIndex - contentHeight + 1
			}
			return true
		}
	case tcell.KeyLeft:
		if p.selectedIndex < len(p.displayLines) {
			node := p.displayLines[p.selectedIndex]
			if node.IsDir && node.Expanded {
				node.Expanded = false
				p.rebuildDisplayLines()
				return true
			}
		}
	case tcell.KeyRight:
		if p.selectedIndex < len(p.displayLines) {
			node := p.displayLines[p.selectedIndex]
			if node.IsDir && !node.Expanded {
				node.Expanded = true
				p.rebuildDisplayLines()
				return true
			}
		}
	case tcell.KeyHome:
		p.scrollOffset = 0
		p.selectedIndex = 0
		return true
	case tcell.KeyEnd:
		p.selectedIndex = len(p.displayLines) - 1
		if maxScroll > 0 && p.selectedIndex > p.scrollOffset+contentHeight-1 {
			p.scrollOffset = maxScroll
		}
		return true
	case tcell.KeyPgUp:
		if p.scrollOffset > 0 {
			p.scrollOffset -= contentHeight
			if p.scrollOffset < 0 {
				p.scrollOffset = 0
			}
			return true
		}
	case tcell.KeyPgDn:
		if p.scrollOffset < maxScroll {
			p.scrollOffset += contentHeight
			if p.scrollOffset > maxScroll {
				p.scrollOffset = maxScroll
			}
			return true
		}
	}
	return false
}

// GetState returns pane state for checkpointing
func (p *ProjectPane) GetState() interface{} {
	return map[string]interface{}{
		"fileCount":      len(p.displayLines),
		"scrollOffset":   p.scrollOffset,
		"selectedIndex":  p.selectedIndex,
		"modifiedCount":  len(p.modifiedFiles),
	}
}
