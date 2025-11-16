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

// Package queuer provides a rate-limiting wrapper around SMTP clients.
//
// This package implements a pooler pattern with configurable throttling to control
// email sending rate. It prevents overwhelming SMTP servers by limiting the number
// of emails sent within a specified time window.
//
// Key features:
//   - Configurable rate limiting (max emails per time duration)
//   - Thread-safe operations with proper mutex protection
//   - Optional callback function for throttle events
//   - Compatible with github.com/nabbar/golib/mail/smtp interface
//   - Independent counter state for each pooler instance
//
// Basic usage:
//
//	// Create configuration with rate limit of 10 emails per second
//	cfg := &queuer.Config{
//	    Max:  10,
//	    Wait: 1 * time.Second,
//	}
//
//	// Create SMTP client
//	smtpCli, _ := smtp.New(smtpConfig, tlsConfig)
//
//	// Create throttled pooler
//	pooler := queuer.New(cfg, smtpCli)
//	defer pooler.Close()
//
//	// Send emails with automatic throttling
//	err := pooler.Send(ctx, "from@example.com", []string{"to@example.com"}, message)
//
// The pooler automatically throttles requests when the maximum number of emails
// per time window is reached. It will sleep until the next available time slot.
//
// See also:
//   - github.com/nabbar/golib/mail/smtp for SMTP client implementation
//   - github.com/nabbar/golib/mail/sender for email composition
package queuer

import (
	libsmtp "github.com/nabbar/golib/mail/smtp"
)

// Pooler defines the interface for a rate-limited SMTP client.
//
// It extends the SMTP interface from github.com/nabbar/golib/mail/smtp with
// additional pooler-specific methods for managing the rate limiter state.
//
// All methods are thread-safe and can be called concurrently from multiple
// goroutines. Each Pooler instance maintains its own independent throttling state.
type Pooler interface {
	// Reset resets the internal rate limiter counter to its maximum value.
	// This allows immediate sending of Max emails without waiting.
	//
	// The Reset operation is thread-safe and can be called while other
	// operations are in progress. If a FuncCaller is configured, it will
	// be invoked during the reset (unless throttling is disabled).
	//
	// Returns an error if the configured FuncCaller returns an error.
	Reset() error

	// NewPooler creates an independent copy of the pooler with its own state.
	//
	// The new pooler:
	//   - Has the same configuration (Max, Wait, FuncCaller)
	//   - Maintains its own independent throttle counter
	//   - Uses a cloned SMTP client (if present)
	//   - Has a fresh time window starting from creation
	//
	// This is useful for creating multiple independent poolers with the same
	// settings but different throttle states.
	NewPooler() Pooler

	// Embedded SMTP interface provides all standard SMTP client operations
	// with automatic rate limiting applied to Send operations.
	//
	// See github.com/nabbar/golib/mail/smtp.SMTP for detailed documentation
	// of available methods: Send, Client, Check, Close, Clone, UpdConfig, Monitor.
	libsmtp.SMTP
}

// New creates a new rate-limited SMTP pooler with the given configuration.
//
// Parameters:
//   - cfg: Configuration specifying rate limits (Max emails per Wait duration).
//     Must not be nil. See Config for details on throttling behavior.
//   - cli: SMTP client to wrap with rate limiting. If nil, a pooler is created
//     but Send/Check/Client operations will return errors until a client is
//     provided via UpdConfig.
//
// The created pooler starts with a full quota (can send Max emails immediately).
// If cli is not nil, it will be cloned to ensure the pooler has its own
// independent SMTP connection state.
//
// Example:
//
//	// Create pooler with rate limit of 100 emails per minute
//	cfg := &queuer.Config{
//	    Max:  100,
//	    Wait: 1 * time.Minute,
//	}
//	pooler := queuer.New(cfg, smtpClient)
//
// Example with callback for monitoring:
//
//	cfg := &queuer.Config{
//	    Max:  50,
//	    Wait: 1 * time.Second,
//	}
//	cfg.SetFuncCaller(func() error {
//	    log.Println("Throttle limit reached, waiting...")
//	    return nil
//	})
//	pooler := queuer.New(cfg, smtpClient)
//
// Returns a Pooler instance ready to use. The returned pooler is thread-safe
// and can be used concurrently from multiple goroutines.
func New(cfg *Config, cli libsmtp.SMTP) Pooler {
	if cli == nil {
		return &pooler{
			s: nil,
			c: newCounter(cfg.Max, cfg.Wait, cfg._fct),
		}
	} else {
		return &pooler{
			s: cli.Clone(),
			c: newCounter(cfg.Max, cfg.Wait, cfg._fct),
		}
	}
}
