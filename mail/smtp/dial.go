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
	"context"
	"crypto/tls"
	"net"
	"net/smtp"

	liberr "github.com/nabbar/golib/errors"
	smtptl "github.com/nabbar/golib/mail/smtp/tlsmode"
)

// dialTLS establishes a TLS-wrapped connection to the SMTP server.
// This is used for "strict TLS" mode where TLS is negotiated from the start.
//
// The method first establishes a plain connection, then wraps it with TLS.
// If any error occurs, the connection is automatically closed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - addr: Server address in "host:port" format
//   - tlsConfig: TLS configuration to use for the connection
//
// Returns:
//   - con: Established TLS connection
//   - err: ErrorSMTPClientStartTLS if TLS wrapping fails; ErrorSMTPDial if connection fails
func (s *smtpClient) dialTLS(ctx context.Context, addr string, tlsConfig *tls.Config) (con net.Conn, err liberr.Error) {
	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	if con, err = s.dial(ctx, addr); err != nil {
		return
	}

	if con = tls.Client(con, tlsConfig); con == nil {
		err = ErrorSMTPClientStartTLS.Error(nil)
	}

	return
}

// dial establishes a plain TCP connection to the SMTP server.
// This is the base connection method used by both plain and TLS connections.
//
// The method uses net.Dialer with the configured network type (tcp4, tcp6, unix, etc.)
// from the SMTP configuration. If an error occurs, the connection is automatically closed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - addr: Server address in "host:port" format
//
// Returns:
//   - con: Established TCP connection
//   - err: ErrorSMTPDial if the connection fails
func (s *smtpClient) dial(ctx context.Context, addr string) (con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	d := net.Dialer{}

	if con, e = d.DialContext(ctx, s.cfg.GetNet().String(), addr); e != nil {
		return con, ErrorSMTPDial.Error(e)
	}

	return
}

// client creates an SMTP client with the appropriate TLS mode.
// It handles three TLS modes:
//  1. TLSStrictTLS: Establishes TLS connection before SMTP handshake
//  2. TLSStartTLS: Uses STARTTLS command after initial connection
//  3. TLSNone: Plain connection, but opportunistically upgrades to STARTTLS if available
//
// The method automatically handles STARTTLS negotiation and can upgrade the
// TLS mode if the server advertises STARTTLS support.
//
// If any error occurs, both client and connection are automatically closed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - addr: Server address in "host:port" format
//   - tlsConfig: TLS configuration; required for STARTTLS and Strict TLS modes
//
// Returns:
//   - cli: Initialized SMTP client
//   - con: Underlying network connection
//   - err: ErrorParamEmpty if tlsConfig is nil for TLS modes;
//     ErrorSMTPClientInit if client creation fails;
//     ErrorSMTPClientStartTLS if STARTTLS fails
func (s *smtpClient) client(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil {
			if cli != nil {
				_ = cli.Close()
			}
			if con != nil {
				_ = con.Close()
			}
		}
	}()

	if s.cfg.GetTlsMode() == smtptl.TLSStartTLS && tlsConfig == nil {
		err = ErrorParamEmpty.Error(nil)
		return
	} else if s.cfg.GetTlsMode() == smtptl.TLSStrictTLS && tlsConfig == nil {
		err = ErrorParamEmpty.Error(nil)
		return
	}

	if s.cfg.GetTlsMode() == smtptl.TLSStrictTLS && tlsConfig != nil {
		if con, err = s.dialTLS(ctx, addr, tlsConfig); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.Error(e)
			return
		}
	} else {
		if con, err = s.dial(ctx, addr); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.Error(e)
			return
		}

		try := s.checkExtension(cli, "STARTTLS")

		if s.cfg.GetTlsMode() == smtptl.TLSStartTLS || try {
			if e = cli.StartTLS(tlsConfig); e != nil && !try {
				err = ErrorSMTPClientStartTLS.Error(e)
				return
			} else if e == nil && try {
				s.cfg.SetTlsMode(smtptl.TLSStartTLS)
			}
		}
	}

	return
}

// tryClient attempts to create an SMTP client with automatic fallback.
// If strict TLS fails, it automatically retries with STARTTLS mode.
//
// This provides resilience when connecting to servers that don't support
// strict TLS but do support STARTTLS. The fallback only occurs for
// TLSStrictTLS mode; other modes don't retry.
//
// Fallback strategy:
//   - TLSStrictTLS fails → Retry with TLSStartTLS
//   - TLSStartTLS fails → No retry
//   - TLSNone fails → No retry
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - addr: Server address in "host:port" format
//   - tlsConfig: TLS configuration
//
// Returns:
//   - cli: Initialized SMTP client
//   - con: Underlying network connection
//   - err: Any error from the final connection attempt
func (s *smtpClient) tryClient(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	cli, con, err = s.client(ctx, addr, tlsConfig)

	if err == nil {
		return
	}

	switch s.cfg.GetTlsMode() {
	case smtptl.TLSStrictTLS:
		s.cfg.SetTlsMode(smtptl.TLSStartTLS)
		return s.tryClient(ctx, addr, tlsConfig)
	case smtptl.TLSStartTLS, smtptl.TLSNone:
		return
	}

	return
}

// auth performs SMTP authentication if credentials are configured.
// Currently supports PLAIN authentication mechanism.
//
// If both username and password are empty, authentication is skipped.
// This allows the client to work with servers that don't require authentication.
//
// Note: CRAM-MD5 authentication is commented out but could be enabled if needed.
// PLAIN authentication should be used over TLS to protect credentials.
//
// Parameters:
//   - cli: Connected SMTP client
//   - addr: Server address used for PLAIN auth identity
//
// Returns:
//   - error: ErrorSMTPClientAuth if authentication fails; nil if successful or not needed
func (s *smtpClient) auth(cli *smtp.Client, addr string) liberr.Error {
	usr := s.cfg.GetUser()
	pwd := s.cfg.GetPass()
	err := ErrorSMTPClientAuth.Error(nil)

	if usr == "" && pwd == "" {
		return nil
	}
	/*
		if e := cli.Auth(smtp.CRAMMD5Auth(usr, pwd)); e != nil {
			err.Add(e)
		} else {
			return nil
		}
	*/
	if e := cli.Auth(smtp.PlainAuth("", usr, pwd, addr)); e != nil {
		err.Add(e)
	} else {
		return nil
	}

	return err
}

// checkExtension checks if the SMTP server supports a specific extension.
// This is used to detect features like STARTTLS, AUTH, SIZE, etc.
//
// The method queries the server's EHLO response for the extension.
//
// Parameters:
//   - cli: Connected SMTP client (after EHLO)
//   - ext: Extension name to check (e.g., "STARTTLS", "AUTH")
//
// Returns:
//   - bool: true if the extension is supported; false otherwise
func (s *smtpClient) checkExtension(cli *smtp.Client, ext string) bool {
	ok, _ := cli.Extension(ext)
	return ok
}
