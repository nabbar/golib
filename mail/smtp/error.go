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
 *
 */

package smtp

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for SMTP operations.
// These codes are registered with github.com/nabbar/golib/errors and can be
// used for error handling, logging, and monitoring.
//
// All error codes start from liberr.MinPkgSMTP to avoid conflicts with other packages.
const (
	// ErrorParamEmpty indicates that a required parameter was nil or empty.
	// This typically occurs when calling New() with a nil configuration.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgSMTP

	// ErrorSMTPDial indicates a failure to establish a TCP connection to the SMTP server.
	// Possible causes: network issues, firewall, incorrect host/port, DNS resolution failure.
	ErrorSMTPDial

	// ErrorSMTPClientInit indicates a failure to initialize the SMTP client after connection.
	// This occurs after successful TCP connection but before SMTP handshake completes.
	ErrorSMTPClientInit

	// ErrorSMTPClientStartTLS indicates a failure during STARTTLS negotiation.
	// Possible causes: TLS handshake failure, certificate issues, unsupported TLS version.
	ErrorSMTPClientStartTLS

	// ErrorSMTPClientAuth indicates authentication failure.
	// Possible causes: incorrect credentials, authentication method not supported,
	// server requires TLS before authentication.
	ErrorSMTPClientAuth

	// ErrorSMTPClientNoop indicates the NOOP command failed.
	// This suggests the connection is broken or the server is not responding properly.
	ErrorSMTPClientNoop

	// ErrorSMTPClientMail indicates the MAIL FROM command failed.
	// Possible causes: sender address rejected, server policy violation.
	ErrorSMTPClientMail

	// ErrorSMTPClientRcpt indicates the RCPT TO command failed for one or more recipients.
	// Possible causes: recipient address rejected, mailbox doesn't exist, relay denied.
	ErrorSMTPClientRcpt

	// ErrorSMTPClientData indicates the DATA command failed.
	// This occurs before transmitting the email body.
	ErrorSMTPClientData

	// ErrorSMTPClientWrite indicates a failure while writing email content.
	// Possible causes: connection lost, timeout, io.WriterTo implementation error.
	ErrorSMTPClientWrite

	// ErrorSMTPLineCRLF indicates that an email address contained CR or LF characters.
	// This is a security measure to prevent SMTP command injection attacks.
	// The from and to addresses must not contain newlines or carriage returns.
	ErrorSMTPLineCRLF
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/smtp"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the human-readable error message for a given error code.
// This function is registered with the error system and is called automatically
// when formatting errors.
//
// Parameters:
//   - code: The error code to get the message for
//
// Returns:
//   - message: Human-readable error description; liberr.NullMessage if code is unknown
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorSMTPDial:
		return "error while trying to dial with SMTP server"
	case ErrorSMTPClientInit:
		return "error while trying to initialize new client for dial connection to SMTP Server"
	case ErrorSMTPClientStartTLS:
		return "error while trying to starttls on SMTP server"
	case ErrorSMTPClientAuth:
		return "error while trying to authenticate to SMTP server"
	case ErrorSMTPClientNoop:
		return "error on sending noop command to check connection with SMTP server"
	case ErrorSMTPClientMail:
		return "cannot initialize new mail command with server"
	case ErrorSMTPClientRcpt:
		return "cannot add recipient to current mail"
	case ErrorSMTPClientData:
		return "cannot call data command to send contents of mail"
	case ErrorSMTPClientWrite:
		return "cannot write data to send contents of mail"
	case ErrorSMTPLineCRLF:
		return "smtp: A line must not contain CR or LF"
	}

	return liberr.NullMessage
}
