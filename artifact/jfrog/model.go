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
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
)

type artifactoryModel struct {
	artcli.ClientHelper
	Do       func(req *http.Request) (*http.Response, error)
	ctx      context.Context
	endpoint *url.URL
	path     []string
	name     string
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

func (a *artifactoryModel) request(uri string, bodyResponse interface{}) liberr.Error {
	var (
		ctx context.Context
		cnl context.CancelFunc
		req *http.Request
		rsp *http.Response

		e error
		u *url.URL
	)

	defer func() {
		if cnl != nil {
			cnl()
		}

		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}

		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	//ctx, cnl = context.WithTimeout(a.ctx, libhtc.TIMEOUT_5_SEC)
	ctx, cnl = context.WithCancel(a.ctx)

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
		return ErrorRequestInit.ErrorParent(e)
	}

	if rsp, e = a.Do(req); e != nil {
		return ErrorRequestDo.ErrorParent(e)
	}

	if rsp.StatusCode >= 400 {
		return ErrorRequestResponse.ErrorParent(fmt.Errorf("status: %v", rsp.Status))
	}

	if rsp.Body == nil {
		return ErrorRequestResponseBodyEmpty.ErrorParent(fmt.Errorf("status: %v", rsp.Status))
	}

	if buf, e := ioutil.ReadAll(rsp.Body); e != nil {
		return ErrorRequestResponseBodyDecode.ErrorParent(e)
	} else if e = json.Unmarshal(buf, bodyResponse); e != nil {
		return ErrorRequestResponseBodyDecode.ErrorParent(e)
	}

	cnl()
	cnl = nil

	return nil

}

func (a *artifactoryModel) getStorageList() (sto []ResponseStorage, err liberr.Error) {
	var (
		lst = ResponseReposStorage{}
	)

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

		if err = a.request(c.Uri, &res); err != nil {
			return nil, err
		}

		if res.size, e = strconv.ParseInt(res.Size, 10, 64); e != nil {
			return nil, ErrorRequestResponseBodyDecode.ErrorParent(e)
		}

		sto = append(sto, res)
	}

	return sto, nil
}

func (a *artifactoryModel) releasesAppendNotExist(releases version.Collection, vers *version.Version) version.Collection {
	for _, k := range releases {
		if k.Equal(vers) {
			return releases
		}
	}

	return append(releases, vers)
}

func (a *artifactoryModel) ListReleases() (releases version.Collection, err liberr.Error) {
	var (
		r   *regexp.Regexp
		sto []ResponseStorage
	)

	if a.regex != "" {
		r = regexp.MustCompile(a.regex)
	}

	if sto, err = a.getStorageList(); err != nil {
		return nil, err
	}

	for _, f := range sto {
		if a.name != "" && !strings.Contains(f.Path, a.name) {
			continue
		}

		if r != nil && !r.MatchString(f.Path) {
			continue
		}

		if v, e := version.NewVersion(f.Path); e != nil {
			continue
		} else if !libart.ValidatePreRelease(v) {
			continue
		} else {
			releases = a.releasesAppendNotExist(releases, v)
		}
	}

	return releases, nil
}

func (a *artifactoryModel) getArtifact(containName string, regexName string, release *version.Version) (art *ResponseStorage, err liberr.Error) {
	var (
		r1 *regexp.Regexp
		r2 *regexp.Regexp

		sto []ResponseStorage
	)

	if a.regex != "" {
		r1 = regexp.MustCompile(a.regex)
	}

	if regexName != "" {
		r2 = regexp.MustCompile(regexName)
	}

	if sto, err = a.getStorageList(); err != nil {
		return nil, err
	}

	for _, f := range sto {
		if a.name != "" && !strings.Contains(f.Path, a.name) {
			continue
		}

		if r1 != nil && !r1.MatchString(f.Path) {
			continue
		}

		if v, e := version.NewVersion(f.Path); e != nil {
			continue
		} else if !libart.ValidatePreRelease(v) {
			continue
		} else if release != nil && !v.Equal(release) {
			continue
		} else if containName != "" && !strings.Contains(f.Path, containName) {
			continue
		} else if r2 != nil && !r2.MatchString(f.Path) {
			continue
		} else {
			return &f, nil
		}
	}

	return nil, ErrorArtifactoryNotFound.Error(nil)
}

func (a *artifactoryModel) GetArtifact(containName string, regexName string, release *version.Version) (link string, err liberr.Error) {
	if art, err := a.getArtifact(containName, regexName, release); err != nil {
		return "", err
	} else {
		return art.DownloadUri, nil
	}
}

func (a *artifactoryModel) Download(dst libiot.FileProgress, containName string, regexName string, release *version.Version) liberr.Error {
	var (
		e error

		art *ResponseStorage
		err liberr.Error
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
	}

	dst.ResetMax(art.size)

	if req, e = http.NewRequestWithContext(a.ctx, http.MethodGet, art.DownloadUri, nil); e != nil {
		return ErrorRequestInit.ErrorParent(e)
	} else if rsp, e = a.Do(req); e != nil {
		return ErrorRequestDo.ErrorParent(e)
	} else if rsp.StatusCode >= 400 {
		return ErrorRequestResponse.ErrorParent(fmt.Errorf("status: %v", rsp.Status))
	} else if rsp.Body == nil {
		return ErrorRequestResponseBodyEmpty.ErrorParent(fmt.Errorf("status: %v", rsp.Status))
	} else if _, e := dst.ReadFrom(rsp.Body); e != nil {
		return ErrorArtifactoryDownload.ErrorParent(e)
	}

	return nil
}
