# Mail Sender Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Test Coverage](https://img.shields.io/badge/coverage-81.4%25-brightgreen)](TESTING.md)

High-level email composition and sending library for Go with SMTP integration, multi-part content, file attachments, and configuration-based setup.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Integration](#integration)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides a production-ready email composition and delivery system for Go applications. It emphasizes ease of use, flexibility, and robustness while supporting all standard email features including multi-part content, attachments, custom headers, and priority management.

### Design Philosophy

1. **Developer-Friendly**: Intuitive API with method chaining and clear semantics
2. **Configuration-First**: JSON/YAML/TOML support for declarative email setup
3. **Type-Safe**: Strong typing with validation at compile and runtime
4. **SMTP Integration**: Seamless integration with the `mail/smtp` package
5. **Standard Compliance**: RFC-compliant email generation (RFC 822, RFC 2045, RFC 2156)

---

## Key Features

- **Multi-Part Content**: Plain text and HTML alternatives in single email
- **File Attachments**: Regular attachments and inline embedded files (e.g., images in HTML)
- **Flexible Addressing**: From, Sender, ReplyTo, ReturnPath with intelligent fallbacks
- **Recipient Management**: To, CC, BCC with automatic deduplication
- **Custom Headers**: Full control over email headers
- **Transfer Encoding**: None, Binary, Base64, Quoted-Printable
- **Priority Levels**: Normal, Low, High with multi-client compatibility
- **Configuration Support**: JSON, YAML, TOML, Viper mapstructure tags
- **Validation**: Comprehensive validation using `go-playground/validator`
- **Error Handling**: Structured error codes with parent error wrapping
- **Stream-Friendly**: io.Reader/io.Writer interfaces for efficient memory usage

---

## Installation

```bash
go get github.com/nabbar/golib/mail/sender
```

**Dependencies:**
- `github.com/nabbar/golib/errors` - Structured error handling
- `github.com/nabbar/golib/mail/smtp` - SMTP client (for sending)
- `github.com/nabbar/golib/file/progress` - File operations
- `github.com/go-playground/validator/v10` - Configuration validation
- `github.com/wneessen/go-mail` - MIME message generation

---

## Architecture

### Package Structure

```
mail/sender/
├── interface.go          # Mail and Email interfaces
├── mail.go              # Mail implementation
├── email.go             # Email address management
├── sender.go            # Sender interface and implementation
├── config.go            # Configuration structs and validation
├── body.go              # Email body parts
├── file.go              # File attachments
├── encoding.go          # Transfer encoding types
├── priority.go          # Priority management
├── contentType.go       # Content type definitions
├── recipientType.go     # Recipient categories
└── error.go             # Error codes and messages
```

### Component Overview

```
┌─────────────────────────────────────────────────────┐
│                   Mail Interface                    │
│  Compose, Configure, Manage Content                 │
└──────────────┬────────────────┬─────────────────────┘
               │                │
      ┌────────▼───────┐  ┌─────▼────────┐
      │     Email      │  │    Sender    │
      │   Addresses    │  │ SMTP Delivery│
      │  To/CC/BCC     │  │  Send/Close  │
      └────────────────┘  └──────────────┘
               │                │
      ┌────────▼────────────────▼─────────┐
      │         Config (Optional)         │
      │   JSON/YAML/TOML Configuration    │
      └───────────────────────────────────┘
```

| Component | Purpose | Memory | Thread-Safe |
|-----------|---------|--------|-------------|
| **`Mail`** | Email composition and configuration | O(1) per email | ⚠️ Clone for concurrency |
| **`Email`** | Address management with fallbacks | O(n) recipients | ⚠️ Part of Mail |
| **`Sender`** | SMTP transmission preparation | O(1) | ⚠️ One-time use |
| **`Config`** | Declarative email setup | O(1) | ✅ Immutable |

### Email Composition Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│  Create  │────▶│ Configure│────▶│ Validate │────▶│  Send   │
│  Mail    │     │  Content │     │  Email   │     │  SMTP   │
└──────────┘     └──────────┘     └──────────┘     └─────────┘
     │                 │                 │               │
   New()          SetSubject()      Sender()       Send(ctx)
                  SetBody()                        SendClose(ctx)
                  AddAttachment()
                  Email().SetFrom()
```

---

## Performance

### Benchmark Results

Tests run on Go 1.21, AMD64 architecture:

| Operation | Throughput | Memory | Notes |
|-----------|------------|--------|-------|
| Mail Creation | ~200 µs | O(1) | Empty mail object |
| Set Properties | <1 µs | O(1) | Subject, encoding, priority |
| Add Recipients | <1 µs | O(n) | Per recipient |
| Add Headers | <1 µs | O(1) | Per header |
| Add Body | <1 µs | O(1) | Body content |
| Add Attachment | 100-300 µs | O(1) | File opening |
| Clone Mail | <1 µs | O(n) | Full deep copy |
| Create Sender | 100-200 µs | O(1) | MIME message prep |
| Config Validation | 100 µs | O(1) | Field validation |
| Config NewMailer | <1 µs | O(1) | Mail from config |
| Encoding Parse | <1 µs | O(1) | String to enum |
| Priority Parse | <1 µs | O(1) | String to enum |

### Test Results

- **Total Tests**: 252 specs
- **Test Duration**: ~0.8s (normal), ~1.5s (with race detector)
- **Coverage**: 81.4% of statements
- **Race Detection**: Zero data races detected

### Memory Characteristics

- **Constant Overhead**: ~2KB per Mail instance
- **Streaming Body**: Body content not loaded into memory
- **File Streaming**: Attachments streamed from io.Reader
- **No Buffering**: Direct passthrough to SMTP client

---

## Use Cases

This package is designed for applications requiring robust email functionality:

**Transactional Emails**
- User registration confirmations
- Password reset emails
- Order confirmations and invoices
- Account notifications

**Notification Systems**
- Application alerts and monitoring
- CI/CD pipeline notifications
- System health reports
- Error reporting

**Newsletter and Marketing**
- HTML newsletters with inline images
- Multi-recipient campaigns (using BCC for privacy)
- Personalized email templates
- Tracking pixels (inline attachments)

**Report Distribution**
- Automated report generation
- PDF invoice attachments
- Data export emails
- Log file delivery

**Support and Communication**
- Support ticket responses
- Customer service emails
- Team collaboration notifications
- Document sharing

---

## Quick Start

### Simple Email

Send a basic plain-text email:

```go
package main

import (
    "context"
    "io"
    "strings"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Create email
    mail := sender.New()
    mail.SetSubject("Hello World")
    mail.SetCharset("UTF-8")
    mail.SetEncoding(sender.EncodingBase64)
    
    // Set sender and recipient
    mail.Email().SetFrom("noreply@example.com")
    mail.Email().AddRecipients(sender.RecipientTo, "user@example.com")
    
    // Add body
    body := io.NopCloser(strings.NewReader("This is a test email."))
    mail.SetBody(sender.ContentPlainText, body)
    
    // Create sender
    snd, _ := mail.Sender()
    defer snd.Close()
    
    // Send via SMTP
    smtpClient, _ := smtp.New(smtpConfig, nil)
    err := snd.Send(context.Background(), smtpClient)
    if err != nil {
        panic(err)
    }
}
```

### HTML Email with Attachment

Send an HTML email with file attachment:

```go
package main

import (
    "context"
    "io"
    "os"
    "strings"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    mail := sender.New()
    mail.SetSubject("Monthly Report")
    mail.SetCharset("UTF-8")
    mail.SetEncoding(sender.EncodingBase64)
    mail.SetPriority(sender.PriorityHigh)
    
    // Addresses
    mail.Email().SetFrom("reports@company.com")
    mail.Email().AddRecipients(sender.RecipientTo, "manager@company.com")
    mail.Email().AddRecipients(sender.RecipientCC, "team@company.com")
    
    // Multi-part body
    plainBody := io.NopCloser(strings.NewReader("Please see the attached report."))
    htmlBody := io.NopCloser(strings.NewReader("<p>Please see the <b>attached report</b>.</p>"))
    
    mail.SetBody(sender.ContentPlainText, plainBody)
    mail.AddBody(sender.ContentHTML, htmlBody)
    
    // Add PDF attachment
    file, _ := os.Open("/path/to/report.pdf")
    mail.AddAttachment("Monthly_Report.pdf", "application/pdf", file, false)
    
    // Send
    snd, _ := mail.Sender()
    defer snd.Close()
    
    smtpClient, _ := smtp.New(smtpConfig, nil)
    _ = snd.SendClose(context.Background(), smtpClient)
}
```

### Configuration-Based Email

Create email from configuration file:

```go
package main

import (
    "context"
    "encoding/json"
    "io"
    "os"
    "strings"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Load configuration from JSON
    var config sender.Config
    data, _ := os.ReadFile("email-config.json")
    json.Unmarshal(data, &config)
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        panic(err)
    }
    
    // Create mail from config
    mail, err := config.NewMailer()
    if err != nil {
        panic(err)
    }
    
    // Add body (not in config)
    body := io.NopCloser(strings.NewReader("Email body content"))
    mail.SetBody(sender.ContentPlainText, body)
    
    // Send
    snd, _ := mail.Sender()
    defer snd.Close()
    
    smtpClient, _ := smtp.New(smtpConfig, nil)
    _ = snd.SendClose(context.Background(), smtpClient)
}
```

**email-config.json:**
```json
{
  "charset": "UTF-8",
  "subject": "Welcome to Our Service",
  "encoding": "Base 64",
  "priority": "Normal",
  "from": "welcome@service.com",
  "to": ["newuser@example.com"],
  "attach": [
    {
      "name": "Welcome_Guide.pdf",
      "mime": "application/pdf",
      "path": "/assets/welcome-guide.pdf"
    }
  ]
}
```

### Inline Images in HTML

Embed images in HTML emails:

```go
package main

import (
    "context"
    "io"
    "os"
    "strings"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    mail := sender.New()
    mail.SetSubject("Newsletter")
    
    // HTML body with inline image reference
    html := `
    <html>
      <body>
        <h1>Company Newsletter</h1>
        <img src="cid:logo.png" alt="Logo"/>
        <p>Welcome to our monthly newsletter!</p>
      </body>
    </html>
    `
    htmlBody := io.NopCloser(strings.NewReader(html))
    mail.SetBody(sender.ContentHTML, htmlBody)
    
    // Add inline image
    logo, _ := os.Open("/assets/logo.png")
    mail.AddAttachment("logo.png", "image/png", logo, true) // inline=true
    
    // Send
    mail.Email().SetFrom("newsletter@company.com")
    mail.Email().AddRecipients(sender.RecipientTo, "subscriber@example.com")
    
    snd, _ := mail.Sender()
    defer snd.Close()
    
    smtpClient, _ := smtp.New(smtpConfig, nil)
    _ = snd.SendClose(context.Background(), smtpClient)
}
```

### Template-Based Emails

Use mail cloning for template-based emails:

```go
package main

import (
    "context"
    "fmt"
    "io"
    "strings"
    
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
)

func main() {
    // Create template
    template := sender.New()
    template.SetSubject("Account Notification")
    template.SetCharset("UTF-8")
    template.SetEncoding(sender.EncodingBase64)
    template.Email().SetFrom("notifications@service.com")
    
    // Send to multiple users
    users := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
    
    for _, user := range users {
        // Clone template for each user
        mail := template.Clone()
        
        // Personalize
        mail.Email().AddRecipients(sender.RecipientTo, user)
        body := io.NopCloser(strings.NewReader(fmt.Sprintf("Hello %s!", user)))
        mail.SetBody(sender.ContentPlainText, body)
        
        // Send
        snd, _ := mail.Sender()
        smtpClient, _ := smtp.New(smtpConfig, nil)
        _ = snd.SendClose(context.Background(), smtpClient)
    }
}
```

---

## Integration

### SMTP Client Integration

This package integrates with the `mail/smtp` package for email delivery:

```go
import (
    "github.com/nabbar/golib/mail/sender"
    "github.com/nabbar/golib/mail/smtp"
    "github.com/nabbar/golib/mail/smtp/config"
    "github.com/nabbar/golib/mail/smtp/tlsmode"
)

// Configure SMTP client
smtpConfig := config.ConfigModel{
    DSN: "tcp(smtp.gmail.com:587)/starttls",
}
cfg, _ := smtpConfig.Config()

// Create SMTP client
smtpClient, _ := smtp.New(cfg, nil)
defer smtpClient.Close()

// Create and send email
mail := sender.New()
// ... configure mail ...

snd, _ := mail.Sender()
defer snd.Close()

err := snd.Send(context.Background(), smtpClient)
```

See [github.com/nabbar/golib/mail/smtp](../smtp/README.md) for SMTP client documentation.

### Configuration File Integration

Supports multiple configuration formats:

**JSON:**
```json
{
  "charset": "UTF-8",
  "subject": "Test Email",
  "encoding": "Base 64",
  "priority": "Normal",
  "from": "sender@example.com",
  "to": ["recipient@example.com"],
  "cc": ["manager@example.com"],
  "headers": {
    "X-Campaign-ID": "2024-Q1"
  }
}
```

**YAML:**
```yaml
charset: UTF-8
subject: Test Email
encoding: Base 64
priority: Normal
from: sender@example.com
to:
  - recipient@example.com
cc:
  - manager@example.com
headers:
  X-Campaign-ID: 2024-Q1
```

**TOML:**
```toml
charset = "UTF-8"
subject = "Test Email"
encoding = "Base 64"
priority = "Normal"
from = "sender@example.com"
to = ["recipient@example.com"]
cc = ["manager@example.com"]

[headers]
X-Campaign-ID = "2024-Q1"
```

### Viper Integration

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/mail/sender"
)

// Load config with Viper
viper.SetConfigFile("email.yaml")
viper.ReadInConfig()

var config sender.Config
viper.Unmarshal(&config)

if err := config.Validate(); err != nil {
    panic(err)
}

mail, _ := config.NewMailer()
```

---

## API Reference

### Core Interfaces

**Mail Interface**
- `New() Mail` - Create new email
- `Clone() Mail` - Deep copy
- `SetSubject(string)` - Set subject
- `SetCharset(string)` - Set character encoding
- `SetEncoding(Encoding)` - Set transfer encoding
- `SetPriority(Priority)` - Set priority
- `SetBody(ContentType, io.ReadCloser)` - Set body
- `AddBody(ContentType, io.ReadCloser)` - Add body part
- `AddAttachment(name, mime string, data io.ReadCloser, inline bool)` - Add file
- `Email() Email` - Access address management
- `Sender() (Sender, error)` - Create sender

**Email Interface**
- `SetFrom(string)` - Set From address
- `SetSender(string)` - Set Sender address
- `SetReplyTo(string)` - Set Reply-To address
- `SetReturnPath(string)` - Set Return-Path
- `AddRecipients(recipientType, ...string)` - Add recipients
- `SetRecipients(recipientType, ...string)` - Replace recipients
- `GetRecipients(recipientType) []string` - Get recipients

**Sender Interface**
- `Send(context.Context, smtp.SMTP) error` - Send email
- `SendClose(context.Context, smtp.SMTP) error` - Send and close
- `Close() error` - Cleanup resources

### Configuration

**Config Struct**
- `Validate() error` - Validate configuration
- `NewMailer() (Mail, error)` - Create mail from config

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/mail/sender) for complete API documentation.

---

## Testing

**Test Suite**: 252 specs using Ginkgo v2 and Gomega

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Test Results:**
- ✅ 252/252 tests passing
- ✅ 81.4% code coverage
- ✅ Zero data races detected
- ✅ All benchmarks passing

**Coverage Areas:**
- Mail interface operations
- Email address management
- Configuration validation
- Body and attachment handling
- Sender creation and lifecycle
- Error handling and edge cases
- Type parsing (encoding, priority)
- Concurrent operations

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Best Practices

**Always Close Resources**
```go
// ✅ Good: Proper cleanup
func sendEmail(mail sender.Mail) error {
    snd, err := mail.Sender()
    if err != nil {
        return err
    }
    defer snd.Close() // Cleanup resources
    
    return snd.SendClose(ctx, smtpClient)
}
```

**Validate Configuration**
```go
// ✅ Good: Validate before use
func loadConfig(path string) (sender.Mail, error) {
    var cfg sender.Config
    // load config...
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return cfg.NewMailer()
}

// ❌ Bad: No validation
func loadConfigBad(path string) sender.Mail {
    var cfg sender.Config
    // load config...
    mail, _ := cfg.NewMailer() // May fail silently
    return mail
}
```

**Use Clone for Templates**
```go
// ✅ Good: Clone template for each recipient
func sendBulk(template sender.Mail, recipients []string) error {
    for _, rcpt := range recipients {
        mail := template.Clone() // Independent copy
        mail.Email().AddRecipients(sender.RecipientTo, rcpt)
        // send mail...
    }
    return nil
}

// ❌ Bad: Reuse same mail object
func sendBulkBad(mail sender.Mail, recipients []string) error {
    for _, rcpt := range recipients {
        mail.Email().AddRecipients(sender.RecipientTo, rcpt)
        // send mail... (recipients accumulate!)
    }
    return nil
}
```

**Handle Errors Properly**
```go
// ✅ Good: Check all errors
func sendWithAttachment(mail sender.Mail, path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer file.Close()
    
    mail.AddAttachment(filepath.Base(path), "application/pdf", file, false)
    
    snd, err := mail.Sender()
    if err != nil {
        return fmt.Errorf("create sender: %w", err)
    }
    defer snd.Close()
    
    if err := snd.Send(ctx, smtpClient); err != nil {
        return fmt.Errorf("send email: %w", err)
    }
    
    return nil
}
```

**Stream Large Files**
```go
// ✅ Good: Stream from file
func addLargeAttachment(mail sender.Mail, path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    // Don't close here - Sender will close it
    
    mail.AddAttachment(filepath.Base(path), "application/zip", file, false)
    return nil
}

// ❌ Bad: Load entire file into memory
func addLargeAttachmentBad(mail sender.Mail, path string) error {
    data, _ := os.ReadFile(path) // Full file in RAM!
    reader := io.NopCloser(bytes.NewReader(data))
    mail.AddAttachment(filepath.Base(path), "application/zip", reader, false)
    return nil
}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥80%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document all public APIs with GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety concerns
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results and coverage
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Email Features**
- Email templates with variable substitution
- Batch sending with rate limiting
- Email scheduling and queuing
- Read receipts and delivery status notifications
- S/MIME encryption and signing
- DKIM signing support

**Performance**
- Connection pooling for SMTP
- Concurrent email sending
- Attachment streaming optimization
- Message caching for templates

**Integration**
- Direct integration with popular SMTP services (SendGrid, Mailgun, AWS SES)
- Webhook support for delivery tracking
- Email analytics and metrics
- Template management system

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
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/mail/sender)
- **SMTP Client**: [mail/smtp Package](../smtp/README.md)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
