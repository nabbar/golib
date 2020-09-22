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

package s3aws

import (
	"fmt"

	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/errors"
)

const (
	ErrorParamsEmpty errors.CodeError = iota + artifact.MIN_ARTIFACT_S3AWS
	ErrorClientInit
	ErrorS3AWSRegex
	ErrorS3AWSFind
	ErrorS3AWSNewVers
	ErrorS3AWSNotFound
	ErrorS3AWSDownloadError
)

var (
	isCodeError    = false
	errRegexGroup  = "regex '%s' has only '%d' group extracted and given group to use '%d'"
	errVersion     = "error with version '%s'"
	errVersRequest = "version requested '%s'"
	errObject      = "object '%s'"
	errObjectEmpty = "object '%s' is empty"
	errObjectSize  = "object size '%s' mismatch with downloaded size"
)

func getError(code string, args ...interface{}) error {
	//nolint #goerr113
	return fmt.Errorf(code, args...)
}

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
	case ErrorClientInit:
		return "initialization of gitlab client failed"
	case ErrorS3AWSRegex:
		return "regex parsing object has not enough matched group to use given group number"
	case ErrorS3AWSFind:
		return "request the object find to aws s3 backend occurs an error"
	case ErrorS3AWSNewVers:
		return "init a new version based on group value extracted occurs an error"
	case ErrorS3AWSNotFound:
		return "requested version is not found"
	case ErrorS3AWSDownloadError:
		return "download object requested occurs error"
	}

	return ""
}
