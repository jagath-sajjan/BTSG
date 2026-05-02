# BTSG AI Explanation System - Flow Diagrams

## 1. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      BTSG Application                         │
│                                                               │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐    │
│  │  Scanner   │─────▶│ Analyzer   │─────▶│  Reporter  │    │
│  │  Module    │      │  Module    │      │  Module    │    │
│  └────────────┘      └──────┬─────┘      └────────────┘    │
│                             │                                │
└─────────────────────────────┼────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│              AI Explanation Engine                            │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Prompt     │  │     Cache    │  │  AI Provider │      │
│  │  Generator   │  │   Manager    │  │  (OpenAI/    │      │
│  │              │  │              │  │   Claude)    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Response   │  │  Validator   │  │   Template   │      │
│  │    Parser    │  │              │  │   Fallback   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────────────────────────────────────────────┘
```

## 2. Explanation Request Flow

```
START: User requests explanation
    │
    ▼
┌─────────────────────────┐
│ Receive Vulnerability   │
│ {id, tool, severity,    │
│  description, code}     │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Generate Cache Key      │
│ SHA256(tool+type+desc)  │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Check Cache             │
│ (Redis/Memory)          │
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Hit│           │Miss
      │           │
      ▼           ▼
┌──────────┐  ┌──────────────────┐
│ Return   │  │ Generate Prompt  │
│ Cached   │  │ - System prompt  │
│ Result   │  │ - User prompt    │
└──────────┘  │ - Context data   │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │ Call AI API      │
              │ - OpenAI GPT-4   │
              │ - Claude 3       │
              │ - Local LLM      │
              └────────┬─────────┘
                       │
                  ┌────┴────┐
                  │         │
              Success    Failure
                  │         │
                  ▼         ▼
          ┌──────────┐  ┌──────────┐
          │ Parse    │  │ Try      │
          │ JSON     │  │ Fallback │
          │ Response │  │ Provider │
          └────┬─────┘  └────┬─────┘
               │             │
               ▼             ▼
          ┌──────────┐  ┌──────────┐
          │ Validate │  │ Template │
          │ Schema   │  │ Based    │
          └────┬─────┘  └────┬─────┘
               │             │
               └──────┬──────┘
                      │
                      ▼
              ┌──────────────────┐
              │ Store in Cache   │
              │ TTL: 24 hours    │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │ Return to User   │
              │ - Simple explain │
              │ - Risk impact    │
              │ - Real example   │
              │ - Remediation    │
              └──────────────────┘
                       │
                       ▼
                     END
```

## 3. Prompt Generation Flow

```
Vulnerability Input
    │
    ▼
┌─────────────────────────┐
│ Extract Key Info        │
│ - Type (code/dep/secret)│
│ - Severity              │
│ - Language/Framework    │
│ - CWE/CVE               │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Select Prompt Template  │
│ Based on Type:          │
│ - Code → Code template  │
│ - Deps → CVE template   │
│ - Secret → Leak template│
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Inject Context Data     │
│ - Vulnerability details │
│ - Code snippet          │
│ - File location         │
│ - CWE/CVE info          │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Add Instructions        │
│ - Output format (JSON)  │
│ - Required sections     │
│ - Tone and style        │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Construct Final Prompt  │
│ System + User messages  │
└───────────┬─────────────┘
            │
            ▼
      Complete Prompt
```

## 4. AI Provider Integration Flow

```
Prompt Ready
    │
    ▼
┌─────────────────────────┐
│ Select Provider         │
│ - Primary: OpenAI       │
│ - Fallback: Claude      │
│ - Last resort: Template │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Prepare API Request     │
│ - Add API key           │
│ - Set parameters        │
│   * temperature: 0.7    │
│   * max_tokens: 2000    │
│   * model: gpt-4        │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Send HTTP Request       │
│ POST /v1/chat/completions│
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Success     Error
      │           │
      ▼           ▼
┌──────────┐  ┌──────────────┐
│ Extract  │  │ Check Error  │
│ Response │  │ Type:        │
│ Content  │  │ - Rate limit │
└────┬─────┘  │ - Timeout    │
      │        │ - Auth fail  │
      │        │ - Invalid    │
      │        └──────┬───────┘
      │               │
      │               ▼
      │        ┌──────────────┐
      │        │ Retry Logic  │
      │        │ - Exponential│
      │        │   backoff    │
      │        │ - Max 3 tries│
      │        └──────┬───────┘
      │               │
      │          ┌────┴────┐
      │          │         │
      │       Success   Failure
      │          │         │
      └──────────┘         ▼
            │        ┌──────────────┐
            │        │ Use Fallback │
            │        │ Provider or  │
            │        │ Template     │
            │        └──────┬───────┘
            │               │
            └───────────────┘
                    │
                    ▼
            Response Ready
```

## 5. Response Parsing Flow

```
AI Response Received
    │
    ▼
┌─────────────────────────┐
│ Extract Content         │
│ from API response       │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Detect Format           │
│ - JSON                  │
│ - Markdown              │
│ - Plain text            │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Parse JSON              │
│ json.Unmarshal()        │
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Success     Error
      │           │
      ▼           ▼
┌──────────┐  ┌──────────────┐
│ Validate │  │ Try Fallback │
│ Schema   │  │ Parsing:     │
└────┬─────┘  │ - Regex      │
      │        │ - Heuristics │
      │        └──────┬───────┘
      │               │
      ▼               ▼
┌──────────────────────────┐
│ Check Required Fields    │
│ - simple_explanation ✓   │
│ - risk_impact ✓          │
│ - remediation_steps ✓    │
└───────────┬──────────────┘
            │
      ┌─────┴─────┐
      │           │
   Valid      Invalid
      │           │
      ▼           ▼
┌──────────┐  ┌──────────────┐
│ Enrich   │  │ Use Partial  │
│ Data     │  │ Data +       │
│ - Add    │  │ Templates    │
│   metadata│  └──────────────┘
│ - Format │
│   text   │
└────┬─────┘
     │
     ▼
Parsed Explanation
```

## 6. Caching Strategy Flow

```
Explanation Request
    │
    ▼
┌─────────────────────────┐
│ Generate Cache Key      │
│ hash(tool+type+desc+cwe)│
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Check Cache Type        │
│ - Memory (fast)         │
│ - Redis (distributed)   │
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Memory      Redis
      │           │
      ▼           ▼
┌──────────┐  ┌──────────┐
│ Local    │  │ Network  │
│ HashMap  │  │ Call     │
│ Lookup   │  │ GET key  │
└────┬─────┘  └────┬─────┘
     │             │
     └──────┬──────┘
            │
      ┌─────┴─────┐
      │           │
    Found      Not Found
      │           │
      ▼           ▼
┌──────────┐  ┌──────────────┐
│ Check    │  │ Generate New │
│ Expiry   │  │ Explanation  │
└────┬─────┘  └──────┬───────┘
     │               │
┌────┴────┐          │
│         │          │
Valid  Expired       │
│         │          │
▼         ▼          ▼
Return  ┌──────────────────┐
Result  │ Store in Cache   │
        │ - Set TTL: 24h   │
        │ - Compress data  │
        │ - Update metrics │
        └────────┬─────────┘
                 │
                 ▼
          ┌──────────────┐
          │ Return Result│
          └──────────────┘
```

## 7. Error Handling Flow

```
Operation Start
    │
    ▼
┌─────────────────────────┐
│ Try Primary Operation   │
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Success     Error
      │           │
      ▼           ▼
   Return    ┌──────────────┐
   Result    │ Classify     │
             │ Error Type   │
             └──────┬───────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
   ┌────────┐  ┌────────┐  ┌────────┐
   │ Rate   │  │Timeout │  │ Auth   │
   │ Limit  │  │        │  │ Error  │
   └───┬────┘  └───┬────┘  └───┬────┘
       │           │           │
       ▼           ▼           ▼
   ┌────────┐  ┌────────┐  ┌────────┐
   │ Wait & │  │ Retry  │  │ Fail   │
   │ Retry  │  │ Once   │  │ Fast   │
   └───┬────┘  └───┬────┘  └───┬────┘
       │           │           │
       └───────────┼───────────┘
                   │
                   ▼
            ┌──────────────┐
            │ Try Fallback │
            │ Provider     │
            └──────┬───────┘
                   │
             ┌─────┴─────┐
             │           │
          Success    Failure
             │           │
             ▼           ▼
          Return    ┌──────────┐
          Result    │ Template │
                    │ Based    │
                    │ Fallback │
                    └────┬─────┘
                         │
                         ▼
                    ┌──────────┐
                    │ Log      │
                    │ Error &  │
                    │ Return   │
                    └──────────┘
```

## 8. Batch Processing Flow

```
Multiple Vulnerabilities
    │
    ▼
┌─────────────────────────┐
│ Group by Similarity     │
│ - Same tool             │
│ - Same type             │
│ - Similar description   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Create Worker Pool      │
│ Size: 5 workers         │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Distribute Work         │
│ - Round robin           │
│ - Load balancing        │
└───────────┬─────────────┘
            │
            ▼
    ┌───────┴───────┐
    │               │
    ▼               ▼
┌─────────┐     ┌─────────┐
│Worker 1 │ ... │Worker 5 │
│Process  │     │Process  │
│Vuln 1,6 │     │Vuln 5,10│
└────┬────┘     └────┬────┘
     │               │
     └───────┬───────┘
             │
             ▼
    ┌────────────────┐
    │ Collect Results│
    │ via Channel    │
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │ Wait for All   │
    │ (WaitGroup)    │
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │ Aggregate &    │
    │ Return Results │
    └────────────────┘
```

## 9. Template Fallback Flow

```
AI API Failed
    │
    ▼
┌─────────────────────────┐
│ Load Template Library   │
│ - Code templates        │
│ - Dependency templates  │
│ - Secret templates      │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Match Template          │
│ by CWE/Type             │
└───────────┬─────────────┘
            │
      ┌─────┴─────┐
      │           │
   Found      Not Found
      │           │
      ▼           ▼
┌──────────┐  ┌──────────┐
│ Load     │  │ Use      │
│ Specific │  │ Generic  │
│ Template │  │ Template │
└────┬─────┘  └────┬─────┘
     │             │
     └──────┬──────┘
            │
            ▼
┌─────────────────────────┐
│ Fill Template Variables │
│ - {{.Description}}      │
│ - {{.Severity}}         │
│ - {{.File}}             │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Add Standard Sections   │
│ - Risk scenarios        │
│ - Remediation steps     │
│ - References            │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Mark as Template-Based  │
│ confidence: 0.6         │
└───────────┬─────────────┘
            │
            ▼
      Template Result
```

## 10. Metrics Collection Flow

```
Every Operation
    │
    ▼
┌─────────────────────────┐
│ Record Start Time       │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Execute Operation       │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Record End Time         │
│ Calculate Duration      │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Update Metrics          │
│ - Total requests        │
│ - Cache hit/miss        │
│ - API success/failure   │
│ - Average latency       │
│ - Tokens used           │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Store in Metrics DB     │
│ - Prometheus            │
│ - InfluxDB              │
│ - CloudWatch            │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Trigger Alerts          │
│ if thresholds exceeded  │
│ - High error rate       │
│ - Slow response         │
│ - Cache miss rate       │
└─────────────────────────┘
```

## Summary

These flow diagrams illustrate:

1. **Request Flow** - End-to-end explanation generation
2. **Prompt Generation** - Template selection and context injection
3. **AI Integration** - Provider communication and retry logic
4. **Response Parsing** - JSON extraction and validation
5. **Caching** - Multi-tier cache strategy
6. **Error Handling** - Fallback mechanisms
7. **Batch Processing** - Concurrent explanation generation
8. **Template Fallback** - Offline explanation generation
9. **Metrics** - Performance monitoring

Each flow is designed for:
- **Reliability** - Multiple fallback options
- **Performance** - Caching and concurrency
- **Observability** - Comprehensive metrics
- **Maintainability** - Clear separation of concerns