# mailer Package Documentation

The `mailer` package provides a flexible API for composing, theming, and rendering transactional emails in Go applications. It leverages the Hermes library for HTML and plain text output, supports dynamic data injection, and allows full customization of product and message content.

---

## Features

- Compose emails with customizable themes and text direction
- Set product information: name, link, logo, copyright, trouble text
- Define email body with intros, outros, tables, actions, and markdown
- Inject dynamic data into all fields
- Generate HTML and plain text content
- Validate configuration with detailed error reporting
- Thread-safe and suitable for production use

---

## Main Types

### Mailer Interface

Defines the main email object with methods to configure and retrieve all aspects of an email:

- `SetTheme(t Themes)` / `GetTheme()`
- `SetTextDirection(d TextDirection)` / `GetTextDirection()`
- `SetBody(b *hermes.Body)` / `GetBody()`
- `SetCSSInline(disable bool)`
- `SetName(name string)` / `GetName()`
- `SetCopyright(copy string)` / `GetCopyright()`
- `SetLink(link string)` / `GetLink()`
- `SetLogo(logoUrl string)` / `GetLogo()`
- `SetTroubleText(text string)` / `GetTroubleText()`
- `ParseData(data map[string]string)` — injects dynamic values into all fields
- `GenerateHTML()` — returns the email as HTML
- `GeneratePlainText()` — returns the email as plain text
- `Clone()` — deep copy of the mailer

### Config Struct

A configuration struct for easy mapping from config files or environment variables:

- Theme, direction, name, link, logo, copyright
- Trouble text, disable CSS inlining
- Body (hermes.Body)
- Validation method to ensure all required fields are set
- `NewMailer()` — creates a Mailer from the config

### Themes and TextDirection

- `Themes`: `ThemeDefault`, `ThemeFlat`
- `TextDirection`: `LeftToRight`, `RightToLeft`
- Parsing helpers: `ParseTheme(string)`, `ParseTextDirection(string)`

---

## Example Usage

```go
import (
    "github.com/nabbar/golib/mailer"
    "github.com/matcornic/hermes"
)

m := mailer.New()
m.SetTheme(mailer.ThemeFlat)
m.SetTextDirection(mailer.LeftToRight)
m.SetName("MyApp")
m.SetLink("https://myapp.example.com")
m.SetLogo("https://myapp.example.com/logo.png")
m.SetCopyright("© 2024 MyApp")
m.SetTroubleText("If you’re having trouble, contact support.")

body := &hermes.Body{
    Name:    "John Doe",
    Intros:  []string{"Welcome to MyApp!"},
    Outros:  []string{"Thank you for joining."},
    Actions: []hermes.Action{{Button: hermes.Button{Text: "Get Started", Link: "https://myapp.example.com/start"}}},
}
m.SetBody(body)

// Inject dynamic data
m.ParseData(map[string]string{"MyApp": "YourApp"})

// Generate HTML and plain text
html, err := m.GenerateHTML()
text, err := m.GeneratePlainText()
```

---

## Configuration Example

```go
cfg := mailer.Config{
    Theme:       "Default",
    Direction:   "Left->Right",
    Name:        "MyApp",
    Link:        "https://myapp.example.com",
    Logo:        "https://myapp.example.com/logo.png",
    Copyright:   "© 2024 MyApp",
    TroubleText: "If you’re having trouble, contact support.",
    Body:        hermes.Body{ /* ... */ },
}
if err := cfg.Validate(); err != nil {
    // handle validation error
}
mailer := cfg.NewMailer()
```

---

## Error Handling

- Errors are returned as custom error types with codes for invalid config, HTML/text generation, or empty parameters.

---

## Notes

- Designed for Go 1.18+.
- All operations are thread-safe.
- Integrates with the [hermes](https://github.com/matcornic/hermes) library for email rendering.
- Suitable for high-concurrency and production environments.
