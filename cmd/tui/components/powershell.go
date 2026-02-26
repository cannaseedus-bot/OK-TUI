// +build windows

package components

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// PowerShellEngine manages PowerShell process execution (Windows only)
type PowerShellEngine struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      *bufio.Reader
	stderr      *bufio.Reader
	workingDir  string
	running     bool
	mutex       sync.Mutex
	outputChan  chan string
	errorChan   chan error
}

// NewPowerShellEngine creates a new PowerShell engine
func NewPowerShellEngine(workingDir string) (*PowerShellEngine, error) {
	ps := &PowerShellEngine{
		workingDir: workingDir,
		outputChan: make(chan string, 100),
		errorChan:  make(chan error, 10),
	}

	// Create PowerShell process
	ps.cmd = exec.Command("powershell", "-NoProfile", "-NoExit", "-Command", "-")
	ps.cmd.Dir = workingDir

	// Set up pipes
	stdin, err := ps.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	ps.stdin = stdin

	stdout, err := ps.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	ps.stdout = bufio.NewReader(stdout)

	stderr, err := ps.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	ps.stderr = bufio.NewReader(stderr)

	// Start process
	if err := ps.cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start PowerShell: %w", err)
	}

	ps.running = true

	// Start output readers
	go ps.readOutput()
	go ps.readErrors()

	return ps, nil
}

// Execute runs a PowerShell script/command
func (ps *PowerShellEngine) Execute(script string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if !ps.running {
		return fmt.Errorf("PowerShell engine not running")
	}

	// Send command with trailing newline
	_, err := io.WriteString(ps.stdin, script+"\n")
	return err
}

// ReadLine reads one line of output (non-blocking)
func (ps *PowerShellEngine) ReadLine() (string, error) {
	select {
	case line := <-ps.outputChan:
		return line, nil
	case err := <-ps.errorChan:
		return "", err
	default:
		return "", nil
	}
}

// Close stops the PowerShell engine
func (ps *PowerShellEngine) Close() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if !ps.running {
		return nil
	}

	ps.running = false

	// Close stdin to signal end
	if ps.stdin != nil {
		ps.stdin.Close()
	}

	// Wait for process
	if ps.cmd.Process != nil {
		ps.cmd.Wait()
	}

	close(ps.outputChan)
	close(ps.errorChan)

	return nil
}

// readOutput continuously reads stdout
func (ps *PowerShellEngine) readOutput() {
	for ps.running {
		line, err := ps.stdout.ReadString('\n')
		if err != nil && err != io.EOF {
			ps.errorChan <- err
			break
		}
		if line != "" {
			ps.outputChan <- strings.TrimSuffix(line, "\n")
		}
		if err == io.EOF {
			break
		}
	}
}

// readErrors continuously reads stderr
func (ps *PowerShellEngine) readErrors() {
	for ps.running {
		line, err := ps.stderr.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		if line != "" {
			ps.outputChan <- "ERROR: " + strings.TrimSuffix(line, "\n")
		}
		if err == io.EOF {
			break
		}
	}
}

// IsAvailable checks if PowerShell is available
func PowerShellIsAvailable() bool {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Write-Host 'ok'")
	err := cmd.Run()
	return err == nil
}
