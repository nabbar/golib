# SMTP Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Production-ready SMTP client library for Go with flexible TLS modes, thread-safe operations, and integrated health monitoring.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Subpackages](#subpackages)
  - [config - Configuration Management](#config-subpackage)
  - [tlsmode - TLS Mode Handling](#tlsmode-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides a high-level SMTP client implementation that wraps the standard `net/smtp` package with enhanced features for production environments. It emphasizes security, reliability, and ease of use while supporting multiple TLS modes and providing comprehensive health monitoring capabilities.

### Design Philosophy

1. **Security-First**: TLS/STARTTLS support with certificate validation and injection attack prevention
2. **Production-Ready**: Thread-safe operations with health monitoring and error tracking
3. **Configuration-Driven**: Flexible DSN-based configuration with programmatic overrides
4. **Monitoring-Aware**: Built-in health checks and integration with monitoring systems
5. **Developer-Friendly**: Clear error messages, comprehensive documentation, and intuitive API

---

## Key Features

- **Multiple TLS Modes**: Plain SMTP (port 25), STARTTLS (port 587), and Strict TLS/SMTPS (port 465)
- **Automatic TLS Fallback**: Intelligent fallback from Strict TLS to STARTTLS if connection fails
- **Thread-Safe Operations**: Mutex-protected connection state with concurrent-safe methods
- **Health Monitoring**: Integrated health checks with `github.com/nabbar/golib/monitor/types`
- **DSN Configuration**: Flexible configuration via Data Source Name strings
- **Security Hardening**: 
  - CR/LF injection prevention in email addresses
  - TLS certificate verification (with optional skip for testing)
  - Server Name Indication (SNI) support
- **PLAIN Authentication**: Password-based authentication (CRAM-MD5 ready but commented)
- **Connection Management**: Automatic connection reuse and lifecycle management
- **Standard Interfaces**: Compatible with `io.WriterTo` for email content streaming

---

## Installation

```bash
go get github.com/nabbar/golib/mail/smtp
```

**Dependencies**:
- Go 1.18 or later
- `github.com/nabbar/golib/errors` - Error management
- `github.com/nabbar/golib/mail/smtp/config` - Configuration handling
- `github.com/nabbar/golib/mail/smtp/tlsmode` - TLS mode types
- `github.com/nabbar/golib/network/protocol` - Network protocol types
- `github.com/nabbar/golib/monitor/types` - Health monitoring (optional)
- `github.com/nabbar/golib/certificates` - TLS configuration (optional)

---

## Architecture

### Package Structure

The package is organized into focused subpackages with clear separation of concerns:

```
smtp/
├── config/              # Configuration parsing and validation
│   ├── doc.go          # Comprehensive package documentation
│   ├── interface.go    # SMTP and Config interfaces
│   ├── model.go        # Configuration implementation
│   └── error.go        # Configuration-specific errors
├── tlsmode/            # TLS mode constants and utilities
│   ├── doc.go          # TLS mode documentation
│   ├── interface.go    # TLSMode interface
│   ├── model.go        # TLS mode implementation
│   └── encode.go       # JSON/YAML/TOML encoding
├── interface.go        # Main SMTP client interface
├── model.go            # Internal client implementation
├── client.go           # Public client methods
├── dial.go             # Connection and authentication
├── monitor.go          # Health monitoring integration
└── error.go            # Error codes and messages
```

### Component Overview

```
┌───────────────────────────────────────────────────────┐
│                   SMTP Client                         │
│  Send(), Check(), Monitor(), Clone(), Close()         │
└──────────────┬─────────────┬─────────────┬────────────┘
               │             │             │
      ┌────────▼─────┐  ┌────▼─────┐  ┌────▼────────┐
      │   config     │  │ tlsmode  │  │  monitor    │
      │              │  │          │  │             │
      │ DSN parsing  │  │ TLS modes│  │ Health check│
      │ Validation   │  │ Encoding │  │ Integration │
      └──────────────┘  └──────────┘  └─────────────┘
```

| Component | Purpose | Coverage | Thread-Safe |
|-----------|---------|----------|-------------|
| **`smtp`** | Main client implementation | 86.6% | ✅ |
| **`config`** | Configuration management | 96.7% | ✅ Read-only |
| **`tlsmode`** | TLS mode handling | 98.8% | ✅ Immutable |
| **`monitor`** | Health monitoring | 100% | ✅ |

### TLS Modes

The library supports three TLS connection modes:

**TLSNone (0)** - Plain SMTP
- No encryption
- Typically port 25
- Suitable for localhost/testing only
- Opportunistic STARTTLS upgrade if available

**TLSStartTLS (1)** - SMTP with STARTTLS
- Starts as plain connection
- Upgrades to TLS via STARTTLS command
- Typically port 587 (submission)
- Recommended for client-to-server communication

**TLSStrictTLS (2)** - Direct TLS (SMTPS)
- TLS from connection start
- Typically port 465
- Also known as SMTPS
- Fallback to STARTTLS if connection fails

### Connection Flow

```
Client Request
      │
      ▼
┌─────────────┐
│ Parse Config│
└──────┬──────┘
       │
       ▼
┌──────────────────┐     No      ┌──────────────┐
│ TLS Mode Check   │────────────▶│ Plain SMTP   │
└──────┬───────────┘             │ (port 25)    │
       │ Yes                     └──────────────┘
       │
       ▼
┌──────────────────┐
│ Strict TLS?      │
└──────┬───────────┘
       │ Yes
       ▼
┌──────────────────┐     Failed   ┌──────────────┐
│ TLS Handshake    │─────────────▶│ Retry with   │
│ (port 465)       │              │ STARTTLS     │
└──────┬───────────┘              └──────────────┘
       │ Success
       ▼
┌──────────────────┐
│ SMTP Handshake   │
│ (EHLO)           │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Authentication   │
│ (if configured)  │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Ready for MAIL   │
│ Commands         │
└──────────────────┘
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/nabbar/golib/mail/smtp"
    "github.com/nabbar/golib/mail/smtp/config"
    "github.com/nabbar/golib/mail/smtp/tlsmode"
)

func main() {
    // Create configuration
    cfg, err := config.New(config.ConfigModel{
        DSN: "user:password@tcp(smtp.gmail.com:587)/starttls",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Create SMTP client
    client, err := smtp.New(cfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Send email
    email := &MyEmail{} // implements io.WriterTo
    err = client.Send(ctx, 
        "sender@example.com",
        []string{"recipient@example.com"},
        email,
    )
    if err != nil {
        log.Fatal(err)
    }
}

// MyEmail implements io.WriterTo
type MyEmail struct {
    content string
}

func (e *MyEmail) WriteTo(w io.Writer) (int64, error) {
    n, err := w.Write([]byte(e.content))
    return int64(n), err
}
```

### Programmatic Configuration

```go
// Create with default DSN
cfg, _ := config.New(config.ConfigModel{
    DSN: "tcp(localhost:25)/",
})

// Modify configuration
cfg.SetHost("smtp.example.com")
cfg.SetPort(587)
cfg.SetUser("user@example.com")
cfg.SetPass("secretpassword")
cfg.SetTlsMode(tlsmode.TLSStartTLS)
cfg.ForceTLSSkipVerify(false) // Verify certificates

// Create client with custom TLS config
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    ServerName: "smtp.example.com",
}

client, err := smtp.New(cfg, tlsConfig)
```

### Health Monitoring

```go
import (
    "github.com/nabbar/golib/version"
)

// Create version info
vrs := version.NewVersion(
    version.License_MIT,
    "myapp",
    "My Application",
    "2024-01-01",
    "abc123",
    "1.0.0",
    "maintainer@example.com",
    "",
    struct{}{},
    0,
)

// Create monitor
monitor, err := client.Monitor(ctx, vrs)
if err != nil {
    log.Fatal(err)
}
defer monitor.Stop(ctx)

// Monitor will perform periodic health checks automatically
// Check current health status
if err := client.Check(ctx); err != nil {
    log.Printf("SMTP health check failed: %v", err)
}
```

### Configuration via DSN

```go
// Basic SMTP (no TLS, port 25)
cfg, _ := config.New(config.ConfigModel{
    DSN: "tcp(localhost:25)/",
})

// STARTTLS (port 587) with authentication
cfg, _ := config.New(config.ConfigModel{
    DSN: "user:password@tcp(smtp.gmail.com:587)/starttls",
})

// Strict TLS (port 465) with SNI
cfg, _ := config.New(config.ConfigModel{
    DSN: "user:password@tcp(mail.example.com:465)/tls?ServerName=smtp.example.com",
})

// Skip certificate verification (TESTING ONLY!)
cfg, _ := config.New(config.ConfigModel{
    DSN: "tcp(smtp.test.local:465)/tls?SkipVerify=true",
})
```

---

## Performance

### Memory Efficiency

The SMTP client maintains minimal memory footprint:

- **Connection Reuse**: Single connection per client instance
- **Lazy Connection**: Connection established only when needed
- **Streaming Email**: Content streamed via `io.WriterTo` interface
- **No Buffering**: Direct passthrough to network socket
- **Example**: Send 100MB email using ~5MB RAM

### Thread Safety

All public methods are thread-safe through:

- **Mutex Protection**: `sync.Mutex` guards connection state
- **Atomic State**: Connection validity checks are atomic
- **Independent Clones**: `Clone()` creates isolated instances
- **Concurrent Safe**: Multiple goroutines can use separate clients

### Throughput

Based on integration tests:

| Operation | Duration | Notes |
|-----------|----------|-------|
| Health Check | 50-100ms | Connect + NOOP + Disconnect |
| Send Email | 150-300ms | Includes authentication and transmission |
| TLS Handshake | 100-200ms | One-time cost per connection |
| Authentication | 50-100ms | PLAIN auth over TLS |

**Network Characteristics**:
- Local SMTP server: ~5ms latency
- Remote SMTP server: 50-200ms latency (depends on network)
- Concurrent sends: Linear scaling with independent clients

---

## Use Cases

### 1. **Transactional Emails**
Send order confirmations, password resets, and notifications.

```go
func SendOrderConfirmation(orderID string, to string) error {
    client, _ := smtp.New(cfg, nil)
    defer client.Close()
    
    email := BuildOrderEmail(orderID)
    return client.Send(ctx, "orders@shop.com", []string{to}, email)
}
```

### 2. **Batch Email Processing**
Send marketing campaigns or newsletters efficiently.

```go
func SendNewsletter(recipients []string) error {
    client, _ := smtp.New(cfg, nil)
    defer client.Close()
    
    for _, recipient := range recipients {
        email := BuildNewsletter(recipient)
        if err := client.Send(ctx, "news@company.com", []string{recipient}, email); err != nil {
            log.Printf("Failed to send to %s: %v", recipient, err)
            continue
        }
        time.Sleep(100 * time.Millisecond) // Rate limiting
    }
    return nil
}
```

### 3. **Application Monitoring**
Monitor SMTP server health as part of system monitoring.

```go
func HealthCheckEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    if err := smtpClient.Check(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error": err.Error(),
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}
```

### 4. **Microservice Communication**
Use SMTP as a notification channel between services.

```go
type NotificationService struct {
    smtp smtp.SMTP
}

func (s *NotificationService) Notify(event Event) error {
    email := BuildNotificationEmail(event)
    return s.smtp.Send(
        context.Background(),
        "system@company.com",
        event.Recipients,
        email,
    )
}
```

### 5. **Testing and Development**
Use local SMTP servers for development and testing.

```go
func TestMode() smtp.SMTP {
    cfg, _ := config.New(config.ConfigModel{
        DSN: "tcp(localhost:1025)/", // Mailhog or similar
    })
    client, _ := smtp.New(cfg, nil)
    return client
}
```

---

## Subpackages

### config Subpackage

**Purpose**: SMTP configuration parsing, validation, and management

**Key Features**:
- DSN parsing with comprehensive format support
- Network protocol handling (tcp, tcp4, tcp6)
- TLS configuration integration
- Thread-safe read access
- Struct validation with detailed error messages

**Documentation**: See [config/doc.go](config/doc.go) for detailed configuration options and examples, including comprehensive DSN format specifications and integration examples.

**Coverage**: 96.7%

### tlsmode Subpackage

**Purpose**: TLS mode type definitions and encoding/decoding

**Key Features**:
- Three TLS modes (None, STARTTLS, Strict TLS)
- Multiple encoding formats (JSON, YAML, TOML, CBOR)
- String and numeric parsing
- Viper integration via decode hooks
- Immutable and thread-safe

**Documentation**: See [tlsmode/doc.go](tlsmode/doc.go) for TLS mode usage, encoding examples, and comprehensive integration guides.

**Coverage**: 98.8%

---

## Best Practices

### Security

**1. Always Use TLS for Authentication**
```go
// ❌ INSECURE - Plain auth without TLS
cfg.SetTlsMode(tlsmode.TLSNone)
cfg.SetUser("user")
cfg.SetPass("password")

// ✅ SECURE - Auth over TLS
cfg.SetTlsMode(tlsmode.TLSStartTLS)
cfg.SetUser("user")
cfg.SetPass("password")
```

**2. Verify Certificates in Production**
```go
// ❌ INSECURE - Skip certificate verification
cfg.ForceTLSSkipVerify(true)

// ✅ SECURE - Verify certificates
cfg.ForceTLSSkipVerify(false)
```

**3. Use SNI for Virtual Hosting**
```go
// When connecting to IP address but server expects specific hostname
cfg.SetTlSServerName("smtp.example.com")
```

**4. Prevent Injection Attacks**
The library automatically validates email addresses for CR/LF characters, but you should still sanitize input:

```go
// The library will reject these automatically
client.Send(ctx, "sender\n@evil.com", ...) // ❌ Rejected
client.Send(ctx, "sender@example.com", []string{"victim\r\n@evil.com"}, ...) // ❌ Rejected
```

### Performance

**1. Reuse Client Instances**
```go
// ❌ Creates new connection each time
for _, email := range emails {
    client, _ := smtp.New(cfg, nil)
    client.Send(ctx, from, []string{to}, email)
    client.Close()
}

// ✅ Reuses connection
client, _ := smtp.New(cfg, nil)
defer client.Close()
for _, email := range emails {
    client.Send(ctx, from, []string{to}, email)
}
```

**2. Use Context Timeouts**
```go
// ✅ Prevent hanging connections
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := client.Send(ctx, from, to, email)
```

**3. Handle Errors Appropriately**
```go
if err := client.Send(ctx, from, to, email); err != nil {
    // Check error type for appropriate action
    if strings.Contains(err.Error(), "authentication") {
        // Credential issue
    } else if strings.Contains(err.Error(), "dial") {
        // Network issue
    }
    // Log and handle accordingly
}
```

### Monitoring

**1. Implement Health Checks**
```go
// Periodic health check
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        if err := client.Check(ctx); err != nil {
            // Alert monitoring system
            alerting.Send("SMTP server unhealthy: " + err.Error())
        }
    }
}()
```

**2. Use Monitoring Integration**
```go
monitor, err := client.Monitor(ctx, versionInfo)
if err == nil {
    // Monitor automatically performs periodic checks
    defer monitor.Stop(ctx)
}
```

### Configuration Management

**1. Use Environment Variables**
```go
dsn := os.Getenv("SMTP_DSN")
if dsn == "" {
    dsn = "tcp(localhost:25)/" // Fallback
}
cfg, _ := config.New(config.ConfigModel{DSN: dsn})
```

**2. Separate Credentials**
```go
// Load credentials from secure store
user := secretManager.Get("smtp_user")
pass := secretManager.Get("smtp_password")

cfg.SetUser(user)
cfg.SetPass(pass)
```

---

## Testing

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

**Quick Summary**:
- Total Specs: 379 (104 SMTP + 110 Config + 165 TLSMode)
- Overall Coverage: ~93.4%
- Zero Data Races
- Execution Time: ~27s (without race), ~40s (with race)

---

## Contributing

We welcome contributions! Please follow these guidelines:

### Code Contributions

**AI Usage Policy**: AI tools may assist with testing, documentation, and bug fixes, but **must not be used** for implementing core package functionality. All code must be written and reviewed by humans to ensure quality, security, and maintainability.

**Process**:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests for your changes
4. Ensure all tests pass with race detection: `CGO_ENABLED=1 go test -race ./...`
5. Update documentation (GoDoc comments and README if needed)
6. Submit a pull request

### Testing and Documentation

AI assistance is **encouraged** for:
- Writing comprehensive tests
- Improving documentation clarity
- Identifying edge cases
- Bug reproduction and analysis

**Requirements**:
- All public APIs must have GoDoc comments
- Complex logic should have inline comments
- Tests must achieve ≥80% coverage
- Zero data races

### Bug Reports

Please include:
- Go version (`go version`)
- Package version
- Minimal reproduction code
- Expected vs actual behavior
- Error messages and logs

---

## Future Enhancements

Potential improvements for future releases:

### Authentication
- [ ] OAuth2 support for Gmail and Office365
- [ ] CRAM-MD5 authentication (currently commented)
- [ ] SCRAM-SHA-256 authentication
- [ ] External SASL mechanisms

### Features
- [ ] Connection pooling for high-throughput scenarios
- [ ] Retry logic with exponential backoff
- [ ] Email template system integration
- [ ] HTML email helpers
- [ ] Attachment handling utilities
- [ ] DKIM signing support
- [ ] SPF validation helpers

### Monitoring
- [ ] Prometheus metrics export
- [ ] Detailed timing metrics per operation
- [ ] Connection pool statistics
- [ ] Email queue monitoring

### Configuration
- [ ] YAML/TOML configuration file support
- [ ] Configuration hot-reload
- [ ] Multiple profile support
- [ ] Credential rotation support

**Note**: These are potential enhancements based on community feedback. Actual implementation depends on demand and maintenance capacity.

---

## License

MIT License - see [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2024 Nicolas JUHEL

---

## Disclaimer

**AI Assistance**: Per EU AI Act Article 50.4, this package's development uses AI assistance for testing, documentation, and bug analysis under human supervision. Core functionality is human-written and reviewed.

---

## Additional Resources

- **Go SMTP RFC**: [RFC 5321](https://tools.ietf.org/html/rfc5321) - Simple Mail Transfer Protocol
- **STARTTLS**: [RFC 3207](https://tools.ietf.org/html/rfc3207) - SMTP Service Extension for Secure SMTP
- **SMTP Authentication**: [RFC 4954](https://tools.ietf.org/html/rfc4954) - SMTP Service Extension for Authentication
- **TLS Best Practices**: [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- **Email Standards**: [RFC 5322](https://tools.ietf.org/html/rfc5322) - Internet Message Format

---

**Questions?** Open an issue on GitHub or consult the [examples](examples/) directory.
