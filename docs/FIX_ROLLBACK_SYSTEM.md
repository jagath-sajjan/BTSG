## BTSG Fix Command - Rollback System Documentation

### Overview

The BTSG fix command includes a comprehensive rollback system that automatically backs up files before applying fixes and can restore them if anything goes wrong.

### Features

#### 1. Automatic Backup Creation
- **When**: Before every fix is applied
- **Where**: `.btsg-backups/` directory (configurable)
- **Format**: `filename.YYYYMMDD-HHMMSS.backup`
- **Example**: `app.py.20260502-113045.backup`

#### 2. Backup on Apply
```go
// Automatically creates backup before applying fix
result, err := fixer.ApplyFix(fix)
if err != nil {
    // Fix failed, backup is preserved
    // Original file unchanged
}
// Success: backup created at result.BackupPath
```

#### 3. Manual Rollback
```go
// Restore from backup
err := fixer.RollbackFix(backupPath)
if err != nil {
    // Rollback failed
}
// File restored to original state
```

#### 4. Automatic Rollback on Failure
The system automatically preserves the original file if:
- Validation fails
- File write fails
- Permission denied
- Any error occurs

### Usage Examples

#### Example 1: Apply Fix with Automatic Backup
```bash
# Fix creates backup automatically
./btsg fix BTSG-001

# Output shows backup location
Backup created: .btsg-backups/app.py.20260502-113045.backup
Fix applied successfully!
```

#### Example 2: Rollback After Fix
```bash
# Apply fix
./btsg fix BTSG-001
# Backup: .btsg-backups/app.py.20260502-113045.backup

# Rollback if needed
./btsg fix --rollback .btsg-backups/app.py.20260502-113045.backup
# File restored to original state
```

#### Example 3: Automatic Failure Protection
```bash
# Try to apply fix
./btsg fix BTSG-001

# If fix fails:
# - Original file is NOT modified
# - Backup is created but not used
# - Error message explains what went wrong
# - You can try again safely
```

### Implementation Details

#### Backup Creation Process
```
1. Read original file
2. Generate backup filename with timestamp
3. Copy original to backup directory
4. Return backup path
5. Proceed with fix
```

#### Rollback Process
```
1. Verify backup file exists
2. Extract original filename from backup
3. Copy backup to original location
4. Verify restoration
5. Return success/error
```

#### Safety Mechanisms

1. **Pre-validation**
   - Check file exists
   - Check permissions
   - Validate fix confidence
   - Check line change limits

2. **Backup Protection**
   - Backup created BEFORE any changes
   - Backup preserved on failure
   - Timestamped for multiple versions
   - Separate directory for organization

3. **Atomic Operations**
   - Read entire file
   - Apply changes in memory
   - Write only if successful
   - Rollback if write fails

4. **Error Handling**
   - Detailed error messages
   - Backup path in error response
   - Original file never corrupted
   - Safe to retry

### Configuration

```go
config := &FixerConfig{
    CreateBackup: true,                    // Enable backups
    BackupDir:    ".btsg-backups",        // Backup directory
    RequireConfirmation: true,             // Ask before applying
    MaxLineChanges: 50,                    // Safety limit
    MinConfidence: 0.7,                    // Quality threshold
}
```

### File Structure

```
project/
├── app.py                          # Original file
├── .btsg-backups/                  # Backup directory
│   ├── app.py.20260502-113045.backup
│   ├── app.py.20260502-114523.backup
│   └── config.py.20260502-115012.backup
└── .gitignore                      # Excludes .btsg-backups/
```

### API Reference

#### CreateBackup
```go
func (f *fixer) createBackup(filePath string) (string, error)
```
- **Input**: Path to file to backup
- **Output**: Path to backup file
- **Error**: If backup creation fails

#### RollbackFix
```go
func (f *fixer) RollbackFix(backupPath string) error
```
- **Input**: Path to backup file
- **Output**: Error if rollback fails
- **Effect**: Restores original file

#### ApplyFix (with automatic backup)
```go
func (f *fixer) ApplyFix(fix *Fix) (*FixResult, error)
```
- **Input**: Fix to apply
- **Output**: Result with backup path
- **Effect**: Creates backup, applies fix, or rolls back on error

### Best Practices

1. **Always Enable Backups**
   ```go
   config.CreateBackup = true  // Recommended
   ```

2. **Keep Backups Organized**
   ```bash
   # Backups are timestamped
   # Clean old backups periodically
   find .btsg-backups -mtime +30 -delete
   ```

3. **Test Fixes First**
   ```bash
   # Preview before applying
   ./btsg fix BTSG-001 --preview
   
   # Apply with confirmation
   ./btsg fix BTSG-001 --interactive
   ```

4. **Verify After Fix**
   ```bash
   # Check if fix worked
   ./btsg scan
   
   # Rollback if needed
   ./btsg fix --rollback <backup-path>
   ```

### Error Scenarios

#### Scenario 1: Validation Fails
```
Input: Fix with low confidence (0.5)
Config: MinConfidence = 0.7

Result:
- Validation fails
- No backup created
- No file modified
- Error: "confidence 0.50 below threshold 0.70"
```

#### Scenario 2: Write Permission Denied
```
Input: Fix for read-only file
Action: Attempt to apply fix

Result:
- Backup created successfully
- Write fails (permission denied)
- Original file unchanged
- Backup preserved
- Error: "failed to write file: permission denied"
```

#### Scenario 3: Disk Full
```
Input: Fix for large file
Action: Attempt to write

Result:
- Backup created
- Write fails (no space)
- Original file unchanged
- Backup preserved
- Error: "failed to write file: no space left on device"
```

### Recovery Procedures

#### If Fix Goes Wrong
```bash
# 1. Find the backup
ls -lt .btsg-backups/

# 2. Identify the correct backup
# Format: filename.YYYYMMDD-HHMMSS.backup

# 3. Rollback
./btsg fix --rollback .btsg-backups/app.py.20260502-113045.backup

# 4. Verify restoration
cat app.py  # Check content
```

#### If Backup Directory Lost
```bash
# Backups are in .btsg-backups/
# If deleted, cannot rollback
# Prevention: Add to .gitignore but don't delete

# Recovery:
# - Use git to restore: git checkout app.py
# - Use system backup if available
# - Manually fix the file
```

### Testing the Rollback System

```bash
# 1. Create test file
echo "original content" > test.py

# 2. Apply fix (creates backup)
./btsg fix TEST-001

# 3. Verify backup exists
ls .btsg-backups/test.py.*.backup

# 4. Check modified file
cat test.py

# 5. Rollback
./btsg fix --rollback .btsg-backups/test.py.*.backup

# 6. Verify restoration
cat test.py  # Should show "original content"
```

### Summary

The rollback system provides:
- ✅ **Automatic backups** before every fix
- ✅ **Safe failure handling** - original never corrupted
- ✅ **Easy restoration** with rollback command
- ✅ **Multiple versions** with timestamps
- ✅ **Organized storage** in dedicated directory
- ✅ **Atomic operations** - all or nothing
- ✅ **Error recovery** - always have a way back

**Golden Rule**: Every fix creates a backup. If anything goes wrong, you can always rollback.

## Made with Bob 🤖