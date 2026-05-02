# BTSG AI Explain Command - Usage Guide

## Overview

The `btsg explain` command uses AI to provide detailed, human-friendly explanations of security vulnerabilities found in your code.

## Features

✨ **AI-Powered Explanations**
- Simple language explanations for developers
- Technical details for security experts
- Real-world examples and incidents
- Actionable remediation steps with code examples

🎯 **Risk Assessment**
- Likelihood and impact analysis
- Attack scenarios
- Data at risk identification

💾 **Smart Caching**
- 24-hour cache for faster responses
- Reduces API costs
- Offline template fallback

## Setup

### 1. Configure Environment Variables

Create a `.env` file in your project root:

```bash
# AI Provider Configuration
AI_PROVIDER=hackclub
AI_API_KEY=your-api-key-here
AI_MODEL=openai/gpt-5.5-pro
AI_BASE_URL=https://ai.hackclub.com/proxy/v1

# Cache Configuration
ENABLE_CACHE=true
CACHE_TTL=24h
```

**Security Note**: Never commit your `.env` file to version control. It's already in `.gitignore`.

### 2. Run a Scan First

The explain command works with scan results:

```bash
btsg scan
```

This creates a `scan-results.json` file with all findings.

## Usage

### Basic Usage

Explain a specific vulnerability by ID:

```bash
btsg explain BTSG-001
```

### Detailed Mode

Get more technical details:

```bash
btsg explain BTSG-001 --detailed
```

### Explain All Vulnerabilities

Explain all findings from the last scan:

```bash
btsg explain --from-scan
```

### Verbose Output

See additional metadata (source, confidence, tokens used):

```bash
btsg explain BTSG-001 --verbose
```

## Output Format

The explain command provides structured output:

```
╔═══════════════════════════════════════════════════════════════╗
║  🔒 Security Vulnerability Explanation                        ║
╚═══════════════════════════════════════════════════════════════╝

📋 ID: BTSG-001
📁 File: app.py:42
🔧 Tool: bandit
⚠️  Severity: HIGH

💡 What is it?
─────────────────────────────────────────────────────────────
[Simple explanation in plain language]

🎯 Risk Assessment
─────────────────────────────────────────────────────────────
Likelihood: high | Impact: critical

Attack Scenarios:
  1. [Scenario 1]
  2. [Scenario 2]

Data at Risk:
  • [Data type 1]
  • [Data type 2]

🌍 Real-World Example
─────────────────────────────────────────────────────────────
Title: [Real incident]
[Description of what happened]

Impact: [Consequences]
Lesson: [Key takeaway]

✅ How to Fix
─────────────────────────────────────────────────────────────
1. [Action] [Priority: immediate, Effort: low]
   [Detailed description]

2. [Action] [Priority: high, Effort: medium]
   [Detailed description]

💻 Code Example
─────────────────────────────────────────────────────────────
Before (Vulnerable):
```python
[Vulnerable code]
```

After (Secure):
```python
[Secure code]
```

Explanation: [What changed and why]
```

## Examples

### Example 1: SQL Injection

```bash
$ btsg explain BTSG-001

╔═══════════════════════════════════════════════════════════════╗
║  🔒 Security Vulnerability Explanation                        ║
╚═══════════════════════════════════════════════════════════════╝

📋 ID: BTSG-001
📁 File: app.py:42
🔧 Tool: bandit
⚠️  Severity: HIGH

💡 What is it?
─────────────────────────────────────────────────────────────
Your code is vulnerable to SQL injection because it directly 
concatenates user input into SQL queries. This allows attackers 
to manipulate your database queries and potentially access, 
modify, or delete data they shouldn't have access to.

🎯 Risk Assessment
─────────────────────────────────────────────────────────────
Likelihood: high | Impact: critical

Attack Scenarios:
  1. Attacker extracts entire user database including passwords
  2. Attacker modifies data to escalate privileges
  3. Attacker deletes critical business data

Data at Risk:
  • User credentials and personal information
  • Financial records
  • Business-critical data

✅ How to Fix
─────────────────────────────────────────────────────────────
1. Use parameterized queries [Priority: immediate, Effort: low]
   Replace string concatenation with parameterized queries or 
   prepared statements that automatically escape user input.

2. Implement input validation [Priority: high, Effort: medium]
   Validate and sanitize all user inputs before using them in 
   database queries.

💻 Code Example
─────────────────────────────────────────────────────────────
Before (Vulnerable):
```python
query = "SELECT * FROM users WHERE id = " + user_id
cursor.execute(query)
```

After (Secure):
```python
query = "SELECT * FROM users WHERE id = ?"
cursor.execute(query, (user_id,))
```

Explanation: Parameterized queries treat user input as data, 
not executable code, preventing SQL injection attacks.
```

### Example 2: Hardcoded Secret

```bash
$ btsg explain BTSG-002

╔═══════════════════════════════════════════════════════════════╗
║  🔒 Security Vulnerability Explanation                        ║
╚═══════════════════════════════════════════════════════════════╝

📋 ID: BTSG-002
📁 File: config.py:15
🔧 Tool: detect-secrets
⚠️  Severity: CRITICAL

💡 What is it?
─────────────────────────────────────────────────────────────
An API key or secret credential is hardcoded directly in your 
source code. If this code is committed to version control or 
shared, the secret becomes permanently exposed and can be used 
by anyone with access to the repository.

🎯 Risk Assessment
─────────────────────────────────────────────────────────────
Likelihood: critical | Impact: critical

Attack Scenarios:
  1. Unauthorized access to your API/service
  2. Data breach through compromised credentials
  3. Financial loss from API abuse

Data at Risk:
  • API access credentials
  • Service authentication tokens
  • Protected resources

🌍 Real-World Example
─────────────────────────────────────────────────────────────
Title: Uber 2016 Data Breach
In 2016, Uber suffered a data breach affecting 57 million users 
because AWS credentials were hardcoded in a GitHub repository.

Impact: $148 million settlement, loss of customer trust
Lesson: Never commit secrets to version control, even private repos

✅ How to Fix
─────────────────────────────────────────────────────────────
1. Rotate the exposed secret immediately [Priority: immediate, Effort: low]
   Generate a new API key/secret and update all systems using it.

2. Remove secret from code [Priority: immediate, Effort: low]
   Delete the hardcoded secret from your source code.

3. Use environment variables [Priority: high, Effort: low]
   Store secrets in environment variables or a secret management 
   service like AWS Secrets Manager, HashiCorp Vault, or .env files.

💻 Code Example
─────────────────────────────────────────────────────────────
Before (Vulnerable):
```python
API_KEY = "sk-1234567890abcdef"
```

After (Secure):
```python
import os
API_KEY = os.getenv("API_KEY")
if not API_KEY:
    raise ValueError("API_KEY environment variable not set")
```

Explanation: Environment variables keep secrets out of source 
code and allow different values per environment (dev, staging, prod).
```

## Command Options

| Option | Short | Description |
|--------|-------|-------------|
| `--detailed` | `-d` | Show detailed technical explanation |
| `--from-scan` | | Explain all vulnerabilities from last scan |
| `--id` | | Specify vulnerability ID to explain |
| `--verbose` | `-v` | Show additional metadata and debug info |

## How It Works

1. **Load Scan Results**: Reads findings from `scan-results.json`
2. **Check Cache**: Looks for cached explanation (24h TTL)
3. **Generate Prompt**: Creates context-aware prompt for AI
4. **Call AI API**: Sends request to Hack Club AI proxy
5. **Parse Response**: Extracts structured explanation
6. **Cache Result**: Stores for future use
7. **Display**: Shows formatted output

## Performance

- **First Request**: ~2-5 seconds (AI generation)
- **Cached Request**: <100ms (instant)
- **Cost**: ~$0.002 per explanation (with caching)
- **Tokens**: ~500-2000 per explanation

## Troubleshooting

### "AI_API_KEY not set"

Make sure you have a `.env` file with your API key:

```bash
AI_API_KEY=your-api-key-here
```

### "No scan results found"

Run a scan first:

```bash
btsg scan
```

### "Vulnerability not found"

Check the vulnerability ID matches one from your scan:

```bash
btsg scan  # Shows all vulnerability IDs
```

### API Timeout

Increase timeout in code or try again. The system has automatic retry logic.

## Advanced Usage

### Custom Configuration

You can customize the AI behavior by modifying environment variables:

```bash
# Use different model
AI_MODEL=openai/gpt-4

# Adjust cache duration
CACHE_TTL=48h

# Disable cache
ENABLE_CACHE=false
```

### Batch Processing

Explain multiple vulnerabilities efficiently:

```bash
# Explain all findings (uses worker pool for concurrency)
btsg explain --from-scan
```

### Integration with CI/CD

```bash
#!/bin/bash
# Scan and explain in CI pipeline

btsg scan --output scan-results.json

# Explain critical vulnerabilities
btsg explain --from-scan | grep -A 50 "CRITICAL"
```

## Best Practices

1. **Run Scans Regularly**: Keep scan results up to date
2. **Review Explanations**: Use AI as a guide, not absolute truth
3. **Verify Fixes**: Test remediation steps before deploying
4. **Cache Wisely**: 24h cache balances freshness and cost
5. **Secure API Keys**: Never commit `.env` to version control

## API Provider Details

### Hack Club AI Proxy

- **Endpoint**: `https://ai.hackclub.com/proxy/v1`
- **Models**: OpenAI GPT-4, GPT-3.5, and more
- **Rate Limits**: Generous for development
- **Cost**: Free for Hack Club members

### Switching Providers

To use OpenAI directly:

```bash
AI_PROVIDER=openai
AI_API_KEY=sk-your-openai-key
AI_MODEL=gpt-4
AI_BASE_URL=https://api.openai.com/v1
```

## Support

For issues or questions:
- Check the main README.md
- Review error messages carefully
- Ensure `.env` is properly configured
- Verify scan results exist

## Made with Bob 🤖