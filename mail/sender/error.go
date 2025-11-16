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

package sender

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const pkgName = "golib/mail"

// Error codes specific to the mail/sender package.
// These codes are used with the github.com/nabbar/golib/errors package
// to provide structured, traceable error handling with parent error wrapping.
//
// All error codes in this package start from liberr.MinPkgMail to avoid
// conflicts with other packages in the golib ecosystem.
//
// Example usage:
//
//	if err := config.Validate(); err != nil {
//	    if libErr, ok := err.(liberr.Error); ok {
//	        if libErr.Code() == sender.ErrorMailConfigInvalid {
//	            // Handle configuration error specifically
//	        }
//	    }
//	}
//
// See github.com/nabbar/golib/errors for more information on error handling.
const (
	// ErrorParamEmpty indicates that required parameters were not provided.
	// This typically occurs when nil or empty values are passed to functions
	// that require valid input.
	//
	// Common causes:
	//   - Passing nil SMTP client to Sender.Send()
	//   - Empty configuration values
	//   - Missing required fields
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgMail

	// ErrorMailConfigInvalid indicates that the email configuration validation failed.
	// This error is returned when Config.Validate() detects invalid settings.
	//
	// Common causes:
	//   - Missing required fields (Charset, Subject, Encoding, Priority, From)
	//   - Invalid email address format
	//   - Invalid encoding or priority strings
	//
	// Check the parent error for specific validation details.
	ErrorMailConfigInvalid

	// ErrorMailIORead indicates a failure reading data from an io.Reader.
	// This typically occurs when reading email body content or attachment data.
	//
	// Common causes:
	//   - File read errors for attachments
	//   - Network issues when reading from streams
	//   - Closed or invalid readers
	ErrorMailIORead

	// ErrorMailIOWrite indicates a failure writing data to an io.Writer.
	// This typically occurs during email message generation.
	//
	// Common causes:
	//   - Disk space issues
	//   - Permission errors
	//   - Network failures during transmission
	ErrorMailIOWrite

	// ErrorMailDateParsing indicates that a date string could not be parsed.
	// This error occurs in Mail.SetDateString() when the provided date string
	// doesn't match the specified layout format.
	//
	// Common causes:
	//   - Incorrect date format string
	//   - Mismatched layout and date string
	//   - Invalid date values
	//
	// Use time.RFC822, time.RFC1123Z, or other standard formats.
	ErrorMailDateParsing

	// ErrorMailSmtpClient indicates an error communicating with the SMTP server.
	// This error occurs during SMTP connection checks or send operations.
	//
	// Common causes:
	//   - SMTP server unavailable or unreachable
	//   - Network connectivity issues
	//   - Authentication failures
	//   - TLS/SSL handshake errors
	//
	// See github.com/nabbar/golib/mail/smtp for SMTP client configuration.
	ErrorMailSmtpClient

	// ErrorMailSenderInit indicates that the SMTP email sender could not be initialized.
	// This error occurs in Mail.Sender() when creating the email message structure.
	//
	// Common causes:
	//   - Invalid email addresses (malformed format)
	//   - Missing required fields (from, recipients)
	//   - Address format validation failures
	//
	// The parent error contains specific details about what validation failed.
	ErrorMailSenderInit

	// ErrorFileOpenCreate indicates that a file could not be opened or created.
	// This typically occurs when adding attachments from file paths.
	//
	// Common causes:
	//   - File does not exist
	//   - Insufficient permissions
	//   - Invalid file path
	//   - File is locked by another process
	//
	// Used in Config.NewMailer() when processing attachment files.
	ErrorFileOpenCreate
)

// init registers all error codes with the liberr error system.
// This ensures that error codes are unique across the entire golib ecosystem
// and provides error message lookup functionality.
//
// Panics if there is an error code collision with another package.
func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package %s", pkgName))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the error message for a given error code.
// This function is used by the liberr package to provide human-readable
// error messages while maintaining structured error codes.
//
// Parameters:
//   - code: The error code to get the message for
//
// Returns:
//   - The error message string, or liberr.NullMessage if the code is unknown
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorMailConfigInvalid:
		return "config is invalid"
	case ErrorMailIORead:
		return "cannot read bytes from io source"
	case ErrorMailIOWrite:
		return "cannot write given string to IO resource"
	case ErrorMailDateParsing:
		return "error occurs while trying to parse a date string"
	case ErrorMailSmtpClient:
		return "error occurs while to checking connection with SMTP server"
	case ErrorMailSenderInit:
		return "error occurs while to preparing SMTP Email sender"
	case ErrorFileOpenCreate:
		return "cannot open/create file"
	}

	return liberr.NullMessage
}
