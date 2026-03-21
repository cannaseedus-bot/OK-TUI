package session

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Checkpoint represents a saved snapshot of session state
type Checkpoint struct {
	SessionID   string    `json:"-"`
	Timestamp   time.Time `json:"timestamp"`
	Filename    string    `json:"-"`
	Size        int64     `json:"-"`
	Session     *Session  `json:"-"` // Transient, not persisted
}

// CheckpointManager handles session checkpointing and recovery
type CheckpointManager struct {
	sessionPath      string
	maxCheckpoints   int
	mu               sync.RWMutex
	lastCheckpoint   map[string]time.Time
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(basePath string) (*CheckpointManager, error) {
	// Create checkpoints directory
	checkpointDir := filepath.Join(basePath, "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	return &CheckpointManager{
		sessionPath:    basePath,
		maxCheckpoints: 10,
		lastCheckpoint: make(map[string]time.Time),
	}, nil
}

// SaveCheckpoint saves a session checkpoint to disk
func (cm *CheckpointManager) SaveCheckpoint(session *Session) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create session checkpoint directory
	sessionCheckpointDir := filepath.Join(cm.sessionPath, "checkpoints", session.ID)
	if err := os.MkdirAll(sessionCheckpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create session checkpoint directory: %w", err)
	}

	// Generate checkpoint filename with timestamp
	timestamp := time.Now()
	checkpointFile := filepath.Join(
		sessionCheckpointDir,
		fmt.Sprintf("checkpoint_%d.gob.gz", timestamp.Unix()),
	)

	// Open file for writing
	f, err := os.Create(checkpointFile)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint file: %w", err)
	}
	defer f.Close()

	// Compress and encode session
	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	encoder := gob.NewEncoder(gzWriter)
	if err := encoder.Encode(session); err != nil {
		os.Remove(checkpointFile) // Clean up on error
		return fmt.Errorf("failed to encode checkpoint: %w", err)
	}

	// Track last checkpoint time
	cm.lastCheckpoint[session.ID] = timestamp

	// Prune old checkpoints
	if err := cm.pruneCheckpoints(sessionCheckpointDir); err != nil {
		fmt.Printf("Warning: failed to prune checkpoints: %v\n", err)
	}

	return nil
}

// LoadCheckpoint loads the most recent checkpoint for a session
func (cm *CheckpointManager) LoadCheckpoint(sessionID string) (*Session, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessionCheckpointDir := filepath.Join(cm.sessionPath, "checkpoints", sessionID)

	// Find most recent checkpoint
	checkpointFile, err := cm.findMostRecentCheckpoint(sessionCheckpointDir)
	if err != nil {
		return nil, err
	}

	if checkpointFile == "" {
		return nil, fmt.Errorf("no checkpoints found for session: %s", sessionID)
	}

	// Open and decompress checkpoint
	f, err := os.Open(checkpointFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Decode session
	var session Session
	decoder := gob.NewDecoder(gzReader)
	if err := decoder.Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode checkpoint: %w", err)
	}

	return &session, nil
}

// ListCheckpoints returns all checkpoints for a session
func (cm *CheckpointManager) ListCheckpoints(sessionID string) ([]Checkpoint, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessionCheckpointDir := filepath.Join(cm.sessionPath, "checkpoints", sessionID)

	files, err := ioutil.ReadDir(sessionCheckpointDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Checkpoint{}, nil
		}
		return nil, err
	}

	checkpoints := make([]Checkpoint, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, _ := file.Stat()
		checkpoint := Checkpoint{
			SessionID: sessionID,
			Filename:  file.Name(),
			Size:      file.Size(),
		}

		// Parse timestamp from filename
		if _, err := fmt.Sscanf(file.Name(), "checkpoint_%d.gob.gz", &checkpoint.Timestamp); err == nil {
			checkpoint.Timestamp = time.Unix(0, 0) // Will be overwritten
			// Extract unix timestamp from filename more carefully
			parts := file.Name()
			fmt.Sscanf(parts, "checkpoint_%d.gob.gz", &info)
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Timestamp.After(checkpoints[j].Timestamp)
	})

	return checkpoints, nil
}

// DeleteCheckpoints deletes all checkpoints for a session
func (cm *CheckpointManager) DeleteCheckpoints(sessionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	sessionCheckpointDir := filepath.Join(cm.sessionPath, "checkpoints", sessionID)
	if err := os.RemoveAll(sessionCheckpointDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(cm.lastCheckpoint, sessionID)
	return nil
}

// RecoveryAvailable checks if a recovery checkpoint is available
func (cm *CheckpointManager) RecoveryAvailable(sessionID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessionCheckpointDir := filepath.Join(cm.sessionPath, "checkpoints", sessionID)
	checkpointFile, _ := cm.findMostRecentCheckpoint(sessionCheckpointDir)
	return checkpointFile != ""
}

// pruneCheckpoints removes old checkpoints, keeping only the most recent maxCheckpoints
func (cm *CheckpointManager) pruneCheckpoints(checkpointDir string) error {
	files, err := ioutil.ReadDir(checkpointDir)
	if err != nil {
		return err
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// Delete old files beyond maxCheckpoints
	for i := cm.maxCheckpoints; i < len(files); i++ {
		oldFile := filepath.Join(checkpointDir, files[i].Name())
		if err := os.Remove(oldFile); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// findMostRecentCheckpoint finds the most recent checkpoint file
func (cm *CheckpointManager) findMostRecentCheckpoint(checkpointDir string) (string, error) {
	files, err := ioutil.ReadDir(checkpointDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var mostRecent os.FileInfo
	var mostRecentPath string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if mostRecent == nil || file.ModTime().After(mostRecent.ModTime()) {
			mostRecent = file
			mostRecentPath = filepath.Join(checkpointDir, file.Name())
		}
	}

	return mostRecentPath, nil
}

// CompressCheckpointFile compresses an existing checkpoint file
func (cm *CheckpointManager) CompressCheckpointFile(checkpointFile string) error {
	// Read existing uncompressed file
	data, err := ioutil.ReadFile(checkpointFile)
	if err != nil {
		return err
	}

	// Create compressed version
	compressedFile := checkpointFile + ".gz"
	f, err := os.Create(compressedFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	_, err = io.WriteString(gzWriter, string(data))
	if err != nil {
		os.Remove(compressedFile)
		return err
	}

	// Remove original uncompressed file
	return os.Remove(checkpointFile)
}
