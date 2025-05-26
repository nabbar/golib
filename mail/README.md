# `mail` Package Documentation

The `mail` package provides a flexible and extensible API for composing, configuring, and sending emails in Go applications. It supports advanced features such as multiple recipients, attachments, inline files, custom headers, content types, encoding, and priority management.

---

## Features

- Compose emails with plain text and HTML bodies
- Add attachments and inline files
- Set custom headers, subject, charset, encoding, and priority
- Manage sender, recipients (To, Cc, Bcc), reply-to, and return path
- Validate email configuration
- Send emails via SMTP with progress tracking and error handling
- Thread-safe and suitable for production use

---

## Main Types

### `Mail` Interface

Defines the main email object with methods to configure and retrieve all aspects of an email:

- Set and get charset, subject, encoding, and priority
- Set and get date/time
- Add and retrieve custom headers
- Manage email bodies (plain text, HTML)
- Add and retrieve attachments and inline files
- Manage sender and recipients
- Create a `Sender` for SMTP delivery

### `Config` Struct

A configuration struct for easy mapping from config files or environment variables:

- Charset, subject, encoding, priority
- Headers (map of key-value pairs)
- From, sender, reply-to, return path
- Lists of recipients (To, Cc, Bcc)
- Attachments and inline files (with name, mime, and path)
- Validation method to ensure all required fields are set

### `Sender` Interface

Handles the actual sending of the composed email via an SMTP client:

- `Send(ctx, smtp)`: Sends the email
- `SendClose(ctx, smtp)`: Sends and closes resources
- `Close()`: Cleans up resources

---

## Content Types and Encoding

- Supports plain text and HTML bodies
- Encoding options: none, binary, base64, quoted-printable
- Priority options: normal, low, high

---

## Example Usage

```go
import (
    "github.com/nabbar/golib/mail"
    "github.com/nabbar/golib/smtp"
    "context"
)

m := mail.New()
m.SetSubject("Test Email")
m.SetCharset("UTF-8")
m.SetPriority(mail.PriorityHigh)
m.Email().SetFrom("sender@example.com")
m.Email().AddRecipients(mail.RecipientTo, "recipient@example.com")
// Add body, attachments, etc.

sender, err := m.Sender()
if err != nil {
    // handle error
}
defer sender.Close()

smtpClient := /* initialize SMTP client */
err = sender.Send(context.Background(), smtpClient)
if err != nil {
    // handle error
}
```

---

## Configuration Example

```go
cfg := mail.Config{
    Charset:  "UTF-8",
    Subject:  "Hello",
    Encoding: "Base 64",
    Priority: "High",
    From:     "me@example.com",
    To:       []string{"you@example.com"},
    Attach:   []mail.ConfigFile{{Name: "file.txt", Mime: "text/plain", Path: "/tmp/file.txt"}},
}
if err := cfg.Validate(); err != nil {
    // handle validation error
}
mailer, err := cfg.NewMailer()
```

---

## Notes

- All operations are safe for concurrent use.
- Designed for Go 1.18+.
- Integrates with external SMTP clients and file progress tracking.
- Provides detailed error codes for robust error handling.

