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

package njs_smtp

import errors "github.com/nabbar/golib/njs-errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_SMTP
	FILE_STAT
	FILE_READ
	CONFIG_INVALID_DSN
	CONFIG_INVALID_NETWORK
	CONFIG_INVALID_PARAMS
	CONFIG_INVALID_HOST
	SMTP_DIAL
	SMTP_SEND
	SMTP_CLIENT_INIT
	SMTP_CLIENT_STARTTLS
	SMTP_CLIENT_AUTH
	SMTP_CLIENT_NOOP
	SMTP_CLIENT_MAIL
	SMTP_CLIENT_RCPT
	SMTP_CLIENT_DATA
	SMTP_CLIENT_EMPTY
	SMTP_CLIENT_SEND_RECOVERED
	SMTP_CLIENT_FROM_EMPTY
	SMTP_CLIENT_TO_EMPTY
	SMTP_CLIENT_SUBJECT_EMPTY
	SMTP_CLIENT_MAILER_EMPTY
	RAND_READER
	BUFFER_EMPTY
	BUFFER_WRITE_STRING
	BUFFER_WRITE_BYTES
	IO_WRITER_MISSING
	IO_WRITER_ERROR
	EMPTY_HTML
	EMPTY_CONTENTS
	TEMPLATE_PARSING
	TEMPLATE_EXECUTE
	TEMPLATE_CLONE
	TEMPLATE_HTML2TEXT
)

func init() {
	errors.RegisterFctMessage(getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case FILE_STAT:
		return "error occurs on getting stat of file"
	case FILE_READ:
		return "error occurs on reading file"
	case CONFIG_INVALID_DSN:
		return "invalid DSN: did you forget to escape a param value"
	case CONFIG_INVALID_NETWORK:
		return "invalid DSN: network address not terminated (missing closing brace)"
	case CONFIG_INVALID_PARAMS:
		return "invalid DSN: parsing uri parameters occurs an error"
	case CONFIG_INVALID_HOST:
		return "invalid DSN: missing the slash ending the host"
	case SMTP_DIAL:
		return "error while trying to dial with SMTP server"
	case SMTP_SEND:
		return "error while sending mail to SMTP server"
	case SMTP_CLIENT_INIT:
		return "error while trying to initialize new client for dial connection to SMTP Server"
	case SMTP_CLIENT_STARTTLS:
		return "error while trying to starttls on SMTP server"
	case SMTP_CLIENT_AUTH:
		return "error while trying to authenticate to SMTP server"
	case SMTP_CLIENT_NOOP:
		return "error on sending noop command to check connection with SMTP server"
	case SMTP_CLIENT_MAIL:
		return "error on sending mail command to initialize new mail transaction with SMTP server"
	case SMTP_CLIENT_RCPT:
		return "error on sending rcpt command to specify add recipient email for the new mail"
	case SMTP_CLIENT_DATA:
		return "error on opening io writer to send data on client"
	case SMTP_CLIENT_EMPTY:
		return "cannot send email without any attachment and contents"
	case SMTP_CLIENT_SEND_RECOVERED:
		return "recovered error while client sending mail"
	case SMTP_CLIENT_FROM_EMPTY:
		return "sender From address cannot be empty"
	case SMTP_CLIENT_TO_EMPTY:
		return "list of recipient To address cannot be empty"
	case SMTP_CLIENT_SUBJECT_EMPTY:
		return "subject of the new mail cannot be empty"
	case SMTP_CLIENT_MAILER_EMPTY:
		return "mailer of the new mail cannot be empty"
	case RAND_READER:
		return "error on reading on random reader io"
	case BUFFER_EMPTY:
		return "buffer is empty"
	case BUFFER_WRITE_STRING:
		return "error on write string into buffer"
	case BUFFER_WRITE_BYTES:
		return "error on write bytes into buffer"
	case IO_WRITER_MISSING:
		return "io writer is not defined"
	case IO_WRITER_ERROR:
		return "error occur on write on io writer"
	case EMPTY_HTML:
		return "text/html content is empty"
	case EMPTY_CONTENTS:
		return "mail content is empty"
	case TEMPLATE_PARSING:
		return "error occur on parsing template"
	case TEMPLATE_EXECUTE:
		return "error occur on execute template"
	case TEMPLATE_CLONE:
		return "error occur while cloning template"
	case TEMPLATE_HTML2TEXT:
		return "error occur on reading io reader html and convert it to text"
	}

	return ""
}
