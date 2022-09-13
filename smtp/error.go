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

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgSMTP
	ErrorConfigValidator
	ErrorConfigInvalidDSN
	ErrorConfigInvalidNetwork
	ErrorConfigInvalidParams
	ErrorConfigInvalidHost
	ErrorSMTPDial
	ErrorSMTPClientInit
	ErrorSMTPClientStartTLS
	ErrorSMTPClientAuth
	ErrorSMTPClientNoop
	ErrorSMTPClientMail
	ErrorSMTPClientRcpt
	ErrorSMTPClientData
	ErrorSMTPClientWrite
	ErrorSMTPLineCRLF
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/smtp"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorConfigValidator:
		return "invalid config, validation error"
	case ErrorConfigInvalidDSN:
		return "invalid DSN: did you forget to escape a param value"
	case ErrorConfigInvalidNetwork:
		return "invalid DSN: network address not terminated (missing closing brace)"
	case ErrorConfigInvalidParams:
		return "invalid DSN: parsing uri parameters occurs an error"
	case ErrorConfigInvalidHost:
		return "invalid DSN: missing the slash ending the host"
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
