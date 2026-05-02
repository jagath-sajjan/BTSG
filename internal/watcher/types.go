package watcher

import (
	"time"

	"btsg/internal/scanner"
)

// WatchConfig holds configuration for the file watcher
type WatchConfig struct {
	// Paths to watch
	Paths []string

	// File patterns to watch (e.g., "*.go", "*.py")
	Patterns []string

	// Paths to ignore
	IgnorePaths []string

	// Debounce duration to avoid multiple scans for rapid changes
	DebounceDelay time.Duration

	// Whether to scan on startup
	ScanOnStart bool

	// Whether to watch subdirectories recursively
	Recursive bool

	// Verbose output
	Verbose bool
}

// DefaultWatchConfig returns default watcher configuration
func DefaultWatchConfig() *WatchConfig {
	return &WatchConfig{
		Paths:         []string{"."},
		Patterns:      []string{"*.go", "*.py", "*.js", "*.ts", "*.java", "*.rb"},
		IgnorePaths:   []string{".git", "node_modules", "vendor", ".btsg-backups", "__pycache__"},
		DebounceDelay: 2 * time.Second,
		ScanOnStart:   true,
		Recursive:     true,
		Verbose:       false,
	}
}

// WatchEvent represents a file system event
type WatchEvent struct {
	Path      string
	Operation string // "create", "write", "remove", "rename", "chmod"
	Timestamp time.Time
}

// ScanResult represents the result of an automatic scan
type ScanResult struct {
	Timestamp time.Time
	Findings  []*scanner.Finding
	Error     error
	Duration  time.Duration
}

// Watcher is the interface for file watching and automatic scanning
type Watcher interface {
	// Start begins watching for file changes
	Start() error

	// Stop stops the watcher
	Stop() error

	// IsRunning returns whether the watcher is currently running
	IsRunning() bool

	// GetStats returns watcher statistics
	GetStats() *WatchStats
}

// WatchStats holds statistics about the watcher
type WatchStats struct {
	StartTime      time.Time
	TotalEvents    int
	TotalScans     int
	LastScanTime   time.Time
	LastScanResult *ScanResult
	WatchedPaths   []string
}

// WatchError represents an error during watching
type WatchError struct {
	Code    string
	Message string
	Cause   error
}

func (e *WatchError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error codes
const (
	ErrCodeWatcherNotRunning = "WATCHER_NOT_RUNNING"
	ErrCodeWatcherFailed     = "WATCHER_FAILED"
	ErrCodeScanFailed        = "SCAN_FAILED"
	ErrCodeInvalidPath       = "INVALID_PATH"
)

// Made with Bob
