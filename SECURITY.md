# Security Best Practices for BTSG

## 🔒 API Key Security

### ✅ DO:
- Store API keys in `.env` file only
- Keep `.env` in `.gitignore` (already configured)
- Use `.env.example` as a template for others
- Rotate API keys regularly
- Use different keys for dev/staging/prod

### ❌ DON'T:
- Never commit `.env` to version control
- Never hardcode API keys in source code
- Never share API keys in documentation
- Never expose keys in logs or error messages
- Never commit keys in comments

## 📁 File Security

### Protected Files (in .gitignore)
```
.env
.env.local
.env.*.local
*.pem
*.key
secrets.yml
credentials.json
```

### Safe to Commit
```
.env.example  ✅ (template without real keys)
go.mod        ✅
go.sum        ✅
*.go          ✅ (source code)
docs/*.md     ✅ (documentation)
```

## 🔐 Configuration Security

### Environment Variables
Always load from `.env`:
```go
// ✅ CORRECT
apiKey := os.Getenv("AI_API_KEY")

// ❌ WRONG
apiKey := "sk-1234567890abcdef"
```

### Checking for Exposed Keys
```bash
# Search for potential exposed keys
git log -p | grep -i "api_key\|secret\|password"

# Check current files
grep -r "sk-" . --exclude-dir=.git --exclude="*.md"
```

## 🛡️ Git Security

### Before Committing
```bash
# Check what you're about to commit
git diff --cached

# Ensure .env is not staged
git status | grep .env
```

### If You Accidentally Commit a Key

1. **Immediately rotate the key** (generate new one)
2. Remove from git history:
```bash
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch .env" \
  --prune-empty --tag-name-filter cat -- --all
```
3. Force push (if already pushed):
```bash
git push origin --force --all
```

## 🔍 Scanning for Secrets

### Using detect-secrets
```bash
# Install
pip install detect-secrets

# Scan
detect-secrets scan

# Audit findings
detect-secrets audit .secrets.baseline
```

### Using git-secrets
```bash
# Install
brew install git-secrets

# Setup
git secrets --install
git secrets --register-aws

# Scan
git secrets --scan
```

## 📋 Security Checklist

Before committing:
- [ ] `.env` is in `.gitignore`
- [ ] No API keys in source code
- [ ] No keys in documentation
- [ ] `.env.example` has placeholder values only
- [ ] Ran `git diff --cached` to review changes
- [ ] No sensitive data in commit message

## 🚨 Incident Response

If a key is exposed:

1. **Immediate Actions** (within 5 minutes)
   - Rotate/revoke the exposed key
   - Generate new key
   - Update `.env` with new key

2. **Investigation** (within 1 hour)
   - Check where it was exposed (commit, docs, logs)
   - Determine exposure duration
   - Review access logs for unauthorized use

3. **Remediation** (within 24 hours)
   - Remove from git history if committed
   - Update all systems using the key
   - Document the incident
   - Review security practices

4. **Prevention** (ongoing)
   - Add pre-commit hooks
   - Enable secret scanning
   - Train team on security
   - Regular security audits

## 🔧 Pre-commit Hook

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash

# Check for .env in staged files
if git diff --cached --name-only | grep -q "^\.env$"; then
    echo "❌ Error: Attempting to commit .env file!"
    echo "Remove it from staging: git reset HEAD .env"
    exit 1
fi

# Check for potential API keys
if git diff --cached | grep -qE "sk-[a-zA-Z0-9]{32,}|ghp_[a-zA-Z0-9]{36}"; then
    echo "❌ Error: Potential API key detected in commit!"
    echo "Review your changes and remove any secrets"
    exit 1
fi

exit 0
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## 📞 Support

If you discover a security issue:
1. Do NOT create a public issue
2. Rotate any exposed credentials immediately
3. Contact the maintainers privately
4. Document the incident

## 🎯 Summary

**Golden Rule**: If it's secret, it stays in `.env` and never leaves your local machine.

- ✅ API keys → `.env` only
- ✅ Passwords → `.env` only  
- ✅ Tokens → `.env` only
- ✅ Credentials → `.env` only

**Remember**: `.env` is in `.gitignore` and should NEVER be committed!

## Made with Bob 🤖