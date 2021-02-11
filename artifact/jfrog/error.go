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
	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/errors"
)

const (
	ErrorParamsEmpty errors.CodeError = iota + artifact.MIN_ARTIFACT_JFORG
	ErrorURLParse
	ErrorClientInit
	ErrorArtifactoryList
	ErrorArtifactoryGetRelease
	ErrorArtifactoryNotFound
	ErrorArtifactoryPatternRegex
	ErrorArtifactoryVersion
	ErrorArtifactoryReleaseNotFound
	ErrorArtifactoryReleaseRequest
	ErrorArtifactoryDownload
)

var isCodeError = false

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
	case ErrorURLParse:
		return "endpoint of JFrog Artifactory seems to be not valid"
	case ErrorClientInit:
		return "initialization of JFrog Artifactory client failed"
	case ErrorArtifactoryList:
		return "request the release list to JFrog Artifactory occurs an error"
	case ErrorArtifactoryGetRelease:
		return "request the release data to JFrog Artifactory occurs an error"
	case ErrorArtifactoryNotFound:
		return "the requested constrains to the release are not matching"
	case ErrorArtifactoryPatternRegex:
		return "regex parsing artifact name has not enough matched group to use given group number"
	case ErrorArtifactoryVersion:
		return "cannot create new version item based on regex group extraction"
	case ErrorArtifactoryReleaseNotFound:
		return "cannot find the requested release"
	case ErrorArtifactoryReleaseRequest:
		return "error on sending request to download artifact"
	case ErrorArtifactoryDownload:
		return "error on downloading artifact"
	}

	return ""
}
