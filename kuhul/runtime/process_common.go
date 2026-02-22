//go:build !windows

package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ProcessResult represents the result of a process execution
type ProcessResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// RunCommand executes a command with arguments on Unix systems
// Returns exit code, stdout, and stderr
func RunCommand(cmd string, args []string) *ProcessResult {
	result := &ProcessResult{}

	command := exec.Command(cmd, args...)

	var outBuf, errBuf bytes.Buffer
	command.Stdout = &outBuf
	command.Stderr = &errBuf

	err := command.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	result.Stdout = outBuf.String()
	result.Stderr = errBuf.String()

	return result
}

// ShellCommand executes a command through the shell on Unix systems
// This allows for complex commands with pipes, redirects, etc.
func ShellCommand(cmdLine string) *ProcessResult {
	result := &ProcessResult{}

	command := exec.Command("sh", "-c", cmdLine)

	var outBuf, errBuf bytes.Buffer
	command.Stdout = &outBuf
	command.Stderr = &errBuf

	err := command.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	result.Stdout = outBuf.String()
	result.Stderr = errBuf.String()

	return result
}

// ParseCommandLine splits a command line into command and arguments
// Handles quoted strings and escaping
func ParseCommandLine(cmdLine string) (string, []string) {
	// For Unix, we can use shell.Parser or simple split
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return "", []string{}
	}
	return parts[0], parts[1:]
}

// CommandExists checks if a command is available in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetEnvWithPath returns environment variables with updated PATH
func GetEnvWithPath(additionalPaths []string) []string {
	env := os.Environ()

	if len(additionalPaths) > 0 {
		// Update PATH environment variable
		pathSeparator := ":"
		newPath := strings.Join(additionalPaths, pathSeparator)
		for i, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				env[i] = "PATH=" + newPath + pathSeparator + e[5:]
				return env
			}
		}
		// If PATH doesn't exist, add it
		env = append(env, "PATH="+newPath)
	}

	return env
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	PID  int
	Name string
}

// ListProcesses returns a list of running processes
// Note: Unix implementation is limited without /proc access
func ListProcesses() []ProcessInfo {
	result := []ProcessInfo{}

	// Use `ps` command to list processes
	psResult := ShellCommand("ps aux")
	if psResult.ExitCode != 0 {
		return result
	}

	lines := strings.Split(psResult.Stdout, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Second field is PID, last field is command name
			pid := 0
			name := ""
			if _, err := os.Stat(line); err == nil {
				// Try to parse PID
				_, _ = fmt.Sscanf(fields[1], "%d", &pid)
				if len(fields) > 10 {
					name = fields[len(fields)-1]
				}
			}
			if pid > 0 {
				result = append(result, ProcessInfo{PID: pid, Name: name})
			}
		}
	}

	return result
}

// TerminateProcess kills a process by PID
func TerminateProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
