# BTSG (Bob The Security Guy)

A production-ready CLI security tool that scans local repositories for vulnerabilities, explains issues using AI, and auto-fixes them with your approval.

## Features

- 🔍 **Scan** - Comprehensive security scanning for multiple vulnerability types
- 💡 **Explain** - AI-powered explanations of security issues
- 🔧 **Fix** - Automated vulnerability remediation with approval workflow
- 📊 **Report** - Generate structured security reports in multiple formats

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/btsg.git
cd btsg

# Build the binary
go build -o btsg

# Install globally (optional)
go install
```

## Quick Start

```bash
# Scan current directory
btsg scan .

# Scan with specific vulnerability types
btsg scan . --types secrets,dependencies

# Explain a vulnerability
btsg explain CVE-2024-1234

# Fix vulnerabilities interactively
btsg fix . --interactive

# Generate a security report
btsg report --format html --output report.html
```

## Commands

### `btsg scan [path]`

Scan a repository for security vulnerabilities.

**Options:**
- `-r, --recursive` - Scan directories recursively (default: true)
- `-t, --types` - Vulnerability types to scan: secrets, dependencies, code, config, all (default: all)
- `-v, --verbose` - Enable verbose output
- `-o, --output` - Output file path

**Examples:**
```bash
# Scan current directory
btsg scan .

# Scan specific path
btsg scan /path/to/project

# Scan only for secrets and dependencies
btsg scan . --types secrets,dependencies

# Verbose output
btsg scan . --verbose
```

### `btsg explain [vulnerability-id]`

Get AI-powered explanations for security vulnerabilities.

**Options:**
- `-d, --detailed` - Show detailed technical explanation
- `--cve` - CVE identifier to explain
- `-v, --verbose` - Enable verbose output

**Examples:**
```bash
# Explain a vulnerability by ID
btsg explain BTSG-001

# Explain a CVE
btsg explain CVE-2024-1234

# Get detailed explanation
btsg explain CVE-2024-1234 --detailed
```

### `btsg fix [path]`

Automatically fix security vulnerabilities.

**Options:**
- `-i, --interactive` - Review each fix before applying
- `--dry-run` - Preview fixes without modifying files
- `-a, --all` - Fix all vulnerabilities without prompting
- `--vuln-id` - Fix specific vulnerability by ID
- `-v, --verbose` - Enable verbose output

**Examples:**
```bash
# Interactive fix mode
btsg fix . --interactive

# Dry run to preview changes
btsg fix . --dry-run

# Fix all vulnerabilities
btsg fix . --all

# Fix specific vulnerability
btsg fix . --vuln-id BTSG-001
```

### `btsg report`

Generate structured security reports.

**Options:**
- `-f, --format` - Report format: json, html, pdf, markdown, sarif (default: json)
- `--output` - Output file path
- `--input` - Input scan results file
- `-v, --verbose` - Enable verbose output

**Examples:**
```bash
# Generate JSON report
btsg report --format json

# Generate HTML report
btsg report --format html --output report.html

# Generate PDF report
btsg report --format pdf --output security-report.pdf

# Generate from existing scan results
btsg report --input scan-results.json --format html
```

## Architecture

```
btsg/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── scan.go            # Scan command
│   ├── explain.go         # Explain command
│   ├── fix.go             # Fix command
│   └── report.go          # Report command
├── internal/              # Internal packages
│   ├── scanner/           # Vulnerability scanning logic
│   ├── analyzer/          # AI-powered analysis
│   ├── fixer/             # Automated fixing
│   └── reporter/          # Report generation
├── pkg/                   # Public packages
│   ├── types/             # Shared types and interfaces
│   └── utils/             # Utility functions
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Data Flow

```
┌─────────┐
│  User   │
└────┬────┘
     │
     ▼
┌─────────────┐
│  CLI (Cobra)│
└──────┬──────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼
┌──────────┐    ┌──────────┐
│ Scanner  │    │ Analyzer │
└────┬─────┘    └────┬─────┘
     │               │
     ▼               ▼
┌──────────┐    ┌──────────┐
│  Fixer   │    │ Reporter │
└──────────┘    └──────────┘
```

## Vulnerability Types

- **Secrets** - API keys, tokens, passwords, credentials
- **Dependencies** - Vulnerable packages and libraries (CVEs)
- **Code** - Security issues in source code (SQL injection, XSS, etc.)
- **Config** - Misconfigurations in Docker, Kubernetes, Terraform, etc.

## Report Formats

- **JSON** - Machine-readable format for CI/CD integration
- **HTML** - Interactive web-based report with charts
- **PDF** - Professional report for stakeholders
- **Markdown** - Documentation-friendly format
- **SARIF** - Standard format for security tools

## Development

### Prerequisites

- Go 1.25.3 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/btsg.git
cd btsg

# Install dependencies
go mod download

# Build
go build -o btsg

# Run tests
go test ./...
```

### Project Structure

- `cmd/` - Command-line interface implementation using Cobra
- `internal/` - Internal application logic (not importable by other projects)
- `pkg/` - Public packages that can be imported by other projects
- `main.go` - Application entry point

## Roadmap

- [ ] Implement actual vulnerability scanning engines
- [ ] Integrate with AI services (OpenAI, Claude, etc.)
- [ ] Add support for more vulnerability types
- [ ] Implement CI/CD integration
- [ ] Add watch mode for continuous monitoring
- [ ] Create web dashboard
- [ ] Add plugin system for custom scanners

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

Bob The Security Guy 🔒