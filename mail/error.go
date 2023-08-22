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

package mail

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const pkgName = "golib/mail"

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgMail
	ErrorMailConfigInvalid
	ErrorMailIORead
	ErrorMailIOWrite
	ErrorMailDateParsing
	ErrorMailSmtpClient
	ErrorMailSenderInit
	ErrorFileOpenCreate
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package %s", pkgName))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

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
