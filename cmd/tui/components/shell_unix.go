// +build !windows

package components

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"sync"
)

// UnixShellEngine manages shell process execution (Unix/Mac)
type UnixShellEngine struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Reader
	stderr     *bufio.Reader
	workingDir string
	running    bool
	mutex      sync.Mutex
	outputChan chan string
	errorChan  chan error
	shell      string
}

// NewUnixShellEngine creates a new shell engine
func NewUnixShellEngine(workingDir string) (*UnixShellEngine, error) {
	// Determine which shell to use
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	us := &UnixShellEngine{
		workingDir: workingDir,
		outputChan: make(chan string, 100),
		errorChan:  make(chan error, 10),
		shell:      shell,
	}

	// Create shell process
	us.cmd = exec.Command(shell, "-i", "-s")
	us.cmd.Dir = workingDir

	// Set up environment
	us.cmd.Env = os.Environ()

	// Set up pipes
	stdin, err := us.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	us.stdin = stdin

	stdout, err := us.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	us.stdout = bufio.NewReader(stdout)

	stderr, err := us.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	us.stderr = bufio.NewReader(stderr)

	// Start process
	if err := us.cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	us.running = true

	// Start output readers
	go us.readOutput()
	go us.readErrors()

	return us, nil
}

// Execute runs a shell command
func (us *UnixShellEngine) Execute(script string) error {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	if !us.running {
		return fmt.Errorf("shell engine not running")
	}

	// Send command with trailing newline
	_, err := io.WriteString(us.stdin, script+"\n")
	return err
}

// ReadLine reads one line of output (non-blocking)
func (us *UnixShellEngine) ReadLine() (string, error) {
	select {
	case line := <-us.outputChan:
		return line, nil
	case err := <-us.errorChan:
		return "", err
	default:
		return "", nil
	}
}

// Close stops the shell engine
func (us *UnixShellEngine) Close() error {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	if !us.running {
		return nil
	}

	us.running = false

	// Close stdin to signal end
	if us.stdin != nil {
		us.stdin.Close()
	}

	// Wait for process
	if us.cmd.Process != nil {
		us.cmd.Wait()
	}

	close(us.outputChan)
	close(us.errorChan)

	return nil
}

// readOutput continuously reads stdout
func (us *UnixShellEngine) readOutput() {
	for us.running {
		line, err := us.stdout.ReadString('\n')
		if err != nil && err != io.EOF {
			us.errorChan <- err
			break
		}
		if line != "" {
			us.outputChan <- strings.TrimSuffix(line, "\n")
		}
		if err == io.EOF {
			break
		}
	}
}

// readErrors continuously reads stderr
func (us *UnixShellEngine) readErrors() {
	for us.running {
		line, err := us.stderr.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		if line != "" {
			us.outputChan <- "ERROR: " + strings.TrimSuffix(line, "\n")
		}
		if err == io.EOF {
			break
		}
	}
}

// GetCurrentUser gets the current user for shell initialization
func GetCurrentUser() (*user.User, error) {
	return user.Current()
}

// ShellIsAvailable checks if shell is available
func ShellIsAvailable() bool {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	_, err := os.Stat(shell)
	return err == nil
}
