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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	hscvrs "github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	libfpg "github.com/nabbar/golib/file/progress"
)

type artifactoryModel struct {
	artcli.ClientHelper
	Do       func(req *http.Request) (*http.Response, error)
	ctx      context.Context
	endpoint *url.URL
	path     []string
	group    int
	regex    string
}

type ResponseChecksum struct {
	Md5    string
	Sha1   string
	Sha256 string
}

type ResponseStorage struct {
	Uri               string
	DownloadUri       string
	Repo              string
	Path              string
	RemoteUrl         string
	Created           time.Time
	CreatedBy         string
	LastModified      time.Time
	ModifiedBy        string
	LastUpdated       time.Time
	Size              string
	size              int64
	MimeType          string
	Checksums         ResponseChecksum
	OriginalChecksums ResponseChecksum
}

type ResponseReposChildrenStorage struct {
	Uri    string
	Folder bool
}

type ResponseReposStorage struct {
	Uri          string
	Repo         string
	Path         string
	Created      time.Time
	CreatedBy    string
	LastModified time.Time
	ModifiedBy   string
	LastUpdated  time.Time
	Children     []ResponseReposChildrenStorage
}

func (a *artifactoryModel) request(uri string, bodyResponse interface{}) error {
	var (
		ctx context.Context
		cnl context.CancelFunc
		req *http.Request
		rsp *http.Response

		e error
		u *url.URL
	)

	//ctx, cnl = context.WithTimeout(a.ctx, libhtc.TIMEOUT_5_SEC)
	ctx, cnl = context.WithCancel(a.ctx)
	defer func() {
		if cnl != nil {
			cnl()
		}
	}()

	u = &url.URL{
		Scheme:      a.endpoint.Scheme,
		Opaque:      a.endpoint.Opaque,
		User:        a.endpoint.User,
		Host:        a.endpoint.Host,
		Path:        a.endpoint.Path,
		RawPath:     a.endpoint.RawPath,
		ForceQuery:  a.endpoint.ForceQuery,
		RawQuery:    a.endpoint.RawQuery,
		Fragment:    a.endpoint.Fragment,
		RawFragment: a.endpoint.RawFragment,
	}

	u.Path += path.Join("api", "storage", path.Join(a.path...))

	if uri != "" {
		u.Path += path.Join(uri)
	}

	u.Path = path.Clean(u.Path)

	if req, e = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil); e != nil {
		return ErrorRequestInit.Error(e)
	}

	defer func() {
		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	if rsp, e = a.Do(req); e != nil {
		return ErrorRequestDo.Error(e)
	}

	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.StatusCode >= http.StatusBadRequest {
		//nolint #goerr113
		return ErrorRequestResponse.Error(fmt.Errorf("status: %v", rsp.Status))
	}

	if rsp.Body == nil {
		//nolint #goerr113
		return ErrorRequestResponseBodyEmpty.Error(fmt.Errorf("status: %v", rsp.Status))
	}

	if buf, e := ioutil.ReadAll(rsp.Body); e != nil {
		return ErrorRequestResponseBodyDecode.Error(e)
	} else if e = json.Unmarshal(buf, bodyResponse); e != nil {
		return ErrorRequestResponseBodyDecode.Error(e)
	}

	cnl()
	cnl = nil

	return nil

}

func (a *artifactoryModel) getStorageList() (sto []ResponseStorage, err error) {
	var (
		lst = ResponseReposStorage{}
		reg = regexp.MustCompile(a.regex)
	)

	if a.regex == "" {
		//nolint #goerr113
		return nil, ErrorParamEmpty.Error(fmt.Errorf("regex is empty: %s", a.regex))
	}

	if a.group < 1 {
		//nolint #goerr113
		return nil, ErrorParamEmpty.Error(fmt.Errorf("group extracted from regex is empty: %s - %v", a.regex, a.group))
	}

	if err = a.request("", &lst); err != nil {
		return nil, err
	} else if len(lst.Children) < 1 {
		return make([]ResponseStorage, 0), nil
	}

	sto = make([]ResponseStorage, 0)

	for _, c := range lst.Children {
		var (
			e   error
			res = ResponseStorage{}
		)

		if c.Folder {
			continue
		}

		if !reg.MatchString(c.Uri) {
			continue
		}

		if err = a.request(c.Uri, &res); err != nil {
			return nil, err
		}

		if res.size, e = strconv.ParseInt(res.Size, 10, 64); e != nil {
			return nil, ErrorRequestResponseBodyDecode.Error(e)
		}

		sto = append(sto, res)
	}

	return sto, nil
}

func (a *artifactoryModel) releasesAppendNotExist(releases hscvrs.Collection, vers *hscvrs.Version) hscvrs.Collection {
	for _, k := range releases {
		if k.Equal(vers) {
			return releases
		}
	}

	return append(releases, vers)
}

func (a *artifactoryModel) ListReleases() (releases hscvrs.Collection, err error) {
	var (
		reg = regexp.MustCompile(a.regex)
		sto []ResponseStorage
	)

	if sto, err = a.getStorageList(); err != nil {
		return nil, err
	}

	for _, f := range sto {
		grp := reg.FindStringSubmatch(f.Path)

		if len(grp) < a.group {
			continue
		}

		if v, e := hscvrs.NewVersion(grp[a.group]); e != nil {
			continue
		} else if !libart.ValidatePreRelease(v) {
			continue
		} else {
			releases = a.releasesAppendNotExist(releases, v)
		}
	}

	return releases, nil
}

func (a *artifactoryModel) getArtifact(containName string, regexName string, release *hscvrs.Version) (art *ResponseStorage, err error) {
	var (
		reg = regexp.MustCompile(a.regex)
		rg2 *regexp.Regexp

		sto []ResponseStorage
	)

	if regexName != "" {
		rg2 = regexp.MustCompile(regexName)
	}

	if sto, err = a.getStorageList(); err != nil {
		return nil, err
	}

	for _, f := range sto {
		grp := reg.FindStringSubmatch(f.Path)

		if len(grp) < a.group {
			continue
		}

		if v, e := hscvrs.NewVersion(grp[a.group]); e != nil {
			continue
		} else if !libart.ValidatePreRelease(v) {
			continue
		} else if release != nil && !v.Equal(release) {
			continue
		} else if containName != "" && !strings.Contains(f.Path, containName) {
			continue
		} else if rg2 != nil && !rg2.MatchString(f.Path) {
			continue
		} else {
			return &f, nil
		}
	}

	return nil, ErrorArtifactoryNotFound.Error(nil)
}

func (a *artifactoryModel) GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err error) {
	if art, err := a.getArtifact(containName, regexName, release); err != nil {
		return "", err
	} else {
		return art.DownloadUri, nil
	}
}

func (a *artifactoryModel) Download(dst libfpg.Progress, containName string, regexName string, release *hscvrs.Version) error {
	var (
		e error
		n int64

		art *ResponseStorage
		err error
		req *http.Request
		rsp *http.Response
	)

	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}

		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	if art, err = a.getArtifact(containName, regexName, release); err != nil {
		return err
	} else {
		dst.Reset(art.size)
	}

	if req, e = http.NewRequestWithContext(a.ctx, http.MethodGet, art.DownloadUri, nil); e != nil {
		return ErrorRequestInit.Error(e)
	} else if rsp, e = a.Do(req); e != nil {
		return ErrorRequestDo.Error(e)
	} else if rsp.StatusCode >= http.StatusBadRequest {
		//nolint #goerr113
		return ErrorRequestResponse.Error(fmt.Errorf("status: %v", rsp.Status))
	} else if rsp.Body == nil {
		//nolint #goerr113
		return ErrorRequestResponseBodyEmpty.Error(fmt.Errorf("status: %v", rsp.Status))
	} else if n, e = io.Copy(dst, rsp.Body); e != nil {
		return ErrorArtifactoryDownload.Error(e)
	} else if n != art.size {
		return ErrorDestinationSize.Error(errMisMatchingSize)
	} else if n != rsp.ContentLength {
		return ErrorDestinationSize.Error(errMisMatchingSize)
	}

	return nil
}
