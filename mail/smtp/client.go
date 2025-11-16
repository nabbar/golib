/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package smtp

import (
	"context"
	"fmt"
	"io"
	"net/smtp"

	libptc "github.com/nabbar/golib/network/protocol"
)

// _client establishes and returns an SMTP client connection.
// It manages connection reuse, reconnection on failures, and TLS configuration.
//
// The method follows this logic:
//  1. Returns existing connection if both client and connection are valid
//  2. Cleans up partial connections (client without connection or vice versa)
//  3. Establishes a new connection with TLS configuration
//  4. Performs authentication if configured
//
// This method is not thread-safe and must be called while holding the mutex.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *smtp.Client: Connected and authenticated SMTP client
//   - error: Any error during connection or authentication
func (s *smtpClient) _client(ctx context.Context) (*smtp.Client, error) {
	if s.cli != nil && s.con != nil {
		return s.cli, nil
	}

	if s.con == nil && s.cli != nil {
		if e := s.cli.Quit(); e != nil {
			_ = s.cli.Close()
		}
	} else if s.con != nil && s.cli == nil {
		_ = s.con.Close()
	}

	s.cli = nil
	s.con = nil

	var (
		addr = s.cfg.GetHost()
		tlsc = s.tls.Clone()
	)

	if s.cfg.GetTlSServerName() != "" && s.cfg.GetNet() != libptc.NetworkUnix && s.cfg.GetNet() != libptc.NetworkUnixGram {
		tlsc.ServerName = s.cfg.GetTlSServerName()
	}

	if s.cfg.IsTLSSkipVerify() && s.cfg.GetNet() != libptc.NetworkUnix && s.cfg.GetNet() != libptc.NetworkUnixGram {
		tlsc.InsecureSkipVerify = true
	}

	if s.cfg.GetPort() > 0 {
		addr = fmt.Sprintf("%s:%v", s.cfg.GetHost(), s.cfg.GetPort())
	}

	if cli, con, err := s.tryClient(ctx, addr, tlsc); err != nil {
		return nil, err
	} else if err = s.auth(cli, addr); err != nil {
		return nil, err
	} else {
		s.con = con
		s.cli = cli
	}

	return s.cli, nil
}

// Client establishes a connection to the SMTP server and returns the client.
// This is the thread-safe public wrapper around _client.
//
// The returned client is connected and ready for SMTP commands. It should not
// be stored for later use as it may become invalid if the connection is closed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *smtp.Client: Connected SMTP client from net/smtp package
//   - error: ErrorSMTPClientInit or other errors during connection establishment
func (s *smtpClient) Client(ctx context.Context) (*smtp.Client, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s._client(ctx)
}

// Check performs a health check by connecting to the SMTP server and sending a NOOP command.
// The connection is closed after the check, regardless of success or failure.
//
// This is useful for monitoring and verifying that the SMTP server is accessible
// and responding to commands without actually sending any email.
//
// The check process:
//  1. Establishes a connection (with authentication if configured)
//  2. Sends a NOOP (No Operation) command to verify server responsiveness
//  3. Closes the connection
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: nil if the server is healthy; ErrorSMTPClientNoop or connection errors otherwise
func (s *smtpClient) Check(ctx context.Context) error {
	s.mut.Lock()
	defer func() {
		s._close()
		s.mut.Unlock()
	}()

	if c, err := s._client(ctx); err != nil {
		return err
	} else if e := c.Noop(); e != nil {
		return ErrorSMTPClientNoop.Error(e)
	}

	return nil
}

// Send transmits an email through the SMTP server.
// It handles the complete SMTP conversation: connection, authentication, and message transmission.
// The connection is always closed after sending, regardless of success or failure.
//
// The send process follows the SMTP protocol:
//  1. Validates from and to addresses for CR/LF injection
//  2. Establishes connection (with authentication if configured)
//  3. Sends NOOP to verify connection
//  4. Issues MAIL FROM command
//  5. Issues RCPT TO commands for each recipient
//  6. Opens DATA command and streams email content
//  7. Closes connection
//
// The email content is streamed via io.WriterTo, which allows efficient
// transmission without loading the entire email into memory.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - from: Sender email address (must not contain CR or LF)
//   - to: List of recipient email addresses (must not contain CR or LF)
//   - data: Email content implementing io.WriterTo interface
//
// Returns:
//   - error: nil if sent successfully; otherwise one of:
//   - ErrorSMTPLineCRLF: if from or to addresses contain invalid characters
//   - ErrorSMTPClientInit: if connection fails
//   - ErrorSMTPClientNoop: if NOOP command fails
//   - ErrorSMTPClientMail: if MAIL FROM fails
//   - ErrorSMTPClientRcpt: if any RCPT TO fails
//   - ErrorSMTPClientData: if DATA command fails
//   - ErrorSMTPClientWrite: if writing email content fails
//
// Note: This implementation is based on net/smtp.SendMail but with improvements
// for connection management and error handling.
func (s *smtpClient) Send(ctx context.Context, from string, to []string, data io.WriterTo) error {

	var (
		e error
		c *smtp.Client
		w io.WriteCloser
	)

	s.mut.Lock()

	defer func() {
		if w != nil {
			_ = w.Close()
		}

		//mandatory for SMTP protocol
		s._close()

		s.mut.Unlock()
	}()

	if e = s._ValidateLine(from); e != nil {
		return e
	}

	for _, recp := range to {
		if e = s._ValidateLine(recp); e != nil {
			return e
		}
	}

	if c, e = s._client(ctx); e != nil {
		return ErrorSMTPClientInit.Error(e)
	}

	if e = c.Noop(); e != nil {
		return ErrorSMTPClientNoop.Error(e)
	}

	if e = c.Mail(from); e != nil {
		return ErrorSMTPClientMail.Error(e)
	}

	for _, addr := range to {
		if e = c.Rcpt(addr); e != nil {
			return ErrorSMTPClientRcpt.Error(e)
		}
	}

	if w, e = c.Data(); e != nil {
		return ErrorSMTPClientData.Error(e)
	}

	if _, e = data.WriteTo(w); e != nil {
		return ErrorSMTPClientWrite.Error(e)
	}

	return nil
}
