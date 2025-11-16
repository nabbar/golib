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

package smtp

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"strings"
	"sync"

	smtpcf "github.com/nabbar/golib/mail/smtp/config"
)

// smtpClient is the internal implementation of the SMTP interface.
// It maintains the connection state, configuration, and synchronization primitives
// to ensure thread-safe operations.
//
// Fields:
//   - mut: Mutex to protect concurrent access to connection state
//   - con: The underlying network connection to the SMTP server
//   - cli: The SMTP protocol client from net/smtp package
//   - tls: TLS configuration for secure connections
//   - cfg: SMTP configuration (host, port, auth, etc.)
type smtpClient struct {
	mut sync.Mutex   // Protects concurrent access to connection state
	con net.Conn     // Underlying TCP/TLS connection
	cli *smtp.Client // SMTP protocol handler
	tls *tls.Config  // TLS configuration
	cfg smtpcf.SMTP  // SMTP server configuration
}

// _ValidateLine checks if a string contains CR (carriage return) or LF (line feed) characters.
// These characters are not allowed in SMTP commands and email addresses as they can
// be used for SMTP command injection attacks.
//
// Parameters:
//   - line: The string to validate
//
// Returns:
//   - error: ErrorSMTPLineCRLF if the line contains CR or LF; nil otherwise
func (s *smtpClient) _ValidateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return ErrorSMTPLineCRLF.Error(nil)
	}

	return nil
}

// _close performs the actual connection cleanup.
// It attempts a graceful QUIT command first, then forcefully closes the connection
// if the QUIT fails. This method is not thread-safe and should only be called
// from methods that hold the mutex lock.
//
// The method sets both cli and con to nil after closing to prevent double-close issues.
func (s *smtpClient) _close() {
	if s.cli != nil {
		// Try graceful shutdown first
		if e := s.cli.Quit(); e != nil {
			// Force close if QUIT fails
			_ = s.cli.Close()
		}
		s.cli = nil
	}

	if s.con != nil {
		_ = s.con.Close()
		s.con = nil
	}
}

// Clone creates a new SMTP client instance with the same configuration.
// The new instance has its own independent connection state (no active connection)
// but shares the same configuration and TLS settings.
//
// This is useful for creating multiple client instances for concurrent operations
// without sharing connection state, which could lead to race conditions.
//
// Returns:
//   - SMTP: A new client instance with the same configuration
func (s *smtpClient) Clone() SMTP {
	return &smtpClient{
		mut: sync.Mutex{},
		con: nil,   // New instance starts with no connection
		cli: nil,   // New instance starts with no SMTP client
		tls: s.tls, // Share TLS config (immutable)
		cfg: s.cfg, // Share SMTP config (immutable via interface)
	}
}

// Close gracefully terminates the SMTP connection and releases all resources.
// It first attempts to send a QUIT command to the server, then closes the
// underlying network connection.
//
// This method is thread-safe and can be called multiple times safely.
// After calling Close, the client should not be used for further operations
// unless UpdConfig is called to re-initialize it.
func (s *smtpClient) Close() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s._close()
}

// UpdConfig updates the SMTP configuration and TLS settings for this client instance.
// It closes any existing connection before applying the new configuration.
//
// This method is thread-safe and allows reconfiguring a client without creating
// a new instance. The new configuration will be used for all subsequent operations.
//
// Parameters:
//   - cfg: The new SMTP configuration
//   - tslConfig: The new TLS configuration
//
// Note: Any active connection will be terminated when this method is called.
func (s *smtpClient) UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config) {
	s._close() // Close existing connection first

	s.mut.Lock()
	defer s.mut.Unlock()

	s.cfg = cfg
	s.tls = tslConfig
}
