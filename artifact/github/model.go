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

package github

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"

	github "github.com/google/go-github/v33/github"
	hscvrs "github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
)

const (
	githubPageSize = 100
)

type githubModel struct {
	artcli.ClientHelper

	c *github.Client
	x context.Context
	o string
	p string
}

func (g *githubModel) ListReleases() (releases hscvrs.Collection, err liberr.Error) {
	var (
		e    error
		lopt = &github.ListOptions{
			Page:    0,
			PerPage: githubPageSize,
		}
		curr = 0
	)

	for {
		var (
			rels []*github.RepositoryRelease
			resp *github.Response
		)

		if rels, resp, e = g.c.Repositories.ListReleases(g.x, g.o, g.p, lopt); e != nil {
			return nil, ErrorGithubList.ErrorParent(e)
		} else {
			curr++
		}

		for _, r := range rels {
			v, _ := hscvrs.NewVersion(*r.TagName)

			if libart.ValidatePreRelease(v) {
				releases = append(releases, v)
			}
		}

		if curr > resp.LastPage {
			_ = sort.Reverse(releases)
			return
		} else {
			lopt.Page = curr
		}
	}
}

func (g *githubModel) GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err liberr.Error) {
	var (
		rels *github.RepositoryRelease
		e    error
	)

	if rels, _, e = g.c.Repositories.GetReleaseByTag(g.x, g.o, g.p, release.Original()); e != nil {
		return "", ErrorGithubGetRelease.ErrorParent(e)
	}

	for _, a := range rels.Assets {
		if containName != "" && strings.Contains(*a.Name, containName) {
			return *a.BrowserDownloadURL, nil
		} else if regexName != "" && libart.CheckRegex(regexName, *a.Name) {
			return *a.BrowserDownloadURL, nil
		}
	}

	return "", ErrorGithubNotFound.Error(nil)
}

func (g *githubModel) Download(dst libfpg.Progress, containName string, regexName string, release *hscvrs.Version) liberr.Error {
	var (
		uri string
		rsp *github.Response
		req *http.Request
		err error
		e   liberr.Error
		n   int64
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
	} else if req, err = g.c.NewRequest(http.MethodGet, uri, nil); err != nil {
		return ErrorGithubRequestNew.ErrorParent(err)
	} else if rsp, err = g.c.Do(g.x, req, nil); err != nil {
		return ErrorGithubRequestRun.ErrorParent(err)
	} else if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		return ErrorGithubResponse.ErrorParent(errResponseCode)
	} else if rsp.ContentLength < 1 {
		return ErrorGithubResponse.ErrorParent(errResponseContents)
	} else if rsp.Body == nil {
		return ErrorGithubResponse.ErrorParent(errResponseBodyEmpty)
	} else {
		dst.Reset(rsp.ContentLength)
	}

	if n, err = io.Copy(dst, rsp.Body); err != nil {
		return ErrorGithubIOCopy.ErrorParent(err)
	} else if n != rsp.ContentLength {
		return ErrorDestinationSize.ErrorParent(errMisMatchingSize)
	}

	return nil
}
