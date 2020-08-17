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

package static

import "github.com/nabbar/golib/errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_Static
	EMPTY_PACKED
	INDEX_NOT_FOUND
	INDEX_REQUESTED_NOT_SET
	FILE_NOT_FOUND
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

	case EMPTY_PACKED:
		return "packed file is empty"

	case INDEX_NOT_FOUND:
		return "mode index is defined but index.(html|htm) is not found"

	case INDEX_REQUESTED_NOT_SET:
		return "request call index but mode index is false"

	case FILE_NOT_FOUND:
		return "requested packed file is not found"

	case errors.UNK_ERROR:
		return ""
	}

	return ""
}
