# BTSG Scanner Implementation Guide

## Quick Start: Implementing a New Scanner

This guide provides step-by-step instructions for implementing a new scanner in BTSG.

## Table of Contents

1. [Scanner Interface](#scanner-interface)
2. [Implementation Steps](#implementation-steps)
3. [Bandit Scanner Example](#bandit-scanner-example)
4. [pip-audit Scanner Example](#pip-audit-scanner-example)
5. [detect-secrets Scanner Example](#detect-secrets-scanner-example)
6. [Testing Your Scanner](#testing-your-scanner)
7. [Best Practices](#best-practices)

## Scanner Interface

Every scanner must implement the `ScannerEngine` interface:

```go
type ScannerEngine interface {
    Name() string
    Version() string
    Type() VulnType
    IsAvailable() bool
    Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error)
    Normalize(raw *RawScanResult) ([]Vulnerability, error)
    GetCommand(config ScanConfig) []string
    GetMetadata() ScannerMetadata
}
```

## Implementation Steps

### Step 1: Create Scanner Struct

```go
package scanner

type BanditScanner struct {
    version string
    path    string
}

func NewBanditScanner() *BanditScanner {
    return &BanditScanner{
        version: "1.7.5",
        path:    "bandit",
    }
}
```

### Step 2: Implement Basic Methods

```go
func (s *BanditScanner) Name() string {
    return "bandit"
}

func (s *BanditScanner) Version() string {
    return s.version
}

func (s *BanditScanner) Type() VulnType {
    return VulnTypeCode
}

func (s *BanditScanner) GetMetadata() ScannerMetadata {
    return ScannerMetadata{
        Name:           s.Name(),
        Version:        s.Version(),
        Type:           s.Type(),
        Description:    "Python security linter",
        SupportedFiles: []string{"*.py"},
        RequiredTools:  []string{"bandit"},
    }
}
```

### Step 3: Implement IsAvailable

```go
func (s *BanditScanner) IsAvailable() bool {
    cmd := exec.Command(s.path, "--version")
    err := cmd.Run()
    return err == nil
}
```

### Step 4: Implement GetCommand

```go
func (s *BanditScanner) GetCommand(config ScanConfig) []string {
    args := []string{
        "-f", "json",  // JSON output
        "-r",          // Recursive
    }
    
    if config.Verbose {
        args = append(args, "-v")
    }
    
    args = append(args, config.Path)
    return append([]string{s.path}, args...)
}
```

### Step 5: Implement Scan

```go
func (s *BanditScanner) Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error) {
    startTime := time.Now()
    
    // Build command
    cmdArgs := s.GetCommand(config)
    cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
    
    // Capture output
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // Execute
    err := cmd.Run()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        }
    }
    
    return &RawScanResult{
        Scanner:  s.Name(),
        ExitCode: exitCode,
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        Duration: time.Since(startTime),
        Error:    err,
        Metadata: s.GetMetadata(),
    }, nil
}
```

### Step 6: Implement Normalize

```go
func (s *BanditScanner) Normalize(raw *RawScanResult) ([]Vulnerability, error) {
    // Parse JSON output
    var banditOutput struct {
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
            CWE             struct {
                ID   int    `json:"id"`
                Link string `json:"link"`
            } `json:"cwe"`
        } `json:"results"`
    }
    
    if err := json.Unmarshal([]byte(raw.Stdout), &banditOutput); err != nil {
        return nil, fmt.Errorf("failed to parse bandit output: %w", err)
    }
    
    // Convert to unified format
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
            CWE:         []string{fmt.Sprintf("CWE-%d", result.CWE.ID)},
            Confidence:  result.IssueConfidence,
            Scanner:     s.Name(),
            DetectedAt:  time.Now(),
            References: []Reference{
                {Type: "CWE", URL: result.CWE.Link},
            },
        }
        vulns = append(vulns, vuln)
    }
    
    return vulns, nil
}

// Helper function to map severity
func mapBanditSeverity(severity string) Severity {
    switch strings.ToUpper(severity) {
    case "HIGH":
        return SeverityHigh
    case "MEDIUM":
        return SeverityMedium
    case "LOW":
        return SeverityLow
    default:
        return SeverityInfo
    }
}
```

## Bandit Scanner Example

Complete implementation:

```go
// internal/scanner/engines/bandit.go
package engines

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
    
    "btsg/pkg/types"
)

type BanditScanner struct {
    version string
    path    string
}

func NewBanditScanner() *BanditScanner {
    return &BanditScanner{
        version: detectBanditVersion(),
        path:    "bandit",
    }
}

func (s *BanditScanner) Name() string { return "bandit" }
func (s *BanditScanner) Version() string { return s.version }
func (s *BanditScanner) Type() types.VulnType { return types.VulnTypeCode }

func (s *BanditScanner) IsAvailable() bool {
    cmd := exec.Command(s.path, "--version")
    return cmd.Run() == nil
}

func (s *BanditScanner) GetCommand(config types.ScanConfig) []string {
    return []string{
        s.path,
        "-f", "json",
        "-r",
        config.Path,
    }
}

func (s *BanditScanner) GetMetadata() types.ScannerMetadata {
    return types.ScannerMetadata{
        Name:           "bandit",
        Version:        s.version,
        Type:           types.VulnTypeCode,
        Description:    "Python security linter from PyCQA",
        SupportedFiles: []string{"*.py"},
        RequiredTools:  []string{"bandit"},
        ConfigOptions: map[string]string{
            "format":    "json",
            "recursive": "true",
        },
    }
}

func (s *BanditScanner) Scan(ctx context.Context, config types.ScanConfig) (*types.RawScanResult, error) {
    startTime := time.Now()
    
    cmdArgs := s.GetCommand(config)
    cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err := cmd.Run()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        }
    }
    
    return &types.RawScanResult{
        Scanner:  s.Name(),
        ExitCode: exitCode,
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        Duration: time.Since(startTime),
        Error:    err,
        Metadata: s.GetMetadata(),
    }, nil
}

func (s *BanditScanner) Normalize(raw *types.RawScanResult) ([]types.Vulnerability, error) {
    var output banditOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &output); err != nil {
        return nil, err
    }
    
    var vulns []types.Vulnerability
    for _, result := range output.Results {
        vulns = append(vulns, types.Vulnerability{
            ID:          generateID(),
            ExternalID:  result.TestID,
            Title:       result.TestName,
            Description: result.IssueText,
            Severity:    mapSeverity(result.IssueSeverity),
            Type:        types.VulnTypeCode,
            File:        result.Filename,
            Line:        result.LineNumber,
            Scanner:     s.Name(),
            DetectedAt:  time.Now(),
        })
    }
    
    return vulns, nil
}

type banditOutput struct {
    Results []struct {
        TestID        string `json:"test_id"`
        TestName      string `json:"test_name"`
        IssueText     string `json:"issue_text"`
        IssueSeverity string `json:"issue_severity"`
        Filename      string `json:"filename"`
        LineNumber    int    `json:"line_number"`
    } `json:"results"`
}

func detectBanditVersion() string {
    cmd := exec.Command("bandit", "--version")
    output, err := cmd.Output()
    if err != nil {
        return "unknown"
    }
    // Parse version from output
    return string(output)
}

func generateID() string {
    return fmt.Sprintf("BTSG-%d", time.Now().UnixNano())
}

func mapSeverity(s string) types.Severity {
    switch s {
    case "HIGH":
        return types.SeverityHigh
    case "MEDIUM":
        return types.SeverityMedium
    case "LOW":
        return types.SeverityLow
    default:
        return types.SeverityInfo
    }
}
```

## pip-audit Scanner Example

```go
// internal/scanner/engines/pipaudit.go
package engines

type PipAuditScanner struct {
    version string
    path    string
}

func NewPipAuditScanner() *PipAuditScanner {
    return &PipAuditScanner{
        version: detectPipAuditVersion(),
        path:    "pip-audit",
    }
}

func (s *PipAuditScanner) Name() string { return "pip-audit" }
func (s *PipAuditScanner) Type() types.VulnType { return types.VulnTypeDependencies }

func (s *PipAuditScanner) GetCommand(config types.ScanConfig) []string {
    return []string{
        s.path,
        "--format", "json",
        "--desc", "on",
    }
}

func (s *PipAuditScanner) Normalize(raw *types.RawScanResult) ([]types.Vulnerability, error) {
    var output pipAuditOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &output); err != nil {
        return nil, err
    }
    
    var vulns []types.Vulnerability
    for _, vuln := range output.Vulnerabilities {
        vulns = append(vulns, types.Vulnerability{
            ID:           generateID(),
            ExternalID:   vuln.ID,
            Title:        fmt.Sprintf("Vulnerable dependency: %s", vuln.Package),
            Description:  vuln.Description,
            Severity:     mapCVSSSeverity(vuln.CVSS),
            Type:         types.VulnTypeDependencies,
            CVE:          []string{vuln.ID},
            CVSS:         vuln.CVSS,
            FixedVersion: vuln.FixedVersion,
            Scanner:      s.Name(),
            DetectedAt:   time.Now(),
            Context: map[string]interface{}{
                "package":          vuln.Package,
                "installed_version": vuln.Version,
                "fixed_version":    vuln.FixedVersion,
            },
        })
    }
    
    return vulns, nil
}

type pipAuditOutput struct {
    Vulnerabilities []struct {
        ID           string  `json:"id"`
        Package      string  `json:"package"`
        Version      string  `json:"version"`
        Description  string  `json:"description"`
        CVSS         float64 `json:"cvss"`
        FixedVersion string  `json:"fixed_version"`
    } `json:"vulnerabilities"`
}

func mapCVSSSeverity(cvss float64) types.Severity {
    switch {
    case cvss >= 9.0:
        return types.SeverityCritical
    case cvss >= 7.0:
        return types.SeverityHigh
    case cvss >= 4.0:
        return types.SeverityMedium
    default:
        return types.SeverityLow
    }
}
```

## detect-secrets Scanner Example

```go
// internal/scanner/engines/detectsecrets.go
package engines

type DetectSecretsScanner struct {
    version string
    path    string
}

func NewDetectSecretsScanner() *DetectSecretsScanner {
    return &DetectSecretsScanner{
        version: detectVersion(),
        path:    "detect-secrets",
    }
}

func (s *DetectSecretsScanner) Name() string { return "detect-secrets" }
func (s *DetectSecretsScanner) Type() types.VulnType { return types.VulnTypeSecrets }

func (s *DetectSecretsScanner) GetCommand(config types.ScanConfig) []string {
    return []string{
        s.path,
        "scan",
        "--all-files",
        config.Path,
    }
}

func (s *DetectSecretsScanner) Normalize(raw *types.RawScanResult) ([]types.Vulnerability, error) {
    var output detectSecretsOutput
    if err := json.Unmarshal([]byte(raw.Stdout), &output); err != nil {
        return nil, err
    }
    
    var vulns []types.Vulnerability
    for filename, secrets := range output.Results {
        for _, secret := range secrets {
            vulns = append(vulns, types.Vulnerability{
                ID:          generateID(),
                ExternalID:  secret.HashedSecret,
                Title:       fmt.Sprintf("Secret detected: %s", secret.Type),
                Description: fmt.Sprintf("Potential %s found in source code", secret.Type),
                Severity:    types.SeverityHigh,
                Type:        types.VulnTypeSecrets,
                Category:    secret.Type,
                File:        filename,
                Line:        secret.LineNumber,
                Scanner:     s.Name(),
                DetectedAt:  time.Now(),
                Remediation: "Remove secret and use environment variables",
            })
        }
    }
    
    return vulns, nil
}

type detectSecretsOutput struct {
    Results map[string][]struct {
        Type         string `json:"type"`
        LineNumber   int    `json:"line_number"`
        HashedSecret string `json:"hashed_secret"`
    } `json:"results"`
}
```

## Testing Your Scanner

### Unit Tests

```go
// internal/scanner/engines/bandit_test.go
package engines

import (
    "context"
    "testing"
    
    "btsg/pkg/types"
)

func TestBanditScanner_Name(t *testing.T) {
    scanner := NewBanditScanner()
    if scanner.Name() != "bandit" {
        t.Errorf("expected 'bandit', got '%s'", scanner.Name())
    }
}

func TestBanditScanner_IsAvailable(t *testing.T) {
    scanner := NewBanditScanner()
    // This will fail if bandit is not installed
    if !scanner.IsAvailable() {
        t.Skip("bandit not available")
    }
}

func TestBanditScanner_Normalize(t *testing.T) {
    scanner := NewBanditScanner()
    
    raw := &types.RawScanResult{
        Stdout: `{
            "results": [{
                "test_id": "B201",
                "test_name": "Test",
                "issue_text": "Issue",
                "issue_severity": "HIGH",
                "filename": "test.py",
                "line_number": 10
            }]
        }`,
    }
    
    vulns, err := scanner.Normalize(raw)
    if err != nil {
        t.Fatalf("normalize failed: %v", err)
    }
    
    if len(vulns) != 1 {
        t.Errorf("expected 1 vulnerability, got %d", len(vulns))
    }
}
```

### Integration Tests

```go
func TestBanditScanner_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    scanner := NewBanditScanner()
    if !scanner.IsAvailable() {
        t.Skip("bandit not available")
    }
    
    config := types.ScanConfig{
        Path: "./testdata/python",
    }
    
    result, err := scanner.Scan(context.Background(), config)
    if err != nil {
        t.Fatalf("scan failed: %v", err)
    }
    
    vulns, err := scanner.Normalize(result)
    if err != nil {
        t.Fatalf("normalize failed: %v", err)
    }
    
    t.Logf("found %d vulnerabilities", len(vulns))
}
```

## Best Practices

### 1. Error Handling

```go
func (s *Scanner) Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error) {
    // Always check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Validate input
    if config.Path == "" {
        return nil, fmt.Errorf("path cannot be empty")
    }
    
    // Handle command execution errors
    cmd := exec.CommandContext(ctx, s.path, args...)
    if err := cmd.Run(); err != nil {
        // Distinguish between different error types
        if exitErr, ok := err.(*exec.ExitError); ok {
            // Non-zero exit code (might be expected)
            return result, nil
        }
        // Actual error
        return nil, fmt.Errorf("command failed: %w", err)
    }
    
    return result, nil
}
```

### 2. Logging

```go
func (s *Scanner) Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error) {
    if config.Verbose {
        log.Printf("[%s] Starting scan of %s", s.Name(), config.Path)
    }
    
    // ... scan logic ...
    
    if config.Verbose {
        log.Printf("[%s] Scan completed in %s", s.Name(), duration)
    }
    
    return result, nil
}
```

### 3. Resource Cleanup

```go
func (s *Scanner) Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error) {
    // Create temporary files if needed
    tmpFile, err := os.CreateTemp("", "btsg-*")
    if err != nil {
        return nil, err
    }
    defer os.Remove(tmpFile.Name()) // Always cleanup
    defer tmpFile.Close()
    
    // ... use tmpFile ...
    
    return result, nil
}
```

### 4. Context Handling

```go
func (s *Scanner) Scan(ctx context.Context, config ScanConfig) (*RawScanResult, error) {
    // Use context for command execution
    cmd := exec.CommandContext(ctx, s.path, args...)
    
    // Monitor context in long-running operations
    done := make(chan error, 1)
    go func() {
        done <- cmd.Run()
    }()
    
    select {
    case <-ctx.Done():
        cmd.Process.Kill()
        return nil, ctx.Err()
    case err := <-done:
        return result, err
    }
}
```

### 5. Version Detection

```go
func detectVersion() string {
    cmd := exec.Command("tool", "--version")
    output, err := cmd.Output()
    if err != nil {
        return "unknown"
    }
    
    // Parse version from output
    // Example: "tool version 1.2.3"
    parts := strings.Fields(string(output))
    if len(parts) >= 3 {
        return parts[2]
    }
    
    return "unknown"
}
```

## Registration

Register your scanner in the orchestrator:

```go
// internal/scanner/orchestrator.go
func NewOrchestrator() *Orchestrator {
    registry := NewRegistry()
    
    // Register all scanners
    registry.Register(engines.NewBanditScanner())
    registry.Register(engines.NewPipAuditScanner())
    registry.Register(engines.NewDetectSecretsScanner())
    
    return &Orchestrator{
        registry: registry,
    }
}
```

## Conclusion

Follow these steps to implement any new scanner:
1. Create scanner struct
2. Implement interface methods
3. Handle errors properly
4. Write tests
5. Register scanner
6. Document usage

For more details, see the full design document in `SCANNER_ENGINE_DESIGN.md`.