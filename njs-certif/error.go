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

package njs_certif

import errors "github.com/nabbar/golib/njs-errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_Certif
	FILE_STAT_ERROR
	FILE_READ_ERROR
	CERT_APPEND_KO
	CERT_LOAD_KEYPAIR
	CERT_PARSE_KEYPAIR
)

func init() {
	errors.RegisterFctMessage(getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case FILE_STAT_ERROR:
		return "cannot get file stat"
	case FILE_READ_ERROR:
		return "cannot read file"
	case CERT_APPEND_KO:
		return "cannot append PEM file"
	case CERT_LOAD_KEYPAIR:
		return "cannot X509 parsing certificate string"
	case CERT_PARSE_KEYPAIR:
		return "cannot x509 loading certificate file"
	}

	return ""
}
