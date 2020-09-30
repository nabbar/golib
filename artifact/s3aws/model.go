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
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/hashicorp/go-version"
	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/artifact/client"
	"github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
)

type s3awsModel struct {
	client.ClientHelper

	c     aws.AWS
	x     context.Context
	regex string
	group int
}

func (s *s3awsModel) ListReleases() (releases version.Collection, err errors.Error) {
	var (
		e errors.Error
		r *regexp.Regexp
		l []string
	)

	if s.regex == "" {
		return nil, ErrorParamsEmpty.Error(e)
	}

	if l, e = s.c.Object().Find(s.regex); e != nil {
		return nil, ErrorS3AWSFind.Error(e)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return nil, ErrorS3AWSRegex.ErrorParent(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e := version.NewVersion(grp[s.group]); e != nil {
			continue
		} else if !artifact.ValidatePreRelease(v) {
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

func (s *s3awsModel) GetArtifact(containName string, regexName string, release *version.Version) (link string, err errors.Error) {
	var (
		e errors.Error
		r *regexp.Regexp
		l []string
	)

	if s.regex == "" {
		return "", ErrorParamsEmpty.Error(e)
	}

	if l, e = s.c.Object().Find(s.regex); e != nil {
		return "", ErrorS3AWSFind.Error(e)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return "", ErrorS3AWSRegex.ErrorParent(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e := version.NewVersion(grp[s.group]); e != nil {
			return "", ErrorS3AWSNewVers.ErrorParent(getError(errVersion, grp[s.group]), e)
		} else if v.Equal(release) {
			uri := s.c.Config().GetEndpoint()
			uri.Path += "/" + s.c.GetBucketName() + "/" + o
			uri.Path = strings.Replace(uri.Path, "//", "/", -1)

			if containName != "" && strings.Contains(o, containName) {
				return uri.String(), nil
			}

			if regexName != "" {
				if ok, e := regexp.MatchString(regexName, o); e == nil && ok {
					return uri.String(), nil
				}
			}

			if containName == "" && regexName == "" {
				return uri.String(), nil
			}
		}
	}

	return "", ErrorS3AWSNotFound.ErrorParent(getError(errVersRequest, release.String()))
}

func (s *s3awsModel) Download(dst ioutils.FileProgress, containName string, regexName string, release *version.Version) errors.Error {
	var (
		e errors.Error
		r *regexp.Regexp
		l []string
	)

	if s.regex == "" {
		return ErrorParamsEmpty.Error(e)
	}

	if l, e = s.c.Object().Find(s.regex); e != nil {
		return ErrorS3AWSFind.Error(e)
	}

	r = regexp.MustCompile(s.regex)

	for _, o := range l {
		grp := r.FindStringSubmatch(o)

		if len(grp) < s.group {
			return ErrorS3AWSRegex.ErrorParent(getError(errRegexGroup, s.regex, len(grp), s.group))
		}

		if v, e := version.NewVersion(grp[s.group]); e != nil {
			return ErrorS3AWSNewVers.ErrorParent(getError(errVersion, grp[s.group]), e)
		} else if v.Equal(release) {
			if containName != "" && strings.Contains(o, containName) {
				return s.downloadObject(dst, o)
			}

			if regexName != "" {
				if ok, e := regexp.MatchString(regexName, o); e == nil && ok {
					return s.downloadObject(dst, o)
				} else if e != nil {
					println(e)
				} else {
					println("regex " + regexName + " => KO - " + o)
				}
			}

			if containName == "" && regexName == "" {
				return s.downloadObject(dst, o)
			}
		}
	}

	return ErrorS3AWSNotFound.ErrorParent(getError(errVersRequest, release.String()))
}

func (s *s3awsModel) downloadObject(dst ioutils.FileProgress, object string) errors.Error {
	var (
		r *s3.GetObjectOutput
		e errors.Error
		i os.FileInfo
		j int64
	)

	defer func() {
		if r != nil && r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	if j, e = s.c.Object().Size(object); e != nil {
		err := ErrorS3AWSDownloadError.ErrorParent(getError(errObject, object))
		err.AddParentError(e)
		return err
	} else if j < 1 {
		return ErrorS3AWSDownloadError.ErrorParent(getError(errObjectEmpty, object))
	}

	dst.ResetMax(j)

	if r, e = s.c.Object().Get(object); e != nil {
		err := ErrorS3AWSDownloadError.ErrorParent(getError(errObject, object))
		err.AddParentError(e)
		return err
	} else if r.Body == nil {
		return ErrorS3AWSIOReaderError.ErrorParent(getError(errObject, object))
	} else if _, err := io.Copy(dst, r.Body); err != nil {
		return ErrorS3AWSDownloadError.ErrorParent(getError(errObject, object), err)
	} else if i, e = dst.FileStat(); e != nil {
		err := ErrorS3AWSDownloadError.ErrorParent(getError(errObject, object))
		err.AddParentError(e)
		return err
	} else if i.Size() != j {
		return ErrorS3AWSDownloadError.ErrorParent(getError(errObjectSize, object))
	}

	return nil
}
