# 🚀 BTSG Demo Ready Guide

## ⚠️ Scanner Installation (REQUIRED FIRST)

**IMPORTANT**: These security scanning tools must be installed BEFORE running `btsg scan`.

### 1️⃣ Quick Install (All Tools)
```bash
pip install bandit pip-audit detect-secrets
```

### 2️⃣ Verify Installation
```bash
bandit --version
pip-audit --version
detect-secrets --version
```

**Expected output**: Each command should display a version number (e.g., `bandit 1.7.5`).

### 3️⃣ Individual Installation (if needed)
If you prefer to install tools separately or if the quick install fails:

- **Bandit** (Python code security scanner):
  ```bash
  pip install bandit
  ```

- **pip-audit** (Dependency vulnerability scanner):
  ```bash
  pip install pip-audit
  ```

- **detect-secrets** (Secret detection tool):
  ```bash
  pip install detect-secrets
  ```

### 4️⃣ Troubleshooting

**Problem: `pip: command not found`**
- **Solution**: Install Python 3.x first from [python.org](https://www.python.org/downloads/)
- **Alternative**: Try `pip3` instead of `pip`:
  ```bash
  pip3 install bandit pip-audit detect-secrets
  ```

**Problem: `Permission denied` or access errors**
- **Solution 1** (Recommended): Install for current user only:
  ```bash
  pip install --user bandit pip-audit detect-secrets
  ```
- **Solution 2**: Use sudo (Linux/macOS):
  ```bash
  sudo pip install bandit pip-audit detect-secrets
  ```

**Problem: Tools installed but not found**
- **Solution**: Add Python's bin directory to PATH:
  ```bash
  # macOS/Linux
  export PATH="$HOME/.local/bin:$PATH"
  
  # Or add to ~/.bashrc or ~/.zshrc for permanent fix
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
  ```

**Problem: Version conflicts**
- **Solution**: Use a virtual environment:
  ```bash
  python3 -m venv btsg-env
  source btsg-env/bin/activate  # On Windows: btsg-env\Scripts\activate
  pip install bandit pip-audit detect-secrets
  ```

---

## Quick Start (2 minutes)

### Prerequisites
1. **Scanner Tools** (See "Scanner Installation" section above):
   ```bash
   # Verify scanners are installed
   bandit --version && pip-audit --version && detect-secrets --version
   ```

2. **API Key Setup** (Choose one):
   ```bash
   # Option 1: OpenAI (Recommended)
   export OPENAI_API_KEY="your-key-here"
   
   # Option 2: Hack Club AI (Free)
   export HACKCLUB_API_KEY="your-key-here"
   ```

3. **Build BTSG**:
   ```bash
   go build -o btsg
   ```

---

## 🎯 4-Command Demo Flow

### 1️⃣ Scan for Vulnerabilities
```bash
./btsg scan vulnerable-demo/test.py
```

**Expected Output:**
```
🔍 Scanning vulnerable-demo/test.py...
✓ Scan complete! Found 8 issues
📊 Results saved to: .btsg/scan-results.json

Summary:
  HIGH: 3 issues
  MEDIUM: 3 issues
  LOW: 2 issues
```

**What it does:** Runs Bandit, pip-audit, and detect-secrets to find security issues.

---

### 2️⃣ Get AI Explanation
```bash
./btsg explain
```

**Expected Output:**
```
🤖 Analyzing security issues with AI...
✓ Generated explanations for 8 issues
📄 Report saved to: security-report.md

Top Issues:
  • SQL Injection (HIGH) - Line 15
  • Hardcoded Password (HIGH) - Line 8
  • Command Injection (HIGH) - Line 23
```

**What it does:** Uses AI to explain each vulnerability in plain English with fix suggestions.

---

### 3️⃣ Auto-Fix Issues
```bash
./btsg fix
```

**Expected Output:**
```
🔧 Applying AI-suggested fixes...
✓ Fixed 6/8 issues
⚠ 2 issues require manual review

Backup created: .btsg/backups/test.py.backup-20260502-064500

Fixed:
  ✓ SQL Injection - Added parameterized query
  ✓ Hardcoded Password - Moved to environment variable
  ✓ Command Injection - Added input sanitization
```

**What it does:** Automatically applies AI-generated fixes with backup creation.

---

### 4️⃣ Generate Report
```bash
./btsg report
```

**Expected Output:**
```
📊 Generating comprehensive security report...
✓ Report generated: security-report.md

Report includes:
  • Executive summary
  • Detailed findings (8 issues)
  • Fix recommendations
  • Compliance mapping (OWASP, CWE)
```

**What it does:** Creates a professional markdown report for stakeholders.

---

## 🎬 Complete Demo Script (30 seconds)

```bash
# 1. Scan
./btsg scan vulnerable-demo/test.py

# 2. Explain
./btsg explain

# 3. Fix
./btsg fix

# 4. Report
./btsg report

# View the report
cat security-report.md
```

---

## 🔧 Troubleshooting

### "API key not found"
```bash
# Check if key is set
echo $OPENAI_API_KEY
# or
echo $HACKCLUB_API_KEY

# Set it if missing
export OPENAI_API_KEY="your-key-here"
```

### "Python tools not found"
```bash
# Install all required tools
pip install bandit pip-audit detect-secrets

# Verify installation
bandit --version
pip-audit --version
detect-secrets --version
```

### "No scan results found"
```bash
# Run scan first
./btsg scan vulnerable-demo/test.py

# Check results file exists
ls -la .btsg/scan-results.json
```

### "Build failed"
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o btsg
```

---

## 📁 Demo Files

- **Vulnerable Code**: `vulnerable-demo/test.py` (8 intentional vulnerabilities)
- **Scan Results**: `.btsg/scan-results.json` (shared between commands)
- **AI Report**: `security-report.md` (generated by explain/report)
- **Backups**: `.btsg/backups/` (created before fixes)

---

## 🎯 Key Features to Highlight

1. **Multi-Tool Integration**: Bandit + pip-audit + detect-secrets
2. **AI-Powered Explanations**: Plain English vulnerability descriptions
3. **Automated Fixes**: Safe, backed-up code modifications
4. **Professional Reports**: Stakeholder-ready documentation
5. **Shared State**: Commands work together seamlessly

---

## 💡 Pro Tips

- **Run commands in order** for best demo flow
- **Show the report** (`cat security-report.md`) to highlight AI quality
- **Emphasize safety**: Backups are created before any fixes
- **Highlight speed**: Full scan + explain + fix in under 10 seconds
- **Mention extensibility**: Easy to add more scanners or AI providers

---

## 🚀 Ready to Demo!

Your BTSG installation is ready. Run the 4-command flow and showcase:
- ✅ Comprehensive security scanning
- ✅ AI-powered vulnerability analysis
- ✅ Automated fix generation
- ✅ Professional reporting

**Time to complete demo**: ~30 seconds  
**Wow factor**: High 🎉