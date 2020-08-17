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

package router

import errors "github.com/nabbar/golib/errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_Router
	HEADER_AUTH_MISSING
	HEADER_AUTH_EMPTY
	HEADER_AUTH_REQUIRE
	HEADER_AUTH_FORBIDDEN
	HEADER_AUTH_ERROR
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = errors.ExistInMapMessage(EMPTY_PARAMS)
	errors.RegisterIdFctMessage(EMPTY_PARAMS, getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"

	case HEADER_AUTH_MISSING:
		return "missing authorization header"

	case HEADER_AUTH_EMPTY:
		return "authorization header is empty"

	case HEADER_AUTH_REQUIRE:
		return "authorization check failed, authorization still require"

	case HEADER_AUTH_FORBIDDEN:
		return "authorization check success but unauthorized client"

	case HEADER_AUTH_ERROR:
		return "authorization check return an invalid response code"

	case errors.UNK_ERROR:
		return ""
	}

	return ""
}
