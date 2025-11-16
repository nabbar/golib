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

package render

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the render package.
// These codes are used with the github.com/nabbar/golib/errors package
// to provide structured error handling.
//
// All error codes start from liberr.MinPkgMailer to avoid conflicts
// with other packages in the golib ecosystem.
const (
	// ErrorParamEmpty indicates that required parameters are missing or empty.
	// This error is returned when mandatory configuration or data is not provided.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgMailer

	// ErrorMailerConfigInvalid indicates that the mailer configuration is invalid.
	// This error is returned by Config.Validate() when validation fails.
	ErrorMailerConfigInvalid

	// ErrorMailerHtml indicates a failure during HTML email generation.
	// This typically occurs when the Hermes library encounters an error
	// while processing the email template.
	ErrorMailerHtml

	// ErrorMailerText indicates a failure during plain text email generation.
	// This typically occurs when the Hermes library encounters an error
	// while converting the email to plain text format.
	ErrorMailerText
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/mailer"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the human-readable error message for a given error code.
// This function is registered with the error system during package initialization
// and is called automatically when formatting errors.
//
// Parameters:
//   - code: The error code to get the message for
//
// Returns:
//   - The human-readable error message, or liberr.NullMessage if the code is unknown
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorMailerConfigInvalid:
		return "config of mailer is invalid"
	case ErrorMailerHtml:
		return "cannot generate html content"
	case ErrorMailerText:
		return "cannot generate plain text content"
	}

	return liberr.NullMessage
}
