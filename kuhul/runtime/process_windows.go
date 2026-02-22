//go:build windows

package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessResult represents the result of a process execution
type ProcessResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// RunCommand executes a command with arguments on Windows
// Handles .exe extensions and batch files
func RunCommand(cmd string, args []string) *ProcessResult {
	result := &ProcessResult{}

	// Normalize command: try with .exe if no extension
	cmdPath := cmd
	if !strings.HasSuffix(strings.ToLower(cmd), ".exe") {
		// First try as-is (might be in PATH)
		if _, err := exec.LookPath(cmd); err != nil {
			// Try with .exe
			exeCmd := cmd + ".exe"
			if _, err := exec.LookPath(exeCmd); err == nil {
				cmdPath = exeCmd
			}
		}
	}

	command := exec.Command(cmdPath, args...)

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

// ShellCommand executes a command through cmd.exe on Windows
// This allows for complex commands with pipes, redirects, etc.
func ShellCommand(cmdLine string) *ProcessResult {
	result := &ProcessResult{}

	// Use cmd.exe /c to execute the command line
	command := exec.Command("cmd.exe", "/c", cmdLine)

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

// ParseCommandLine splits a command line into command and arguments on Windows
// Handles quoted strings with spaces
func ParseCommandLine(cmdLine string) (string, []string) {
	// Simple Windows command line parsing
	// Does not handle all edge cases but covers common scenarios
	parts := []string{}
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(cmdLine); i++ {
		ch := cmdLine[i]

		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return "", []string{}
	}

	return parts[0], parts[1:]
}

// CommandExists checks if a command is available in PATH on Windows
// Also checks with .exe, .bat, .cmd extensions
func CommandExists(cmd string) bool {
	// First try as-is
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}

	// Try with common Windows extensions
	for _, ext := range []string{".exe", ".bat", ".cmd", ".com"} {
		if _, err := exec.LookPath(cmd + ext); err == nil {
			return true
		}
	}

	return false
}

// GetEnvWithPath returns environment variables with updated PATH on Windows
func GetEnvWithPath(additionalPaths []string) []string {
	env := os.Environ()

	if len(additionalPaths) > 0 {
		// Update PATH environment variable (Windows uses ; as separator)
		pathSeparator := ";"
		newPath := strings.Join(additionalPaths, pathSeparator)

		for i, e := range env {
			if strings.EqualFold(e[:5], "PATH=") {
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

// ListProcesses returns a list of running processes on Windows
// Uses Get-Process PowerShell command
func ListProcesses() []ProcessInfo {
	result := []ProcessInfo{}

	// Use PowerShell Get-Process command
	psResult := ShellCommand("powershell -Command \"Get-Process | Select-Object Id,ProcessName\"")
	if psResult.ExitCode != 0 {
		return result
	}

	lines := strings.Split(psResult.Stdout, "\n")
	for i, line := range lines {
		if i < 2 {
			continue // Skip header lines
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid := 0
			name := fields[len(fields)-1]

			if _, err := fmt.Sscanf(fields[0], "%d", &pid); err == nil && pid > 0 {
				result = append(result, ProcessInfo{PID: pid, Name: name})
			}
		}
	}

	return result
}

// TerminateProcess kills a process by PID on Windows
// Uses taskkill command
func TerminateProcess(pid int) error {
	result := ShellCommand(fmt.Sprintf("taskkill /pid %d /f", pid))
	if result.ExitCode != 0 {
		return fmt.Errorf("taskkill failed: %s", result.Stderr)
	}
	return nil
}

// CheckBatchFileType determines if a file is a batch file
func CheckBatchFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".bat" || ext == ".cmd" || ext == ".com"
}
