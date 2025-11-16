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

package sender

import "slices"

const (
	headerFrom       = "From"
	headerSender     = "Sender"
	headerReplyTo    = "Reply-To"
	headerReturnPath = "Return-Path"
	headerTo         = "To"
	headerCc         = "Cc"
	headerBcc        = "Bcc"
)

type email struct {
	from       string
	sender     string
	replyTo    string
	returnPath string

	to  []string
	cc  []string
	bcc []string
}

func (e *email) SetFrom(mail string) {
	e.from = mail
}

func (e *email) SetSender(mail string) {
	e.sender = mail
}

func (e *email) SetReplyTo(mail string) {
	e.replyTo = mail
}

func (e *email) SetReturnPath(mail string) {
	e.returnPath = mail
}

func (e *email) GetFrom() string {
	if e.from != "" {
		return e.from
	}

	return ""
}

func (e *email) GetSender() string {
	if e.sender != "" {
		return e.sender
	}

	if e.replyTo != "" {
		return e.replyTo
	}

	if e.returnPath != "" {
		return e.returnPath
	}

	return ""
}

func (e *email) GetReplyTo() string {
	if e.replyTo != "" {
		return e.replyTo
	}

	if e.sender != "" {
		return e.sender
	}

	if e.returnPath != "" {
		return e.returnPath
	}

	return ""
}

func (e *email) GetReturnPath() string {
	if e.returnPath != "" {
		return e.returnPath
	}

	if e.sender != "" {
		return e.sender
	}

	if e.replyTo != "" {
		return e.replyTo
	}

	return ""
}

func (e *email) GetRecipients(rt recipientType) []string {
	switch rt {
	case RecipientTo:
		return e.to
	case RecipientCC:
		return e.cc
	case RecipientBCC:
		return e.bcc
	}

	return make([]string, 0)
}

func (e *email) SetRecipients(rt recipientType, rcpt ...string) {
	switch rt {
	case RecipientTo:
		e.to = make([]string, 0)
	case RecipientCC:
		e.cc = make([]string, 0)
	case RecipientBCC:
		e.bcc = make([]string, 0)
	default:
		return
	}

	e.AddRecipients(rt, rcpt...)
}

func (e *email) AddRecipients(rt recipientType, rcpt ...string) {
	for _, s := range rcpt {
		switch rt {
		case RecipientTo:
			if !slices.Contains(e.to, s) {
				e.to = append(e.to, s)
			}
		case RecipientCC:
			if !slices.Contains(e.cc, s) {
				e.cc = append(e.cc, s)
			}
		case RecipientBCC:
			if !slices.Contains(e.bcc, s) {
				e.bcc = append(e.bcc, s)
			}
		}
	}
}

func (e *email) getHeader(h func(key string, values ...string)) {
	h(headerFrom, e.GetFrom())
	h(headerSender, e.GetSender())
	h(headerReplyTo, e.GetReplyTo())
	h(headerReturnPath, e.GetReturnPath())
	h(headerTo, e.GetRecipients(RecipientTo)...)
	h(headerCc, e.GetRecipients(RecipientCC)...)
	h(headerBcc, e.GetRecipients(RecipientBCC)...)
}
