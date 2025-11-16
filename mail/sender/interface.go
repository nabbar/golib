/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

// Package sender provides a high-level API for composing and sending emails via SMTP.
//
// This package simplifies email composition by providing an intuitive interface for:
//   - Setting email headers (subject, priority, encoding, etc.)
//   - Managing recipients (To, CC, BCC)
//   - Adding body content (plain text and/or HTML)
//   - Attaching files (regular and inline attachments)
//   - Sending emails through SMTP servers
//
// # Basic Usage
//
// Creating and sending a simple email:
//
//	import (
//	    "github.com/nabbar/golib/mail/sender"
//	    "github.com/nabbar/golib/mail/smtp"
//	    "strings"
//	)
//
//	// Create a new email
//	mail := sender.New()
//	mail.SetSubject("Hello World")
//	mail.SetCharset("UTF-8")
//	mail.SetEncoding(sender.EncodingBase64)
//	mail.SetPriority(sender.PriorityNormal)
//
//	// Set sender and recipients
//	mail.Email().SetFrom("sender@example.com")
//	mail.Email().AddRecipients(sender.RecipientTo, "recipient@example.com")
//
//	// Add body content
//	body := strings.NewReader("This is the email body")
//	mail.SetBody(sender.ContentPlainText, io.NopCloser(body))
//
//	// Send the email
//	smtpClient, _ := smtp.New(smtpConfig, nil)
//	sender, _ := mail.Sender()
//	err := sender.SendClose(ctx, smtpClient)
//
// # Configuration-Based Usage
//
// Create emails from configuration:
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
//	    log.Fatal(err)
//	}
//
// # Advanced Features
//
// Multiple body parts (text + HTML):
//
//	mail.SetBody(sender.ContentPlainText, io.NopCloser(strings.NewReader("Plain text version")))
//	mail.AddBody(sender.ContentHTML, io.NopCloser(strings.NewReader("<p>HTML version</p>")))
//
// File attachments:
//
//	file, _ := os.Open("document.pdf")
//	mail.AddAttachment("document.pdf", "application/pdf", file, false)
//
// Inline attachments (for HTML emails):
//
//	image, _ := os.Open("logo.png")
//	mail.AddAttachment("logo.png", "image/png", image, true)
//
// # SMTP Integration
//
// This package integrates with github.com/nabbar/golib/mail/smtp for sending emails.
// See that package for SMTP client configuration, TLS settings, and authentication.
//
// # Error Handling
//
// All errors implement github.com/nabbar/golib/errors.Error interface, providing:
//   - Structured error codes
//   - Parent error wrapping
//   - Stack traces
//
// See error.go for specific error codes and their meanings.
//
// # Thread Safety
//
// Mail and Email objects are NOT thread-safe. Create separate instances
// for concurrent operations, or use the Clone() method:
//
//	mail1 := mail.Clone()
//	mail2 := mail.Clone()
//	// mail1 and mail2 can now be used independently
package sender

import (
	"io"
	"net/textproto"
	"time"
)

// Mail defines the main interface for composing email messages.
//
// This interface provides methods for setting all aspects of an email, including:
//   - Metadata (charset, subject, priority, encoding, date)
//   - Headers (custom and standard)
//   - Body content (plain text and/or HTML)
//   - Attachments (regular and inline)
//   - Recipients and sender information (via Email interface)
//
// Create a new Mail instance using New():
//
//	mail := sender.New()
//
// Mail objects are not thread-safe. Use Clone() for concurrent operations.
type Mail interface {
	// Clone creates a deep copy of the Mail object.
	// The cloned mail can be modified independently without affecting the original.
	//
	// This is useful for:
	//   - Sending similar emails to different recipients
	//   - Concurrent email operations
	//   - Template-based email generation
	//
	// Returns a new Mail instance with all fields copied.
	Clone() Mail

	// SetCharset sets the character encoding for the email (e.g., "UTF-8", "ISO-8859-1").
	// Default is "UTF-8", which supports all Unicode characters.
	//
	// Common charsets:
	//   - "UTF-8" (recommended, universal support)
	//   - "ISO-8859-1" (Western European)
	//   - "ISO-8859-15" (Western European with Euro symbol)
	SetCharset(charset string)

	// GetCharset returns the current character encoding setting.
	GetCharset() string

	// SetPriority sets the urgency level of the email.
	// See Priority type for available values (Normal, Low, High).
	//
	// Example:
	//	mail.SetPriority(sender.PriorityHigh)
	SetPriority(p Priority)

	// GetPriority returns the current priority setting.
	GetPriority() Priority

	// SetSubject sets the email subject line.
	// The subject appears in the recipient's inbox and email client.
	//
	// Example:
	//	mail.SetSubject("Monthly Report - January 2024")
	SetSubject(subject string)

	// GetSubject returns the current subject line.
	GetSubject() string

	// SetEncoding sets the transfer encoding for the email body.
	// See Encoding type for available values (None, Binary, Base64, QuotedPrintable).
	//
	// Recommended encodings:
	//   - EncodingBase64: For binary data and non-ASCII text
	//   - EncodingQuotedPrintable: For mostly ASCII text with occasional special characters
	//
	// Example:
	//	mail.SetEncoding(sender.EncodingBase64)
	SetEncoding(enc Encoding)

	// GetEncoding returns the current transfer encoding setting.
	GetEncoding() Encoding

	// SetDateTime sets the email date/time using a time.Time value.
	// If not set, the sending time will be used.
	//
	// Example:
	//	mail.SetDateTime(time.Now())
	SetDateTime(datetime time.Time)

	// GetDateTime returns the current date/time setting.
	// Returns zero time if not explicitly set.
	GetDateTime() time.Time

	// SetDateString parses and sets the email date/time from a string.
	// The layout parameter follows time.Parse format (e.g., time.RFC1123Z).
	//
	// Returns ErrorMailDateParsing if the string cannot be parsed.
	//
	// Example:
	//	err := mail.SetDateString(time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700")
	SetDateString(layout, datetime string) error

	// GetDateString returns the date/time as a formatted string.
	// Returns empty string if date is not set.
	GetDateString() string

	// AddHeader adds a custom header to the email.
	// Multiple values can be added for the same header key.
	//
	// Standard headers (To, From, Subject, etc.) are set via other methods.
	// Use this for custom headers like "X-Custom-ID", "Reply-To", etc.
	//
	// Example:
	//	mail.AddHeader("X-Campaign-ID", "2024-Q1-Newsletter")
	//	mail.AddHeader("X-Mailer", "MyApp/1.0")
	AddHeader(key string, values ...string)

	// GetHeader retrieves all values for a specific header key.
	// Returns empty slice if the header doesn't exist.
	GetHeader(key string) []string

	// GetHeaders returns all email headers as a textproto.MIMEHeader.
	// This includes both custom headers and standard headers set by the package.
	GetHeaders() textproto.MIMEHeader

	// SetBody sets the email body, replacing any existing body of the same content type.
	// The body is provided as an io.ReadCloser for efficient streaming.
	//
	// Parameters:
	//   - ct: Content type (ContentPlainText or ContentHTML)
	//   - body: Body content as an io.ReadCloser
	//
	// Example:
	//	body := io.NopCloser(strings.NewReader("Email content"))
	//	mail.SetBody(sender.ContentPlainText, body)
	SetBody(ct ContentType, body io.ReadCloser)

	// AddBody adds an additional body part without removing existing ones.
	// Use this to provide both plain text and HTML versions of the email.
	//
	// Best practice: Always include a plain text version for maximum compatibility.
	//
	// Example:
	//	mail.SetBody(sender.ContentPlainText, plainBody)
	//	mail.AddBody(sender.ContentHTML, htmlBody)
	AddBody(ct ContentType, body io.ReadCloser)

	// GetBody returns all body parts added to the email.
	GetBody() []Body

	// SetAttachment sets or replaces an attachment with the given name.
	// If an attachment with the same name exists, it will be replaced.
	//
	// Parameters:
	//   - name: Filename as it appears in the email
	//   - mime: MIME type (e.g., "application/pdf", "image/png")
	//   - data: File content as io.ReadCloser
	//   - inline: true for inline attachments (embedded in HTML), false for regular attachments
	//
	// The data ReadCloser will be closed automatically after sending.
	SetAttachment(name string, mime string, data io.ReadCloser, inline bool)

	// AddAttachment adds an attachment without checking for duplicates.
	// Multiple attachments with the same name can exist.
	//
	// Parameters match SetAttachment.
	//
	// Example:
	//	file, _ := os.Open("report.pdf")
	//	mail.AddAttachment("report.pdf", "application/pdf", file, false)
	AddAttachment(name string, mime string, data io.ReadCloser, inline bool)

	// AttachFile is a convenience method for adding file attachments.
	// The filename is extracted from the filepath parameter.
	//
	// Deprecated: Use AddAttachment directly for better control over the filename.
	AttachFile(filepath string, data io.ReadCloser, inline bool)

	// GetAttachment returns all attachments of the specified type.
	//
	// Parameters:
	//   - inline: true to get inline attachments, false for regular attachments
	//
	// Returns a slice of File objects.
	GetAttachment(inline bool) []File

	// Email returns the Email interface for managing sender and recipient addresses.
	// Use this to set From, To, CC, BCC, ReplyTo, and other address fields.
	//
	// Example:
	//	mail.Email().SetFrom("sender@example.com")
	//	mail.Email().AddRecipients(sender.RecipientTo, "user@example.com")
	Email() Email

	// Sender creates a Sender instance that can send the email via SMTP.
	//
	// This method validates the email structure (addresses, recipients) and
	// prepares it for transmission. Returns ErrorMailSenderInit if validation fails.
	//
	// Example:
	//	sender, err := mail.Sender()
	//	if err != nil {
	//	    return err
	//	}
	//	defer sender.Close()
	//	err = sender.Send(ctx, smtpClient)
	//
	// Returns a Sender instance and nil error on success, or nil and an error on failure.
	Sender() (Sender, error)
}

// New creates a new Mail instance with default settings.
//
// The returned Mail is initialized with:
//   - Charset: UTF-8
//   - Encoding: EncodingNone
//   - MIME-Version: 1.0 header
//   - Empty recipient lists
//   - No attachments or body content
//
// After creation, configure the email by setting:
//   - Subject, priority, and encoding
//   - Sender and recipient addresses (via Email())
//   - Body content and attachments
//
// Example:
//
//	mail := sender.New()
//	mail.SetSubject("Welcome")
//	mail.SetEncoding(sender.EncodingBase64)
//	mail.Email().SetFrom("noreply@example.com")
//	mail.Email().AddRecipients(sender.RecipientTo, "user@example.com")
//
// Returns a new Mail instance ready for configuration.
func New() Mail {
	m := &mail{
		headers:  make(textproto.MIMEHeader),
		charset:  "UTF-8",
		encoding: EncodingNone,
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

	return m
}

// Clone creates a deep copy of the mail object.
// Implementation of the Mail.Clone() interface method.
//
// All fields are copied including:
//   - Headers, charset, subject, date
//   - Priority and encoding settings
//   - Email addresses (from, sender, replyTo, returnPath, recipients)
//   - Attachments and inline files
//   - Body content
//
// Note: File readers (attachments and body) are shared between the original
// and clone. If you need independent file access, reopen the files.
//
// Returns a new Mail instance with copied values.
func (m *mail) Clone() Mail {
	return &mail{
		date:    m.date,
		attach:  m.attach,
		inline:  m.inline,
		body:    m.body,
		charset: m.charset,
		subject: m.subject,
		headers: m.headers,
		address: &email{
			from:       m.address.from,
			sender:     m.address.sender,
			replyTo:    m.address.replyTo,
			returnPath: m.address.returnPath,
			to:         m.address.to,
			cc:         m.address.cc,
			bcc:        m.address.bcc,
		},
		encoding: m.encoding,
		priority: m.priority,
	}
}

// Email defines the interface for managing email addresses.
//
// This interface handles all address-related operations:
//   - From: The sender's email address (required)
//   - Sender: Optional explicit sender (if different from From)
//   - ReplyTo: Where replies should be sent (defaults to From)
//   - ReturnPath: Bounce address (defaults to From)
//   - Recipients: To, CC, and BCC recipient lists
//
// Access via Mail.Email():
//
//	mail.Email().SetFrom("sender@example.com")
//	mail.Email().AddRecipients(sender.RecipientTo, "user1@example.com", "user2@example.com")
//	mail.Email().AddRecipients(sender.RecipientCC, "manager@example.com")
//
// # Address Fallback Behavior
//
// If optional address fields are not set, they fall back to other addresses:
//   - Sender: Falls back to ReplyTo, then ReturnPath, then From
//   - ReplyTo: Falls back to Sender, then ReturnPath, then From
//   - ReturnPath: Falls back to ReplyTo, then Sender, then From
//
// This ensures that email headers always have valid addresses even if
// only From is explicitly set.
type Email interface {
	// SetFrom sets the primary sender email address.
	// This is the address that appears in the "From" header and is required.
	//
	// The From address is used as a fallback for Sender, ReplyTo, and ReturnPath
	// if those fields are not explicitly set.
	//
	// Example:
	//	mail.Email().SetFrom("noreply@example.com")
	SetFrom(mail string)

	// GetFrom returns the From email address.
	GetFrom() string

	// SetSender sets the actual sender email address if different from From.
	// This appears in the "Sender" header and is used when sending on behalf of someone.
	//
	// Use cases:
	//   - Mailing list managers sending on behalf of list members
	//   - Automated systems sending on behalf of users
	//   - Delegation scenarios
	//
	// If not set, falls back to ReplyTo, then ReturnPath, then From.
	//
	// Example:
	//	mail.Email().SetFrom("user@example.com")
	//	mail.Email().SetSender("system@example.com")  // System sending for user
	SetSender(mail string)

	// GetSender returns the Sender address, with fallback to other addresses if not set.
	GetSender() string

	// SetReplyTo sets where replies to this email should be sent.
	// This appears in the "Reply-To" header.
	//
	// Use cases:
	//   - Directing replies to a different address than From
	//   - Reply-to addresses for no-reply emails
	//   - Department or team email addresses
	//
	// If not set, falls back to Sender, then ReturnPath, then From.
	//
	// Example:
	//	mail.Email().SetFrom("noreply@example.com")
	//	mail.Email().SetReplyTo("support@example.com")  // Replies go to support
	SetReplyTo(mail string)

	// GetReplyTo returns the ReplyTo address, with fallback to other addresses if not set.
	GetReplyTo() string

	// SetReturnPath sets the return path (bounce address) for the email.
	// This is used for delivery failure notifications.
	//
	// Use cases:
	//   - Dedicated bounce processing addresses
	//   - Separating bounce handling from regular email
	//   - Email deliverability monitoring
	//
	// If not set, falls back to ReplyTo, then Sender, then From.
	//
	// Example:
	//	mail.Email().SetFrom("noreply@example.com")
	//	mail.Email().SetReturnPath("bounces@example.com")  // Bounces go here
	SetReturnPath(mail string)

	// GetReturnPath returns the ReturnPath address, with fallback to other addresses if not set.
	GetReturnPath() string

	// SetRecipients replaces all recipients of the specified type with the provided list.
	// This removes any existing recipients of that type.
	//
	// Use this when you want to completely replace the recipient list.
	// For adding to existing recipients, use AddRecipients instead.
	//
	// Parameters:
	//   - rt: Recipient type (RecipientTo, RecipientCC, or RecipientBCC)
	//   - rcpt: Email addresses (can be empty to clear all recipients)
	//
	// Example:
	//	mail.Email().SetRecipients(sender.RecipientTo, "user1@example.com", "user2@example.com")
	//	mail.Email().SetRecipients(sender.RecipientCC)  // Clear all CC recipients
	SetRecipients(rt recipientType, rcpt ...string)

	// AddRecipients adds recipients to the specified type without removing existing ones.
	// Duplicate addresses are automatically prevented.
	//
	// Use this when you want to add recipients while preserving existing ones.
	//
	// Parameters:
	//   - rt: Recipient type (RecipientTo, RecipientCC, or RecipientBCC)
	//   - rcpt: Email addresses to add
	//
	// Example:
	//	mail.Email().AddRecipients(sender.RecipientTo, "user@example.com")
	//	mail.Email().AddRecipients(sender.RecipientTo, "another@example.com")  // Both preserved
	//	mail.Email().AddRecipients(sender.RecipientCC, "manager@example.com")
	AddRecipients(rt recipientType, rcpt ...string)

	// GetRecipients returns all recipients of the specified type.
	//
	// Parameters:
	//   - rt: Recipient type (RecipientTo, RecipientCC, or RecipientBCC)
	//
	// Returns a slice of email addresses, or empty slice if none exist.
	//
	// Example:
	//	toList := mail.Email().GetRecipients(sender.RecipientTo)
	//	ccList := mail.Email().GetRecipients(sender.RecipientCC)
	GetRecipients(rt recipientType) []string
}
