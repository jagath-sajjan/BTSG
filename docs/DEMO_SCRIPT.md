# BTSG Hackathon Demo Script

**Duration:** < 2 minutes  
**Goal:** Show real security impact with AI-powered vulnerability management

---

## 🎯 Demo Flow

### Setup (Pre-demo)
```bash
# Clone vulnerable demo repository
git clone https://github.com/vulnerable-demo/python-app demo-app
cd demo-app

# Ensure BTSG is installed
btsg --version
```

---

## 📝 Script (90 seconds)

### **[0:00-0:15] Introduction (15s)**

> "Hi! I'm demonstrating **BTSG - Bob The Security Guy**, an AI-powered security scanner that finds, explains, and fixes vulnerabilities automatically."
>
> "Let's scan this Python application for security issues."

---

### **[0:15-0:30] Step 1: Scan (15s)**

```bash
btsg scan .
```

**Expected Output:**
```
🔍 BTSG Security Scanner
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Scanning with bandit... ✓
Scanning with pip-audit... ✓
Scanning with detect-secrets... ✓

Found 5 Vulnerabilities
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

#  Severity    File                Line  Description
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1  🔴 HIGH     app.py              23    SQL Injection vulnerability
2  🟠 MEDIUM   config.py           15    Hardcoded API key detected
3  🟡 LOW      utils.py            42    Weak cryptographic algorithm
```

> "We found **5 vulnerabilities** including a HIGH severity SQL injection. Let's understand what's wrong."

---

### **[0:30-0:50] Step 2: Explain (20s)**

```bash
btsg explain --finding-id 1
```

**Expected Output:**
```
🤖 AI-Powered Vulnerability Explanation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Location: app.py:23
🔴 Severity: HIGH
🏷️  Type: SQL Injection (CWE-89)

❓ What's Wrong?
The code directly concatenates user input into SQL queries without 
sanitization, allowing attackers to inject malicious SQL commands.

💥 Impact:
- Data theft: Attackers can extract sensitive database information
- Data manipulation: Unauthorized modification or deletion of records
- Authentication bypass: Potential admin access without credentials

🔧 How to Fix:
Use parameterized queries or an ORM like SQLAlchemy:

  # Bad (Current)
  query = f"SELECT * FROM users WHERE id = {user_id}"
  
  # Good (Recommended)
  query = "SELECT * FROM users WHERE id = ?"
  cursor.execute(query, (user_id,))

📚 References:
- OWASP SQL Injection: https://owasp.org/www-community/attacks/SQL_Injection
- CWE-89: https://cwe.mitre.org/data/definitions/89.html
```

> "BTSG's AI explains the vulnerability, its impact, and provides a secure code example. Now let's fix it automatically."

---

### **[0:50-1:10] Step 3: Fix (20s)**

```bash
btsg fix --vuln-id 1 --preview
```

**Expected Output:**
```
🔧 Proposed Fix for SQL Injection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📄 File: app.py
📍 Line: 23
🎯 Confidence: 95%

Diff Preview:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- query = f"SELECT * FROM users WHERE id = {user_id}"
- cursor.execute(query)
+ query = "SELECT * FROM users WHERE id = ?"
+ cursor.execute(query, (user_id,))
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Apply this fix? (y/n): y

✅ Fix applied successfully!
💾 Backup created at: .btsg-backups/app.py.2024-01-15-143022
```

> "The fix is applied with automatic backup. If something goes wrong, we can rollback instantly."

---

### **[1:10-1:25] Step 4: Report (15s)**

```bash
btsg report --format html --output security-report.html
```

**Expected Output:**
```
📊 Generating Security Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Analyzing findings...
✓ Generating charts...
✓ Creating HTML report...

✅ Report generated: security-report.html

Summary:
  Total Vulnerabilities: 5
  Critical: 0
  High: 1 (Fixed: 1)
  Medium: 2
  Low: 2
```

> "BTSG generates a comprehensive HTML report with charts, trends, and remediation status."

---

### **[1:25-1:30] Bonus: Watch Mode (5s)**

```bash
btsg watch --verbose
```

**Expected Output:**
```
🔍 BTSG File Watcher
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Watching 3 paths for changes...
   - ./src
   - ./config
   - ./utils

📊 Running initial scan...
✅ No vulnerabilities found (scan took 2.3s)

📡 Watching for changes... (Press Ctrl+C to stop)
```

> "And with watch mode, BTSG continuously monitors your code and scans automatically on file changes."

---

## 🎬 Closing (5s)

> "That's BTSG - **Find, Explain, Fix, Report** - all powered by AI. Security made simple."

---

## 📋 Demo Checklist

### Before Demo
- [ ] Install BTSG: `go install`
- [ ] Set up `.env` with Hack Club API key
- [ ] Clone vulnerable demo repository
- [ ] Test all commands once
- [ ] Prepare backup slides (in case of network issues)
- [ ] Have security-report.html pre-generated as backup

### During Demo
- [ ] Clear terminal before starting
- [ ] Use large font size (18-20pt)
- [ ] Speak clearly and maintain pace
- [ ] Point to key outputs on screen
- [ ] Have backup plan if API fails

### After Demo
- [ ] Show GitHub repository
- [ ] Mention open-source and extensible
- [ ] Share documentation link
- [ ] Answer questions

---

## 🎯 Key Talking Points

1. **Problem**: Manual security reviews are slow and require expertise
2. **Solution**: BTSG automates finding, explaining, and fixing vulnerabilities
3. **AI-Powered**: Uses Gemini 2.5 Pro for intelligent explanations
4. **Safe**: Automatic backups and rollback capabilities
5. **Real-time**: Watch mode for continuous security monitoring
6. **Comprehensive**: Integrates multiple security tools (Bandit, pip-audit, detect-secrets)

---

## 🚀 Impressive Stats to Mention

- **3 Security Tools** integrated (Bandit, pip-audit, detect-secrets)
- **AI-Powered** explanations using Google Gemini 2.5 Pro
- **Automatic Fixes** with 95%+ confidence
- **Real-time Monitoring** with file watching
- **Comprehensive Reports** in multiple formats (HTML, JSON, Markdown)
- **Production-Ready** with rollback and backup systems

---

## 💡 Backup Talking Points (If Time Permits)

### Architecture Highlights
- Modular scanner engine (easy to add new tools)
- Caching layer for faster repeated scans
- Concurrent scanning for performance
- Extensible via MCP (Model Context Protocol)

### Use Cases
- **Development**: Catch vulnerabilities during coding
- **CI/CD**: Integrate into build pipelines
- **Security Audits**: Generate compliance reports
- **Learning**: Understand security issues with AI explanations

---

## 🎥 Visual Tips

1. **Terminal Setup**
   - Use dark theme with high contrast
   - Font: Fira Code or JetBrains Mono
   - Size: 18-20pt
   - Colors: Enable ANSI colors

2. **Screen Recording** (if needed)
   - Use Asciinema for terminal recordings
   - Record at 1920x1080 resolution
   - Keep recordings under 2 minutes

3. **Presentation Flow**
   - Start with problem statement
   - Show live demo
   - End with impact/results
   - Have backup slides ready

---

## 📞 Q&A Preparation

**Q: How accurate are the AI explanations?**  
A: We use Google Gemini 2.5 Pro with carefully crafted prompts. Explanations include CWE references and OWASP guidelines for accuracy.

**Q: Can it fix all vulnerabilities?**  
A: BTSG focuses on high-confidence fixes (95%+). Complex issues get detailed explanations for manual review.

**Q: What languages are supported?**  
A: Currently Python (Bandit, pip-audit) and secrets detection (all languages). Easy to extend with new scanners.

**Q: Is it production-ready?**  
A: Yes! Includes automatic backups, rollback capabilities, and comprehensive error handling.

**Q: How does it compare to other tools?**  
A: BTSG combines scanning, AI explanations, and automatic fixes in one tool. Most tools only scan.

---

## 🔗 Resources

- **GitHub**: https://github.com/jagath-sajjan/BTSG
- **Documentation**: See `/docs` folder
- **Demo Repository**: https://github.com/vulnerable-demo/python-app
- **Hack Club AI**: https://hackclub.com/ai

---

**Made with ❤️ by Bob (The Security Guy)**