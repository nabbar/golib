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

package artifact

import (
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/nabbar/golib/artifact/client"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
)

const (
	MIN_ARTIFACT_ARTIFAC = errors.MIN_PKG_Artifact + 10
	MIN_ARTIFACT_GITLAB  = errors.MIN_PKG_Artifact + 20
	MIN_ARTIFACT_GITHUB  = errors.MIN_PKG_Artifact + 40
	MIN_ARTIFACT_JFORG   = errors.MIN_PKG_Artifact + 60
	MIN_ARTIFACT_S3AWS   = errors.MIN_PKG_Artifact + 80
)

type Client interface {
	client.ArtifactManagement

	ListReleases() (releases version.Collection, err errors.Error)
	GetArtifact(containName string, regexName string, release *version.Version) (link string, err errors.Error)
	Download(dst ioutils.FileProgress, containName string, regexName string, release *version.Version) errors.Error
}

func CheckRegex(name, regex string) bool {
	if ok, _ := regexp.MatchString(regex, name); ok {
		return ok
	}

	return false
}

func DownloadRelease(link string) (file os.File, err errors.Error) {
	panic("not implemented")
}

func ValidatePreRelease(version *version.Version) bool {
	p := strings.ToLower(version.Prerelease())

	if strings.Contains(p, "alpha") {
		return false
	} else if strings.Contains(p, "beta") {
		return false
	} else if strings.Contains(p, "rc") {
		return false
	} else if strings.Contains(p, "dev") {
		return false
	} else if strings.Contains(p, "test") {
		return false
	} else if strings.Contains(p, "master") {
		return false
	}

	return true
}
