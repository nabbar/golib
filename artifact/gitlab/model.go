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
	"sort"
	"strings"

	hschtc "github.com/hashicorp/go-retryablehttp"
	hscvrs "github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	gitlabPageSize = 100
)

type gitlabModel struct {
	artcli.ClientHelper

	c *gitlab.Client
	x context.Context
	p int
}

func (g *gitlabModel) ListReleases() (releases hscvrs.Collection, err liberr.Error) {
	var (
		e    error
		lopt = &gitlab.ListReleasesOptions{
			ListOptions: gitlab.ListOptions{
				Page:    0,
				PerPage: gitlabPageSize,
			},
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
			v, _ := hscvrs.NewVersion(r.TagName)

			if libart.ValidatePreRelease(v) {
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

func (g *gitlabModel) GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err liberr.Error) {
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
		} else if regexName != "" && libart.CheckRegex(regexName, l.Name) {
			return l.URL, nil
		}
	}

	return "", ErrorGitlabNotFound.Error(nil)
}

func (g *gitlabModel) Download(dst libfpg.Progress, containName string, regexName string, release *hscvrs.Version) liberr.Error {
	var (
		uri string
		rsp *gitlab.Response
		req *hschtc.Request
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
	} else {
		dst.Reset(rsp.ContentLength)
	}

	if n, err = io.Copy(dst, rsp.Body); err != nil {
		return ErrorGitlabIOCopy.ErrorParent(err)
	} else if n != rsp.ContentLength {
		return ErrorDestinationSize.ErrorParent(errMisMatchingSize)
	}

	return nil
}
