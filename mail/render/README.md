# Mail Render Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Production-ready email template rendering library for Go with theme support, bidirectional text, template variables, and HTML/plain text generation using the Hermes v2 engine.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Template System](#template-system)
- [Themes & Localization](#themes--localization)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The `mail/render` package provides a high-level API for generating professional transactional emails using the [Hermes v2](https://github.com/go-hermes/hermes) template engine. It wraps Hermes with additional features like theme management, text direction control, template variable replacement, and configuration-based initialization.

### Design Philosophy

1. **Template-First**: Pre-designed themes for consistent branding
2. **Localization-Ready**: Built-in RTL support for international audiences
3. **Type-Safe**: Structured configuration with validation
4. **Production-Ready**: Deep cloning for concurrent operations
5. **Framework-Agnostic**: Works with any SMTP/email sending library

---

## Key Features

- **Multiple Themes**: Default (classic) and Flat (modern) visual styles
- **Bidirectional Text**: LTR/RTL support for internationalization
- **Template Variables**: `{{variable}}` replacement across all content
- **Rich Content**: Tables, actions/buttons, dictionaries, custom markdown
- **Dual Output**: Generate both HTML and plain text versions
- **Configuration-Based**: JSON/YAML/TOML with validation
- **Thread-Safe Cloning**: Deep copy for concurrent email generation
- **CSS Inlining**: Optional inline CSS for maximum email client compatibility
- **Structured Errors**: Integration with `github.com/nabbar/golib/errors`

---

## Installation

```bash
go get github.com/nabbar/golib/mail/render
```

Dependencies:
- `github.com/go-hermes/hermes/v2` - Email template engine
- `github.com/go-playground/validator/v10` - Configuration validation
- `github.com/nabbar/golib/errors` - Structured error handling

---

## Architecture

### Package Structure

The package is organized around the `Mailer` interface with support components:

```
render/
├── interface.go         # Mailer interface and New() constructor
├── email.go            # Internal email implementation
├── config.go           # Configuration struct with validation
├── render.go           # HTML/text generation and ParseData
├── themes.go           # Theme enumeration and parsing
├── direction.go        # Text direction enumeration
├── error.go            # Error codes and messages
└── doc.go             # Package documentation
```

### Component Overview

```
┌───────────────────────────────────────────────────┐
│                 Mailer Interface                  │
│  Theme, Direction, Product, Body, Generate()      │
└──────────────┬─────────────┬────────────┬─────────┘
               │             │            │
      ┌────────▼─────┐  ┌────▼────┐  ┌────▼─────┐
      │   Config     │  │  email  │  │  Hermes  │
      │              │  │         │  │          │
      │ Validation   │  │ State   │  │ Renderer │
      │ NewMailer()  │  │ Clone() │  │ Themes   │
      └──────────────┘  └─────────┘  └──────────┘
```

| Component | Purpose | Thread-Safe |
|-----------|---------|-------------|
| **`Mailer`** | Interface for email configuration and generation | Via Clone() |
| **`Config`** | Structured configuration with validation | ✅ |
| **`email`** | Internal state management | ❌ (use Clone()) |
| **Hermes** | Template rendering engine | ✅ |

### Workflow Diagram

```
   User Code
       │
       ▼
   New() / Config.NewMailer()
       │
       ▼
   Configure (SetTheme, SetBody, etc.)
       │
       ▼
   ParseData(variables) [Optional]
       │
       ▼
   GenerateHTML() / GeneratePlainText()
       │
       ▼
   Hermes Rendering Engine
       │
       ▼
   HTML/Text Buffer Output
       │
       ▼
   Send via SMTP (external)
```

---

## Performance

### Generation Benchmarks

Based on comprehensive testing with **123 test specs** and **89.6% code coverage**:

| Operation | Duration (Normal) | Duration (Race) | Memory | Notes |
|-----------|-------------------|-----------------|--------|-------|
| **New()** | ~100 ns | ~1.5 µs | O(1) | Struct initialization |
| **Config.NewMailer()** | ~300 ns | ~4 µs | O(1) | With validation |
| **Clone() Simple** | ~1 µs | ~10 µs | O(1) | Empty body |
| **Clone() Complex** | ~10 µs | ~100 µs | O(n) | Deep copy of tables |
| **ParseData Simple** | ~350 ns | ~4 µs | O(1) | Few variables |
| **ParseData Complex** | ~1.2 µs | ~20 µs | O(n*m) | Many variables |
| **GenerateHTML Simple** | ~2.7 ms | ~47 ms | ~100 KB | Basic email |
| **GenerateHTML Complex** | ~3.1 ms | ~51 ms | ~200 KB | Tables + actions |
| **GeneratePlainText** | ~3.6 ms | ~48 ms | ~50 KB | Text conversion |
| **ParseTheme** | ~83 ns | ~1.5 µs | O(1) | String parsing |
| **ParseTextDirection** | ~54 ns | ~495 ns | O(1) | String parsing |
| **Config.Validate()** | ~31 µs | ~112 µs | O(1) | Field validation |
| **Complete Workflow** | ~3.1 ms | ~51 ms | ~250 KB | Config → HTML |

*Measured on Linux/AMD64, Go 1.21+*

### Memory Efficiency

- **Email Generation**: ~100-200 KB per email (HTML)
- **Plain Text**: ~50 KB per email
- **Clone Operations**: Proportional to body content size
- **ParseData**: In-place replacements (no extra allocation)

### Throughput Capacity

```
Performance Characteristics:
├─ Email Generation: ~300 emails/second (single thread)
├─ With Concurrency: ~3000 emails/second (10 goroutines)
├─ ParseData: ~1M operations/second
└─ Clone: ~100K clones/second (simple bodies)

Bottleneck: Hermes rendering (HTML generation)
```

### Thread Safety Notes

- **Not Thread-Safe**: Individual `Mailer` instances
- **Concurrent Pattern**: Clone() for each goroutine
- **Race Detection**: Zero data races (verified with `-race`)
- **Production**: Use worker pool pattern for high-volume

---

## Use Cases

This library is designed for transactional email scenarios:

**User Onboarding**
- Welcome emails with verification links
- Account activation and password reset
- Multi-step onboarding sequences
- Personalized greetings and instructions

**Notifications & Alerts**
- Order confirmations and invoices
- Shipping notifications with tracking
- Account activity alerts (login, changes)
- System status updates

**Marketing Automation**
- Newsletter templates
- Promotional campaigns
- Event invitations
- Product announcements

**E-Commerce**
- Order receipts with itemized tables
- Cart abandonment reminders
- Product review requests
- Loyalty program updates

**SaaS Applications**
- Trial expiration reminders
- Feature announcement emails
- Usage reports and analytics
- Subscription renewals

---

## Quick Start

### Basic Email Generation

Simple email with intro and outro:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/mail/render"
    "github.com/go-hermes/hermes/v2"
)

func main() {
    // Create mailer
    mailer := render.New()
    mailer.SetName("My Company")
    mailer.SetLink("https://example.com")
    mailer.SetLogo("https://example.com/logo.png")
    mailer.SetTheme(render.ThemeFlat)

    // Configure email body
    body := &hermes.Body{
        Name:   "John Doe",
        Intros: []string{"Welcome to our service!"},
        Outros: []string{"Thank you for signing up."},
    }
    mailer.SetBody(body)

    // Generate HTML
    htmlBuf, err := mailer.GenerateHTML()
    if err != nil {
        panic(err)
    }

    fmt.Println("HTML length:", htmlBuf.Len())
    
    // Generate plain text
    textBuf, _ := mailer.GeneratePlainText()
    fmt.Println("Text length:", textBuf.Len())
}
```

### Configuration-Based Initialization

Using structured configuration (suitable for JSON/YAML/TOML):

```go
package main

import (
    "encoding/json"
    "os"
    "github.com/nabbar/golib/mail/render"
    "github.com/go-hermes/hermes/v2"
)

func main() {
    // Load from JSON config file
    data, _ := os.ReadFile("email-config.json")
    
    var config render.Config
    json.Unmarshal(data, &config)
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        panic(err)
    }
    
    // Create mailer from config
    mailer := config.NewMailer()
    
    // Generate email
    htmlBuf, _ := mailer.GenerateHTML()
}
```

Example `email-config.json`:

```json
{
  "theme": "flat",
  "direction": "ltr",
  "name": "My Company",
  "link": "https://example.com",
  "logo": "https://example.com/logo.png",
  "copyright": "© 2024 My Company",
  "troubleText": "Need help? Contact support@example.com",
  "disableCSSInline": false,
  "body": {
    "name": "{{user}}",
    "intros": ["Welcome to our service!"],
    "outros": ["Thank you for signing up."]
  }
}
```

### Email with Tables and Actions

Rich email with itemized data and call-to-action:

```go
body := &hermes.Body{
    Name:   "John Doe",
    Intros: []string{"Your order has been confirmed:"},
    Dictionary: []hermes.Entry{
        {Key: "Order ID", Value: "ORD-123456"},
        {Key: "Date", Value: "2024-01-15"},
        {Key: "Total", Value: "$129.99"},
    },
    Tables: []hermes.Table{{
        Data: [][]hermes.Entry{
            {
                {Key: "Item", Value: "Product A"},
                {Key: "Quantity", Value: "2"},
                {Key: "Price", Value: "$49.99"},
            },
            {
                {Key: "Item", Value: "Product B"},
                {Key: "Quantity", Value: "1"},
                {Key: "Price", Value: "$30.00"},
            },
        },
        Columns: hermes.Columns{
            CustomWidth: map[string]string{
                "Item":     "50%",
                "Quantity": "20%",
                "Price":    "30%",
            },
        },
    }},
    Actions: []hermes.Action{
        {
            Instructions: "View your order details:",
            Button: hermes.Button{
                Text: "View Order",
                Link: "https://example.com/orders/123456",
                Color: "#3869D4",
            },
        },
    },
    Outros: []string{"Need help? Reply to this email."},
}

mailer.SetBody(body)
htmlBuf, _ := mailer.GenerateHTML()
```

### Template Variable Replacement

Dynamic content with variable substitution:

```go
// Configure email with placeholders
mailer.SetName("{{company}}")
body := &hermes.Body{
    Name:   "{{username}}",
    Intros: []string{"Your verification code is {{code}}"},
    Actions: []hermes.Action{
        {
            Instructions: "Click below to verify:",
            Button: hermes.Button{
                Text: "Verify Email",
                Link: "{{verification_url}}",
            },
        },
    },
}
mailer.SetBody(body)

// Replace variables
mailer.ParseData(map[string]string{
    "{{company}}":          "Acme Inc",
    "{{username}}":         "John Doe",
    "{{code}}":             "837492",
    "{{verification_url}}": "https://example.com/verify?token=abc123",
})

// Generate final email
htmlBuf, _ := mailer.GenerateHTML()
```

### Concurrent Email Generation

Thread-safe pattern for bulk email sending:

```go
package main

import (
    "sync"
    "github.com/nabbar/golib/mail/render"
    "github.com/go-hermes/hermes/v2"
)

func main() {
    // Base template
    baseMailer := render.New()
    baseMailer.SetName("My Company")
    baseMailer.SetTheme(render.ThemeFlat)
    
    users := []struct{ Name, Email string }{
        {"John Doe", "john@example.com"},
        {"Jane Smith", "jane@example.com"},
        // ... more users
    }
    
    var wg sync.WaitGroup
    for _, user := range users {
        wg.Add(1)
        go func(name, email string) {
            defer wg.Done()
            
            // Clone for thread safety
            mailer := baseMailer.Clone()
            
            // Customize per user
            body := &hermes.Body{
                Name:   name,
                Intros: []string{"Welcome!"},
            }
            mailer.SetBody(body)
            
            // Generate and send
            htmlBuf, _ := mailer.GenerateHTML()
            sendEmail(email, htmlBuf.String())
        }(user.Name, user.Email)
    }
    
    wg.Wait()
}
```

---

## Template System

### Template Variables

Variables use the `{{variable}}` syntax and can appear anywhere in the email content:

**Supported Locations**:
- Product fields (name, link, logo, copyright, troubleText)
- Body fields (name, greeting, signature, title)
- Intro and outro texts
- Dictionary entries (keys and values)
- Table data (keys and values)
- Action instructions and button properties
- Free-form Markdown content

**Example**:

```go
// Define template
body := &hermes.Body{
    Name:   "{{user_name}}",
    Title:  "Order {{order_id}}",
    Intros: []string{"Hi {{user_name}}, your order for {{product}} is ready!"},
    Dictionary: []hermes.Entry{
        {Key: "Order", Value: "{{order_id}}"},
        {Key: "Total", Value: "{{total}}"},
    },
}

// Populate variables
mailer.ParseData(map[string]string{
    "{{user_name}}": "John Doe",
    "{{order_id}}":  "ORD-12345",
    "{{product}}":   "Widget Pro",
    "{{total}}":     "$99.99",
})
```

### Email Body Structure

The `hermes.Body` struct supports rich content:

```go
type hermes.Body struct {
    Name         string           // Recipient name
    Intros       []string         // Opening paragraphs
    Dictionary   []Entry          // Key-value pairs
    Tables       []Table          // Data tables
    Actions      []Action         // Call-to-action buttons
    Outros       []string         // Closing paragraphs
    Greeting     string           // Custom greeting
    Signature    string           // Custom signature
    Title        string           // Email title/subject
    FreeMarkdown Markdown         // Custom HTML/Markdown
}
```

**Content Flow**:
1. Greeting (or default "Hi")
2. Name
3. Title (optional)
4. Intros
5. Dictionary (if present)
6. Tables (if present)
7. Actions (if present)
8. Outros
9. Signature (or default "Yours truly")

### Tables

Tables display structured data with optional column customization:

```go
table := hermes.Table{
    Title: "Order Items",  // Optional table title
    Data: [][]hermes.Entry{
        {{Key: "Product", Value: "Widget A"}, {Key: "Price", Value: "$10"}},
        {{Key: "Product", Value: "Widget B"}, {Key: "Price", Value: "$20"}},
    },
    Columns: hermes.Columns{
        CustomWidth: map[string]string{
            "Product": "70%",
            "Price":   "30%",
        },
        CustomAlignment: map[string]string{
            "Price": "right",
        },
    },
}
```

### Actions and Buttons

Actions create call-to-action sections with buttons:

```go
action := hermes.Action{
    Instructions: "Click below to confirm:",
    InviteCode:   "ABC-123",  // Optional code display
    Button: hermes.Button{
        Text:      "Confirm Email",
        Link:      "https://example.com/confirm",
        Color:     "#3869D4",     // Primary button color
        TextColor: "#FFFFFF",     // Button text color
    },
}
```

---

## Themes & Localization

### Available Themes

**ThemeDefault** - Classic email design with centered layout

```
┌──────────────────────────────────────────┐
│              [LOGO]                      │
│                                          │
│  Hi John Doe,                            │
│                                          │
│  Welcome to our service! We're excited   │
│  to have you on board.                   │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │     [VIEW ORDER]                   │  │
│  └────────────────────────────────────┘  │
│                                          │
│  Thank you,                              │
│  My Company Team                         │
│                                          │
│  © 2024 My Company                       │
└──────────────────────────────────────────┘
```

**ThemeFlat** - Modern, minimalist design

```
───────────────────────────────────────────────
[LOGO]  My Company

Hi John Doe,

Welcome to our service! We're excited to have
you on board.

┌─────────────────────┐
│  VIEW ORDER →       │
└─────────────────────┘

Thank you,
My Company Team

───────────────────────────────────────────────
© 2024 My Company
```

### Theme Selection

```go
// Via constant
mailer.SetTheme(render.ThemeDefault)
mailer.SetTheme(render.ThemeFlat)

// Via string (case-insensitive)
theme := render.ParseTheme("flat")
mailer.SetTheme(theme)

// From configuration
config := render.Config{Theme: "default", ...}
mailer := config.NewMailer()
```

### Text Direction (Internationalization)

Support for both LTR and RTL languages:

**LeftToRight** (default) - Western languages

```go
mailer.SetTextDirection(render.LeftToRight)
// English, French, Spanish, German, Italian, etc.
```

**RightToLeft** - Middle Eastern languages

```go
mailer.SetTextDirection(render.RightToLeft)
// Arabic, Hebrew, Persian, Urdu, etc.
```

**Parsing from String**:

```go
// Supported formats (case-insensitive)
dir := render.ParseTextDirection("ltr")     // LeftToRight
dir = render.ParseTextDirection("rtl")      // RightToLeft
dir = render.ParseTextDirection("left-to-right")
dir = render.ParseTextDirection("right-to-left")
```

**RTL Email Example**:

```go
mailer := render.New()
mailer.SetTextDirection(render.RightToLeft)
mailer.SetName("شركتي")  // Arabic: "My Company"

body := &hermes.Body{
    Name:   "أحمد",        // Ahmed
    Intros: []string{"مرحبا بك"},  // Welcome
}
mailer.SetBody(body)
```

### CSS Inlining

Control CSS inlining for maximum email client compatibility:

```go
// Enable inline CSS (default, recommended)
mailer.SetCSSInline(false)  // false = inline enabled

// Disable inline CSS (for modern clients)
mailer.SetCSSInline(true)   // true = inline disabled
```

**When to disable CSS inlining**:
- Testing during development
- Modern email clients only
- Custom CSS post-processing

**Why inline CSS matters**:
- Gmail strips `<style>` tags
- Outlook has limited CSS support
- Mobile clients vary widely
- Inlining ensures consistent rendering

---

## Best Practices

### Configuration Management

**✅ Use Config Struct**
```go
// Store in config file (JSON/YAML)
config := render.Config{
    Theme: "flat",
    Name:  "My Company",
    // ... other fields
}

// Validate before use
if err := config.Validate(); err != nil {
    log.Fatalf("invalid config: %v", err)
}

mailer := config.NewMailer()
```

**❌ Avoid Hardcoding**
```go
// Don't hardcode in every function
func sendEmail() {
    mailer := render.New()
    mailer.SetName("My Company")  // Repeated everywhere
    mailer.SetLink("https://...")
    // ...
}
```

### Error Handling

**✅ Check All Errors**
```go
htmlBuf, err := mailer.GenerateHTML()
if err != nil {
    if err.Code() == render.ErrorMailerHtml {
        // Handle HTML generation error
    }
    return fmt.Errorf("generate email: %w", err)
}
```

**❌ Ignore Errors**
```go
htmlBuf, _ := mailer.GenerateHTML()  // Don't do this!
```

### Resource Cleanup

**✅ Always Close Buffers**
```go
func generateEmail(mailer render.Mailer) (*bytes.Buffer, error) {
    htmlBuf, err := mailer.GenerateHTML()
    if err != nil {
        return nil, err
    }
    // Buffer returned for use
    return htmlBuf, nil
}
```

### Concurrent Operations

**✅ Clone for Each Goroutine**
```go
baseMailer := render.New()
// Configure base template...

for _, user := range users {
    go func(u User) {
        mailer := baseMailer.Clone()  // Independent copy
        // Customize and send...
    }(user)
}
```

**❌ Share Mailer Across Goroutines**
```go
mailer := render.New()

for _, user := range users {
    go func(u User) {
        mailer.SetBody(...)  // RACE CONDITION!
    }(user)
}
```

### Template Organization

**✅ Reusable Templates**
```go
// Define template factory
func newWelcomeEmail(name, email string) render.Mailer {
    mailer := render.New()
    // Base configuration
    mailer.SetName("My Company")
    mailer.SetTheme(render.ThemeFlat)
    
    // User-specific
    body := &hermes.Body{
        Name:   name,
        Intros: []string{"Welcome!"},
    }
    mailer.SetBody(body)
    return mailer
}

// Use
mailer := newWelcomeEmail("John", "john@example.com")
```

### Variable Naming

**✅ Clear Delimiters**
```go
// Use double braces for clarity
"{{user_name}}"
"{{order_id}}"
"{{verification_token}}"
```

**❌ Ambiguous Patterns**
```go
// Avoid single braces or short names
"{name}"  // Might conflict with JSON
"$n"      // Unclear what this represents
```

### Testing Emails

**✅ Test Both Formats**
```go
func TestEmail(t *testing.T) {
    mailer := setupTestMailer()
    
    // Test HTML
    htmlBuf, err := mailer.GenerateHTML()
    assert.NoError(t, err)
    assert.Contains(t, htmlBuf.String(), "Welcome")
    
    // Test plain text
    textBuf, err := mailer.GeneratePlainText()
    assert.NoError(t, err)
    assert.Contains(t, textBuf.String(), "Welcome")
}
```

### Performance Optimization

**✅ Reuse Base Mailers**
```go
// Create once, clone many
var baseMailer = initializeBase()

func sendToUser(user User) {
    mailer := baseMailer.Clone()
    // Fast, only clones state
}
```

**✅ Batch Variable Replacement**
```go
// Single ParseData call
mailer.ParseData(map[string]string{
    "{{name}}":  name,
    "{{email}}": email,
    "{{code}}":  code,
    // ... all variables at once
})
```

**❌ Multiple ParseData Calls**
```go
// Inefficient: iterates content multiple times
mailer.ParseData(map[string]string{"{{name}}": name})
mailer.ParseData(map[string]string{"{{email}}": email})
mailer.ParseData(map[string]string{"{{code}}": code})
```

---

## Testing

**Test Suite**: 123 specs using Ginkgo v2 and Gomega with gmeasure benchmarks (89.6% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# With benchmarks
go test -bench=. ./...
```

**Coverage Areas**:
- Mailer interface operations
- Theme and text direction parsing
- Configuration validation
- HTML and plain text generation
- Template variable replacement
- Deep cloning and thread safety
- Error handling and edge cases
- Concurrent operations

**Quality Assurance**:
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe Clone() operations
- ✅ Deep copy validation
- ✅ 89.6% code coverage

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `CGO_ENABLED=1 go test -race`
- Maintain or improve test coverage (≥85%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Update GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add benchmarks for performance-critical code

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Additional Themes**
- Corporate theme (formal business style)
- Newsletter theme (multi-section layout)
- Notification theme (minimal alert style)
- Custom theme builder API

**Enhanced Localization**
- Pre-translated greeting/signature templates
- Date/time formatting per locale
- Number formatting (currency, decimals)
- Language-specific typography

**Template Features**
- Conditional sections (if/else)
- Loops for repeated content
- Nested variable support
- Template inheritance

**Rich Content**
- Image embedding (data URIs)
- Custom fonts
- Progress bars
- Star ratings
- Social media buttons

**Performance**
- Template caching
- Compiled templates
- Streaming generation (large emails)
- Parallel rendering

**Integrations**
- Direct SMTP sending (combine with `mail/smtp`)
- Email validation
- Preview in browser
- A/B testing support

**Developer Experience**
- Email preview server
- Visual template editor
- CLI tool for testing
- Migration helpers (from other libraries)

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
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/mail/render)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Hermes Library**: [github.com/go-hermes/hermes](https://github.com/go-hermes/hermes)
- **Related Packages**:
  - [mail/smtp](../smtp) - SMTP client for sending emails
  - [mail/sender](../sender) - High-level email sending with attachments
  - [errors](../../errors) - Structured error handling
