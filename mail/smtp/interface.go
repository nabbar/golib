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

// Package smtp provides a high-level SMTP client implementation with support for
// TLS/STARTTLS connections, authentication, and health monitoring.
//
// This package wraps the standard net/smtp package to provide a more convenient
// and feature-rich interface for sending emails through SMTP servers.
//
// Key features:
//   - Multiple TLS modes (None, STARTTLS, Strict TLS)
//   - Connection pooling and reuse
//   - Health check monitoring via github.com/nabbar/golib/monitor/types
//   - Thread-safe operations
//   - Configurable via github.com/nabbar/golib/mail/smtp/config
//
// Example usage:
//
//	cfg, err := config.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	cfg.SetHost("smtp.example.com")
//	cfg.SetPort(587)
//	cfg.SetTLSMode(tlsmode.TLSModeSTARTTLS)
//
//	client, err := smtp.New(cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	email := &MyEmail{} // implements io.WriterTo
//	err = client.Send(ctx, "from@example.com", []string{"to@example.com"}, email)
package smtp

import (
	"context"
	"crypto/tls"
	"io"
	"net/smtp"
	"sync"

	smtpcf "github.com/nabbar/golib/mail/smtp/config"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// SMTP defines the interface for SMTP client operations.
// All methods are thread-safe and can be called concurrently.
//
// The SMTP client manages connections to SMTP servers, handles TLS negotiation,
// authentication, and provides health monitoring capabilities.
//
// See github.com/nabbar/golib/mail/smtp/config for configuration options.
type SMTP interface {
	// Clone creates a deep copy of the SMTP client with the same configuration.
	// The cloned client maintains its own connection state and can be used
	// independently from the original client.
	//
	// This is useful for concurrent operations or when you need multiple
	// independent clients with the same configuration.
	Clone() SMTP

	// Close terminates the SMTP connection and releases all associated resources.
	// After calling Close, the client should not be used for further operations.
	//
	// It is safe to call Close multiple times; subsequent calls will be no-ops.
	Close()

	// UpdConfig updates the client's configuration and TLS settings.
	// This method is thread-safe and will apply the new configuration
	// for subsequent operations. Existing connections may be closed and
	// re-established with the new configuration.
	//
	// Parameters:
	//   - cfg: The new SMTP configuration (see github.com/nabbar/golib/mail/smtp/config)
	//   - tlsConfig: The new TLS configuration; if nil, uses secure defaults (TLS 1.2+)
	UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config)

	// Client establishes a connection to the SMTP server and returns the underlying
	// net/smtp.Client for advanced operations.
	//
	// The returned client should be used immediately and not stored for later use.
	// The connection will be closed when the SMTP client is closed.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//
	// Returns:
	//   - *smtp.Client: The connected SMTP client from net/smtp package
	//   - error: Any error encountered during connection establishment
	Client(ctx context.Context) (*smtp.Client, error)

	// Check performs a health check by connecting to the SMTP server and
	// verifying that it responds correctly. This does not send any email.
	//
	// This is useful for monitoring purposes and can be called periodically
	// to ensure the SMTP server is accessible and responding.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//
	// Returns:
	//   - error: nil if the server is healthy; otherwise an error describing the issue
	Check(ctx context.Context) error

	// Send transmits an email through the SMTP server.
	//
	// The email content is provided via an io.WriterTo implementation, which allows
	// for efficient streaming of email data without loading everything into memory.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - from: The sender's email address (MAIL FROM)
	//   - to: List of recipient email addresses (RCPT TO)
	//   - data: Email content implementing io.WriterTo interface
	//
	// Returns:
	//   - error: nil if the email was sent successfully; otherwise an error
	//
	// Note: The from and to addresses should not contain CR or LF characters.
	// The SMTP client handles authentication if configured.
	Send(ctx context.Context, from string, to []string, data io.WriterTo) error

	// Monitor creates a health monitoring instance for this SMTP client.
	//
	// The monitor can be used to integrate with monitoring systems and perform
	// periodic health checks. See github.com/nabbar/golib/monitor/types for
	// more information on monitoring capabilities.
	//
	// Parameters:
	//   - ctx: Context for the monitor lifecycle
	//   - vrs: Version information for monitoring metadata
	//
	// Returns:
	//   - montps.Monitor: A monitor instance for this SMTP client
	//   - error: Any error encountered during monitor creation
	Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error)
}

// New creates a new SMTP client with the given configuration and TLS settings.
//
// This is the primary constructor for creating SMTP clients. The client will be
// ready to use immediately after creation, though actual connections are established
// lazily when needed.
//
// Parameters:
//   - cfg: SMTP configuration defining host, port, auth credentials, and TLS mode.
//     See github.com/nabbar/golib/mail/smtp/config for configuration details.
//     Must not be nil.
//   - tlsConfig: TLS configuration for secure connections. If nil, a default
//     secure configuration will be used (TLS 1.2 minimum, TLS 1.3 maximum).
//
// Returns:
//   - SMTP: A configured SMTP client ready for use
//   - error: ErrorParamEmpty if cfg is nil; nil otherwise
//
// Example:
//
//	cfg, _ := config.New()
//	cfg.SetHost("smtp.gmail.com")
//	cfg.SetPort(587)
//	cfg.SetTLSMode(tlsmode.TLSStartTLS)
//	cfg.SetAuth("username", "password")
//
//	client, err := smtp.New(cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func New(cfg smtpcf.SMTP, tlsConfig *tls.Config) (SMTP, error) {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}
	}

	if cfg == nil {
		return nil, ErrorParamEmpty.Error(nil)
	} else {
		return &smtpClient{
			mut: sync.Mutex{},
			cfg: cfg,
			tls: tlsConfig,
		}, nil
	}
}
