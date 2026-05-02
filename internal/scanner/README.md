# BTSG Scanner Module

## Overview

The scanner module provides a unified interface for running multiple security scanning tools and normalizing their output into a consistent format.

## Architecture

```
Scanner (Orchestrator)
    ├── BanditScanner (Python code analysis)
    ├── PipAuditScanner (Python dependency vulnerabilities)
    └── DetectSecretsScanner (Secret detection)
```

## Components

### 1. Core Types (`types.go`)

- **ScannerEngine**: Interface that all scanners must implement
- **ScanConfig**: Configuration for scan execution
- **RawScanResult**: Raw output from scanner tools
- **Finding**: Unified vulnerability format
- **ScanResults**: Aggregated scan results

### 2. Scanner Orchestrator (`scanner.go`)

Main orchestrator that:
- Registers available scanner engines
- Executes scanners concurrently
- Aggregates results
- Handles errors gracefully

### 3. Scanner Implementations

#### Bandit Scanner (`bandit.go`)
- **Tool**: Bandit (Python security linter)
- **Type**: Code vulnerabilities
- **Output**: JSON format
- **Detects**: Security issues in Python code

#### pip-audit Scanner (`pipaudit.go`)
- **Tool**: pip-audit (Python dependency scanner)
- **Type**: Dependency vulnerabilities
- **Output**: JSON format
- **Detects**: CVEs in Python packages

#### detect-secrets Scanner (`detectsecrets.go`)
- **Tool**: detect-secrets (Secret detection)
- **Type**: Secrets
- **Output**: JSON format
- **Detects**: API keys, tokens, passwords in code

### 4. Utilities (`utils.go`)

Helper functions for:
- Generating unique IDs
- Sorting findings by severity
- Deduplicating results
- Counting by severity/tool

## Usage

### Basic Usage

```go
import (
    "context"
    "btsg/internal/scanner"
    "time"
)

// Create scanner configuration
config := &scanner.ScanConfig{
    Path:      "./myproject",
    Recursive: true,
    Verbose:   true,
    Timeout:   5 * time.Minute,
}

// Initialize scanner
s := scanner.New(config)

// Run scan
ctx := context.Background()
results, err := s.Scan(ctx)
if err != nil {
    log.Fatal(err)
}

// Process results
for _, finding := range results.Findings {
    fmt.Printf("[%s] %s: %s\n", 
        finding.Severity, 
        finding.Tool, 
        finding.Description)
}
```

### Advanced Usage

```go
// Create custom scanner
s := &scanner.Scanner{}

// Register only specific scanners
s.RegisterEngine(scanner.NewBanditScanner())
s.RegisterEngine(scanner.NewDetectSecretsScanner())

// Run with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

results, err := s.Scan(ctx)

// Sort and deduplicate
scanner.SortFindingsBySeverity(results.Findings)
results.Findings = scanner.DeduplicateFindings(results.Findings)

// Get statistics
severityCounts := scanner.CountBySeverity(results.Findings)
toolCounts := scanner.CountByTool(results.Findings)
```

## Output Format

### Finding Structure

```json
{
  "id": "BTSG-a1b2c3d4",
  "tool": "bandit",
  "severity": "HIGH",
  "file": "app/views.py",
  "line": 42,
  "description": "[B201] Use of insecure pickle module",
  "code": "import pickle\ndata = pickle.loads(user_input)",
  "cwe": "CWE-502",
  "confidence": "HIGH"
}
```

### Scan Results Structure

```json
{
  "findings": [...],
  "total_scanned": 15,
  "duration": "2.5s",
  "errors": []
}
```

## Installing Scanner Tools

### Bandit (Python)

```bash
pip install bandit
bandit --version
```

### pip-audit (Python)

```bash
pip install pip-audit
pip-audit --version
```

### detect-secrets (Python)

```bash
pip install detect-secrets
detect-secrets --version
```

## Implementing a New Scanner

### Step 1: Create Scanner File

```go
// internal/scanner/myscan.go
package scanner

type MyScanner struct {
    path    string
    version string
}

func NewMyScanner() *MyScanner {
    return &MyScanner{
        path:    "myscan",
        version: detectVersion(),
    }
}
```

### Step 2: Implement Interface

```go
func (s *MyScanner) Name() string { return "myscan" }
func (s *MyScanner) Version() string { return s.version }
func (s *MyScanner) Type() string { return "code" }

func (s *MyScanner) IsAvailable() bool {
    cmd := exec.Command(s.path, "--version")
    return cmd.Run() == nil
}

func (s *MyScanner) GetCommand(config *ScanConfig) []string {
    return []string{s.path, "--json", config.Path}
}

func (s *MyScanner) Scan(ctx context.Context, config *ScanConfig) (*RawScanResult, error) {
    // Execute command and capture output
    // Return RawScanResult
}

func (s *MyScanner) Normalize(raw *RawScanResult) ([]*Finding, error) {
    // Parse raw output
    // Convert to Finding format
    // Return findings
}
```

### Step 3: Register Scanner

```go
// internal/scanner/scanner.go
func New(config *ScanConfig) *Scanner {
    s := &Scanner{config: config}
    s.RegisterEngine(NewBanditScanner())
    s.RegisterEngine(NewPipAuditScanner())
    s.RegisterEngine(NewDetectSecretsScanner())
    s.RegisterEngine(NewMyScanner()) // Add your scanner
    return s
}
```

## Error Handling

The scanner module handles errors gracefully:

1. **Tool Not Available**: Scanner is skipped if tool is not installed
2. **Execution Errors**: Errors are collected and reported in results
3. **Parse Errors**: Failed normalizations are logged as errors
4. **Timeout**: Context timeout cancels long-running scans

## Concurrency

Scanners run concurrently using goroutines:
- Each scanner runs in its own goroutine
- Results are collected via channels
- WaitGroup ensures all scanners complete
- Context allows cancellation

## Testing

### Unit Tests

```go
func TestBanditScanner_Name(t *testing.T) {
    scanner := NewBanditScanner()
    if scanner.Name() != "bandit" {
        t.Errorf("expected 'bandit', got '%s'", scanner.Name())
    }
}
```

### Integration Tests

```go
func TestScanner_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    config := &ScanConfig{
        Path:    "./testdata",
        Timeout: 1 * time.Minute,
    }
    
    s := New(config)
    results, err := s.Scan(context.Background())
    
    if err != nil {
        t.Fatalf("scan failed: %v", err)
    }
    
    t.Logf("found %d findings", len(results.Findings))
}
```

## Performance

- **Concurrent Execution**: All scanners run in parallel
- **Timeout Control**: Configurable timeout per scanner
- **Resource Efficient**: Streams output instead of buffering
- **Fast Startup**: Lazy initialization of scanners

## Limitations

1. Requires external tools to be installed
2. Output format depends on tool versions
3. Some tools may have platform-specific behavior
4. Large codebases may take time to scan

## Future Enhancements

- [ ] Add more scanner engines (Semgrep, Trivy, etc.)
- [ ] Implement result caching
- [ ] Add incremental scanning
- [ ] Support custom scanner plugins
- [ ] Add scanner health checks
- [ ] Implement retry logic for failed scans
- [ ] Add progress reporting
- [ ] Support parallel file scanning

## Contributing

To add a new scanner:
1. Create a new file `internal/scanner/toolname.go`
2. Implement the `ScannerEngine` interface
3. Register in `scanner.go`
4. Add tests
5. Update documentation

## License

MIT License - See LICENSE file for details