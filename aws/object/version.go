/*
 *  MIT License
 *
 *  Copyright (c) 2022 Nicolas JUHEL
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

package object

import (
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err liberr.Error) {
	in := sdksss.ListObjectVersionsInput{
		Bucket:  cli.GetBucketAws(),
		MaxKeys: 1000,
	}

	if prefix != "" {
		in.Prefix = sdkaws.String(prefix)
	}

	if keyMarker != "" && markerId != "" {
		in.KeyMarker = sdkaws.String(keyMarker)
		in.VersionIdMarker = sdkaws.String(markerId)
	}

	out, e := cli.s3.ListObjectVersions(cli.GetContext(), &in)

	if e != nil {
		return nil, nil, "", "", 0, cli.GetError(e)
	} else if out.IsTruncated {
		return out.Versions, out.DeleteMarkers, *out.NextKeyMarker, *out.NextVersionIdMarker, int64(out.MaxKeys), nil
	} else {
		return out.Versions, out.DeleteMarkers, "", "", int64(out.MaxKeys), nil
	}
}

func (cli *client) VersionWalk(fv VersionWalkFunc, fd DelMakWalkFunc) liberr.Error {
	return cli.VersionWalkPrefix("", fv, fd)
}

func (cli *client) VersionWalkPrefix(prefix string, fv VersionWalkFunc, fd DelMakWalkFunc) liberr.Error {
	in := sdksss.ListObjectVersionsInput{
		Bucket:  cli.GetBucketAws(),
		MaxKeys: 1000,
	}

	if prefix != "" {
		in.Prefix = sdkaws.String(prefix)
	}

	var (
		e  liberr.Error
		km = sdkaws.String("")
		mi = sdkaws.String("")
	)

	for {
		if len(*km) > 0 && len(*mi) > 0 {
			in.KeyMarker = km
			in.VersionIdMarker = mi
		}

		out, err := cli.s3.ListObjectVersions(cli.GetContext(), &in)

		if err != nil {
			return cli.GetError(err)
		}

		for _, o := range out.Versions {
			if o.Key == nil || len(*o.Key) < 3 {
				continue
			} else if o.VersionId == nil || len(*o.VersionId) < 1 {
				continue
			}

			if fv != nil {
				e = fv(e, o)
			}
		}

		for _, o := range out.DeleteMarkers {
			if o.Key == nil || len(*o.Key) < 3 {
				continue
			} else if o.VersionId == nil || len(*o.VersionId) < 1 {
				continue
			}

			if fd != nil {
				e = fd(e, o)
			}
		}

		if out.IsTruncated {
			km = out.NextKeyMarker
			mi = out.NextVersionIdMarker
		} else {
			return e
		}
	}
}

func (cli *client) VersionGet(object, version string) (*sdksss.GetObjectOutput, liberr.Error) {
	in := sdksss.GetObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, err := cli.s3.GetObject(cli.GetContext(), &in)

	if err != nil {
		defer func() {
			if out != nil && out.Body != nil {
				_ = out.Body.Close()
			}
		}()
		return nil, cli.GetError(err)
	} else if out.Body == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}

}

func (cli *client) VersionHead(object, version string) (*sdksss.HeadObjectOutput, liberr.Error) {
	in := sdksss.HeadObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, e := cli.s3.HeadObject(cli.GetContext(), &in)

	if e != nil {
		return nil, cli.GetError(e)
	} else if out.ETag == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}
}

func (cli *client) VersionSize(object, version string) (size int64, err liberr.Error) {
	var (
		h *sdksss.HeadObjectOutput
	)

	if h, err = cli.VersionHead(object, version); err != nil {
		return
	} else {
		return h.ContentLength, nil
	}
}

func (cli *client) VersionDelete(check bool, object, version string) liberr.Error {
	if check {
		if _, err := cli.VersionHead(object, version); err != nil {
			return err
		}
	}

	in := sdksss.DeleteObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	_, err := cli.s3.DeleteObject(cli.GetContext(), &in)

	if !check && err != nil {
		e := err.Error()
		if strings.Contains(e, "api error NoSuchKey") {
			return nil
		}
	}

	return cli.GetError(err)
}
