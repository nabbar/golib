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

package httpserver

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgHttpServer
	ErrorInvalidInstance
	ErrorHTTP2Configure
	ErrorServerValidate
	ErrorServerStart
	ErrorInvalidAddress
	ErrorPortUse
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/httpserver"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorInvalidInstance:
		return "invalid instance"
	case ErrorHTTP2Configure:
		return "cannot initialize http2 over http srv"
	case ErrorServerValidate:
		return "config srv seems to be not valid"
	case ErrorServerStart:
		return "server killed : server start but not listen"
	case ErrorInvalidAddress:
		return "address seems to be invalid"
	case ErrorPortUse:
		return "srv port is still used"
	}

	return liberr.NullMessage
}
