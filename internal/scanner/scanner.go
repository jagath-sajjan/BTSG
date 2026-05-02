package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Scanner orchestrates multiple scanner engines
type Scanner struct {
	engines []ScannerEngine
	config  *ScanConfig
	mu      sync.RWMutex
}

// scanResult holds the result from a single scanner execution
type scanResult struct {
	scanner  string
	findings []*Finding
	duration time.Duration
	err      error
}

// New creates a new scanner instance
func New(config *ScanConfig) *Scanner {
	s := &Scanner{
		config:  config,
		engines: []ScannerEngine{},
	}

	// Register available scanners
	s.RegisterEngine(NewBanditScanner())
	s.RegisterEngine(NewPipAuditScanner())
	s.RegisterEngine(NewDetectSecretsScanner())

	return s
}

// RegisterEngine adds a scanner engine
func (s *Scanner) RegisterEngine(engine ScannerEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if engine.IsAvailable() {
		s.engines = append(s.engines, engine)
		if s.config.Verbose {
			fmt.Printf("Registered scanner: %s (v%s)\n", engine.Name(), engine.Version())
		}
	}
}

// Scan executes all available scanners concurrently
func (s *Scanner) Scan(ctx context.Context) (*ScanResults, error) {
	startTime := time.Now()

	if s.config.Verbose {
		fmt.Printf("Starting scan of %s...\n", s.config.Path)
		fmt.Printf("Available scanners: %d\n", len(s.engines))
	}

	// Check if we have any scanners
	if len(s.engines) == 0 {
		return &ScanResults{
			Findings:     []*Finding{},
			TotalScanned: 0,
			Duration:     time.Since(startTime),
			Errors:       []string{"No scanners available. Please install security scanning tools."},
		}, nil
	}

	// Create buffered channels for results
	resultsChan := make(chan *scanResult, len(s.engines))

	// Use WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Launch scanner goroutines
	for _, engine := range s.engines {
		wg.Add(1)
		go s.runScanner(ctx, engine, resultsChan, &wg)
	}

	// Close results channel when all scanners complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results with timeout protection
	return s.collectResults(ctx, resultsChan, startTime)
}

// runScanner executes a single scanner in a goroutine with panic recovery
func (s *Scanner) runScanner(ctx context.Context, engine ScannerEngine, resultsChan chan<- *scanResult, wg *sync.WaitGroup) {
	defer wg.Done()

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in %s scanner: %v", engine.Name(), r)
			resultsChan <- &scanResult{
				scanner: engine.Name(),
				err:     err,
			}
		}
	}()

	scanStart := time.Now()
	result := &scanResult{
		scanner: engine.Name(),
	}

	if s.config.Verbose {
		fmt.Printf("→ Running %s scanner...\n", engine.Name())
	}

	// Create context with timeout for this specific scanner
	scanCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Execute scanner
	raw, err := engine.Scan(scanCtx, s.config)
	if err != nil {
		// Check if it's a context error
		if ctx.Err() != nil {
			result.err = fmt.Errorf("%s: context cancelled", engine.Name())
		} else if scanCtx.Err() == context.DeadlineExceeded {
			result.err = fmt.Errorf("%s: timeout after %s", engine.Name(), s.config.Timeout)
		} else {
			result.err = fmt.Errorf("%s: %v", engine.Name(), err)
		}
		result.duration = time.Since(scanStart)
		resultsChan <- result
		return
	}

	// Normalize results
	findings, err := engine.Normalize(raw)
	if err != nil {
		result.err = fmt.Errorf("%s normalization: %v", engine.Name(), err)
		result.duration = time.Since(scanStart)
		resultsChan <- result
		return
	}

	// Success
	result.findings = findings
	result.duration = time.Since(scanStart)
	resultsChan <- result

	if s.config.Verbose {
		fmt.Printf("✓ %s completed in %s (%d findings)\n",
			engine.Name(), result.duration, len(findings))
	}
}

// collectResults safely collects results from all scanners
func (s *Scanner) collectResults(ctx context.Context, resultsChan <-chan *scanResult, startTime time.Time) (*ScanResults, error) {
	var (
		allFindings []*Finding
		errors      []string
		mu          sync.Mutex
	)

	// Collect results from channel
	for result := range resultsChan {
		mu.Lock()
		if result.err != nil {
			errors = append(errors, result.err.Error())
		} else {
			allFindings = append(allFindings, result.findings...)
		}
		mu.Unlock()
	}

	// Deduplicate findings
	allFindings = DeduplicateFindings(allFindings)

	// Sort by severity
	SortFindingsBySeverity(allFindings)

	results := &ScanResults{
		Findings:     allFindings,
		TotalScanned: len(allFindings),
		Duration:     time.Since(startTime),
		Errors:       errors,
	}

	if s.config.Verbose {
		fmt.Printf("\nScan completed in %s\n", results.Duration)
		fmt.Printf("Total findings: %d (after deduplication)\n", len(allFindings))
		if len(errors) > 0 {
			fmt.Printf("Errors encountered: %d\n", len(errors))
		}
	}

	return results, nil
}

// ListAvailableScanners returns all available scanner engines
func (s *Scanner) ListAvailableScanners() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var names []string
	for _, engine := range s.engines {
		names = append(names, engine.Name())
	}
	return names
}

// GetScannerInfo returns detailed information about registered scanners
func (s *Scanner) GetScannerInfo() []map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var info []map[string]string
	for _, engine := range s.engines {
		info = append(info, map[string]string{
			"name":    engine.Name(),
			"version": engine.Version(),
			"type":    engine.Type(),
		})
	}
	return info
}

// ScanWithProgress executes scanners and reports progress via callback
func (s *Scanner) ScanWithProgress(ctx context.Context, progressFn func(scanner string, status string)) (*ScanResults, error) {
	startTime := time.Now()

	if len(s.engines) == 0 {
		return &ScanResults{
			Findings:     []*Finding{},
			TotalScanned: 0,
			Duration:     time.Since(startTime),
			Errors:       []string{"No scanners available"},
		}, nil
	}

	resultsChan := make(chan *scanResult, len(s.engines))
	var wg sync.WaitGroup

	for _, engine := range s.engines {
		wg.Add(1)
		go func(e ScannerEngine) {
			defer wg.Done()

			if progressFn != nil {
				progressFn(e.Name(), "starting")
			}

			scanStart := time.Now()
			result := &scanResult{scanner: e.Name()}

			scanCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
			defer cancel()

			raw, err := e.Scan(scanCtx, s.config)
			if err != nil {
				result.err = err
				if progressFn != nil {
					progressFn(e.Name(), "failed")
				}
			} else {
				findings, err := e.Normalize(raw)
				if err != nil {
					result.err = err
					if progressFn != nil {
						progressFn(e.Name(), "failed")
					}
				} else {
					result.findings = findings
					if progressFn != nil {
						progressFn(e.Name(), "completed")
					}
				}
			}

			result.duration = time.Since(scanStart)
			resultsChan <- result
		}(engine)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	return s.collectResults(ctx, resultsChan, startTime)
}

// Made with Bob
