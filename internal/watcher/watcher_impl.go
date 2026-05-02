package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"btsg/internal/scanner"
)

// watcherImpl implements the Watcher interface
type watcherImpl struct {
	config   *WatchConfig
	scanner  *scanner.Scanner
	watcher  *fsnotify.Watcher
	stats    *WatchStats
	running  bool
	stopChan chan struct{}
	mu       sync.RWMutex

	// Debouncing
	debounceTimer *time.Timer
	pendingEvents map[string]*WatchEvent
	eventsMu      sync.Mutex
}

// NewWatcher creates a new file watcher instance
func NewWatcher(config *WatchConfig, scnr *scanner.Scanner) (*watcherImpl, error) {
	if config == nil {
		config = DefaultWatchConfig()
	}

	if scnr == nil {
		return nil, &WatchError{
			Code:    ErrCodeWatcherFailed,
			Message: "scanner cannot be nil",
		}
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, &WatchError{
			Code:    ErrCodeWatcherFailed,
			Message: "failed to create fsnotify watcher",
			Cause:   err,
		}
	}

	return &watcherImpl{
		config:        config,
		scanner:       scnr,
		watcher:       fsWatcher,
		stats:         &WatchStats{},
		stopChan:      make(chan struct{}),
		pendingEvents: make(map[string]*WatchEvent),
	}, nil
}

// Start begins watching for file changes
func (w *watcherImpl) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return &WatchError{
			Code:    ErrCodeWatcherFailed,
			Message: "watcher is already running",
		}
	}
	w.running = true
	w.stats.StartTime = time.Now()
	w.mu.Unlock()

	// Add paths to watch
	for _, path := range w.config.Paths {
		if err := w.addPath(path); err != nil {
			w.Stop()
			return err
		}
	}

	if w.config.Verbose {
		fmt.Printf("🔍 Watching %d paths for changes...\n", len(w.stats.WatchedPaths))
		for _, path := range w.stats.WatchedPaths {
			fmt.Printf("   - %s\n", path)
		}
	}

	// Run initial scan if configured
	if w.config.ScanOnStart {
		if w.config.Verbose {
			fmt.Println("\n📊 Running initial scan...")
		}
		w.triggerScan("initial scan")
	}

	// Start event loop
	go w.eventLoop()

	return nil
}

// Stop stops the watcher
func (w *watcherImpl) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return &WatchError{
			Code:    ErrCodeWatcherNotRunning,
			Message: "watcher is not running",
		}
	}

	w.running = false
	close(w.stopChan)

	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}

	if err := w.watcher.Close(); err != nil {
		return &WatchError{
			Code:    ErrCodeWatcherFailed,
			Message: "failed to close watcher",
			Cause:   err,
		}
	}

	if w.config.Verbose {
		fmt.Println("\n✅ Watcher stopped")
	}

	return nil
}

// IsRunning returns whether the watcher is currently running
func (w *watcherImpl) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetStats returns watcher statistics
func (w *watcherImpl) GetStats() *WatchStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := *w.stats
	statsCopy.WatchedPaths = make([]string, len(w.stats.WatchedPaths))
	copy(statsCopy.WatchedPaths, w.stats.WatchedPaths)

	return &statsCopy
}

// addPath adds a path to the watcher
func (w *watcherImpl) addPath(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return &WatchError{
			Code:    ErrCodeInvalidPath,
			Message: fmt.Sprintf("path does not exist: %s", path),
			Cause:   err,
		}
	}

	// If it's a directory and recursive is enabled, walk the tree
	if info.IsDir() && w.config.Recursive {
		return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip ignored paths
			if w.shouldIgnore(walkPath) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Only watch directories
			if info.IsDir() {
				if err := w.watcher.Add(walkPath); err != nil {
					return err
				}
				w.stats.WatchedPaths = append(w.stats.WatchedPaths, walkPath)
			}

			return nil
		})
	}

	// Add single path
	if err := w.watcher.Add(path); err != nil {
		return &WatchError{
			Code:    ErrCodeWatcherFailed,
			Message: fmt.Sprintf("failed to watch path: %s", path),
			Cause:   err,
		}
	}

	w.stats.WatchedPaths = append(w.stats.WatchedPaths, path)
	return nil
}

// shouldIgnore checks if a path should be ignored
func (w *watcherImpl) shouldIgnore(path string) bool {
	for _, ignore := range w.config.IgnorePaths {
		if strings.Contains(path, ignore) {
			return true
		}
	}
	return false
}

// shouldWatch checks if a file should trigger a scan based on patterns
func (w *watcherImpl) shouldWatch(path string) bool {
	if len(w.config.Patterns) == 0 {
		return true
	}

	for _, pattern := range w.config.Patterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}

	return false
}

// eventLoop processes file system events
func (w *watcherImpl) eventLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			if w.config.Verbose {
				fmt.Printf("⚠️  Watcher error: %v\n", err)
			}

		case <-w.stopChan:
			return
		}
	}
}

// handleEvent processes a single file system event
func (w *watcherImpl) handleEvent(event fsnotify.Event) {
	// Skip if path should be ignored
	if w.shouldIgnore(event.Name) {
		return
	}

	// Skip if file pattern doesn't match
	if !w.shouldWatch(event.Name) {
		return
	}

	w.mu.Lock()
	w.stats.TotalEvents++
	w.mu.Unlock()

	// Create watch event
	watchEvent := &WatchEvent{
		Path:      event.Name,
		Operation: w.getOperationName(event.Op),
		Timestamp: time.Now(),
	}

	if w.config.Verbose {
		fmt.Printf("📝 %s: %s\n", watchEvent.Operation, watchEvent.Path)
	}

	// Add to pending events for debouncing
	w.eventsMu.Lock()
	w.pendingEvents[event.Name] = watchEvent
	w.eventsMu.Unlock()

	// Reset debounce timer
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}

	w.debounceTimer = time.AfterFunc(w.config.DebounceDelay, func() {
		w.eventsMu.Lock()
		eventCount := len(w.pendingEvents)
		w.pendingEvents = make(map[string]*WatchEvent)
		w.eventsMu.Unlock()

		if eventCount > 0 {
			w.triggerScan(fmt.Sprintf("%d file(s) changed", eventCount))
		}
	})
}

// getOperationName converts fsnotify operation to string
func (w *watcherImpl) getOperationName(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "create"
	case op&fsnotify.Write == fsnotify.Write:
		return "write"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "unknown"
	}
}

// triggerScan runs a security scan
func (w *watcherImpl) triggerScan(reason string) {
	startTime := time.Now()

	if w.config.Verbose {
		fmt.Printf("\n🔍 Triggering scan: %s\n", reason)
	}

	// Run scan with context
	ctx := context.Background()
	scanResults, err := w.scanner.Scan(ctx)
	duration := time.Since(startTime)

	result := &ScanResult{
		Timestamp: startTime,
		Findings:  scanResults.Findings,
		Error:     err,
		Duration:  duration,
	}

	// Update stats
	w.mu.Lock()
	w.stats.TotalScans++
	w.stats.LastScanTime = startTime
	w.stats.LastScanResult = result
	w.mu.Unlock()

	// Display results
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		return
	}

	findings := scanResults.Findings

	if w.config.Verbose {
		fmt.Printf("✅ Scan completed in %s\n", duration)
		fmt.Printf("   Found %d vulnerabilities\n", len(findings))

		if len(findings) > 0 {
			fmt.Println("\n📋 Vulnerabilities:")
			for i, finding := range findings {
				if i >= 5 {
					fmt.Printf("   ... and %d more\n", len(findings)-5)
					break
				}
				fmt.Printf("   - [%s] %s:%d - %s\n",
					finding.Severity,
					finding.File,
					finding.Line,
					finding.Description)
			}
		}
		fmt.Println()
	} else {
		// Minimal output
		if len(findings) > 0 {
			fmt.Printf("⚠️  Found %d vulnerabilities (scan took %s)\n", len(findings), duration)
		} else {
			fmt.Printf("✅ No vulnerabilities found (scan took %s)\n", duration)
		}
	}
}

// Made with Bob
