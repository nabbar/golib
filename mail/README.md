# Mail Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Test Coverage](https://img.shields.io/badge/coverage-85.5%25-brightgreen)](TESTING.md)

Comprehensive email solution for Go with SMTP client, template rendering, email composition, and rate-limited sending capabilities.

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
  - [smtp - SMTP Client](#smtp-subpackage)
  - [sender - Email Composition](#sender-subpackage)
  - [render - Template Rendering](#render-subpackage)
  - [queuer - Rate Limiting](#queuer-subpackage)
- [Integration Patterns](#integration-patterns)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The mail package provides a complete, production-ready email solution for Go applications. It combines four specialized subpackages to handle every aspect of email operations: SMTP connectivity, HTML/text template rendering, email composition with attachments, and rate-limited sending for high-volume scenarios.

### Design Philosophy

1. **Modular**: Each subpackage can be used independently or combined
2. **Production-Ready**: Thread-safe operations with comprehensive error handling
3. **Security-First**: TLS support, injection prevention, and certificate validation
4. **Developer-Friendly**: Intuitive APIs with method chaining and clear semantics
5. **Standards-Compliant**: Full RFC compliance (RFC 822, RFC 2045, RFC 5321)

---

## Key Features

**SMTP Operations**
- Multiple TLS modes (None, STARTTLS, Strict TLS)
- Automatic TLS fallback mechanisms
- Connection pooling and lifecycle management
- Health monitoring integration

**Email Composition**
- Multi-part content (HTML + plain text)
- File attachments and inline embeddings
- Custom headers and priority levels
- Transfer encoding options (Base64, Quoted-Printable)

**Template Rendering**
- Professional transactional email themes
- Variable substitution with `{{placeholders}}`
- Bidirectional text support (LTR/RTL)
- Dual output generation (HTML + text)

**Rate Limiting**
- Configurable throttling (e.g., 100 emails/minute)
- Context-aware cancellation
- Thread-safe concurrent sending
- Optional monitoring callbacks

---

## Installation

```bash
# Install all subpackages
go get github.com/nabbar/golib/mail/...

# Or install individually
go get github.com/nabbar/golib/mail/smtp
go get github.com/nabbar/golib/mail/sender
go get github.com/nabbar/golib/mail/render
go get github.com/nabbar/golib/mail/queuer
```

---

## Architecture

### Package Structure

```
mail/
├── smtp/                # SMTP client implementation
│   ├── config/         # Configuration parsing
│   └── tlsmode/        # TLS mode types
├── sender/             # Email composition and sending
├── render/             # HTML/text template rendering
└── queuer/             # Rate-limited SMTP wrapper
```

### Component Flow Diagram

```
┌────────────────────────────────────────────────────────┐
│                   Application Layer                    │
└──────┬──────────────┬──────────────┬──────────────┬────┘
       │              │              │              │
┌──────▼──────┐  ┌───▼──────┐  ┌───▼──────┐  ┌───▼──────┐
│   render    │  │  sender  │  │  queuer  │  │   smtp   │
│             │  │          │  │          │  │          │
│ Templates   │──▶ Compose  │──▶ Throttle │──▶ Send     │
│ Variables   │  │ Attach   │  │ Monitor  │  │ TLS      │
└─────────────┘  └──────────┘  └──────────┘  └──────────┘
```

### Integration Patterns

| Pattern | Components | Use Case |
|---------|------------|----------|
| **Simple** | `smtp` only | Basic transactional emails |
| **Composed** | `smtp` + `sender` | Emails with attachments |
| **Templated** | `render` + `sender` + `smtp` | Branded transactional emails |
| **High-Volume** | `queuer` + `smtp` | Bulk sending with rate limits |
| **Full-Stack** | All four | Complete email platform |

---

## Quick Start

### Simple Email (SMTP Only)

Send a basic email with SMTP:

```go
package main

import (
    "context"
    "strings"
    
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Create SMTP client
    client := smtp.New()
    
    // Configure
    cfg := smtp.NewConfig()
    cfg.SetHost("smtp.example.com")
    cfg.SetPort(587)
    cfg.SetUser("user@example.com")
    cfg.SetPass("password")
    cfg.SetTLSMode(smtp.TLSStartTLS)
    
    client.SetConfig(cfg)
    
    // Send email
    message := strings.NewReader("Subject: Test\r\n\r\nHello!")
    ctx := context.Background()
    
    err := client.Send(ctx, "from@example.com", []string{"to@example.com"}, message)
    if err != nil {
        panic(err)
    }
}
```

### Email with Attachments (Sender + SMTP)

Compose an email with attachments:

```go
package main

import (
    "context"
    "os"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Create sender
    mail := sender.New()
    
    // Configure email
    mail.SetFrom("sender@example.com", "Sender Name")
    mail.AddTo("recipient@example.com", "Recipient")
    mail.SetSubject("Document Attached")
    
    // Add content
    mail.SetBodyString(sender.TypeTextPlain, "Please find the document attached.")
    
    // Add attachment
    file, _ := os.Open("document.pdf")
    defer file.Close()
    mail.AttachReader("document.pdf", file)
    
    // Create SMTP client
    client := smtp.New()
    cfg := smtp.NewConfig()
    cfg.SetDSN("smtp://user:pass@smtp.example.com:587?tls=starttls")
    client.SetConfig(cfg)
    
    // Send
    ctx := context.Background()
    err := mail.Send(ctx, client)
    if err != nil {
        panic(err)
    }
}
```

### Templated Email (Render + Sender + SMTP)

Generate and send templated emails:

```go
package main

import (
    "context"
    
    "github.com/nabbar/golib/mail/render"
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Configure template renderer
    cfg := render.Config{
        Theme:     render.ThemeFlat,
        Direction: render.DirectionLTR,
        Product: render.Product{
            Name:      "MyApp",
            Link:      "https://myapp.com",
            Logo:      "https://myapp.com/logo.png",
            Copyright: "© 2024 MyApp Inc.",
        },
    }
    
    mailer, _ := cfg.NewMailer()
    
    // Create email body
    body := render.Body{
        Name: "John Doe",
        Intros: []string{
            "Welcome to MyApp!",
            "We're excited to have you on board.",
        },
        Actions: []render.Action{
            {
                Instructions: "Click below to get started:",
                Button: render.Button{
                    Color: "#3498db",
                    Text:  "Get Started",
                    Link:  "https://myapp.com/start",
                },
            },
        },
        Outros: []string{
            "Need help? Just reply to this email.",
        },
    }
    
    mailer.SetBody(body)
    
    // Generate HTML and text
    htmlContent, _ := mailer.GenerateHTML()
    textContent, _ := mailer.GeneratePlainText()
    
    // Compose email
    mail := sender.New()
    mail.SetFrom("no-reply@myapp.com", "MyApp")
    mail.AddTo("user@example.com", "")
    mail.SetSubject("Welcome to MyApp!")
    mail.SetBodyString(sender.TypeTextHTML, htmlContent)
    mail.SetBodyString(sender.TypeTextPlain, textContent)
    
    // Send via SMTP
    client := smtp.New()
    cfg := smtp.NewConfig()
    cfg.SetDSN("smtp://user:pass@smtp.example.com:587?tls=starttls")
    client.SetConfig(cfg)
    
    ctx := context.Background()
    err := mail.Send(ctx, client)
    if err != nil {
        panic(err)
    }
}
```

### Rate-Limited Sending (Queuer + SMTP)

Send emails with rate limiting:

```go
package main

import (
    "context"
    "time"
    
    "github.com/nabbar/golib/mail/queuer"
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Create SMTP client
    client := smtp.New()
    cfg := smtp.NewConfig()
    cfg.SetDSN("smtp://user:pass@smtp.example.com:587?tls=starttls")
    client.SetConfig(cfg)
    
    // Wrap with rate limiter (100 emails per minute)
    poolCfg := queuer.Config{
        Max:  100,
        Wait: time.Minute,
    }
    pooler := queuer.New(poolCfg, client)
    
    // Send multiple emails (automatically throttled)
    ctx := context.Background()
    
    for i := 0; i < 200; i++ {
        mail := sender.New()
        mail.SetFrom("sender@example.com", "Sender")
        mail.AddTo("user@example.com", "")
        mail.SetSubject("Bulk Email")
        mail.SetBodyString(sender.TypeTextPlain, "This is email #" + string(i))
        
        // Will automatically wait if rate limit reached
        err := mail.Send(ctx, pooler)
        if err != nil {
            panic(err)
        }
    }
}
```

---

## Performance

### Throughput Benchmarks

| Operation | Throughput | Notes |
|-----------|------------|-------|
| SMTP Send | ~50-100 emails/s | Network limited |
| Template Render | ~1000 renders/s | CPU limited |
| Email Compose | ~5000 emails/s | Memory limited |
| Rate Limiter | Configurable | Application controlled |

### Memory Efficiency

- **SMTP**: O(1) - Connection reuse with minimal buffering
- **Sender**: O(n) - Proportional to attachment sizes
- **Render**: O(1) - Template caching, no intermediate files
- **Queuer**: O(1) - Atomic counters, no message buffering

### Thread Safety

All subpackages are fully thread-safe:

- **SMTP**: Mutex-protected connection state
- **Sender**: Immutable message construction
- **Render**: Stateless template rendering
- **Queuer**: Atomic operations with mutex-protected counters

Verified with `go test -race` across 967 specs with zero data races.

---

## Use Cases

### Transactional Emails
**Components**: `render` + `sender` + `smtp`
- Welcome emails
- Password resets
- Order confirmations
- Invoice notifications

### Bulk Campaigns
**Components**: `queuer` + `sender` + `smtp`
- Newsletter distribution
- Marketing campaigns
- System notifications
- Scheduled reports

### User Communications
**Components**: All four subpackages
- Dynamic content per recipient
- Rate-limited to respect provider limits
- Professional branding via templates
- Attachment support (PDFs, reports)

### System Alerts
**Components**: `sender` + `smtp`
- Server monitoring alerts
- Error notifications
- Health check failures
- Security alerts

### SaaS Applications
**Components**: Full integration
- Multi-tenant email sending
- Template customization per tenant
- Rate limiting per account
- Usage tracking and monitoring

---

## Subpackages

### `smtp` Subpackage

Production-ready SMTP client with flexible TLS support.

**Features**
- Multiple TLS modes: None (port 25), STARTTLS (port 587), Strict TLS (port 465)
- Automatic TLS fallback from Strict to STARTTLS
- DSN-based configuration: `smtp://user:pass@host:port?tls=starttls`
- Security hardening: CR/LF injection prevention, certificate validation
- Health monitoring integration
- Thread-safe connection management

**Quick Example**
```go
client := smtp.New()
cfg := smtp.NewConfig()
cfg.SetDSN("smtp://user:pass@smtp.gmail.com:587?tls=starttls")
client.SetConfig(cfg)

ctx := context.Background()
message := strings.NewReader("Subject: Test\r\n\r\nBody")
err := client.Send(ctx, "from@example.com", []string{"to@example.com"}, message)
```

**Subpackages**
- `smtp/config` - Configuration parsing and validation
- `smtp/tlsmode` - TLS mode constants and utilities

**Documentation**: [smtp/README.md](smtp/README.md)  
**Test Coverage**: 80.6% (104 specs)

---

### `sender` Subpackage

High-level email composition with multi-part content and attachments.

**Features**
- Multi-part emails (HTML + plain text alternatives)
- File attachments with Base64/Quoted-Printable encoding
- Inline embedded files (images in HTML)
- Custom headers and priority levels (Low/Normal/High)
- Flexible addressing: From, Sender, ReplyTo, ReturnPath
- Recipient deduplication (To, CC, BCC)
- Configuration via JSON/YAML/TOML
- RFC-compliant message generation

**Quick Example**
```go
mail := sender.New()
mail.SetFrom("sender@example.com", "Sender Name")
mail.AddTo("recipient@example.com", "")
mail.SetSubject("Hello")
mail.SetBodyString(sender.TypeTextHTML, "<h1>Hello!</h1>")
mail.AttachFile("/path/to/file.pdf")

// Send via SMTP
err := mail.Send(ctx, smtpClient)
```

**Integration**: Works with any SMTP client implementing the `smtp.SMTP` interface

**Documentation**: [sender/README.md](sender/README.md)  
**Test Coverage**: 81.4% (252 specs)

---

### `render` Subpackage

Professional email template rendering with Hermes v2 engine.

**Features**
- Pre-designed themes: Default (classic) and Flat (modern)
- Variable substitution: `{{username}}` → `John Doe`
- Bidirectional text: LTR (left-to-right) and RTL (right-to-left)
- Rich content: Tables, action buttons, dictionaries, custom markdown
- Dual output: HTML and plain text from single template
- CSS inlining for email client compatibility
- Configuration validation with `go-playground/validator`
- Thread-safe cloning for concurrent rendering

**Quick Example**
```go
cfg := render.Config{
    Theme: render.ThemeFlat,
    Product: render.Product{
        Name: "MyApp",
        Link: "https://myapp.com",
    },
}

mailer, _ := cfg.NewMailer()
mailer.SetBody(render.Body{
    Name: "{{username}}",
    Intros: []string{"Welcome!"},
    Actions: []render.Action{{
        Button: render.Button{
            Text: "Get Started",
            Link: "https://myapp.com/start",
        },
    }},
})

htmlContent, _ := mailer.GenerateHTML()
textContent, _ := mailer.GeneratePlainText()

// Parse variables
data := map[string]string{"username": "John Doe"}
htmlContent = mailer.ParseData(htmlContent, data)
```

**Documentation**: [render/README.md](render/README.md)  
**Test Coverage**: 89.6% (123 specs)

---

### `queuer` Subpackage

Rate-limited SMTP wrapper for throttled email sending.

**Features**
- Configurable rate limiting (max emails per time window)
- Context-aware throttling with cancellation support
- Thread-safe atomic counters with mutex protection
- Zero-configuration mode (disable by setting Max or Wait to 0)
- Independent instances via cloning
- Optional monitoring callbacks for throttle events
- Transparent SMTP interface implementation
- Health check integration

**Quick Example**
```go
// Configure rate limiter (100 emails per minute)
cfg := queuer.Config{
    Max:  100,
    Wait: time.Minute,
}

// Wrap existing SMTP client
pooler := queuer.New(cfg, smtpClient)

// Send with automatic throttling
ctx := context.Background()
for i := 0; i < 200; i++ {
    // Automatically waits after 100 emails in the first minute
    err := pooler.Send(ctx, from, to, message)
}
```

**Throttling Algorithm**
```
1. Increment counter atomically
2. If counter ≤ Max: Send immediately
3. If counter > Max:
   a. Calculate wait time until window resets
   b. Check context cancellation
   c. Sleep until reset
   d. Reset counter and retry
```

**Documentation**: [queuer/README.md](queuer/README.md)  
**Test Coverage**: 90.8% (101 specs, 1 skipped)

---

## Integration Patterns

### Pattern 1: Simple Transactional

For basic email sending:

```go
// SMTP only
client := smtp.New()
// Configure and send
```

**Use Case**: System alerts, simple notifications

---

### Pattern 2: Composed Emails

For emails with attachments:

```go
// Sender + SMTP
mail := sender.New()
mail.AttachFile("invoice.pdf")
mail.Send(ctx, smtpClient)
```

**Use Case**: Invoices, reports, documents

---

### Pattern 3: Templated Transactional

For branded emails:

```go
// Render → Sender → SMTP
mailer := render.NewMailer(cfg)
html, _ := mailer.GenerateHTML()

mail := sender.New()
mail.SetBodyString(sender.TypeTextHTML, html)
mail.Send(ctx, smtpClient)
```

**Use Case**: Welcome emails, password resets

---

### Pattern 4: Rate-Limited Bulk

For high-volume sending:

```go
// Queuer wrapping SMTP
pooler := queuer.New(cfg, smtpClient)
mail.Send(ctx, pooler) // Automatically throttled
```

**Use Case**: Newsletters, campaigns

---

### Pattern 5: Full-Stack Solution

Complete email platform:

```go
// Render → Sender → Queuer → SMTP
mailer := render.NewMailer(renderCfg)
pooler := queuer.New(queueCfg, smtpClient)

for _, user := range users {
    mailer.SetBody(createBody(user))
    html, _ := mailer.GenerateHTML()
    
    mail := sender.New()
    mail.SetFrom("no-reply@app.com", "MyApp")
    mail.AddTo(user.Email, user.Name)
    mail.SetBodyString(sender.TypeTextHTML, html)
    
    mail.Send(ctx, pooler) // Templated, throttled, sent
}
```

**Use Case**: SaaS platforms, enterprise applications

---

## Best Practices

### Connection Management

```go
// ✅ Good: Reuse SMTP connections
client := smtp.New()
client.SetConfig(cfg)

for _, mail := range emails {
    mail.Send(ctx, client) // Reuses connection
}

// ❌ Bad: Create new client per email
for _, mail := range emails {
    client := smtp.New() // Unnecessary overhead
    client.SetConfig(cfg)
    mail.Send(ctx, client)
}
```

### Error Handling

```go
// ✅ Good: Check and wrap errors
err := mail.Send(ctx, client)
if err != nil {
    return fmt.Errorf("send email to %s: %w", recipient, err)
}

// ❌ Bad: Ignore errors
mail.Send(ctx, client) // Silent failure
```

### Context Usage

```go
// ✅ Good: Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := mail.Send(ctx, client)

// ✅ Good: Respect cancellation in queuer
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Will stop waiting during throttle
}()

err := mail.Send(ctx, pooler)
```

### Template Caching

```go
// ✅ Good: Reuse mailer instance
mailer := render.NewMailer(cfg)

for _, user := range users {
    mailer.SetBody(createBody(user))
    html, _ := mailer.GenerateHTML()
    // Use html...
}

// ❌ Bad: Create new mailer per email
for _, user := range users {
    mailer := render.NewMailer(cfg) // Unnecessary setup
    mailer.SetBody(createBody(user))
    html, _ := mailer.GenerateHTML()
}
```

### Rate Limiting

```go
// ✅ Good: Configure appropriate limits
cfg := queuer.Config{
    Max:  100,              // Gmail limit: 500/day
    Wait: time.Minute,      // Distribute evenly
}

// ❌ Bad: No rate limiting for bulk
// May get blocked by provider or marked as spam
for i := 0; i < 10000; i++ {
    mail.Send(ctx, client) // Too fast!
}
```

### Thread Safety

```go
// ✅ Good: Independent instances per goroutine
var wg sync.WaitGroup
for _, user := range users {
    wg.Add(1)
    go func(u User) {
        defer wg.Done()
        mail := sender.New() // Each goroutine gets own instance
        mail.Send(ctx, client) // Client is thread-safe
    }(user)
}
wg.Wait()

// ⚠️ Note: sender.Mail is not thread-safe
// Do not share across goroutines without synchronization
```

### TLS Configuration

```go
// ✅ Good: Use appropriate TLS mode
cfg.SetTLSMode(smtp.TLSStartTLS) // Port 587

// ✅ Good: Validate certificates in production
cfg.SetTLSSkip(false)

// ⚠️ Warning: Only skip validation in testing
cfg.SetTLSSkip(true) // DO NOT use in production
```

---

## Testing

**Test Suite**: 967 specs across all subpackages  
**Coverage**: 85.5% average, 90.8% maximum  
**Race Detection**: ✅ Zero data races

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

### Test Coverage by Subpackage

| Subpackage | Specs | Passed | Coverage | Duration |
|------------|-------|--------|----------|----------|
| `queuer` | 102 | 101 (1 skipped) | 90.8% | 8.6s |
| `smtp/config` | 222 | 222 | 92.7% | 0.2s |
| `smtp/tlsmode` | 165 | 165 | 98.8% | 0.04s |
| `smtp` | 104 | 104 | 80.6% | 26.8s |
| `render` | 123 | 123 | 89.6% | 2.0s |
| `sender` | 252 | 252 | 81.4% | 0.9s |
| **Total** | **967** | **966** | **85.5%** | **38.5s** |

### Race Detection Results

With `CGO_ENABLED=1 go test -race ./...`:
- **Duration**: ~45s (1.2x slower, expected)
- **Data Races**: 0
- **Status**: ✅ All thread-safety guarantees verified

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI to generate package implementation code**
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥80%)
- Follow existing code style and patterns

**Testing Requirements**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add benchmarks for performance-critical code

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep subpackage documentation synchronized
- Document breaking changes

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results (coverage + race detection)
- Update TESTING.md with new test information

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**SMTP Enhancements**
- OAuth2 authentication (Gmail, Outlook)
- Connection pooling for parallel sending
- Automatic retry with exponential backoff
- Support for more authentication mechanisms (CRAM-MD5, DIGEST-MD5)

**Sender Features**
- S/MIME encryption and signing
- DKIM signature support
- Template variables directly in sender
- Message ID tracking
- Bounce handling

**Render Improvements**
- Custom theme creation
- More built-in themes
- Component library for common elements
- Markdown parsing for custom sections
- Multi-language support

**Queuer Capabilities**
- Priority queues (high/normal/low)
- Persistent queue with disk backing
- Distributed rate limiting (Redis-backed)
- Failed message retry logic
- Queue monitoring and metrics

**General**
- Unified configuration for all subpackages
- Event hooks for lifecycle monitoring
- Plugin system for extensions
- CLI tool for testing email configuration
- Web UI for template preview

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

**Documentation**
- [GoDoc - mail](https://pkg.go.dev/github.com/nabbar/golib/mail)
- [GoDoc - smtp](https://pkg.go.dev/github.com/nabbar/golib/mail/smtp)
- [GoDoc - sender](https://pkg.go.dev/github.com/nabbar/golib/mail/sender)
- [GoDoc - render](https://pkg.go.dev/github.com/nabbar/golib/mail/render)
- [GoDoc - queuer](https://pkg.go.dev/github.com/nabbar/golib/mail/queuer)

**Subpackage Documentation**
- [SMTP Package](smtp/README.md)
- [Sender Package](sender/README.md)
- [Render Package](render/README.md)
- [Queuer Package](queuer/README.md)

**Testing**
- [Testing Guide](TESTING.md)
- [SMTP Tests](smtp/TESTING.md)
- [Sender Tests](sender/TESTING.md)
- [Render Tests](render/TESTING.md)

**External Resources**
- [RFC 5321 - SMTP](https://tools.ietf.org/html/rfc5321)
- [RFC 822 - Email Format](https://tools.ietf.org/html/rfc822)
- [RFC 2045 - MIME](https://tools.ietf.org/html/rfc2045)
- [Hermes Email Templates](https://github.com/matcornic/hermes)

**Support**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [Contributing Guide](../../CONTRIBUTING.md)
