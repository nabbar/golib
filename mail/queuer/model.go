/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
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

package queuer

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"

	libsmtp "github.com/nabbar/golib/mail/smtp"
	smtpcf "github.com/nabbar/golib/mail/smtp/config"
)

// pooler is the internal implementation of the Pooler interface.
//
// It wraps an SMTP client with rate limiting functionality by composing:
//   - s: The underlying SMTP client (can be nil initially)
//   - c: The rate limiting counter that enforces throttling
//
// The pooler delegates most SMTP operations to the underlying client
// while intercepting Send() to apply rate limiting through the counter.
type pooler struct {
	s libsmtp.SMTP // Underlying SMTP client, nil if not yet configured
	c Counter      // Rate limiting counter, never nil
}

// Reset resets the internal rate limiter counter to allow immediate operations.
//
// This method delegates to the counter's Reset() which:
//   - Resets quota to maximum value
//   - Clears the time window
//   - Invokes FuncCaller callback if configured
//
// The reset operation works even if the SMTP client is nil, as it only
// affects the counter state.
//
// Returns:
//   - nil on successful reset
//   - Error from FuncCaller if the callback fails
//
// Thread-safe: Can be called concurrently with other operations.
func (p *pooler) Reset() error {
	// Reset the counter regardless of SMTP client state
	// The counter operates independently from the SMTP client
	if err := p.c.Reset(); err != nil {
		return err
	}

	return nil
}

// NewPooler creates an independent copy of this pooler with fresh state.
//
// The new pooler:
//   - Has the same configuration (Max, Wait, FuncCaller)
//   - Uses a cloned SMTP client (independent connection if client exists)
//   - Has its own counter with fresh throttle state
//   - Can be used concurrently with the original pooler
//
// This is useful for:
//   - Creating multiple poolers with the same settings
//   - Isolating rate limits across different contexts
//   - Testing with independent pooler instances
//
// Returns a new Pooler instance that operates independently.
//
// Thread-safe: Creates a deep copy with no shared mutable state.
func (p *pooler) NewPooler() Pooler {
	if p.s == nil {
		return &pooler{
			s: nil,
			c: p.c.Clone(),
		}
	} else {
		return &pooler{
			s: p.s.Clone(),
			c: p.c.Clone(),
		}
	}
}

// Send transmits an email through the SMTP server with rate limiting applied.
//
// This method:
//  1. Checks that SMTP client is configured (returns error if nil)
//  2. Calls Pool() to enforce rate limiting (may block if quota exhausted)
//  3. Delegates actual sending to the underlying SMTP client
//
// The rate limiting is transparent - the method blocks if necessary
// until quota is available or context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation and timeout. If cancelled during throttling,
//     returns ErrorMailPoolerContext.
//   - from: Sender email address (MAIL FROM)
//   - to: List of recipient email addresses (RCPT TO)
//   - data: Email content as io.WriterTo. See github.com/nabbar/golib/mail/sender
//     for creating email messages.
//
// Returns:
//   - nil if email sent successfully
//   - ErrorParamEmpty if SMTP client is not configured
//   - ErrorMailPoolerContext if context is cancelled
//   - Error from FuncCaller if callback fails during throttling
//   - SMTP-related errors from the underlying client
//
// Thread-safe: Multiple goroutines can call this concurrently, but they will
// be throttled according to the configured rate limits.
//
// Example:
//
//	err := pooler.Send(ctx, "sender@example.com",
//	    []string{"recipient@example.com"},
//	    emailMessage)
func (p *pooler) Send(ctx context.Context, from string, to []string, data io.WriterTo) error {
	if p.s == nil {
		return ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	if err := p.c.Pool(ctx); err != nil {
		return err
	}

	return p.s.Send(ctx, from, to, data)
}

// Client returns the underlying net/smtp.Client for advanced operations.
//
// This method provides access to the raw SMTP client for operations not
// directly exposed by the Pooler interface. Use with caution as direct
// client operations bypass rate limiting.
//
// Parameters:
//   - ctx: Context for connection establishment
//
// Returns:
//   - *smtp.Client if SMTP client is configured
//   - ErrorParamEmpty if SMTP client is nil
//   - Connection errors from the underlying SMTP client
//
// Note: Rate limiting is NOT applied to operations performed directly
// on the returned client. Use the pooler's Send() method for rate-limited sending.
//
// See github.com/nabbar/golib/mail/smtp for SMTP client documentation.
func (p *pooler) Client(ctx context.Context) (*smtp.Client, error) {
	if p.s == nil {
		return nil, ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	return p.s.Client(ctx)
}

// Close terminates the SMTP connection and releases resources.
//
// This method safely closes the underlying SMTP client if it exists.
// It's safe to call Close() multiple times or on a pooler with nil client.
//
// After calling Close(), the pooler should not be used for further operations
// unless UpdConfig() is called to set a new client.
//
// Thread-safe: Can be called concurrently, though the behavior of concurrent
// operations during close is undefined.
func (p *pooler) Close() {
	if p.s != nil {
		p.s.Close()
	}
}

// Check performs a health check by connecting to the SMTP server.
//
// This method verifies that the SMTP server is accessible and responding
// correctly without sending any actual email. Useful for monitoring and
// health check endpoints.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - nil if server is healthy and accessible
//   - ErrorParamEmpty if SMTP client is not configured
//   - Connection or protocol errors from the SMTP server
//
// Note: This operation does NOT consume throttle quota. Health checks
// are not rate-limited.
//
// See github.com/nabbar/golib/mail/smtp for more details on health checking.
func (p *pooler) Check(ctx context.Context) error {
	if p.s == nil {
		return ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	return p.s.Check(ctx)
}

// Clone creates a copy of this pooler implementing the SMTP interface.
//
// This is an alias for NewPooler() that returns the more general SMTP
// interface type instead of the Pooler interface. This allows the pooler
// to satisfy the SMTP interface's Clone() method.
//
// Returns a new SMTP client (actually a Pooler) with independent state.
//
// See NewPooler() for detailed behavior.
func (p *pooler) Clone() libsmtp.SMTP {
	return p.NewPooler()
}

// UpdConfig updates or initializes the SMTP client configuration.
//
// This method allows changing the SMTP server settings after the pooler
// has been created. It can also be used to initialize the SMTP client
// if the pooler was created with New(cfg, nil).
//
// Behavior:
//   - If SMTP client exists: Updates its configuration
//   - If SMTP client is nil: Attempts to create a new client
//   - Creation errors are silently ignored (as per SMTP interface contract)
//
// Parameters:
//   - cfg: New SMTP configuration (host, port, auth, etc.)
//     See github.com/nabbar/golib/mail/smtp/config for configuration details
//   - tlsConfig: TLS configuration for secure connections. Can be nil for
//     default secure settings.
//
// Thread-safe: Can be called concurrently, though behavior of operations
// during config update is undefined. It's recommended to call UpdConfig()
// before starting concurrent operations.
//
// Note: This method does not return errors as per the SMTP interface contract.
// If client creation fails, the pooler remains with nil client and subsequent
// Send/Check operations will return ErrorParamEmpty.
func (p *pooler) UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config) {
	if p.s != nil {
		p.s.UpdConfig(cfg, tslConfig)
	} else {
		// If client doesn't exist yet, try to create it
		// Silently ignore errors as per interface contract
		if cli, err := libsmtp.New(cfg, tslConfig); err == nil {
			p.s = cli
		}
	}
}
