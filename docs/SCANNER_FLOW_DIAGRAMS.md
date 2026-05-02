# BTSG Scanner Engine - Flow Diagrams

## 1. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         BTSG CLI                                 │
│                    (User Interface)                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Scanner Orchestrator                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Registry   │  │  Executor    │  │  Aggregator  │         │
│  │   Manager    │  │   Pool       │  │   Engine     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
         ┌──────────┐ ┌──────────┐ ┌──────────┐
         │  Bandit  │ │pip-audit │ │ detect-  │
         │  Engine  │ │  Engine  │ │ secrets  │
         └────┬─────┘ └────┬─────┘ └────┬─────┘
              │            │            │
              └────────────┴────────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │   Normalizer    │
                 │     Layer       │
                 └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │ Unified Results │
                 │   (JSON/SARIF)  │
                 └─────────────────┘
```

## 2. Scanner Execution Flow

```
START
  │
  ├─► Initialize Orchestrator
  │   ├─ Load Configuration
  │   ├─ Initialize Registry
  │   └─ Validate Scanners
  │
  ├─► Pre-Scan Phase
  │   ├─ Discover Files
  │   │  ├─ Walk Directory Tree
  │   │  ├─ Apply Exclusions
  │   │  └─ Filter by Type
  │   │
  │   ├─ Select Scanners
  │   │  ├─ Check Availability
  │   │  ├─ Match File Types
  │   │  └─ Apply User Filters
  │   │
  │   └─ Create Scan Plan
  │      ├─ Assign Files to Scanners
  │      ├─ Set Priorities
  │      └─ Allocate Resources
  │
  ├─► Execution Phase (Concurrent)
  │   │
  │   ├─► Scanner 1: Bandit
  │   │   ├─ Build Command
  │   │   ├─ Execute Process
  │   │   ├─ Capture Output
  │   │   ├─ Parse Results
  │   │   └─ Return Raw Data
  │   │
  │   ├─► Scanner 2: pip-audit
  │   │   ├─ Build Command
  │   │   ├─ Execute Process
  │   │   ├─ Capture Output
  │   │   ├─ Parse Results
  │   │   └─ Return Raw Data
  │   │
  │   └─► Scanner 3: detect-secrets
  │       ├─ Build Command
  │       ├─ Execute Process
  │       ├─ Capture Output
  │       ├─ Parse Results
  │       └─ Return Raw Data
  │
  ├─► Normalization Phase
  │   ├─ Collect Raw Results
  │   ├─ For Each Scanner Result:
  │   │  ├─ Parse Output Format
  │   │  ├─ Map to Unified Schema
  │   │  ├─ Enrich Metadata
  │   │  └─ Validate Data
  │   │
  │   └─ Merge All Results
  │
  ├─► Post-Processing Phase
  │   ├─ Deduplicate Findings
  │   │  ├─ Hash Vulnerabilities
  │   │  ├─ Compare Signatures
  │   │  └─ Merge Duplicates
  │   │
  │   ├─ Enrich Data
  │   │  ├─ Add CVE Details
  │   │  ├─ Calculate CVSS
  │   │  └─ Add References
  │   │
  │   ├─ Sort & Filter
  │   │  ├─ By Severity
  │   │  ├─ By Type
  │   │  └─ Apply Filters
  │   │
  │   └─ Generate Statistics
  │      ├─ Count by Severity
  │      ├─ Count by Type
  │      └─ Scanner Performance
  │
  └─► Output Phase
      ├─ Format Results
      │  ├─ JSON
      │  ├─ SARIF
      │  └─ Custom
      │
      └─ Return to CLI
         └─ Display/Save Results
```

## 3. Scanner Interface Interaction

```
┌─────────────────────────────────────────────────────────────┐
│                    Orchestrator                              │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ 1. GetMetadata()
                        ├──────────────────────►┐
                        │                        │
                        │ 2. IsAvailable()       │
                        ├──────────────────────► │
                        │                        │
                        │ 3. GetCommand(config)  │
                        ├──────────────────────► │
                        │                        │
                        │ 4. Scan(ctx, config)   │
                        ├──────────────────────► │
                        │                        │  Scanner
                        │ ◄──────────────────────┤  Engine
                        │   RawScanResult        │
                        │                        │
                        │ 5. Normalize(raw)      │
                        ├──────────────────────► │
                        │                        │
                        │ ◄──────────────────────┤
                        │   []Vulnerability      │
                        │                        │
                        ▼                        ▼
```

## 4. Data Transformation Pipeline

```
Raw Scanner Output
       │
       ├─► Bandit JSON
       │   {
       │     "results": [{
       │       "test_id": "B201",
       │       "issue_text": "...",
       │       "filename": "app.py",
       │       "line_number": 45
       │     }]
       │   }
       │
       ├─► pip-audit JSON
       │   {
       │     "vulnerabilities": [{
       │       "id": "CVE-2024-1234",
       │       "package": "requests",
       │       "version": "2.25.0"
       │     }]
       │   }
       │
       └─► detect-secrets JSON
           {
             "results": {
               "config.py": [{
                 "type": "AWS Access Key",
                 "line_number": 10
               }]
             }
           }
       │
       ▼
┌──────────────────┐
│   Normalizer     │
│   - Parse        │
│   - Map Fields   │
│   - Validate     │
└────────┬─────────┘
         │
         ▼
Unified Vulnerability Schema
{
  "id": "BTSG-001",
  "external_id": "B201",
  "title": "Use of insecure pickle",
  "severity": "HIGH",
  "type": "code",
  "file": "app.py",
  "line": 45,
  "scanner": "bandit",
  "cwe": ["CWE-502"],
  "remediation": "..."
}
         │
         ▼
┌──────────────────┐
│  Deduplicator    │
│  - Hash          │
│  - Compare       │
│  - Merge         │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│    Enricher      │
│  - Add CVE Info  │
│  - Add CWE Info  │
│  - Add Refs      │
└────────┬─────────┘
         │
         ▼
Final Unified Results
{
  "scan_id": "...",
  "vulnerabilities": [...],
  "stats": {...},
  "scanners": [...]
}
```

## 5. Concurrent Execution Model

```
Orchestrator
     │
     ├─► Create Worker Pool (size: N)
     │
     ├─► Distribute Scanners
     │   │
     │   ├─► Worker 1 ──► Bandit ──────┐
     │   │                              │
     │   ├─► Worker 2 ──► pip-audit ───┤
     │   │                              ├─► Results Channel
     │   ├─► Worker 3 ──► detect-secrets┤
     │   │                              │
     │   └─► Worker N ──► Custom ───────┘
     │
     ├─► Wait for Completion
     │   ├─ Timeout: 5 minutes
     │   ├─ Cancel on Error (optional)
     │   └─ Collect Results
     │
     └─► Aggregate Results
         └─► Return Combined Output

Timeline:
0s    ├─ Start All Scanners
      │
2s    ├─ detect-secrets completes ✓
      │
5s    ├─ Bandit completes ✓
      │
8s    ├─ pip-audit completes ✓
      │
8s    └─ All Complete → Aggregate
```

## 6. Error Handling Flow

```
Scanner Execution
      │
      ├─► Try Execute
      │   │
      │   ├─ Success? ──► Return Results
      │   │
      │   └─ Error?
      │      │
      │      ├─► Check Error Type
      │      │   │
      │      │   ├─ Timeout ──────────┐
      │      │   ├─ Not Found ────────┤
      │      │   ├─ Permission ───────┤
      │      │   ├─ Invalid Output ───┤
      │      │   └─ Unknown ──────────┤
      │      │                        │
      │      ├─► Is Retryable? ◄──────┘
      │      │   │
      │      │   ├─ Yes ──► Retry (max 3)
      │      │   │          │
      │      │   │          ├─ Wait (exponential backoff)
      │      │   │          └─ Try Again
      │      │   │
      │      │   └─ No ───► Log Error
      │      │              │
      │      └──────────────┴─► Continue or Fail?
      │                         │
      │                         ├─ Continue ──► Next Scanner
      │                         │
      │                         └─ Fail ──────► Abort Scan
      │
      └─► Return Partial Results
```

## 7. Registry System

```
┌─────────────────────────────────────────────────────────┐
│                  Scanner Registry                        │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │           Registered Scanners Map              │    │
│  │                                                 │    │
│  │  "bandit" ──────► BanditScanner Instance       │    │
│  │  "pip-audit" ───► PipAuditScanner Instance     │    │
│  │  "detect-secrets"► DetectSecretsScanner        │    │
│  │  "custom" ───────► CustomScanner Instance      │    │
│  │                                                 │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
│  Operations:                                             │
│  ├─ Register(scanner) ──► Add to map                    │
│  ├─ Get(name) ──────────► Retrieve scanner              │
│  ├─ List() ─────────────► All scanners                  │
│  ├─ ListAvailable() ────► Only available                │
│  └─ GetForPath(path) ───► Applicable scanners           │
│                                                          │
└─────────────────────────────────────────────────────────┘

Query Flow:
User Request
     │
     ├─► GetScannersForPath("./project", ["code", "secrets"])
     │
     ├─► Registry Filters:
     │   ├─ Check IsAvailable()
     │   ├─ Match VulnType
     │   ├─ Check File Support
     │   └─ Apply User Filters
     │
     └─► Return: [BanditScanner, DetectSecretsScanner]
```

## 8. Configuration Cascade

```
┌─────────────────────────────────────────────────────────┐
│                   Configuration Layers                   │
└─────────────────────────────────────────────────────────┘

1. Default Config (Built-in)
   ├─ Timeout: 5 minutes
   ├─ MaxRetries: 2
   ├─ Concurrent: 3
   └─ MinSeverity: INFO
         │
         ▼
2. Config File (.btsg.yml)
   ├─ Timeout: 10 minutes  ← Override
   ├─ Exclude: [node_modules, vendor]
   └─ Scanners: [bandit, pip-audit]
         │
         ▼
3. Environment Variables
   ├─ BTSG_TIMEOUT=15m  ← Override
   └─ BTSG_VERBOSE=true
         │
         ▼
4. CLI Flags
   ├─ --timeout 20m  ← Override
   ├─ --types code,secrets
   └─ --verbose
         │
         ▼
Final Merged Config
   ├─ Timeout: 20 minutes
   ├─ MaxRetries: 2
   ├─ Concurrent: 3
   ├─ MinSeverity: INFO
   ├─ Exclude: [node_modules, vendor]
   ├─ Scanners: [bandit, pip-audit]
   ├─ Types: [code, secrets]
   └─ Verbose: true
```

## 9. Deduplication Algorithm

```
Input: Multiple Vulnerability Lists
  │
  ├─► Bandit Results: [V1, V2, V3]
  ├─► pip-audit Results: [V4, V5]
  └─► detect-secrets Results: [V6, V2*]  (* duplicate of V2)
  │
  ▼
Step 1: Generate Signatures
  │
  ├─► V1: hash(file + line + type + title)
  ├─► V2: hash(file + line + type + title)
  ├─► V3: hash(file + line + type + title)
  ├─► V4: hash(package + version + cve)
  ├─► V5: hash(package + version + cve)
  └─► V6: hash(file + line + type + title) ← Same as V2!
  │
  ▼
Step 2: Group by Signature
  │
  ├─► Signature_A: [V1]
  ├─► Signature_B: [V2, V6]  ← Duplicates!
  ├─► Signature_C: [V3]
  ├─► Signature_D: [V4]
  └─► Signature_E: [V5]
  │
  ▼
Step 3: Merge Duplicates
  │
  └─► For Signature_B:
      ├─ Keep V2 (first found)
      ├─ Merge metadata from V6
      │  ├─ Add scanner: "detect-secrets"
      │  ├─ Increase confidence
      │  └─ Merge context
      └─ Discard V6
  │
  ▼
Output: Deduplicated List
  └─► [V1, V2_merged, V3, V4, V5]
```

## 10. Output Format Selection

```
Scan Results
     │
     ├─► Format Selection
     │   │
     │   ├─ JSON (default)
     │   │  └─► Standard BTSG format
     │   │
     │   ├─ SARIF
     │   │  └─► Static Analysis Results Interchange Format
     │   │      ├─ Industry standard
     │   │      ├─ Tool integration
     │   │      └─ CI/CD compatible
     │   │
     │   ├─ JUnit XML
     │   │  └─► Test framework format
     │   │      └─ CI/CD integration
     │   │
     │   └─ Custom
     │      └─► User-defined template
     │
     └─► Output Destination
         │
         ├─ stdout (console)
         ├─ File (--output file.json)
         └─ Multiple (--output-formats json,sarif)
```

## Summary

These diagrams illustrate:
1. **Architecture**: How components interact
2. **Execution Flow**: Step-by-step process
3. **Data Pipeline**: Transformation stages
4. **Concurrency**: Parallel execution model
5. **Error Handling**: Retry and recovery logic
6. **Registry**: Scanner management
7. **Configuration**: Multi-layer config system
8. **Deduplication**: Finding merge algorithm
9. **Output**: Format selection and routing

Each diagram provides a visual representation of the scanner engine's operation, making it easier to understand and implement the system.