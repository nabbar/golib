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

package smtp

import "strings"

type TLSMode uint8

const (
	TLS_NONE TLSMode = iota
	TLS_STARTTLS
	TLS_TLS
)

func parseTLSMode(str string) TLSMode {
	switch strings.ToLower(str) {
	case TLS_TLS.string():
		return TLS_TLS
	case TLS_STARTTLS.string():
		return TLS_STARTTLS
	}

	return TLS_NONE
}

func (tlm TLSMode) string() string {
	switch tlm {
	case TLS_TLS:
		return "tls"
	case TLS_STARTTLS:
		return "starttls"
	case TLS_NONE:
		return "none"
	}

	return TLS_NONE.string()
}
