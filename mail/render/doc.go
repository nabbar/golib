/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

// Package render provides email template rendering functionality using the Hermes library.
// It offers a high-level API for generating both HTML and plain text email content with
// support for themes, text direction, and template variable replacement.
//
// # Overview
//
// The render package wraps the github.com/go-hermes/hermes/v2 library to provide an
// easy-to-use interface for creating professional-looking transactional emails.
// It supports multiple themes, bidirectional text, and rich content including tables,
// actions, and custom formatting.
//
// # Key Features
//
//   - Multiple visual themes (Default, Flat)
//   - Bidirectional text support (LTR/RTL)
//   - Template variable replacement with {{variable}} syntax
//   - Rich content support (tables, actions, buttons, dictionary entries)
//   - HTML and plain text generation
//   - Configuration via struct with validation
//   - Deep cloning for concurrent operations
//   - Integration with github.com/nabbar/golib/errors for structured error handling
//
// # Basic Usage
//
// Creating a simple email:
//
//	mailer := render.New()
//	mailer.SetName("My Company")
//	mailer.SetLink("https://example.com")
//	mailer.SetLogo("https://example.com/logo.png")
//	mailer.SetTheme(render.ThemeFlat)
//
//	body := &hermes.Body{
//	    Name:   "John Doe",
//	    Intros: []string{"Welcome to our service!"},
//	    Outros: []string{"Thank you for signing up."},
//	}
//	mailer.SetBody(body)
//
//	htmlBuf, err := mailer.GenerateHTML()
//	textBuf, err := mailer.GeneratePlainText()
//
// # Configuration-Based Usage
//
// Using a configuration struct (suitable for JSON/YAML/TOML):
//
//	config := render.Config{
//	    Theme:       "flat",
//	    Direction:   "ltr",
//	    Name:        "My Company",
//	    Link:        "https://example.com",
//	    Logo:        "https://example.com/logo.png",
//	    Copyright:   "© 2024 My Company",
//	    TroubleText: "Need help? Contact support@example.com",
//	    Body: hermes.Body{
//	        Name:   "John Doe",
//	        Intros: []string{"Welcome!"},
//	    },
//	}
//
//	if err := config.Validate(); err != nil {
//	    // Handle validation error
//	}
//
//	mailer := config.NewMailer()
//	htmlBuf, err := mailer.GenerateHTML()
//
// # Template Variable Replacement
//
// The package supports template variable replacement with the {{variable}} syntax:
//
//	mailer := render.New()
//	mailer.SetName("{{company}}")
//
//	body := &hermes.Body{
//	    Name:   "{{user}}",
//	    Intros: []string{"Your verification code is {{code}}"},
//	    Actions: []hermes.Action{
//	        {
//	            Instructions: "Click below to verify:",
//	            Button: hermes.Button{
//	                Text: "Verify Email",
//	                Link: "{{verification_link}}",
//	            },
//	        },
//	    },
//	}
//	mailer.SetBody(body)
//
//	mailer.ParseData(map[string]string{
//	    "{{company}}":           "Acme Inc",
//	    "{{user}}":              "John Doe",
//	    "{{code}}":              "123456",
//	    "{{verification_link}}": "https://example.com/verify?token=abc123",
//	})
//
//	htmlBuf, err := mailer.GenerateHTML()
//
// # Advanced Content
//
// Creating emails with tables and actions:
//
//	body := &hermes.Body{
//	    Name:   "John Doe",
//	    Intros: []string{"Here is your monthly report:"},
//	    Dictionary: []hermes.Entry{
//	        {Key: "Transaction ID", Value: "TXN-123456"},
//	        {Key: "Date", Value: "2024-01-15"},
//	    },
//	    Tables: []hermes.Table{{
//	        Data: [][]hermes.Entry{
//	            {
//	                {Key: "Item", Value: "Product A"},
//	                {Key: "Quantity", Value: "2"},
//	                {Key: "Price", Value: "$50.00"},
//	            },
//	            {
//	                {Key: "Item", Value: "Product B"},
//	                {Key: "Quantity", Value: "1"},
//	                {Key: "Price", Value: "$30.00"},
//	            },
//	        },
//	        Columns: hermes.Columns{
//	            CustomWidth: map[string]string{
//	                "Item":     "50%",
//	                "Quantity": "20%",
//	                "Price":    "30%",
//	            },
//	        },
//	    }},
//	    Actions: []hermes.Action{
//	        {
//	            Instructions: "View full details:",
//	            Button: hermes.Button{
//	                Text: "View Invoice",
//	                Link: "https://example.com/invoice/123456",
//	            },
//	        },
//	    },
//	    Outros: []string{"Thank you for your business!"},
//	}
//
//	mailer.SetBody(body)
//	htmlBuf, err := mailer.GenerateHTML()
//
// # Themes
//
// The package supports multiple themes:
//
//   - ThemeDefault: Classic, centered email design
//   - ThemeFlat: Modern, minimalist email design
//
// Theme selection:
//
//	mailer.SetTheme(render.ThemeFlat)
//	// or
//	theme := render.ParseTheme("flat")
//	mailer.SetTheme(theme)
//
// # Text Direction
//
// Support for both LTR and RTL languages:
//
//	// For LTR languages (English, French, Spanish, etc.)
//	mailer.SetTextDirection(render.LeftToRight)
//
//	// For RTL languages (Arabic, Hebrew, Persian, etc.)
//	mailer.SetTextDirection(render.RightToLeft)
//
//	// Or parse from string
//	direction := render.ParseTextDirection("rtl")
//	mailer.SetTextDirection(direction)
//
// # Thread Safety
//
// Mailer instances are not thread-safe. For concurrent operations, use Clone():
//
//	baseMailer := render.New()
//	baseMailer.SetName("My Company")
//	baseMailer.SetTheme(render.ThemeFlat)
//
//	// Create independent copies for concurrent use
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func(index int) {
//	        defer wg.Done()
//	        mailer := baseMailer.Clone()
//	        body := &hermes.Body{
//	            Name: fmt.Sprintf("User %d", index),
//	        }
//	        mailer.SetBody(body)
//	        htmlBuf, _ := mailer.GenerateHTML()
//	        // Send email...
//	    }(i)
//	}
//	wg.Wait()
//
// # Error Handling
//
// The package uses github.com/nabbar/golib/errors for structured error handling:
//
//	htmlBuf, err := mailer.GenerateHTML()
//	if err != nil {
//	    if err.Code() == render.ErrorMailerHtml {
//	        // Handle HTML generation error
//	    }
//	    log.Printf("Error: %v", err)
//	}
//
// Error codes:
//   - ErrorParamEmpty: Required parameters are missing
//   - ErrorMailerConfigInvalid: Configuration validation failed
//   - ErrorMailerHtml: HTML generation failed
//   - ErrorMailerText: Plain text generation failed
//
// # Integration with Email Sending
//
// The generated content can be used with various email sending packages:
//
//	mailer := render.New()
//	// ... configure mailer ...
//
//	htmlBuf, err := mailer.GenerateHTML()
//	textBuf, err := mailer.GeneratePlainText()
//
//	// Using standard net/smtp
//	msg := "From: sender@example.com\r\n" +
//	      "To: recipient@example.com\r\n" +
//	      "Subject: Welcome\r\n" +
//	      "MIME-Version: 1.0\r\n" +
//	      "Content-Type: text/html; charset=UTF-8\r\n\r\n" +
//	      htmlBuf.String()
//
//	// Or use github.com/nabbar/golib/mail/smtp for full SMTP support
//
// # Dependencies
//
// This package depends on:
//   - github.com/go-hermes/hermes/v2: Email template rendering engine
//   - github.com/nabbar/golib/errors: Structured error handling
//   - github.com/go-playground/validator/v10: Configuration validation
//
// # Performance Considerations
//
//   - Email generation typically takes 1-5ms depending on content complexity
//   - Clone operations are fast (~1µs for simple, ~10µs for complex content)
//   - Template variable replacement is performed in-place with O(n*m) complexity
//   - CSS inlining (default) adds processing time but improves email client compatibility
//
// # Best Practices
//
//   - Always validate configuration before creating a mailer from Config
//   - Use Clone() for concurrent operations to avoid race conditions
//   - Generate both HTML and plain text versions for better compatibility
//   - Test emails in multiple email clients (Gmail, Outlook, Apple Mail, etc.)
//   - Keep logo images under 200KB for faster loading
//   - Use HTTPS URLs for all links and images
//   - Include clear call-to-action buttons with descriptive text
//   - Provide a plain text alternative for accessibility
//
// # Related Packages
//
// For complete email workflow:
//   - github.com/nabbar/golib/mail/smtp: SMTP client for sending emails
//   - github.com/nabbar/golib/mail/sender: High-level email sending with attachments
//   - github.com/nabbar/golib/errors: Error handling and logging
//
// # References
//
// For more information on email body structure and available fields:
//   - Hermes documentation: https://github.com/go-hermes/hermes
//   - Email best practices: https://www.campaignmonitor.com/best-practices/
package render
