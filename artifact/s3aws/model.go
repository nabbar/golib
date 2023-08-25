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

package s3aws

import (
	"context"
	"io"
	"regexp"
	"strings"

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	hscvrs "github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	libaws "github.com/nabbar/golib/aws"
	libfpg "github.com/nabbar/golib/file/progress"
)

type s3awsModel struct {
	artcli.ClientHelper

	c     libaws.AWS
	x     context.Context
	regex string
	group int
}

func (s *s3awsModel) ListReleases() (releases hscvrs.Collection, err error) {
	var (
		e error
		r *regexp.Regexp
		l []string
	)

	if s.regex == "" {
		return nil, ErrorParamEmpty.Error(nil)
	}

	if l, e = s.c.Object().Find(s.regex); e != nil {
		return nil, ErrorS3AWSFind.Error(e)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return nil, ErrorS3AWSRegex.Error(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e := hscvrs.NewVersion(grp[s.group]); e != nil {
			continue
		} else if !libart.ValidatePreRelease(v) {
			continue
		} else {
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

func (s *s3awsModel) GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err error) {
	var (
		e error
		k bool
		l []string
		r *regexp.Regexp
		v *hscvrs.Version
	)

	if s.regex == "" {
		return "", ErrorParamEmpty.Error(nil)
	}

	if l, err = s.c.Object().Find(s.regex); err != nil {
		return "", ErrorS3AWSFind.Error(err)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return "", ErrorS3AWSRegex.Error(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e = hscvrs.NewVersion(grp[s.group]); e != nil {
			return "", ErrorS3AWSNewVers.Error(getError(errVersion, grp[s.group]), e)
		} else if v.Equal(release) {
			uri := s.c.Config().GetEndpoint()
			uri.Path += "/" + s.c.GetBucketName() + "/" + o
			uri.Path = strings.Replace(uri.Path, "//", "/", -1)

			if containName != "" && strings.Contains(o, containName) {
				return uri.String(), nil
			}

			if regexName != "" {
				if k, e = regexp.MatchString(regexName, o); e == nil && k {
					return uri.String(), nil
				}
			}

			if containName == "" && regexName == "" {
				return uri.String(), nil
			}
		}
	}

	return "", ErrorS3AWSNotFound.Error(getError(errVersRequest, release.String()))
}

func (s *s3awsModel) Download(dst libfpg.Progress, containName string, regexName string, release *hscvrs.Version) error {
	var (
		e error
		r *regexp.Regexp
		l []string
		v *hscvrs.Version
		k bool

		err error
	)

	if s.regex == "" {
		return ErrorParamEmpty.Error(nil)
	}

	if l, err = s.c.Object().Find(s.regex); e != nil {
		return ErrorS3AWSFind.Error(err)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return ErrorS3AWSRegex.Error(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e = hscvrs.NewVersion(grp[s.group]); e != nil {
			return ErrorS3AWSNewVers.Error(getError(errVersion, grp[s.group]), e)
		} else if v.Equal(release) {
			if containName != "" && strings.Contains(o, containName) {
				return s.downloadObject(dst, o)
			}

			if regexName != "" {
				if k, e = regexp.MatchString(regexName, o); e == nil && k {
					return s.downloadObject(dst, o)
				}
			}

			if containName == "" && regexName == "" {
				return s.downloadObject(dst, o)
			}
		}
	}

	return ErrorS3AWSNotFound.Error(getError(errVersRequest, release.String()))
}

func (s *s3awsModel) downloadObject(dst libfpg.Progress, object string) error {
	var (
		r *sdksss.GetObjectOutput
		e error
		j int64
		n int64

		err error
	)

	defer func() {
		if r != nil && r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	if j, err = s.c.Object().Size(object); err != nil {
		er := ErrorS3AWSDownloadError.Error(getError(errObject, object))
		er.Add(err)
		return er
	} else if j < 1 {
		return ErrorS3AWSDownloadError.Error(getError(errObjectEmpty, object))
	} else {
		dst.Reset(j)
	}

	if r, err = s.c.Object().Get(object); err != nil {
		er := ErrorS3AWSDownloadError.Error(getError(errObject, object))
		er.Add(err)
		return er
	} else if r.Body == nil {
		return ErrorS3AWSIOReaderError.Error(getError(errObject, object))
	} else if n, e = io.Copy(dst, r.Body); e != nil {
		return ErrorS3AWSDownloadError.Error(getError(errObject, object), e)
	} else if n != j {
		return ErrorS3AWSDownloadError.Error(getError(errObjectSize, object))
	}

	return nil
}
