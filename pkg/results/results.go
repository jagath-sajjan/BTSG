package results

import (
	"btsg/internal/scanner"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	ResultsDir  = ".btsg"
	ResultsFile = "results.json"
)

// ScanResultsFile represents the structure saved to .btsg/results.json
type ScanResultsFile struct {
	Metadata ScanMetadata       `json:"metadata"`
	Findings []*scanner.Finding `json:"findings"`
}

// ScanMetadata holds information about the scan
type ScanMetadata struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Duration  string    `json:"duration"`
	Scanner   string    `json:"scanner"`
	Version   string    `json:"version"`
}

// Save saves scan results to .btsg/results.json
func Save(results *scanner.ScanResults, scanPath string) error {
	// Create .btsg directory if it doesn't exist
	if err := os.MkdirAll(ResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", ResultsDir, err)
	}

	// Assign IDs to findings if not already assigned
	for i, finding := range results.Findings {
		if finding.ID == "" {
			finding.ID = fmt.Sprintf("BTSG-%03d", i+1)
		}
	}

	// Create results file structure
	resultsFile := &ScanResultsFile{
		Metadata: ScanMetadata{
			Timestamp: time.Now(),
			Path:      scanPath,
			Duration:  results.Duration.String(),
			Scanner:   "BTSG",
			Version:   "1.0.0",
		},
		Findings: results.Findings,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(resultsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Write to file
	resultsPath := filepath.Join(ResultsDir, ResultsFile)
	if err := os.WriteFile(resultsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}

// Load loads scan results from .btsg/results.json
func Load() (*ScanResultsFile, error) {
	resultsPath := filepath.Join(ResultsDir, ResultsFile)

	// Check if file exists
	if _, err := os.Stat(resultsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no scan results found. Run 'btsg scan' first")
	}

	// Read file
	data, err := os.ReadFile(resultsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read results file: %w", err)
	}

	// Unmarshal JSON
	var resultsFile ScanResultsFile
	if err := json.Unmarshal(data, &resultsFile); err != nil {
		return nil, fmt.Errorf("failed to parse results file: %w", err)
	}

	return &resultsFile, nil
}

// FindByID finds a specific finding by its ID
func FindByID(id string) (*scanner.Finding, error) {
	resultsFile, err := Load()
	if err != nil {
		return nil, err
	}

	for _, finding := range resultsFile.Findings {
		if finding.ID == id {
			return finding, nil
		}
	}

	return nil, fmt.Errorf("finding with ID %s not found", id)
}

// GetResultsPath returns the full path to the results file
func GetResultsPath() string {
	return filepath.Join(ResultsDir, ResultsFile)
}

// Made with Bob
