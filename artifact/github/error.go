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

package github

import (
	"errors"
	"fmt"

	libart "github.com/nabbar/golib/artifact"
	liberr "github.com/nabbar/golib/errors"
)

const pkgName = "golib/artifact/github"

const (
	ErrorParamEmpty liberr.CodeError = iota + libart.MinArtifactGithub
	ErrorURLParse
	ErrorClientInit
	ErrorGithubList
	ErrorGithubGetRelease
	ErrorGithubNotFound
	ErrorGithubRequestNew
	ErrorGithubRequestRun
	ErrorGithubResponse
	ErrorGithubIOCopy
	ErrorDestinationStat
	ErrorDestinationSize
)

var (
	errResponseCode      = errors.New("response status code %s")
	errResponseContents  = errors.New("response contents is empty")
	errResponseBodyEmpty = errors.New("empty body response")
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
		return "github endpoint seems to be not valid"
	case ErrorClientInit:
		return "initialization of github client failed"
	case ErrorGithubList:
		return "request the release list to github occurs an error"
	case ErrorGithubGetRelease:
		return "request the release data to github occurs an error"
	case ErrorGithubNotFound:
		return "the requested constrains to the release are not matching"
	case ErrorGithubRequestNew:
		return "error on creating new github http request"
	case ErrorGithubRequestRun:
		return "error on running new github http request"
	case ErrorGithubResponse:
		return "response error on http request"
	case ErrorGithubIOCopy:
		return "io copy from github response occurs an error"
	case ErrorDestinationStat:
		return "error occurs on calling destination stats"
	case ErrorDestinationSize:
		return "mismatching size between downloaded contents and github http response header"
	}

	return liberr.NullMessage
}
