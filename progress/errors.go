/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package progress

import liberr "github.com/nabbar/golib/errors"

const (
	ErrorParamsEmpty liberr.CodeError = iota + liberr.MinPkgNutsDB
	ErrorParamsMissing
	ErrorParamsMismatching
	ErrorBarNotInitialized
)

var isCodeError = false

func IsCodeError() bool {
	return isCodeError
}

func init() {
	isCodeError = liberr.ExistInMapMessage(ErrorParamsEmpty)
	liberr.RegisterIdFctMessage(ErrorParamsEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case liberr.UNK_ERROR:
		return ""
	case ErrorParamsEmpty:
		return "at least on given parameters is empty"
	case ErrorParamsMissing:
		return "at least on given parameters is missing"
	case ErrorParamsMismatching:
		return "at least on given parameters is mismatching awaiting type"
	case ErrorBarNotInitialized:
		return "progress bar not initialized"
	}

	return ""
}
