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
)

// art implements the artifact.Client interface for JFrog Artifactory.
// It uses the Artifactory Storage API to list and download artifacts.
// Version extraction is performed via regex matching on file names.
type art struct {
	artcli.Helper
	Do       func(req *http.Request) (*http.Response, error) // HTTP client Do function
	ctx      context.Context                                 // Request context
	endpoint *url.URL                                        // Artifactory base URL
	path     []string                                        // Repository path segments
	group    int                                             // Regex capture group for version
	regex    string                                          // Regex pattern to match artifacts
}

// Checksum contains file integrity checksums returned by Artifactory.
type Checksum struct {
	Md5    string // MD5 hash
	Sha1   string // SHA-1 hash
	Sha256 string // SHA-256 hash
}

// Storage represents a file object in Artifactory with its metadata.
// This structure is returned by the Storage API for individual files.
type Storage struct {
	Uri               string    // API URI
	DownloadUri       string    // Download URL
	Repo              string    // Repository name
	Path              string    // File path
	RemoteUrl         string    // Remote URL (if proxy)
	Created           time.Time // Creation timestamp
	CreatedBy         string    // Creator username
	LastModified      time.Time // Last modification timestamp
	ModifiedBy        string    // Last modifier username
	LastUpdated       time.Time // Last update timestamp
	Size              string    // Size as string
	size              int64     // Size as int64 (internal)
	MimeType          string    // MIME type
	Checksums         Checksum  // File checksums
	OriginalChecksums Checksum  // Original checksums (if proxy)
}

// ChildStore represents a child item (file or folder) in a repository listing.
type ChildStore struct {
	Uri    string // Relative URI
	Folder bool   // True if item is a folder
}

// RepoStore represents a repository or folder listing in Artifactory.
// This structure is returned by the Storage API for directory listings.
type RepoStore struct {
	Uri          string       // API URI
	Repo         string       // Repository name
	Path         string       // Folder path
	Created      time.Time    // Creation timestamp
	CreatedBy    string       // Creator username
	LastModified time.Time    // Last modification timestamp
	ModifiedBy   string       // Last modifier username
	LastUpdated  time.Time    // Last update timestamp
	Children     []ChildStore // Child items
}

func (a *art) request(uri string, bodyResponse interface{}) error {
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

	if buf, e := io.ReadAll(rsp.Body); e != nil {
		return ErrorRequestResponseBodyDecode.Error(e)
	} else if e = json.Unmarshal(buf, bodyResponse); e != nil {
		return ErrorRequestResponseBodyDecode.Error(e)
	}

	cnl()
	cnl = nil

	return nil

}

func (a *art) getStorageList() (sto []Storage, err error) {
	var (
		lst = RepoStore{}
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
		return make([]Storage, 0), nil
	}

	sto = make([]Storage, 0)

	for _, c := range lst.Children {
		var (
			e   error
			res = Storage{}
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

func (a *art) releasesAppendNotExist(releases hscvrs.Collection, vers *hscvrs.Version) hscvrs.Collection {
	for _, k := range releases {
		if k.Equal(vers) {
			return releases
		}
	}

	return append(releases, vers)
}

func (a *art) ListReleases() (releases hscvrs.Collection, err error) {
	var (
		reg = regexp.MustCompile(a.regex)
		sto []Storage
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

func (a *art) getArtifact(containName string, regexName string, release *hscvrs.Version) (art *Storage, err error) {
	var (
		reg = regexp.MustCompile(a.regex)
		rg2 *regexp.Regexp

		sto []Storage
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

func (a *art) GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err error) {
	if art, err := a.getArtifact(containName, regexName, release); err != nil {
		return "", err
	} else {
		return art.DownloadUri, nil
	}
}

func (a *art) Download(containName string, regexName string, release *hscvrs.Version) (int64, io.ReadCloser, error) {
	var (
		e error

		art *Storage
		err error
		req *http.Request
		rsp *http.Response
	)

	defer func() {
		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	if art, err = a.getArtifact(containName, regexName, release); err != nil {
		return 0, nil, err
	}

	if req, e = http.NewRequestWithContext(a.ctx, http.MethodGet, art.DownloadUri, nil); e != nil {
		return 0, nil, ErrorRequestInit.Error(e)
	} else if rsp, e = a.Do(req); e != nil {
		return 0, nil, ErrorRequestDo.Error(e)
	} else if rsp.StatusCode >= http.StatusBadRequest {
		//nolint #goerr113
		return 0, nil, ErrorRequestResponse.Error(fmt.Errorf("status: %v", rsp.Status))
	} else if rsp.Body == nil {
		//nolint #goerr113
		return 0, nil, ErrorRequestResponseBodyEmpty.Error(fmt.Errorf("status: %v", rsp.Status))
	} else if art.size != rsp.ContentLength {
		_ = rsp.Body.Close()
		return 0, nil, ErrorDestinationSize.Error(errMisMatchingSize)
	} else {
		return rsp.ContentLength, rsp.Body, nil
	}
}
