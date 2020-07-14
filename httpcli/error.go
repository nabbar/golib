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

package httpcli

import errors "github.com/nabbar/golib/errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_Httpcli
	URL_PARSE
	HTTP_CLIENT
	HTTP_REQUEST
	HTTP_DO
	IO_READ
	BUFFER_WRITE
	HTTP2_CONFIGURE
)

func init() {
	errors.RegisterFctMessage(getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case URL_PARSE:
		return "uri/url parse error"
	case HTTP_CLIENT:
		return "error on creating a new http/http2 client"
	case HTTP_REQUEST:
		return "error on creating a new http/http2 request"
	case HTTP_DO:
		return "error on sending a http/http2 request"
	case IO_READ:
		return "error on reading i/o stream"
	case BUFFER_WRITE:
		return "error on writing bytes on buffer"
	case HTTP2_CONFIGURE:
		return "error while configure http2 transport for client"
	}

	return ""
}
