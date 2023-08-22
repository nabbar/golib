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

package jfrog

import (
	"errors"
	"fmt"

	libart "github.com/nabbar/golib/artifact"
	liberr "github.com/nabbar/golib/errors"
)

const pkgName = "golib/artifact/jfrog"

const (
	ErrorParamEmpty liberr.CodeError = iota + libart.MinArtifactJfrog
	ErrorURLParse
	ErrorRequestInit
	ErrorRequestDo
	ErrorRequestResponse
	ErrorRequestResponseBodyEmpty
	ErrorRequestResponseBodyDecode
	ErrorArtifactoryNotFound
	ErrorArtifactoryDownload
	ErrorDestinationSize
)

var (
	errMisMatchingSize = errors.New("destination size and contentLength header are not matching")
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package %s", pkgName))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case liberr.UnknownError:
		return liberr.NullMessage
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorURLParse:
		return "endpoint of JFrog Artifactory seems to be not valid"
	case ErrorRequestInit:
		return "cannot create new request to JFrog Artifactory"
	case ErrorRequestDo:
		return "cannot send request to JFrog Artifactory"
	case ErrorRequestResponse:
		return "error on sending request to JFrog Artifactory"
	case ErrorRequestResponseBodyEmpty:
		return "empty response while requesting JFrog Artifactory"
	case ErrorRequestResponseBodyDecode:
		return "cannot decode response from JFrog Artifactory"
	case ErrorArtifactoryNotFound:
		return "the requested constrains to the release are not matching"
	case ErrorArtifactoryDownload:
		return "error on downloading artifact"
	}

	return liberr.NullMessage
}
