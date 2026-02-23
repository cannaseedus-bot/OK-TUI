package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestPathHandling tests cross-platform path handling
func TestPathHandling(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		check func(string) bool
	}{
		{
			name: "normalize_path",
			path: "/home/user/file.txt",
			check: func(p string) bool {
				// Should not be empty
				return len(p) > 0
			},
		},
		{
			name: "join_paths",
			path: filepath.Join("home", "user", "file.txt"),
			check: func(p string) bool {
				return strings.Contains(p, "file.txt")
			},
		},
		{
			name: "absolute_path",
			path: "/absolute/path",
			check: func(p string) bool {
				return filepath.IsAbs(p)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.path)
			if !tt.check(result) {
				t.Errorf("NormalizePath(%q) = %q, check failed", tt.path, result)
			}
		})
	}
}

// TestEnvironmentVariables tests environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Test getting environment variables
	home := GetHomeDir()
	if home == "" {
		t.Fatal("GetHomeDir() returned empty string")
	}

	t.Logf("Home directory: %s", home)

	// Test AppData directory
	appdata := GetAppDataDir()
	if appdata == "" {
		t.Error("GetAppDataDir() returned empty string")
	}
	t.Logf("AppData directory: %s", appdata)
}

// TestFileOperations tests cross-platform file operations
func TestFileOperations(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, cross-platform world!"

	// Test write
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(testFile)

	// Test read
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}

	// Test listdir
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "test.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Test file not found in directory listing")
	}
}

// TestProcessExecution tests process execution
func TestProcessExecution(t *testing.T) {
	// Test a simple command that works on all platforms
	var result *ProcessResult

	if runtime.GOOS == "windows" {
		result = RunCommand("cmd.exe", []string{"/c", "echo", "test"})
	} else {
		result = RunCommand("echo", []string{"test"})
	}

	if result == nil {
		t.Fatal("RunCommand returned nil result")
	}

	if result.ExitCode != 0 {
		t.Errorf("RunCommand returned non-zero exit code: %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	if !strings.Contains(result.Stdout, "test") {
		t.Errorf("Expected 'test' in stdout, got: %s", result.Stdout)
	}

	t.Logf("Process output: %s", strings.TrimSpace(result.Stdout))
}

// TestCommandExists tests command lookup
func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	exists := CommandExists("echo")
	if !exists {
		t.Errorf("CommandExists('echo') returned false, but echo should exist")
	}

	// Test with a command that shouldn't exist
	notExists := CommandExists("nonexistent_command_12345")
	if notExists {
		t.Errorf("CommandExists('nonexistent_command_12345') returned true, but command should not exist")
	}
}

// TestNetworkPortManagement tests port availability checks
func TestNetworkPortManagement(t *testing.T) {
	// Test with a high port number that's likely to be free
	testPort := 65432

	// Try multiple times to find an available port
	for testPort < 65500 {
		if PortAvailable(testPort) {
			t.Logf("Found available port: %d", testPort)
			break
		}
		testPort++
	}

	if testPort >= 65500 {
		t.Skip("Could not find available port for testing")
	}

	// Verify the port is available
	if !PortAvailable(testPort) {
		t.Errorf("PortAvailable(%d) returned false after finding it available", testPort)
	}

	// Well-known port should not be available (or may be)
	wellKnownPort := 80
	result := PortAvailable(wellKnownPort)
	t.Logf("Port 80 availability: %v (may vary by system)", result)
}

// TestServiceDiscovery tests service discovery functionality
func TestServiceDiscovery(t *testing.T) {
	// This test is informational - it just checks the structure
	// Real services may not be running

	// Test port detection
	testPorts := []int{11434, 61683, 8080}

	for _, port := range testPorts {
		available := PortAvailable(port)
		t.Logf("Port %d available: %v", port, available)
	}
}

// TestPathSeparator tests OS-specific path separator handling
func TestPathSeparator(t *testing.T) {
	sep := PathSeparator()

	switch runtime.GOOS {
	case "windows":
		if sep != "\\" {
			t.Errorf("Expected Windows path separator '\\', got %q", sep)
		}
	case "linux", "darwin":
		if sep != "/" {
			t.Errorf("Expected Unix path separator '/', got %q", sep)
		}
	}

	t.Logf("Path separator for %s: %q", runtime.GOOS, sep)
}

// TestPathExpansion tests environment variable expansion in paths
func TestPathExpansion(t *testing.T) {
	// Test home directory expansion
	homePath := "~/testfile.txt"
	expanded := ExpandPath(homePath)

	home := GetHomeDir()
	if !strings.Contains(expanded, home) {
		t.Errorf("ExpandPath(%q) = %q, expected to contain %q", homePath, expanded, home)
	}

	t.Logf("Expanded home path: %s -> %s", homePath, expanded)

	// Test regular paths without expansion
	regularPath := "/absolute/path/file.txt"
	expanded = ExpandPath(regularPath)

	if expanded != regularPath {
		t.Errorf("ExpandPath(%q) = %q, expected unchanged", regularPath, expanded)
	}

	t.Logf("Regular path: %s -> %s", regularPath, expanded)
}

// TestBuiltinIntegration tests that builtins are properly registered
func TestBuiltinIntegration(t *testing.T) {
	systemFuncs := []string{
		"sys.os",
		"sys.env_get",
		"sys.env_home",
		"sys.path_sep",
		"sys.readfile",
		"sys.writefile",
		"sys.proc_run",
	}

	for _, fn := range systemFuncs {
		if _, ok := Builtins[fn]; !ok {
			t.Errorf("Builtin function %q not registered", fn)
		}
	}

	t.Logf("All %d system builtins properly registered", len(systemFuncs))
}

// TestCrossPlatformConsistency ensures consistent behavior across platforms
func TestCrossPlatformConsistency(t *testing.T) {
	// Create consistent test directory
	tmpDir := t.TempDir()

	// Test path joining consistency
	joined := JoinPath(tmpDir, "subdir", "file.txt")
	if joined == "" {
		t.Error("JoinPath returned empty string")
	}

	// Test absolute path detection
	absPath, err := filepath.Abs(joined)
	if err != nil {
		t.Errorf("filepath.Abs failed: %v", err)
	}
	if !IsAbsolutePath(absPath) {
		t.Errorf("IsAbsolutePath(%q) returned false for absolute path", absPath)
	}

	// Test relative path detection
	relPath := "relative/path.txt"
	if IsAbsolutePath(relPath) {
		t.Errorf("IsAbsolutePath(%q) returned true for relative path", relPath)
	}

	t.Logf("Cross-platform consistency checks passed")
}

// TestProcessInfoStructure tests process information structure
func TestProcessInfoStructure(t *testing.T) {
	// Just verify the struct exists and can be instantiated
	procInfo := ProcessInfo{
		PID:  1234,
		Name: "test.exe",
	}

	if procInfo.PID != 1234 {
		t.Error("ProcessInfo PID not set correctly")
	}

	if procInfo.Name != "test.exe" {
		t.Error("ProcessInfo Name not set correctly")
	}

	t.Logf("ProcessInfo structure working correctly")
}

// BenchmarkPathNormalization benchmarks path normalization
func BenchmarkPathNormalization(b *testing.B) {
	testPath := "/home/user/projects/kuhul/runtime/test/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizePath(testPath)
	}
}

// BenchmarkPathJoin benchmarks path joining
func BenchmarkPathJoin(b *testing.B) {
	parts := []string{"/home", "user", "projects", "kuhul", "file.txt"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JoinPath(parts...)
	}
}

// BenchmarkCommandExists benchmarks command lookup
func BenchmarkCommandExists(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CommandExists("echo")
	}
}

// TestRuntimePlatformInfo displays runtime platform information
func TestRuntimePlatformInfo(t *testing.T) {
	fmt.Printf("\n=== Runtime Platform Information ===\n")
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Number of CPUs: %d\n", runtime.NumCPU())

	fmt.Printf("\nPlatform Abstraction Functions:\n")
	fmt.Printf("  PathSeparator: %q\n", PathSeparator())
	fmt.Printf("  HomeDir: %s\n", GetHomeDir())
	fmt.Printf("  AppDataDir: %s\n", GetAppDataDir())
	fmt.Printf("  CacheDir: %s\n", GetCacheDir())

	fmt.Printf("\n=== End Platform Information ===\n\n")
}
