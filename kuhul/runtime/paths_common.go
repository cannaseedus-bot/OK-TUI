//go:build !windows

package runtime

import (
	"os"
	"path/filepath"
)

// PathSeparator returns the OS-specific path separator
func PathSeparator() string {
	return string(filepath.Separator)
}

// NormalizePath normalizes a path for the current OS
func NormalizePath(path string) string {
	return filepath.FromSlash(path)
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback for Unix systems
		if h, ok := os.LookupEnv("HOME"); ok {
			return h
		}
		return ""
	}
	return home
}

// GetAppDataDir returns the application data directory
// On Unix systems, this is typically ~/.local/share or ~/.config
func GetAppDataDir() string {
	home := GetHomeDir()
	if home == "" {
		return ""
	}

	// Check XDG_CONFIG_HOME first
	if xdg, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return xdg
	}

	// Default to ~/.config on Unix
	return filepath.Join(home, ".config")
}

// GetCacheDir returns the cache directory
// On Unix systems, typically ~/.cache or $XDG_CACHE_HOME
func GetCacheDir() string {
	home := GetHomeDir()
	if home == "" {
		return ""
	}

	// Check XDG_CACHE_HOME first
	if xdg, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return xdg
	}

	// Default to ~/.cache on Unix
	return filepath.Join(home, ".cache")
}

// JoinPath joins path components with the OS-specific separator
func JoinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// IsAbsolutePath checks if a path is absolute
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// ExpandPath expands user home directory (~) in path
func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home := GetHomeDir()
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
