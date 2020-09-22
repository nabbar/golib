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

import "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty errors.CodeError = iota + errors.MIN_PKG_SMTP
	ErrorFileStat
	ErrorFileRead
	ErrorConfigInvalidDSN
	ErrorConfigInvalidNetwork
	ErrorConfigInvalidParams
	ErrorConfigInvalidHost
	ErrorSMTPDial
	ErrorSMTPSend
	ErrorSMTPClientInit
	ErrorSMTPClientStartTLS
	ErrorSMTPClientAuth
	ErrorSMTPClientNoop
	ErrorSMTPClientMail
	ErrorSMTPClientRcpt
	ErrorSMTPClientData
	ErrorSMTPClientEmpty
	ErrorSMTPClientSendRecovered
	ErrorSMTPClientEmptyFrom
	ErrorSMTPClientEmptyTo
	ErrorSMTPClientEmptySubject
	ErrorSMTPClientEmptyMailer
	ErrorRandReader
	ErrorBufferEmpty
	ErrorBufferWriteString
	ErrorBufferWriteBytes
	ErrorIOWriter
	ErrorIOWriterMissing
	ErrorEmptyHtml
	ErrorEmptyContents
	ErrorTemplateParsing
	ErrorTemplateExecute
	ErrorTemplateClone
	ErrorTemplateHtml2Text
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = errors.ExistInMapMessage(ErrorParamsEmpty)
	errors.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case errors.UNK_ERROR:
		return ""
	case ErrorParamsEmpty:
		return "given parameters is empty"
	case ErrorFileStat:
		return "error occurs on getting stat of file"
	case ErrorFileRead:
		return "error occurs on reading file"
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
	case ErrorSMTPSend:
		return "error while sending mail to SMTP server"
	case ErrorSMTPClientInit:
		return "error while trying to initialize new client for dial connection to SMTP Server"
	case ErrorSMTPClientStartTLS:
		return "error while trying to starttls on SMTP server"
	case ErrorSMTPClientAuth:
		return "error while trying to authenticate to SMTP server"
	case ErrorSMTPClientNoop:
		return "error on sending noop command to check connection with SMTP server"
	case ErrorSMTPClientMail:
		return "error on sending mail command to initialize new mail transaction with SMTP server"
	case ErrorSMTPClientRcpt:
		return "error on sending rcpt command to specify add recipient email for the new mail"
	case ErrorSMTPClientData:
		return "error on opening io writer to send data on client"
	case ErrorSMTPClientEmpty:
		return "cannot send email without any attachment and contents"
	case ErrorSMTPClientSendRecovered:
		return "recovered error while client sending mail"
	case ErrorSMTPClientEmptyFrom:
		return "sender From address cannot be empty"
	case ErrorSMTPClientEmptyTo:
		return "list of recipient To address cannot be empty"
	case ErrorSMTPClientEmptySubject:
		return "subject of the new mail cannot be empty"
	case ErrorSMTPClientEmptyMailer:
		return "mailer of the new mail cannot be empty"
	case ErrorRandReader:
		return "error on reading on random reader io"
	case ErrorBufferEmpty:
		return "buffer is empty"
	case ErrorBufferWriteString:
		return "error on write string into buffer"
	case ErrorBufferWriteBytes:
		return "error on write bytes into buffer"
	case ErrorIOWriterMissing:
		return "io writer is not defined"
	case ErrorIOWriter:
		return "error occur on write on io writer"
	case ErrorEmptyHtml:
		return "text/html content is empty"
	case ErrorEmptyContents:
		return "mail content is empty"
	case ErrorTemplateParsing:
		return "error occur on parsing template"
	case ErrorTemplateExecute:
		return "error occur on execute template"
	case ErrorTemplateClone:
		return "error occur while cloning template"
	case ErrorTemplateHtml2Text:
		return "error occur on reading io reader html and convert it to text"
	}

	return ""
}
