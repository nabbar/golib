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

package helper

import (
	errors "github.com/nabbar/golib/errors"
)

const (
	// minmal are errors.MIN_AVAILABLE + get a hope free range 1000 + 10 for aws-config errors.
	ErrorResponse errors.CodeError = iota + errors.MinPkgAws + 60
	ErrorConfigEmpty
	ErrorAwsEmpty
	ErrorAws
	ErrorBucketNotFound
	ErrorParamsEmpty
)

var isErrInit = errors.ExistInMapMessage(ErrorResponse)

func init() {
	errors.RegisterIdFctMessage(ErrorResponse, getMessage)
}

func IsErrorInit() bool {
	return isErrInit
}

func getMessage(code errors.CodeError) string {
	switch code {
	case ErrorResponse:
		return "calling aws api occurred a response error"
	case ErrorConfigEmpty:
		return "the given config is empty or invalid"
	case ErrorAws:
		return "the aws request sent to aws API occurred an error"
	case ErrorAwsEmpty:
		return "the aws request sent to aws API occurred an empty result"
	case ErrorBucketNotFound:
		return "the specified bucket is not found"
	case ErrorParamsEmpty:
		return "at least one parameters needed is empty"
	}

	return errors.UNK_MESSAGE
}
