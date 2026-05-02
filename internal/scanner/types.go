package scanner

import (
	"context"
	"time"
)

// ScannerEngine defines the interface that all scanners must implement
type ScannerEngine interface {
	// Name returns the scanner name (e.g., "bandit", "pip-audit")
	Name() string

	// Version returns the scanner version
	Version() string

	// Type returns the vulnerability type this scanner detects
	Type() string

	// IsAvailable checks if the scanner is installed and available
	IsAvailable() bool

	// Scan executes the scanner and returns raw results
	Scan(ctx context.Context, config *ScanConfig) (*RawScanResult, error)

	// Normalize converts raw scanner output to unified format
	Normalize(raw *RawScanResult) ([]*Finding, error)

	// GetCommand returns the command that will be executed
	GetCommand(config *ScanConfig) []string
}

// ScanConfig holds configuration for a scan
type ScanConfig struct {
	Path      string
	Recursive bool
	Verbose   bool
	Timeout   time.Duration
}

// RawScanResult holds the raw output from a scanner
type RawScanResult struct {
	Scanner  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Error    error
}

// Finding represents a security finding in unified format
type Finding struct {
	ID          string `json:"id"`
	Tool        string `json:"tool"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Code        string `json:"code,omitempty"`
	CWE         string `json:"cwe,omitempty"`
	Confidence  string `json:"confidence,omitempty"`
}

// ScanResults holds the complete scan results
type ScanResults struct {
	Findings     []*Finding
	TotalScanned int
	Duration     time.Duration
	Errors       []string
}

// Made with Bob
