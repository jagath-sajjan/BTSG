package ui

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps progressbar with custom styling
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new styled progress bar
func NewProgressBar(max int, description string) *ProgressBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
	)

	return &ProgressBar{bar: bar}
}

// NewSpinner creates a spinner for indeterminate progress
func NewSpinner(description string) *ProgressBar {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(true),
	)

	return &ProgressBar{bar: bar}
}

// Add increments the progress bar
func (p *ProgressBar) Add(n int) error {
	return p.bar.Add(n)
}

// Set sets the progress bar to a specific value
func (p *ProgressBar) Set(n int) error {
	return p.bar.Set(n)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() error {
	return p.bar.Finish()
}

// Clear clears the progress bar
func (p *ProgressBar) Clear() error {
	return p.bar.Clear()
}

// Describe updates the description
func (p *ProgressBar) Describe(description string) {
	p.bar.Describe(description)
}

// ProgressTracker tracks progress for multiple tasks
type ProgressTracker struct {
	tasks    map[string]*ProgressBar
	writer   io.Writer
	finished int
	total    int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		tasks:  make(map[string]*ProgressBar),
		writer: os.Stderr,
		total:  total,
	}
}

// StartTask starts tracking a new task
func (pt *ProgressTracker) StartTask(name, description string, max int) {
	bar := NewProgressBar(max, description)
	pt.tasks[name] = bar
}

// UpdateTask updates a task's progress
func (pt *ProgressTracker) UpdateTask(name string, n int) error {
	if bar, ok := pt.tasks[name]; ok {
		return bar.Add(n)
	}
	return fmt.Errorf("task not found: %s", name)
}

// FinishTask marks a task as complete
func (pt *ProgressTracker) FinishTask(name string) error {
	if bar, ok := pt.tasks[name]; ok {
		pt.finished++
		return bar.Finish()
	}
	return fmt.Errorf("task not found: %s", name)
}

// IsComplete returns whether all tasks are complete
func (pt *ProgressTracker) IsComplete() bool {
	return pt.finished >= pt.total
}

// ScanProgress represents progress for a security scan
type ScanProgress struct {
	Total     int
	Completed int
	Current   string
	bar       *ProgressBar
}

// NewScanProgress creates a new scan progress tracker
func NewScanProgress(total int) *ScanProgress {
	return &ScanProgress{
		Total:     total,
		Completed: 0,
		bar:       NewProgressBar(total, "Scanning..."),
	}
}

// Update updates the scan progress
func (sp *ScanProgress) Update(scanner string) error {
	sp.Completed++
	sp.Current = scanner
	sp.bar.Describe(fmt.Sprintf("Scanning with %s", scanner))
	return sp.bar.Add(1)
}

// Finish completes the scan progress
func (sp *ScanProgress) Finish() error {
	return sp.bar.Finish()
}

// Made with Bob
