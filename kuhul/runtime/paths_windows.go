//go:build windows

package runtime

import (
	"os"
	"path/filepath"
	"strings"
)

// PathSeparator returns the OS-specific path separator
func PathSeparator() string {
	return string(filepath.Separator)
}

// NormalizePath normalizes a path for Windows
// Converts forward slashes to backslashes and handles UNC paths
func NormalizePath(path string) string {
	// Handle UNC paths (\\server\share)
	if strings.HasPrefix(path, "\\\\") {
		return path
	}
	// Convert forward slashes to backslashes
	return filepath.FromSlash(path)
}

// GetHomeDir returns the user's home directory on Windows
// Falls back to USERPROFILE, then HOMEDRIVE+HOMEPATH
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Try USERPROFILE (Windows env var for user home)
		if h, ok := os.LookupEnv("USERPROFILE"); ok {
			return h
		}
		// Try HOMEDRIVE + HOMEPATH combo
		if drive, ok := os.LookupEnv("HOMEDRIVE"); ok {
			if path, ok := os.LookupEnv("HOMEPATH"); ok {
				return filepath.Join(drive, path)
			}
		}
		return ""
	}
	return home
}

// GetAppDataDir returns the Windows AppData directory
// Tries APPDATA first, then falls back to USERPROFILE\AppData\Roaming
func GetAppDataDir() string {
	if appdata, ok := os.LookupEnv("APPDATA"); ok {
		return appdata
	}

	home := GetHomeDir()
	if home == "" {
		return ""
	}

	return filepath.Join(home, "AppData", "Roaming")
}

// GetLocalAppDataDir returns the Windows LocalAppData directory
// This is for per-user application data that doesn't roam
func GetLocalAppDataDir() string {
	if localappdata, ok := os.LookupEnv("LOCALAPPDATA"); ok {
		return localappdata
	}

	home := GetHomeDir()
	if home == "" {
		return ""
	}

	return filepath.Join(home, "AppData", "Local")
}

// GetCacheDir returns the cache directory on Windows
// Typically uses LocalAppData\Cache
func GetCacheDir() string {
	localdata := GetLocalAppDataDir()
	if localdata != "" {
		return filepath.Join(localdata, "Temp")
	}
	return ""
}

// JoinPath joins path components with backslashes on Windows
func JoinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// IsAbsolutePath checks if a path is absolute
// Handles both regular absolute paths (C:\...) and UNC paths (\\server\...)
func IsAbsolutePath(path string) bool {
	if filepath.IsAbs(path) {
		return true
	}
	// Check for UNC path
	return strings.HasPrefix(path, "\\\\")
}

// ExpandPath expands user home directory (~) and Windows environment variables in path
func ExpandPath(path string) string {
	// Expand environment variables first
	expanded := os.ExpandEnv(path)

	// Then expand ~ to home directory
	if len(expanded) > 0 && expanded[0] == '~' {
		home := GetHomeDir()
		if home != "" {
			return filepath.Join(home, expanded[1:])
		}
	}

	return expanded
}

// GetDriveLetter returns the drive letter for a path (e.g., "C" for "C:\Users\...")
// Returns empty string if path is not on a Windows drive
func GetDriveLetter(path string) string {
	if len(path) >= 2 && path[1] == ':' {
		return strings.ToUpper(string(path[0]))
	}
	return ""
}

// ConvertToUNC converts a local path to a UNC path
// Useful for accessing network shares
func ConvertToUNC(server, share, path string) string {
	parts := []string{"\\\\" + server, share}
	if path != "" {
		parts = append(parts, path)
	}
	return filepath.Join(parts...)
}
