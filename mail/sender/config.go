/*
 * MIT License
 *
 * Copyright (c) 2021 Nicolas JUHEL
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
 *
 */

package sender

import (
	"fmt"
	"net/textproto"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
)

// Config provides a declarative way to configure email messages.
// It supports JSON, YAML, TOML, and Viper mapstructure tags for easy
// integration with configuration management systems.
//
// All fields are validated using github.com/go-playground/validator/v10.
// Use Config.Validate() to check for errors before creating emails.
//
// Example JSON configuration:
//
//	{
//	  "charset": "UTF-8",
//	  "subject": "Welcome Email",
//	  "encoding": "Base 64",
//	  "priority": "Normal",
//	  "from": "noreply@example.com",
//	  "to": ["user@example.com"],
//	  "attach": [
//	    {"name": "report.pdf", "mime": "application/pdf", "path": "/path/to/report.pdf"}
//	  ]
//	}
//
// Create a Mail instance from configuration:
//
//	config := sender.Config{ /* ... */ }
//	if err := config.Validate(); err != nil {
//	    return err
//	}
//	mail, err := config.NewMailer()
type Config struct {
	// Charset specifies the character encoding for the email.
	// Required. Common values: "UTF-8", "ISO-8859-1", "ISO-8859-15".
	//
	// Example: "UTF-8"
	Charset string `json:"charset" yaml:"charset" toml:"charset" mapstructure:"charset" validate:"required"`

	// Subject is the email subject line.
	// Required. Appears in the recipient's inbox.
	//
	// Example: "Monthly Report - January 2024"
	Subject string `json:"subject" yaml:"subject" toml:"subject" mapstructure:"subject" validate:"required"`

	// Encoding specifies the transfer encoding for the email body.
	// Required. Valid values: "None", "Binary", "Base 64", "Quoted Printable".
	// Case-insensitive. Parsed by ParseEncoding().
	//
	// Recommended:
	//   - "Base 64" for binary data and non-ASCII text
	//   - "Quoted Printable" for mostly ASCII text
	//
	// Example: "Base 64"
	Encoding string `json:"encoding" yaml:"encoding" toml:"encoding" mapstructure:"encoding" validate:"required"`

	// Priority specifies the urgency level of the email.
	// Required. Valid values: "Normal", "Low", "High".
	// Case-insensitive. Parsed by ParsePriority().
	//
	// Example: "Normal"
	Priority string `json:"priority" yaml:"priority" toml:"priority" mapstructure:"priority" validate:"required"`

	// Headers is a map of custom email headers to add.
	// Optional. Use for custom headers like "X-Campaign-ID", "X-Mailer", etc.
	// Standard headers (To, From, Subject) are set via other fields.
	//
	// Example: {"X-Campaign-ID": "2024-Q1", "X-Mailer": "MyApp/1.0"}
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty" mapstructure:"headers,omitempty"`

	// From is the sender's email address (required).
	// This address appears in the "From" header and is mandatory.
	// Used as fallback for Sender, ReplyTo, and ReturnPath if they're not set.
	//
	// Validated as a proper email address format.
	//
	// Example: "noreply@example.com"
	From string `json:"from" yaml:"from" toml:"from" mapstructure:"from" validate:"required,email"`

	// Sender is the actual sender email if different from From (optional).
	// Use when sending on behalf of someone else (e.g., mailing lists, delegation).
	// If not set, falls back to ReplyTo, ReturnPath, then From.
	//
	// Validated as email if provided.
	//
	// Example: "system@example.com"
	Sender string `json:"sender,omitempty" yaml:"sender,omitempty" toml:"sender,omitempty" mapstructure:"sender,omitempty" validate:"omitempty,email"`

	// ReplyTo specifies where replies should be sent (optional).
	// If not set, falls back to Sender, ReturnPath, then From.
	//
	// Validated as email if provided.
	//
	// Example: "support@example.com"
	ReplyTo string `json:"replyTo,omitempty" yaml:"replyTo,omitempty" toml:"replyTo,omitempty" mapstructure:"replyTo,omitempty" validate:"omitempty,email"`

	// ReturnPath specifies the bounce address (optional).
	// Used for delivery failure notifications.
	// If not set, falls back to ReplyTo, Sender, then From.
	//
	// Useful when the sender's IP is not public or for bounce tracking.
	//
	// Example: "bounces@example.com"
	ReturnPath string `json:"returnPath,omitempty" yaml:"returnPath,omitempty" toml:"returnPath,omitempty" mapstructure:"returnPath,omitempty"`

	// To is the list of primary recipients (optional).
	// Each email address is validated for proper format.
	//
	// Example: ["user1@example.com", "user2@example.com"]
	To []string `json:"to,omitempty" yaml:"to,omitempty" toml:"to,omitempty" mapstructure:"to,omitempty" validate:"dive,email"`

	// Cc is the list of carbon copy recipients (optional).
	// These recipients are visible to all other recipients.
	// Each email address is validated for proper format.
	//
	// Example: ["manager@example.com"]
	Cc []string `json:"cc,omitempty" yaml:"cc,omitempty" toml:"cc,omitempty" mapstructure:"cc,omitempty" validate:"dive,email"`

	// Bcc is the list of blind carbon copy recipients (optional).
	// These recipients are hidden from all other recipients.
	// Each email address is validated for proper format.
	//
	// Example: ["archive@example.com", "audit@example.com"]
	Bcc []string `json:"bcc,omitempty" yaml:"bcc,omitempty" toml:"bcc,omitempty" mapstructure:"bcc,omitempty" validate:"dive,email"`

	// Attach is the list of files to attach to the email (optional).
	// Files are attached as regular attachments visible in the attachment list.
	// Each ConfigFile is validated for required fields and file existence.
	//
	// See ConfigFile for file specification format.
	Attach []ConfigFile `json:"attach,omitempty" yaml:"attach,omitempty" toml:"attach,omitempty" mapstructure:"attach,omitempty" validate:"dive"`

	// Inline is the list of files to embed inline in the email body (optional).
	// Inline files are typically images referenced in HTML content.
	// Each ConfigFile is validated for required fields and file existence.
	//
	// Use this for images embedded in HTML emails: <img src="cid:logo.png">
	//
	// See ConfigFile for file specification format.
	Inline []ConfigFile `json:"inline,omitempty" yaml:"inline,omitempty" toml:"inline,omitempty" mapstructure:"inline,omitempty" validate:"dive"`
}

// ConfigFile specifies a file attachment or inline file configuration.
//
// Example JSON:
//
//	{
//	  "name": "report.pdf",
//	  "mime": "application/pdf",
//	  "path": "/var/data/reports/monthly.pdf"
//	}
//
// Common MIME types:
//   - "application/pdf" - PDF documents
//   - "image/jpeg", "image/png" - Images
//   - "text/plain" - Text files
//   - "application/zip" - ZIP archives
type ConfigFile struct {
	// Name is the filename as it will appear in the email.
	// Required.
	//
	// Example: "report.pdf"
	Name string `json:"name" yaml:"name" toml:"name" mapstructure:"name" validate:"required"`

	// Mime is the MIME type of the file.
	// Required.
	//
	// Example: "application/pdf"
	Mime string `json:"mime" yaml:"mime" toml:"mime" mapstructure:"mime" validate:"required"`

	// Path is the filesystem path to the file.
	// Required. Must be an existing file accessible to the application.
	// Validated using the "file" constraint.
	//
	// Example: "/var/data/reports/monthly.pdf"
	Path string `json:"path" yaml:"path" toml:"path" mapstructure:"path" validate:"required,file"`
}

// Validate checks the Config for errors using struct validation tags.
//
// This method uses github.com/go-playground/validator/v10 to validate:
//   - Required fields (Charset, Subject, Encoding, Priority, From)
//   - Email address formats (From, Sender, ReplyTo, To, Cc, Bcc)
//   - File paths for attachments (Attach, Inline)
//
// Returns nil if validation passes, or ErrorMailConfigInvalid with
// detailed parent errors describing what validation failed.
//
// Example:
//
//	config := sender.Config{
//	    Charset:  "UTF-8",
//	    Subject:  "Test",
//	    Encoding: "Base 64",
//	    Priority: "Normal",
//	    From:     "sender@example.com",
//	}
//	if err := config.Validate(); err != nil {
//	    // Handle validation error
//	    log.Printf("Config validation failed: %v", err)
//	    return err
//	}
//
// See github.com/go-playground/validator/v10 for validation tag documentation.
func (c Config) Validate() liberr.Error {
	err := ErrorMailConfigInvalid.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

// NewMailer creates a Mail instance from the configuration.
//
// This method:
//  1. Creates a new Mail with default settings
//  2. Applies all configuration values (charset, subject, encoding, priority)
//  3. Sets email addresses (from, sender, replyTo, returnPath, recipients)
//  4. Opens and attaches configured files
//  5. Adds custom headers
//
// File handling:
//   - Files specified in Attach and Inline are opened using github.com/nabbar/golib/file/progress
//   - If a file cannot be opened, returns ErrorFileOpenCreate with parent error
//   - The calling code is responsible for eventually closing file handles via Sender.Close()
//
// Address fallback logic:
//   - If optional addresses (Sender, ReplyTo, ReturnPath) are not set,
//     they fall back to other addresses according to Email interface rules
//
// Returns:
//   - Mail instance and nil error on success
//   - nil and ErrorFileOpenCreate if a file cannot be opened
//
// Example:
//
//	config := sender.Config{
//	    Charset:  "UTF-8",
//	    Subject:  "Welcome",
//	    Encoding: "Base 64",
//	    Priority: "Normal",
//	    From:     "noreply@example.com",
//	    To:       []string{"user@example.com"},
//	}
//
//	mail, err := config.NewMailer()
//	if err != nil {
//	    return err
//	}
//
//	// Add body and send
//	mail.SetBody(sender.ContentPlainText, body)
//	sender, _ := mail.Sender()
//	err = sender.SendClose(ctx, smtpClient)
//
// See github.com/nabbar/golib/file/progress for file handling details.
func (c Config) NewMailer() (Mail, liberr.Error) {
	m := &mail{
		headers:  make(textproto.MIMEHeader),
		charset:  "UTF-8",
		encoding: ParseEncoding(c.Encoding),
		priority: ParsePriority(c.Priority),
		address: &email{
			from:       "",
			sender:     "",
			replyTo:    "",
			returnPath: "",
			to:         make([]string, 0),
			cc:         make([]string, 0),
			bcc:        make([]string, 0),
		},
		attach: make([]File, 0),
		inline: make([]File, 0),
		body:   make([]Body, 0),
	}

	m.headers.Set("MIME-Version", "1.0")
	if len(c.Headers) > 0 {
		for k, v := range c.Headers {
			m.headers.Set(k, v)
		}
	}

	if c.Charset != "" {
		m.charset = c.Charset
	}

	if c.Subject != "" {
		m.SetSubject(c.Subject)
	}

	m.Email().SetFrom(c.From)

	if c.Sender != "" {
		m.Email().SetSender(c.Sender)
	}

	if c.ReplyTo != "" {
		m.Email().SetReplyTo(c.ReplyTo)
	} else if c.Sender != "" {
		m.Email().SetReplyTo(c.Sender)
	}

	if c.ReturnPath != "" {
		m.Email().SetReturnPath(c.ReturnPath)
	} else if c.ReplyTo != "" {
		m.Email().SetReturnPath(c.ReplyTo)
	} else if c.Sender != "" {
		m.Email().SetReturnPath(c.Sender)
	}

	if len(c.To) > 0 {
		m.Email().AddRecipients(RecipientTo, c.To...)
	}

	if len(c.Cc) > 0 {
		m.Email().AddRecipients(RecipientCC, c.Cc...)
	}

	if len(c.Bcc) > 0 {
		m.Email().AddRecipients(RecipientBCC, c.Bcc...)
	}

	if len(c.Attach) > 0 {
		for _, f := range c.Attach {
			if h, e := libfpg.Open(f.Path); e != nil {
				return nil, ErrorFileOpenCreate.Error(e)
			} else {
				m.AddAttachment(f.Name, f.Mime, h, false)
			}
		}
	}

	if len(c.Inline) > 0 {
		for _, f := range c.Inline {
			if h, e := libfpg.Open(f.Path); e != nil {
				return nil, ErrorFileOpenCreate.Error(e)
			} else {
				m.AddAttachment(f.Name, f.Mime, h, true)
			}
		}
	}

	return m, nil
}
