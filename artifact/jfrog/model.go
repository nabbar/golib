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

package jfrog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	jauth "github.com/jfrog/jfrog-client-go/auth"

	"github.com/hashicorp/go-version"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
)

const (
	gitlabPageSize = 100
)

type artifactoryModel struct {
	artcli.ClientHelper

	c artifactory.ArtifactoryServicesManager
	a jauth.ServiceDetails
	o Options
	x context.Context
	r string
}

func (g *artifactoryModel) ListReleases() (releases version.Collection, err liberr.Error) {
	var (
		search = services.NewSearchParams()
		reader *content.ContentReader
		e      error
	)

	if g.o.GetRecursive() {
		search.Recursive = true
	} else {
		search.Recursive = false
	}

	if g.o.GetPattern() != "" {
		search.Pattern = path.Join(g.r, g.o.GetPattern())
	}

	if g.o.GetExcludePattern() != "" {
		search.Exclusions = []string{g.o.GetExcludePattern()}
	}

	if g.o.LenProps() > 0 {
		search.Props = g.o.EncProps()
	}

	if g.o.LenExcludeProps() > 0 {
		search.ExcludeProps = g.o.EncExcludeProps()
	}

	if reader, e = g.c.SearchFiles(search); e != nil {
		return nil, ErrorArtifactoryList.ErrorParent(e)
	} else if reader == nil {
		return nil, ErrorArtifactoryList.ErrorParent(e)
	}

	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()

	for {
		if v, _, e := g.getContentSearch(reader); errors.Is(e, io.EOF) {
			break
		} else if v != nil {
			var found bool
			for _, k := range releases {
				if k.Equal(v) {
					found = true
					break
				}
			}
			if !found {
				releases = append(releases, v)
			}
		}
	}

	return releases, nil
}

func (g *artifactoryModel) getContentSearch(reader *content.ContentReader) (*version.Version, *utils.ResultItem, error) {
	var res = utils.ResultItem{}

	if err := reader.NextRecord(&res); err != nil {
		return nil, nil, err
	}

	if reader.IsEmpty() {
		return nil, nil, nil
	}

	if v, e := version.NewVersion(res.Name); e != nil {
		return nil, nil, ErrorArtifactoryVersion.ErrorParent(e)
	} else if !artifact.ValidatePreRelease(v) {
		return nil, &res, nil
	} else {
		return v, &res, nil
	}
}

func (g *artifactoryModel) GetArtifact(containName string, regexName string, release *version.Version) (link string, err liberr.Error) {
	var (
		res *utils.ResultItem
		uri *url.URL
	)

	if u, e := url.Parse(g.a.GetUrl()); e != nil {
		return "", ErrorURLParse.ErrorParent(e)
	} else {
		uri = u
	}

	if res, err = g.getArtifact(containName, regexName, release); err != nil {
		return "", err
	}

	uri.Path += "/" + res.GetItemRelativeLocation()
	uri.Path = strings.Replace(uri.Path, "//", "/", -1)

	return uri.String(), nil
}

func (g *artifactoryModel) Download(dst libiot.FileProgress, containName string, regexName string, release *version.Version) liberr.Error {
	var (
		uri *url.URL
		err liberr.Error
		res *utils.ResultItem
	)

	if u, e := url.Parse(g.a.GetUrl()); e != nil {
		return ErrorURLParse.ErrorParent(e)
	} else {
		uri = u
	}

	if res, err = g.getArtifact(containName, regexName, release); err != nil {
		return err
	}

	uri.Path += "/" + res.GetItemRelativeLocation()
	uri.Path = strings.Replace(uri.Path, "//", "/", -1)

	dst.ResetMax(res.Size)

	return g.dwdArtifact(dst, res)
}

func (g *artifactoryModel) dwdArtifact(dst libiot.FileProgress, art *utils.ResultItem) liberr.Error {
	var (
		e   error
		lnk string
		rsp *http.Response
	)

	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	dst.ResetMax(art.Size)

	if rsp, e = http.Get(lnk); e != nil {
		return ErrorArtifactoryReleaseRequest.ErrorParent(e)
	} else if rsp.StatusCode >= 400 {
		return ErrorArtifactoryReleaseRequest.ErrorParent(fmt.Errorf("status code: %s", rsp.Status))
	} else if rsp.Body == nil {
		return ErrorArtifactoryReleaseRequest.ErrorParent(fmt.Errorf("status code: %s, body is empty", rsp.Status))
	} else if _, e := dst.ReadFrom(rsp.Body); e != nil {
		return ErrorArtifactoryDownload.ErrorParent(e)
	}

	return nil
}

func (g *artifactoryModel) getArtifact(containName string, regexName string, release *version.Version) (*utils.ResultItem, liberr.Error) {
	var (
		search = services.NewSearchParams()
		reader *content.ContentReader
		regex  *regexp.Regexp
		e      error
	)

	if regexName != "" {
		regex = regexp.MustCompile(regexName)
	}

	if g.o.GetRecursive() {
		search.Recursive = true
	} else {
		search.Recursive = false
	}

	if g.o.GetPattern() != "" {
		search.Pattern = g.o.GetPattern()
	}

	if g.o.GetExcludePattern() != "" {
		search.Exclusions = []string{g.o.GetExcludePattern()}
	}

	if g.o.LenProps() > 0 {
		search.Props = g.o.EncProps()
	}

	if g.o.LenExcludeProps() > 0 {
		search.ExcludeProps = g.o.EncExcludeProps()
	}

	if reader, e = g.c.SearchFiles(search); e != nil {
		return nil, ErrorArtifactoryList.ErrorParent(e)
	} else if reader == nil {
		return nil, ErrorArtifactoryList.ErrorParent(e)
	}

	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()

	for {
		if v, r, e := g.getContentSearch(reader); errors.Is(e, io.EOF) {
			break
		} else if v != nil && r != nil && v.Equal(release) {
			if containName != "" && strings.Contains(r.Name, containName) {
				return r, nil
			}

			if regex != nil && regex.MatchString(r.Name) {
				return r, nil
			}

			if containName == "" && regexName == "" {
				return r, nil
			}
		}
	}

	return nil, ErrorArtifactoryReleaseNotFound.Error(nil)
}
