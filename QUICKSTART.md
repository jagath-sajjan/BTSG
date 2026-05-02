# BTSG Quick Start Guide

## Installation

### Option 1: Run from source directory
```bash
cd /Users/jagath-sajjan/btsg
./btsg --help
```

### Option 2: Install globally
```bash
cd /Users/jagath-sajjan/btsg
go install
# Now you can run: btsg --help
```

### Option 3: Add to PATH
```bash
# Add to your ~/.zshrc or ~/.bashrc
export PATH="$PATH:/Users/jagath-sajjan/btsg"

# Reload shell
source ~/.zshrc

# Now you can run: btsg --help
```

## Quick Test

### 1. Create a test file with a vulnerability
```bash
cat > test_vuln.py << 'EOF'
import os

# Hardcoded API key - SECURITY ISSUE!
API_KEY = "sk-1234567890abcdef"

def get_user(user_id):
    # SQL injection vulnerability
    query = "SELECT * FROM users WHERE id = " + user_id
    return execute_query(query)
EOF
```

### 2. Run scan (when scanner is implemented)
```bash
./btsg scan
```

### 3. Test explain command
```bash
# Explain a specific vulnerability
./btsg explain BTSG-001

# Explain with details
./btsg explain BTSG-001 --detailed

# Explain all vulnerabilities
./btsg explain --from-scan

# Verbose mode
./btsg explain BTSG-001 -v
```

## Configuration

Your `.env` file is already configured:
```bash
AI_PROVIDER=hackclub
AI_API_KEY=your-api-key-here
AI_MODEL=openai/gpt-5.5-pro
AI_BASE_URL=https://ai.hackclub.com/proxy/v1
ENABLE_CACHE=true
CACHE_TTL=24h
```

**Note**: The actual API key is stored securely in your `.env` file (not committed to git).

## Usage Examples

### Scan a project
```bash
./btsg scan ./myproject
./btsg scan --recursive
```

### Explain vulnerabilities
```bash
# Single vulnerability
./btsg explain BTSG-001

# All vulnerabilities from last scan
./btsg explain --from-scan

# With technical details
./btsg explain BTSG-001 --detailed
```

### Generate reports
```bash
./btsg report --format json
./btsg report --format html
```

## Next Steps

1. ✅ Binary is built and working
2. ✅ AI explain command is implemented
3. ✅ Configuration is set up
4. 🔄 Need to run a scan to generate findings
5. 🔄 Then test explain command with real vulnerabilities

## Troubleshooting

### "command not found: btsg"
Use `./btsg` instead of `btsg`, or install globally with `go install`

### "No scan results found"
Run `./btsg scan` first to generate findings

### "AI_API_KEY not set"
Check that `.env` file exists in the current directory

## Made with Bob 🤖