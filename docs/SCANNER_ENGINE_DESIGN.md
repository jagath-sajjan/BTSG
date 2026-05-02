# BTSG Scanner Engine Design

## Overview

The BTSG Scanner Engine is designed to run multiple security scanners (Bandit, pip-audit, detect-secrets, etc.) and normalize their output into a unified JSON format. This document outlines the architecture, interfaces, data schemas, and execution flow.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Scanner Orchestrator                      │
│  - Manages scanner lifecycle                                 │
│  - Coordinates concurrent execution                          │
│  - Aggregates results                                        │
└────────────┬────────────────────────────────────────────────┘
             │
             ├──────────────┬──────────────┬──────────────┐
             ▼              ▼              ▼              ▼
      ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
      │  Bandit  │   │pip-audit │   │  detect  │   │  Custom  │
      │ Scanner  │   │ Scanner  │   │ -secrets │   │ Scanner  │
      └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
           │              │              │              │
           └──────────────┴──────────────┴──────────────┘
                          │
                          ▼
                ┌──────────────────┐
                │    Normalizer    │
                │  - Parse output  │
                │  - Map to schema │
                │  - Enrich data   │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │  Unified Results │
                │   (JSON Schema)  │
                └──────────────────┘
```

## 1. Scanner Interface Design

### Core Interface

```go
// ScannerEngine defines the interface that all scanners must implement
type ScannerEngine interface {
    // Name returns the scanner name (e.g., "bandit", "pip-audit")
    Name() string
    
    // Version returns the scanner version
    Version() string
    
    // Type returns the vulnerability type this scanner detects
    Type() VulnType
    
    // IsAvailable checks if the scanner is installed and available
    IsAvailable() bool
    
    // Scan executes the scanner and returns raw results
    Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error)
    
    // Normalize converts raw scanner output to unified format
    Normalize(raw *RawScanResult) ([]Vulnerability, error)
    
    // GetCommand returns the command that will be executed
    GetCommand(config ScanConfig) []string
}
```

### Scanner Metadata

```go
// ScannerMetadata provides information about a scanner
type ScannerMetadata struct {
    Name            string            `json:"name"`
    Version         string            `json:"version"`
    Type            VulnType          `json:"type"`
    Description     string            `json:"description"`
    SupportedFiles  []string          `json:"supported_files"`
    RequiredTools   []string          `json:"required_tools"`
    ConfigOptions   map[string]string `json:"config_options"`
}
```

### Raw Scan Result

```go
// RawScanResult holds the raw output from a scanner
type RawScanResult struct {
    Scanner   string          `json:"scanner"`
    ExitCode  int             `json:"exit_code"`
    Stdout    string          `json:"stdout"`
    Stderr    string          `json:"stderr"`
    Duration  time.Duration   `json:"duration"`
    Error     error           `json:"error,omitempty"`
    Metadata  ScannerMetadata `json:"metadata"`
}
```

## 2. Unified Vulnerability Data Schema

### Core Vulnerability Schema

```go
// Vulnerability represents a unified security vulnerability
type Vulnerability struct {
    // Identification
    ID          string    `json:"id"`           // Unique identifier (BTSG-XXX)
    ExternalID  string    `json:"external_id"`  // Scanner's original ID
    
    // Classification
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Severity    Severity  `json:"severity"`     // CRITICAL, HIGH, MEDIUM, LOW, INFO
    Type        VulnType  `json:"type"`         // secrets, dependencies, code, config
    Category    string    `json:"category"`     // More specific category
    
    // Location
    File        string    `json:"file"`
    Line        int       `json:"line"`
    LineEnd     int       `json:"line_end,omitempty"`
    Column      int       `json:"column"`
    ColumnEnd   int       `json:"column_end,omitempty"`
    CodeSnippet string    `json:"code_snippet,omitempty"`
    
    // Security Information
    CVE         []string  `json:"cve,omitempty"`
    CWE         []string  `json:"cwe,omitempty"`
    CVSS        float64   `json:"cvss,omitempty"`
    CVSSVector  string    `json:"cvss_vector,omitempty"`
    
    // Remediation
    Remediation string    `json:"remediation"`
    FixAvailable bool     `json:"fix_available"`
    FixedVersion string   `json:"fixed_version,omitempty"`
    
    // References
    References  []Reference `json:"references,omitempty"`
    
    // Metadata
    Scanner     string    `json:"scanner"`      // Which scanner found this
    Confidence  string    `json:"confidence"`   // HIGH, MEDIUM, LOW
    DetectedAt  time.Time `json:"detected_at"`
    
    // Additional Context
    Context     map[string]interface{} `json:"context,omitempty"`
}

// Reference represents an external reference
type Reference struct {
    Type string `json:"type"` // CVE, CWE, OWASP, Documentation, etc.
    URL  string `json:"url"`
}
```

### Scan Results Schema

```go
// ScanResults holds the complete scan results
type ScanResults struct {
    // Metadata
    ScanID      string    `json:"scan_id"`
    Path        string    `json:"path"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Duration    time.Duration `json:"duration"`
    
    // Statistics
    Stats       ScanStatistics `json:"stats"`
    
    // Results
    Vulnerabilities []Vulnerability `json:"vulnerabilities"`
    
    // Scanner Information
    Scanners    []ScannerResult `json:"scanners"`
    
    // Errors
    Errors      []ScanError `json:"errors,omitempty"`
}

// ScanStatistics provides summary statistics
type ScanStatistics struct {
    FilesScanned    int            `json:"files_scanned"`
    TotalVulns      int            `json:"total_vulnerabilities"`
    BySeverity      map[Severity]int `json:"by_severity"`
    ByType          map[VulnType]int `json:"by_type"`
    ByScanner       map[string]int   `json:"by_scanner"`
}

// ScannerResult holds results from a specific scanner
type ScannerResult struct {
    Name         string        `json:"name"`
    Version      string        `json:"version"`
    Duration     time.Duration `json:"duration"`
    VulnsFound   int           `json:"vulnerabilities_found"`
    Success      bool          `json:"success"`
    Error        string        `json:"error,omitempty"`
}

// ScanError represents an error during scanning
type ScanError struct {
    Scanner   string    `json:"scanner"`
    Message   string    `json:"message"`
    File      string    `json:"file,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

## 3. Scanner Execution Flow

### Orchestration Flow

```
1. Initialize
   ├─ Load scanner registry
   ├─ Validate configuration
   └─ Check scanner availability

2. Pre-Scan
   ├─ Discover files to scan
   ├─ Filter by scanner capabilities
   └─ Create scan plan

3. Execute Scanners (Concurrent)
   ├─ Bandit Scanner
   │  ├─ Check availability
   │  ├─ Build command
   │  ├─ Execute
   │  └─ Capture output
   │
   ├─ pip-audit Scanner
   │  ├─ Check availability
   │  ├─ Build command
   │  ├─ Execute
   │  └─ Capture output
   │
   └─ detect-secrets Scanner
      ├─ Check availability
      ├─ Build command
      ├─ Execute
      └─ Capture output

4. Normalize Results
   ├─ Parse raw output
   ├─ Map to unified schema
   ├─ Enrich with metadata
   └─ Deduplicate findings

5. Post-Process
   ├─ Calculate statistics
   ├─ Sort by severity
   ├─ Generate scan ID
   └─ Create final report

6. Return Results
   └─ Unified JSON output
```

### Detailed Execution Steps

```go
// Orchestrator manages the scanning process
type Orchestrator struct {
    registry  *ScannerRegistry
    config    OrchestratorConfig
    logger    Logger
}

// Execute runs all configured scanners
func (o *Orchestrator) Execute(ctx context.Context, scanConfig ScanConfig) (*ScanResults, error) {
    // 1. Initialize
    scanID := generateScanID()
    startTime := time.Now()
    
    // 2. Get applicable scanners
    scanners := o.registry.GetScannersForPath(scanConfig.Path, scanConfig.Types)
    
    // 3. Execute scanners concurrently
    results := make(chan *ScannerResult, len(scanners))
    errors := make(chan error, len(scanners))
    
    var wg sync.WaitGroup
    for _, scanner := range scanners {
        wg.Add(1)
        go func(s ScannerEngine) {
            defer wg.Done()
            result := o.executeSingleScanner(ctx, s, scanConfig)
            results <- result
        }(scanner)
    }
    
    // Wait for completion
    go func() {
        wg.Wait()
        close(results)
        close(errors)
    }()
    
    // 4. Collect and normalize results
    allVulns := []Vulnerability{}
    scannerResults := []ScannerResult{}
    
    for result := range results {
        scannerResults = append(scannerResults, *result)
        if result.Success {
            vulns, err := result.Scanner.Normalize(result.RawResult)
            if err == nil {
                allVulns = append(allVulns, vulns...)
            }
        }
    }
    
    // 5. Deduplicate and enrich
    allVulns = o.deduplicateVulnerabilities(allVulns)
    allVulns = o.enrichVulnerabilities(allVulns)
    
    // 6. Build final results
    return &ScanResults{
        ScanID:          scanID,
        Path:            scanConfig.Path,
        StartTime:       startTime,
        EndTime:         time.Now(),
        Duration:        time.Since(startTime),
        Vulnerabilities: allVulns,
        Scanners:        scannerResults,
        Stats:           o.calculateStatistics(allVulns, scannerResults),
    }, nil
}
```

## 4. Scanner Registry System

```go
// ScannerRegistry manages available scanners
type ScannerRegistry struct {
    scanners map[string]ScannerEngine
    mu       sync.RWMutex
}

// Register adds a scanner to the registry
func (r *ScannerRegistry) Register(scanner ScannerEngine) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    name := scanner.Name()
    if _, exists := r.scanners[name]; exists {
        return fmt.Errorf("scanner %s already registered", name)
    }
    
    r.scanners[name] = scanner
    return nil
}

// GetScannersForPath returns scanners applicable for the given path
func (r *ScannerRegistry) GetScannersForPath(path string, types []VulnType) []ScannerEngine {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var applicable []ScannerEngine
    for _, scanner := range r.scanners {
        if scanner.IsAvailable() && r.isApplicable(scanner, path, types) {
            applicable = append(applicable, scanner)
        }
    }
    return applicable
}

// ListAvailable returns all available scanners
func (r *ScannerRegistry) ListAvailable() []ScannerMetadata {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var metadata []ScannerMetadata
    for _, scanner := range r.scanners {
        if scanner.IsAvailable() {
            metadata = append(metadata, scanner.GetMetadata())
        }
    }
    return metadata
}
```

## 5. Output Normalization Layer

### Bandit Normalizer

```go
// BanditNormalizer converts Bandit output to unified format
type BanditNormalizer struct{}

func (n *BanditNormalizer) Normalize(raw *RawScanResult) ([]Vulnerability, error) {
    var banditOutput BanditOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &banditOutput); err != nil {
        return nil, err
    }
    
    var vulns []Vulnerability
    for _, result := range banditOutput.Results {
        vuln := Vulnerability{
            ID:          generateBTSGID(),
            ExternalID:  result.TestID,
            Title:       result.TestName,
            Description: result.IssueText,
            Severity:    mapBanditSeverity(result.IssueSeverity),
            Type:        VulnTypeCode,
            Category:    result.TestID,
            File:        result.Filename,
            Line:        result.LineNumber,
            LineEnd:     result.LineRange[1],
            CodeSnippet: result.Code,
            CWE:         []string{result.CWE},
            Confidence:  result.IssueConfidence,
            Scanner:     "bandit",
            DetectedAt:  time.Now(),
            Remediation: generateRemediation(result.TestID),
        }
        vulns = append(vulns, vuln)
    }
    
    return vulns, nil
}

// BanditOutput represents Bandit's JSON output structure
type BanditOutput struct {
    Results []struct {
        TestID          string   `json:"test_id"`
        TestName        string   `json:"test_name"`
        IssueText       string   `json:"issue_text"`
        IssueSeverity   string   `json:"issue_severity"`
        IssueConfidence string   `json:"issue_confidence"`
        Filename        string   `json:"filename"`
        LineNumber      int      `json:"line_number"`
        LineRange       []int    `json:"line_range"`
        Code            string   `json:"code"`
        CWE             string   `json:"cwe"`
    } `json:"results"`
}
```

### pip-audit Normalizer

```go
// PipAuditNormalizer converts pip-audit output to unified format
type PipAuditNormalizer struct{}

func (n *PipAuditNormalizer) Normalize(raw *RawScanResult) ([]Vulnerability, error) {
    var pipOutput PipAuditOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &pipOutput); err != nil {
        return nil, err
    }
    
    var vulns []Vulnerability
    for _, vuln := range pipOutput.Vulnerabilities {
        v := Vulnerability{
            ID:           generateBTSGID(),
            ExternalID:   vuln.ID,
            Title:        fmt.Sprintf("Vulnerable dependency: %s", vuln.Package),
            Description:  vuln.Description,
            Severity:     mapCVSSSeverity(vuln.CVSS),
            Type:         VulnTypeDependencies,
            File:         "requirements.txt", // or detect from scan
            CVE:          []string{vuln.ID},
            CVSS:         vuln.CVSS,
            CVSSVector:   vuln.CVSSVector,
            FixedVersion: vuln.FixedVersion,
            FixAvailable: vuln.FixedVersion != "",
            Scanner:      "pip-audit",
            DetectedAt:   time.Now(),
            References:   mapReferences(vuln.References),
            Context: map[string]interface{}{
                "package":         vuln.Package,
                "installed_version": vuln.Version,
                "fixed_version":   vuln.FixedVersion,
            },
        }
        vulns = append(vulns, v)
    }
    
    return vulns, nil
}

// PipAuditOutput represents pip-audit's JSON output
type PipAuditOutput struct {
    Vulnerabilities []struct {
        ID           string   `json:"id"`
        Package      string   `json:"package"`
        Version      string   `json:"version"`
        Description  string   `json:"description"`
        CVSS         float64  `json:"cvss"`
        CVSSVector   string   `json:"cvss_vector"`
        FixedVersion string   `json:"fixed_version"`
        References   []string `json:"references"`
    } `json:"vulnerabilities"`
}
```

### detect-secrets Normalizer

```go
// DetectSecretsNormalizer converts detect-secrets output to unified format
type DetectSecretsNormalizer struct{}

func (n *DetectSecretsNormalizer) Normalize(raw *RawScanResult) ([]Vulnerability, error) {
    var secretsOutput DetectSecretsOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &secretsOutput); err != nil {
        return nil, err
    }
    
    var vulns []Vulnerability
    for filename, secrets := range secretsOutput.Results {
        for _, secret := range secrets {
            vuln := Vulnerability{
                ID:          generateBTSGID(),
                ExternalID:  secret.HashedSecret,
                Title:       fmt.Sprintf("Secret detected: %s", secret.Type),
                Description: fmt.Sprintf("Potential %s found in source code", secret.Type),
                Severity:    SeverityHigh,
                Type:        VulnTypeSecrets,
                Category:    secret.Type,
                File:        filename,
                Line:        secret.LineNumber,
                Scanner:     "detect-secrets",
                DetectedAt:  time.Now(),
                Remediation: "Remove the secret from source code and use environment variables or a secrets manager",
                Context: map[string]interface{}{
                    "secret_type": secret.Type,
                    "is_verified": secret.IsVerified,
                },
            }
            vulns = append(vulns, vuln)
        }
    }
    
    return vulns, nil
}

// DetectSecretsOutput represents detect-secrets JSON output
type DetectSecretsOutput struct {
    Results map[string][]struct {
        Type         string `json:"type"`
        LineNumber   int    `json:"line_number"`
        HashedSecret string `json:"hashed_secret"`
        IsVerified   bool   `json:"is_verified"`
    } `json:"results"`
}
```

## 6. Error Handling and Retry Logic

```go
// ExecutionPolicy defines retry and timeout policies
type ExecutionPolicy struct {
    MaxRetries      int
    RetryDelay      time.Duration
    Timeout         time.Duration
    FailFast        bool
    ContinueOnError bool
}

// executeSingleScanner runs a scanner with retry logic
func (o *Orchestrator) executeSingleScanner(
    ctx context.Context,
    scanner ScannerEngine,
    config ScanConfig,
) *ScannerResult {
    result := &ScannerResult{
        Name:    scanner.Name(),
        Version: scanner.Version(),
    }
    
    startTime := time.Now()
    defer func() {
        result.Duration = time.Since(startTime)
    }()
    
    // Apply timeout
    ctx, cancel := context.WithTimeout(ctx, o.config.Policy.Timeout)
    defer cancel()
    
    // Retry logic
    var lastErr error
    for attempt := 0; attempt <= o.config.Policy.MaxRetries; attempt++ {
        if attempt > 0 {
            time.Sleep(o.config.Policy.RetryDelay)
        }
        
        rawResult, err := scanner.Scan(ctx, config)
        if err == nil {
            result.Success = true
            result.RawResult = rawResult
            result.VulnsFound = len(rawResult.Vulnerabilities)
            return result
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryable(err) {
            break
        }
    }
    
    result.Success = false
    result.Error = lastErr.Error()
    return result
}
```

## 7. Configuration System

```go
// OrchestratorConfig configures the scanner orchestrator
type OrchestratorConfig struct {
    // Execution
    Policy          ExecutionPolicy
    MaxConcurrent   int
    
    // Scanners
    EnabledScanners []string
    ScannerConfigs  map[string]interface{}
    
    // Output
    Verbose         bool
    IncludeContext  bool
    DeduplicateResults bool
    
    // Filtering
    MinSeverity     Severity
    ExcludePatterns []string
}

// ScanConfig holds configuration for a specific scan
type ScanConfig struct {
    Path            string
    Recursive       bool
    Types           []VulnType
    Exclude         []string
    IncludeHidden   bool
    FollowSymlinks  bool
    MaxFileSize     int64
    
    // Scanner-specific options
    ScannerOptions  map[string]interface{}
}
```

## 8. Usage Examples

### Basic Usage

```go
// Initialize registry
registry := scanner.NewRegistry()

// Register scanners
registry.Register(scanner.NewBanditScanner())
registry.Register(scanner.NewPipAuditScanner())
registry.Register(scanner.NewDetectSecretsScanner())

// Create orchestrator
orchestrator := scanner.NewOrchestrator(registry, scanner.OrchestratorConfig{
    MaxConcurrent: 3,
    Policy: scanner.ExecutionPolicy{
        MaxRetries: 2,
        Timeout:    5 * time.Minute,
    },
})

// Execute scan
results, err := orchestrator.Execute(context.Background(), scanner.ScanConfig{
    Path:      "./myproject",
    Recursive: true,
    Types:     []types.VulnType{types.VulnTypeCode, types.VulnTypeDependencies, types.VulnTypeSecrets},
})

// Output as JSON
jsonOutput, _ := json.MarshalIndent(results, "", "  ")
fmt.Println(string(jsonOutput))
```

### Advanced Usage with Custom Scanner

```go
// Implement custom scanner
type CustomScanner struct {
    name    string
    version string
}

func (s *CustomScanner) Name() string { return s.name }
func (s *CustomScanner) Version() string { return s.version }
func (s *CustomScanner) Type() types.VulnType { return types.VulnTypeCode }
func (s *CustomScanner) IsAvailable() bool { return true }

func (s *CustomScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.RawScanResult, error) {
    // Custom scanning logic
    return &scanner.RawScanResult{
        Scanner:  s.name,
        ExitCode: 0,
        Stdout:   `{"findings": [...]}`,
    }, nil
}

func (s *CustomScanner) Normalize(raw *scanner.RawScanResult) ([]types.Vulnerability, error) {
    // Custom normalization logic
    return []types.Vulnerability{}, nil
}

// Register and use
registry.Register(&CustomScanner{name: "custom", version: "1.0.0"})
```

## 9. Output Example

```json
{
  "scan_id": "scan_20260502_074950_abc123",
  "path": "./myproject",
  "start_time": "2026-05-02T07:49:50Z",
  "end_time": "2026-05-02T07:50:15Z",
  "duration": "25s",
  "stats": {
    "files_scanned": 42,
    "total_vulnerabilities": 8,
    "by_severity": {
      "CRITICAL": 1,
      "HIGH": 3,
      "MEDIUM": 2,
      "LOW": 2
    },
    "by_type": {
      "code": 3,
      "dependencies": 3,
      "secrets": 2
    },
    "by_scanner": {
      "bandit": 3,
      "pip-audit": 3,
      "detect-secrets": 2
    }
  },
  "vulnerabilities": [
    {
      "id": "BTSG-001",
      "external_id": "B201",
      "title": "Use of insecure pickle module",
      "description": "Pickle library appears to be in use, possible security issue",
      "severity": "HIGH",
      "type": "code",
      "category": "B201",
      "file": "app/utils.py",
      "line": 45,
      "line_end": 47,
      "code_snippet": "import pickle\ndata = pickle.loads(user_input)",
      "cwe": ["CWE-502"],
      "confidence": "HIGH",
      "scanner": "bandit",
      "detected_at": "2026-05-02T07:49:55Z",
      "remediation": "Use safer serialization formats like JSON"
    }
  ],
  "scanners": [
    {
      "name": "bandit",
      "version": "1.7.5",
      "duration": "8s",
      "vulnerabilities_found": 3,
      "success": true
    },
    {
      "name": "pip-audit",
      "version": "2.6.1",
      "duration": "12s",
      "vulnerabilities_found": 3,
      "success": true
    },
    {
      "name": "detect-secrets",
      "version": "1.4.0",
      "duration": "5s",
      "vulnerabilities_found": 2,
      "success": true
    }
  ]
}
```

## 10. Implementation Roadmap

1. **Phase 1: Core Infrastructure**
   - Implement scanner interface
   - Create registry system
   - Build orchestrator

2. **Phase 2: Scanner Implementations**
   - Bandit scanner
   - pip-audit scanner
   - detect-secrets scanner

3. **Phase 3: Normalization**
   - Output parsers
   - Schema mapping
   - Deduplication logic

4. **Phase 4: Advanced Features**
   - Concurrent execution
   - Retry logic
   - Error handling

5. **Phase 5: Integration**
   - CLI integration
   - Configuration system
   - Testing and validation

## Conclusion

This design provides a robust, extensible scanning engine that can integrate multiple security scanners while maintaining a unified output format. The modular architecture allows for easy addition of new scanners and customization of the scanning process.