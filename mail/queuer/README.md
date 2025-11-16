# Mail Queuer Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

High-performance, rate-limited SMTP client wrapper for Go with thread-safe operations, context-aware throttling, and transparent email sending control.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides production-ready rate limiting for SMTP email sending operations. It wraps any SMTP client implementing the `github.com/nabbar/golib/mail/smtp` interface with configurable throttling to prevent overwhelming mail servers and comply with provider sending limits.

### Design Philosophy

1. **Transparent**: Implements the same SMTP interface, acting as a drop-in replacement
2. **Thread-Safe**: All operations protected by atomic operations and mutex synchronization
3. **Context-Aware**: Respects context cancellation during throttling waits
4. **Independent**: Each pooler maintains its own quota and time window
5. **Observable**: Optional callbacks for monitoring throttle events

---

## Key Features

- **Rate Limiting**: Configurable maximum emails per time window (e.g., 100 emails/minute)
- **Thread-Safe Operations**: Atomic state management (`sync.Mutex`) for concurrent sending
- **Context Support**: Cancellation-aware throttling with `context.Context`
- **Zero Configuration**: Disable throttling by setting `Max` or `Wait` to zero
- **Independent Instances**: Clone poolers for isolated rate limits
- **Monitoring Callbacks**: Optional `FuncCaller` invoked on throttle events
- **SMTP Compatible**: Full implementation of `github.com/nabbar/golib/mail/smtp.SMTP` interface
- **Health Checking**: Built-in monitoring and connection verification

---

## Installation

```bash
go get github.com/nabbar/golib/mail/queuer
```

---

## Architecture

### Component Structure

```
mail/queuer/
├── interface.go         # Pooler interface and constructor
├── config.go            # Configuration and callback types
├── counter.go           # Rate limiting logic
├── model.go             # SMTP wrapper implementation
├── monitor.go           # Health check integration
└── error.go             # Custom error codes
```

### Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Application Code                     │
│              pooler.Send(ctx, from, to, msg)            │
└──────────────────────────┬──────────────────────────────┘
                           │
                  ┌────────▼────────┐
                  │     Pooler      │
                  │  (Throttling)   │
                  └────────┬────────┘
                           │
          ┌────────────────┴────────────────┐
          │                                 │
   ┌──────▼──────┐                 ┌───────▼────────┐
   │   Counter   │                 │   SMTP Client  │
   │ (Rate Limit)│                 │  (Send Email)  │
   └─────────────┘                 └────────────────┘
```

### Throttling Algorithm

```
Time Window (Wait duration):
├─────────────────────────────────────────┤
│ Email 1 | Email 2 | ... | Email Max     │ ← Quota available
└─────────────────────────────────────────┘

When Max reached:
├─────────────────────────────────────────┤ ← Wait until window expires
                            ▲
                         Sleep here
                      Call FuncCaller
```

| Component | Responsibility | Thread-Safe |
|-----------|---------------|-------------|
| **Pooler** | SMTP operations + throttling | ✅ |
| **Counter** | Quota tracking and enforcement | ✅ |
| **Config** | Rate limit parameters | Read-only |
| **FuncCaller** | Throttle event callbacks | ✅ |

---

## Performance

### Throughput Benchmarks

Measured performance with Ginkgo `gmeasure` on AMD64, Go 1.21:

| Scenario | Goroutines | Messages/sec | Notes |
|----------|------------|--------------|-------|
| No throttle | 1 | ~3,000 | Direct SMTP sending |
| No throttle | 32 | ~1,100 | Concurrent sending |
| Throttled (100/50ms) | 1 | ~1,300 | Rate limit enforced |
| Throttled (100/50ms) | 32 | ~1,100 | Concurrent throttling |

### Memory Efficiency

- **Constant Memory**: O(1) regardless of email count
- **Lightweight State**: ~100 bytes per pooler instance
- **No Buffering**: Direct passthrough to SMTP client
- **Example**: Send 1 million emails using ~100KB RAM for pooler state

### Thread Safety

All operations are thread-safe through:

- **Mutex Protection**: `sync.Mutex` for counter state
- **Atomic State**: Consistent quota management
- **Context Handling**: Safe cancellation during waits
- **Race-Free**: Verified with `go test -race` (zero races)

### Scalability Characteristics

```
Throughput vs. Goroutines (100 emails, 50ms window):

3000 msg/s ┤
          │ ╭──╮
2500 msg/s┤ │  ╰─╮
          │ │    ╰──╮
2000 msg/s┤╭╯       ╰──╮
          ││            ╰───╮
1500 msg/s┤│                ╰────╮
          │                      ╰────────
1000 msg/s┤
          └─┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──
           1  2  4  8  16 32 64 128 256 512
                    Concurrent Goroutines
```

*Optimal concurrency: 2-4 goroutines for typical SMTP operations*

---

## Use Cases

This package is designed for scenarios requiring controlled email sending:

**Marketing Campaigns**
- Respect provider rate limits (e.g., SendGrid: 100 emails/10 seconds)
- Prevent blacklisting from burst sending
- Monitor throttling events for capacity planning

**Transactional Emails**
- SaaS applications with multiple tenants
- Isolate rate limits per tenant using cloned poolers
- Context-aware cancellation for timed-out requests

**Bulk Email Processing**
- Parallel job processing with controlled throughput
- Independent workers with shared rate limit configuration
- Callback-based monitoring for observability

**SMTP Server Protection**
- Prevent overwhelming internal SMTP relays
- Graceful degradation under high load
- Health checking without consuming quota

**Multi-Provider Strategies**
- Different rate limits per provider
- Fallback logic with independent poolers
- Per-provider monitoring and alerting

---

## Quick Start

### Basic Rate Limiting

Send emails with automatic throttling:

```go
package main

import (
    "context"
    "time"
    
    "github.com/nabbar/golib/mail/queuer"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Configure rate limit: 100 emails per minute
    cfg := &queuer.Config{
        Max:  100,
        Wait: 1 * time.Minute,
    }
    
    // Create SMTP client
    smtpClient, _ := smtp.New(smtpConfig, tlsConfig)
    
    // Wrap with rate limiting
    pooler := queuer.New(cfg, smtpClient)
    defer pooler.Close()
    
    // Send emails - automatically throttled
    ctx := context.Background()
    for i := 0; i < 150; i++ {
        err := pooler.Send(ctx, "from@example.com", 
            []string{"to@example.com"}, emailMessage)
        if err != nil {
            panic(err)
        }
        // First 100: immediate
        // Next 50: waits until next time window
    }
}
```

### Monitoring Throttle Events

Track when rate limiting occurs:

```go
cfg := &queuer.Config{
    Max:  50,
    Wait: 10 * time.Second,
}

// Set callback for throttle events
cfg.SetFuncCaller(func() error {
    log.Printf("Rate limit reached, waiting for next window...")
    metrics.IncrementThrottleCounter()
    return nil
})

pooler := queuer.New(cfg, smtpClient)
```

### Context-Aware Sending

Cancel slow operations:

```go
func sendWithTimeout(pooler queuer.Pooler, email Email) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := pooler.Send(ctx, email.From, email.To, email.Message)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("send timeout: %w", err)
        }
        return err
    }
    return nil
}
```

### Parallel Sending with Shared Limit

Multiple workers respecting a shared rate limit:

```go
func parallelSend(pooler queuer.Pooler, emails []Email) error {
    var wg sync.WaitGroup
    errs := make(chan error, len(emails))
    
    for _, email := range emails {
        wg.Add(1)
        go func(e Email) {
            defer wg.Done()
            
            ctx := context.Background()
            if err := pooler.Send(ctx, e.From, e.To, e.Message); err != nil {
                errs <- err
            }
        }(email)
    }
    
    wg.Wait()
    close(errs)
    
    for err := range errs {
        if err != nil {
            return err // First error wins
        }
    }
    return nil
}
```

### Per-Tenant Rate Limiting

Isolate limits across tenants:

```go
type TenantEmailer struct {
    poolers map[string]queuer.Pooler
    mu      sync.RWMutex
}

func (te *TenantEmailer) GetPooler(tenantID string) queuer.Pooler {
    te.mu.RLock()
    pooler, exists := te.poolers[tenantID]
    te.mu.RUnlock()
    
    if !exists {
        // Create tenant-specific pooler
        cfg := &queuer.Config{
            Max:  tenant.GetLimit(tenantID),
            Wait: 1 * time.Minute,
        }
        pooler = queuer.New(cfg, smtpClient)
        
        te.mu.Lock()
        te.poolers[tenantID] = pooler
        te.mu.Unlock()
    }
    
    return pooler
}
```

---

## API Reference

### Pooler Interface

```go
type Pooler interface {
    Reset() error
    NewPooler() Pooler
    libsmtp.SMTP // Send, Client, Check, Close, Clone, UpdConfig, Monitor
}
```

**`Reset() error`**
- Resets the rate limiter counter to maximum quota
- Invokes `FuncCaller` if configured and throttling enabled
- Thread-safe, can be called during active sending
- Returns error only if `FuncCaller` returns error

**`NewPooler() Pooler`**
- Creates independent copy with same configuration
- Fresh quota and time window
- Clones underlying SMTP client
- Thread-safe, reads state under mutex protection

**SMTP Methods** (inherited from `libsmtp.SMTP`)
- `Send(ctx, from, to, data)` - Send email with throttling
- `Client(ctx)` - Get raw SMTP client (bypasses throttling)
- `Check(ctx)` - Health check (no throttling)
- `Close()` - Close SMTP connection
- `Clone()` - Alias for `NewPooler()`
- `UpdConfig(cfg, tlsConfig)` - Update SMTP settings
- `Monitor(ctx, version)` - Create monitoring instance

See [github.com/nabbar/golib/mail/smtp](https://pkg.go.dev/github.com/nabbar/golib/mail/smtp) for complete SMTP interface documentation.

### Config Struct

```go
type Config struct {
    Max  int           // Maximum operations per Wait duration
    Wait time.Duration // Time window duration
}

func (c *Config) SetFuncCaller(fct FuncCaller)
```

**Fields**

| Field | Type | Description | Special Values |
|-------|------|-------------|----------------|
| `Max` | `int` | Maximum emails per window | `≤0`: No limit |
| `Wait` | `time.Duration` | Time window duration | `≤0`: No limit |

**`SetFuncCaller(fct FuncCaller)`**
- Sets callback invoked on throttle events
- Called when: (1) Rate limit reached, (2) `Reset()` called
- Must be lightweight (holds mutex during execution)
- Return error to abort throttling operation

### FuncCaller Type

```go
type FuncCaller func() error
```

Callback function invoked during throttle events:

```go
cfg.SetFuncCaller(func() error {
    // Log event
    log.Printf("Throttle event at %v", time.Now())
    
    // Update metrics
    prometheus.ThrottleCounter.Inc()
    
    // Conditional error injection (testing)
    if testMode && shouldFail {
        return errors.New("throttle error")
    }
    
    return nil
})
```

**Use Cases**
- Logging and observability
- Metrics collection (Prometheus, Datadog)
- Alerting on sustained throttling
- Testing error scenarios

### Error Codes

```go
const (
    ErrorParamEmpty         // Required parameter missing/nil
    ErrorMailPooler         // Generic pooler error
    ErrorMailPoolerContext  // Context cancelled during throttling
)
```

**Error Handling**

```go
err := pooler.Send(ctx, from, to, message)
if err != nil {
    switch {
    case errors.Is(err, queuer.ErrorParamEmpty):
        // SMTP client not configured
    case errors.Is(err, queuer.ErrorMailPoolerContext):
        // Context cancelled during throttling
    default:
        // Other SMTP errors
    }
}
```

---

## Best Practices

**Choose Appropriate Limits**
```go
// ✅ Good: Respect provider limits
cfg := &queuer.Config{
    Max:  90,  // 10% under provider limit
    Wait: 1 * time.Minute,
}

// ❌ Bad: Exceeds provider limit
cfg := &queuer.Config{
    Max:  110, // Over limit, risk blacklisting
    Wait: 1 * time.Minute,
}
```

**Handle Context Cancellation**
```go
// ✅ Good: Graceful timeout
func sendEmail(pooler queuer.Pooler) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := pooler.Send(ctx, from, to, message)
    if errors.Is(err, queuer.ErrorMailPoolerContext) {
        return fmt.Errorf("send cancelled: %w", err)
    }
    return err
}

// ❌ Bad: Ignore cancellation
func sendEmailBad(pooler queuer.Pooler) error {
    ctx := context.Background() // Never cancelled
    return pooler.Send(ctx, from, to, message)
}
```

**Monitor Throttling**
```go
// ✅ Good: Observability
cfg.SetFuncCaller(func() error {
    metrics.throttleEvents.Inc()
    if metrics.throttleEvents.Value() > threshold {
        alert.Send("High throttling rate detected")
    }
    return nil
})

// ❌ Bad: Silent throttling
cfg := &queuer.Config{Max: 100, Wait: time.Minute}
// No visibility into throttling behavior
```

**Resource Cleanup**
```go
// ✅ Good: Always close
func process(cfg *queuer.Config, smtp libsmtp.SMTP) error {
    pooler := queuer.New(cfg, smtp)
    defer pooler.Close() // Closes SMTP connection
    
    return pooler.Send(ctx, from, to, message)
}

// ❌ Bad: Connection leak
func processBad(cfg *queuer.Config, smtp libsmtp.SMTP) error {
    pooler := queuer.New(cfg, smtp)
    return pooler.Send(ctx, from, to, message) // Connection left open
}
```

**Concurrent Safety**
```go
// ✅ Good: Shared pooler across goroutines
var pooler = queuer.New(cfg, smtp)

func worker(id int) {
    // Thread-safe concurrent access
    pooler.Send(ctx, from, to, message)
}

// ❌ Bad: Creating pooler per goroutine (inefficient)
func workerBad(id int) {
    pooler := queuer.New(cfg, smtp) // Unnecessary duplication
    defer pooler.Close()
    pooler.Send(ctx, from, to, message)
}
```

**Disable Throttling in Tests**
```go
// ✅ Good: Fast tests
func TestEmailSending(t *testing.T) {
    cfg := &queuer.Config{
        Max:  0, // Disable throttling
        Wait: 0,
    }
    pooler := queuer.New(cfg, mockSMTP)
    // Tests run at full speed
}
```

---

## Testing

**Test Suite**: 101 specs using Ginkgo v2 and Gomega (90.8% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**
- Counter throttling logic and time windows
- Pooler SMTP operations (send, check, client)
- Configuration scenarios (zero limits, callbacks)
- Concurrency and race condition testing
- Context cancellation handling
- Error scenarios and edge cases

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Context cancellation respected
- ✅ Mutex-protected state access

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `CGO_ENABLED=1 go test -race`
- Maintain or improve test coverage (≥90%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document thread-safety implications

**Testing**
- Write tests for all new features
- Test concurrent scenarios explicitly
- Verify thread safety with race detector
- Add benchmarks for performance-critical code

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results with race detection
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Advanced Throttling**
- Token bucket algorithm for burst allowance
- Sliding window rate limiting
- Dynamic rate adjustment based on server response
- Per-recipient rate limiting

**Monitoring**
- Prometheus metrics integration
- Grafana dashboard templates
- Real-time throttle statistics
- Queue depth monitoring

**Features**
- Retry logic with exponential backoff
- Priority queueing for urgent emails
- Circuit breaker pattern integration
- Multi-provider load balancing

**Performance**
- Batch sending optimization
- Connection pooling across poolers
- Adaptive throttling based on latency

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/mail/queuer)
- **SMTP Package**: [github.com/nabbar/golib/mail/smtp](https://pkg.go.dev/github.com/nabbar/golib/mail/smtp)
- **Sender Package**: [github.com/nabbar/golib/mail/sender](https://pkg.go.dev/github.com/nabbar/golib/mail/sender)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
