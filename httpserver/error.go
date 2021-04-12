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

import "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty errors.CodeError = iota + errors.MinPkgHttpServer
	ErrorHTTP2Configure
	ErrorPoolAdd
	ErrorPoolValidate
	ErrorPoolListen
	ErrorServerValidate
	ErrorPortUse
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
	case ErrorHTTP2Configure:
		return "cannot initialize http2 over http server"
	case ErrorPoolAdd:
		return "cannot add server on pool"
	case ErrorPoolValidate:
		return "at least one config server seems to be not valid"
	case ErrorPoolListen:
		return "at least one server has listen error"
	case ErrorServerValidate:
		return "config server seems to be not valid"
	case ErrorPortUse:
		return "server port is still used"
	}

	return ""
}
