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
	"io"
	"os"
	"regexp"
	"strings"

	hscvrs "github.com/hashicorp/go-version"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
)

const subUp = 20

const (
	MinArtifactGitlab = subUp + liberr.MinPkgArtifact
	MinArtifactGithub = subUp + MinArtifactGitlab
	MinArtifactJfrog  = subUp + MinArtifactGithub
	MinArtifactS3AWS  = subUp + MinArtifactJfrog
)

type Client interface {
	artcli.ArtifactManagement

	ListReleases() (releases hscvrs.Collection, err error)
	GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err error)
	Download(containName string, regexName string, release *hscvrs.Version) (int64, io.ReadCloser, error)
}

func CheckRegex(name, regex string) bool {
	if ok, _ := regexp.MatchString(regex, name); ok {
		return ok
	}

	return false
}

func DownloadRelease(link string) (file os.File, err error) {
	panic("not implemented")
}

func ValidatePreRelease(version *hscvrs.Version) bool {
	var (
		p = strings.ToLower(version.Prerelease())
		s = []string{
			"a", "alpha",
			"b", "beta",
			"rc",
			"dev",
			"test",
			"draft",
			"master",
			"main",
		}
	)

	for _, i := range s {
		if strings.Contains(p, i) {
			return false
		}
	}

	return true
}
