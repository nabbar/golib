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

package gitlab

import (
	"context"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/go-version"
	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/artifact/client"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
	"github.com/xanzy/go-gitlab"
)

const (
	gitlabPageSize = 100
)

type gitlabModel struct {
	client.ClientHelper

	c *gitlab.Client
	x context.Context
	p int
}

func (g *gitlabModel) ListReleases() (releases version.Collection, err errors.Error) {
	var (
		e    error
		lopt = &gitlab.ListReleasesOptions{
			Page:    0,
			PerPage: gitlabPageSize,
		}
	)

	for {
		var (
			rels []*gitlab.Release
			resp *gitlab.Response
		)

		if rels, resp, e = g.c.Releases.ListReleases(g.p, lopt, gitlab.WithContext(g.x)); e != nil {
			return nil, ErrorGitlabList.ErrorParent(e)
		}

		for _, r := range rels {
			v, _ := version.NewVersion(r.TagName)

			if artifact.ValidatePreRelease(v) {
				releases = append(releases, v)
			}
		}

		if resp.NextPage <= resp.CurrentPage {
			_ = sort.Reverse(releases)
			return
		} else {
			lopt.Page = resp.NextPage
		}
	}
}

func (g *gitlabModel) GetArtifact(containName string, regexName string, release *version.Version) (link string, err errors.Error) {
	var (
		vers *gitlab.Release
		e    error
	)

	if vers, _, e = g.c.Releases.GetRelease(g.p, release.Original(), gitlab.WithContext(g.x)); e != nil {
		return "", ErrorGitlabGetRelease.ErrorParent(e)
	}

	for _, l := range vers.Assets.Links {
		if containName != "" && strings.Contains(l.Name, containName) {
			return l.URL, nil
		} else if regexName != "" && artifact.CheckRegex(regexName, l.Name) {
			return l.URL, nil
		}
	}

	return "", ErrorGitlabNotFound.Error(nil)
}

func (g *gitlabModel) Download(dst ioutils.FileProgress, containName string, regexName string, release *version.Version) errors.Error {
	var (
		uri string
		inf os.FileInfo
		rsp *gitlab.Response
		req *retryablehttp.Request
		err error
		e   errors.Error
	)

	defer func() {
		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if uri, e = g.GetArtifact(containName, regexName, release); e != nil {
		return e
	} else if req, err = g.c.NewRequest(http.MethodGet, uri, nil, nil); err != nil {
		return ErrorGitlabRequestNew.ErrorParent(err)
	} else if rsp, err = g.c.Do(req, nil); err != nil {
		return ErrorGitlabRequestRun.ErrorParent(err)
	} else if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		return ErrorGitlabResponse.ErrorParent(errResponseCode)
	} else if rsp.ContentLength < 1 {
		return ErrorGitlabResponse.ErrorParent(errResponseContents)
	} else if rsp.Body == nil {
		return ErrorGitlabResponse.ErrorParent(errResponseBodyEmpty)
	} else if _, err = io.Copy(dst, rsp.Body); err != nil {
		return ErrorGitlabIOCopy.ErrorParent(err)
	} else if inf, err = dst.FileStat(); err != nil {
		return ErrorDestinationStat.ErrorParent(err)
	} else if inf.Size() != rsp.ContentLength {
		return ErrorDestinationSize.ErrorParent(errMisMatchingSize)
	}

	return nil
}
