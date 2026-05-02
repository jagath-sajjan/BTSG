# BTSG Architecture Documentation

## Overview

BTSG (Bob The Security Guy) is a modular, production-ready CLI security tool built with Go and Cobra. This document describes the architecture, design decisions, and data flow.

## Design Principles

1. **Modularity** - Each component has a single responsibility
2. **Extensibility** - Easy to add new scanners, fixers, and report formats
3. **Testability** - Clear interfaces and dependency injection
4. **Performance** - Concurrent scanning and efficient file processing
5. **User Experience** - Clear output, interactive modes, and helpful error messages

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                     │
│  ┌──────┐  ┌─────────┐  ┌──────┐  ┌────────┐              │
│  │ scan │  │ explain │  │ fix  │  │ report │              │
│  └──┬───┘  └────┬────┘  └───┬──┘  └───┬────┘              │
└─────┼──────────┼────────────┼─────────┼────────────────────┘
      │          │            │         │
      ▼          ▼            ▼         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Business Logic Layer                     │
│  ┌─────────┐  ┌──────────┐  ┌───────┐  ┌──────────┐       │
│  │ Scanner │  │ Analyzer │  │ Fixer │  │ Reporter │       │
│  └────┬────┘  └─────┬────┘  └───┬───┘  └─────┬────┘       │
└───────┼─────────────┼───────────┼────────────┼─────────────┘
        │             │           │            │
        ▼             ▼           ▼            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Shared Layer                            │
│  ┌───────┐  ┌───────┐  ┌──────────────┐                    │
│  │ Types │  │ Utils │  │ File System  │                    │
│  └───────┘  └───────┘  └──────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
btsg/
├── cmd/                          # CLI commands (Cobra)
│   ├── root.go                   # Root command, global flags
│   ├── scan.go                   # Scan command implementation
│   ├── explain.go                # Explain command implementation
│   ├── fix.go                    # Fix command implementation
│   └── report.go                 # Report command implementation
│
├── internal/                     # Internal packages (not importable)
│   ├── scanner/                  # Vulnerability scanning
│   │   ├── scanner.go           # Main scanner logic
│   │   ├── secrets.go           # Secret detection (future)
│   │   ├── dependencies.go      # Dependency scanning (future)
│   │   ├── code.go              # Code analysis (future)
│   │   └── config.go            # Config scanning (future)
│   │
│   ├── analyzer/                 # AI-powered analysis
│   │   ├── analyzer.go          # Main analyzer logic
│   │   ├── ai_client.go         # AI service integration (future)
│   │   └── cache.go             # Explanation caching (future)
│   │
│   ├── fixer/                    # Automated fixing
│   │   ├── fixer.go             # Main fixer logic
│   │   ├── strategies.go        # Fix strategies (future)
│   │   └── backup.go            # Backup management (future)
│   │
│   └── reporter/                 # Report generation
│       ├── reporter.go          # Main reporter logic
│       ├── json.go              # JSON formatter (future)
│       ├── html.go              # HTML formatter (future)
│       ├── pdf.go               # PDF formatter (future)
│       └── sarif.go             # SARIF formatter (future)
│
├── pkg/                          # Public packages (importable)
│   ├── types/                    # Shared types and interfaces
│   │   └── types.go             # Common types
│   │
│   └── utils/                    # Utility functions
│       └── utils.go             # File system, formatting utils
│
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── README.md                     # User documentation
└── ARCHITECTURE.md               # This file
```

## Component Details

### 1. CLI Layer (cmd/)

**Responsibility:** Handle user input, parse flags, orchestrate business logic

**Key Files:**
- `root.go` - Root command, global flags (verbose, output)
- `scan.go` - Scan command with path, recursive, types flags
- `explain.go` - Explain command with detailed, CVE flags
- `fix.go` - Fix command with interactive, dry-run, all flags
- `report.go` - Report command with format, output, input flags

**Design Pattern:** Command Pattern (Cobra framework)

### 2. Scanner Module (internal/scanner/)

**Responsibility:** Detect security vulnerabilities in code

**Key Components:**
- `Scanner` struct - Main scanning orchestrator
- `ScanResults` - Results container
- `Vulnerability` - Vulnerability representation

**Scanning Types:**
1. **Secrets** - API keys, tokens, passwords
2. **Dependencies** - CVEs in packages
3. **Code** - SAST issues (SQL injection, XSS, etc.)
4. **Config** - Misconfigurations (Docker, K8s, Terraform)

**Future Integrations:**
- Trivy (container/filesystem scanning)
- Semgrep (SAST)
- OSV-Scanner (dependency vulnerabilities)
- Custom regex patterns

### 3. Analyzer Module (internal/analyzer/)

**Responsibility:** Provide AI-powered explanations

**Key Components:**
- `Analyzer` struct - Main analysis orchestrator
- `Explanation` - Detailed vulnerability explanation

**Features:**
- Plain language explanations
- Impact analysis
- Exploitation scenarios
- Remediation advice
- Reference links

**Future Integrations:**
- OpenAI GPT-4
- Anthropic Claude
- Local LLMs (Ollama)
- Explanation caching

### 4. Fixer Module (internal/fixer/)

**Responsibility:** Automatically remediate vulnerabilities

**Key Components:**
- `Fixer` struct - Main fixing orchestrator
- `FixResults` - Fix operation results
- `FixedVuln`, `SkippedVuln`, `FailedVuln` - Result types

**Features:**
- Interactive approval workflow
- Dry-run mode
- Automatic backups
- Rollback capability

**Fix Strategies:**
1. **Secrets** - Move to env vars, add to .gitignore
2. **Dependencies** - Update to secure versions
3. **Code** - Apply secure coding patterns
4. **Config** - Apply security hardening

### 5. Reporter Module (internal/reporter/)

**Responsibility:** Generate security reports

**Key Components:**
- `Reporter` struct - Main reporting orchestrator
- `ReportResult` - Generated report metadata
- `ReportData` - Complete report data structure

**Supported Formats:**
1. **JSON** - Machine-readable, CI/CD integration
2. **HTML** - Interactive web report with charts
3. **PDF** - Professional stakeholder reports
4. **Markdown** - Documentation-friendly
5. **SARIF** - Standard security tool format

### 6. Shared Types (pkg/types/)

**Responsibility:** Common types and interfaces

**Key Types:**
- `Vulnerability` - Core vulnerability type
- `Severity` - Severity levels enum
- `VulnType` - Vulnerability types enum
- Configuration structs

### 7. Utilities (pkg/utils/)

**Responsibility:** Common utility functions

**Functions:**
- File system operations
- Path manipulation
- Format conversion
- Exclusion patterns

## Data Flow

### Scan Flow

```
User Input (btsg scan .)
    ↓
Parse CLI flags
    ↓
Initialize Scanner with Config
    ↓
Walk directory tree
    ↓
For each file:
    - Check exclusions
    - Detect file type
    - Run applicable scanners
    - Collect vulnerabilities
    ↓
Aggregate results
    ↓
Display to user
```

### Explain Flow

```
User Input (btsg explain CVE-2024-1234)
    ↓
Parse vulnerability ID
    ↓
Initialize Analyzer
    ↓
Fetch vulnerability details
    ↓
Generate AI explanation
    ↓
Format and display
```

### Fix Flow

```
User Input (btsg fix . --interactive)
    ↓
Run scan to find vulnerabilities
    ↓
For each vulnerability:
    - Determine fix strategy
    - Generate fix code
    - If interactive: prompt user
    - If approved: create backup
    - Apply fix
    - Verify fix
    ↓
Display results
```

### Report Flow

```
User Input (btsg report --format html)
    ↓
Load scan results (or run new scan)
    ↓
Initialize Reporter with format
    ↓
Generate report content
    ↓
Apply formatting (HTML/PDF/JSON/etc.)
    ↓
Save to file or display
```

## Configuration

### Global Flags
- `--verbose, -v` - Enable verbose output
- `--output, -o` - Output file path

### Environment Variables (Future)
- `BTSG_AI_PROVIDER` - AI service provider
- `BTSG_AI_API_KEY` - AI service API key
- `BTSG_CONFIG_PATH` - Custom config file path

### Config File (Future)
```yaml
# .btsg.yml
scanner:
  exclude:
    - node_modules
    - vendor
  types:
    - secrets
    - dependencies
    - code
    - config

analyzer:
  provider: openai
  model: gpt-4

fixer:
  auto_backup: true
  backup_dir: .btsg-backups

reporter:
  default_format: html
  include_code: true
```

## Error Handling

1. **Graceful Degradation** - Continue scanning even if one file fails
2. **Clear Error Messages** - User-friendly error descriptions
3. **Exit Codes** - Standard exit codes for CI/CD integration
4. **Logging** - Structured logging for debugging

## Performance Considerations

1. **Concurrent Scanning** - Scan multiple files in parallel
2. **Streaming Results** - Display results as they're found
3. **Caching** - Cache AI explanations and scan results
4. **Incremental Scanning** - Only scan changed files (future)

## Security Considerations

1. **No Data Leakage** - Don't send sensitive code to AI without consent
2. **Secure Backups** - Encrypt backup files
3. **API Key Management** - Secure storage of API keys
4. **Audit Logging** - Log all fix operations

## Testing Strategy

1. **Unit Tests** - Test individual functions
2. **Integration Tests** - Test component interactions
3. **E2E Tests** - Test complete workflows
4. **Benchmark Tests** - Performance testing

## Future Enhancements

1. **Watch Mode** - Continuous monitoring
2. **Web Dashboard** - Visual interface
3. **Plugin System** - Custom scanners and fixers
4. **CI/CD Integration** - GitHub Actions, GitLab CI
5. **Team Features** - Shared configurations, reports
6. **Machine Learning** - Learn from fix approvals
7. **Compliance Reports** - SOC2, PCI-DSS, etc.

## Contributing

See the main README.md for contribution guidelines.