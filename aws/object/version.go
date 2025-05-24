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
	"path"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err error) {
	in := sdksss.ListObjectVersionsInput{
		Bucket:  cli.GetBucketAws(),
		MaxKeys: sdkaws.Int32(1000),
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
	}

	var maxKeys int32
	if out != nil && out.MaxKeys != nil {
		maxKeys = *out.MaxKeys
	}

	if out != nil && out.IsTruncated != nil && *out.IsTruncated {
		return out.Versions, out.DeleteMarkers, *out.NextKeyMarker, *out.NextVersionIdMarker, int64(maxKeys), nil
	} else {
		return out.Versions, out.DeleteMarkers, "", "", int64(maxKeys), nil
	}
}

func (cli *client) VersionWalk(md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error {
	return cli.VersionWalkPrefix("", md, fv, fd)
}

func (cli *client) VersionWalkPrefix(prefix string, md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error {
	in := sdksss.ListObjectVersionsInput{
		Bucket:  cli.GetBucketAws(),
		MaxKeys: sdkaws.Int32(1000),
	}

	if prefix != "" {
		in.Prefix = sdkaws.String(prefix)
	}

	var (
		e  error
		km = sdkaws.String("")
		mi = sdkaws.String("")

		okm = true
		fmd = func(md Metadata) bool {
			return false
		}

		okv = true
		fov = func(obj sdktps.ObjectVersion) bool {
			return false
		}

		okd = true
		fod = func(del sdktps.DeleteMarkerEntry) bool {
			return false
		}
	)

	if md != nil {
		fmd = md
	}

	if fv != nil {
		fov = fv
	}

	if fd != nil {
		fod = fd
	}

	for {
		if len(*km) > 0 && len(*mi) > 0 {
			in.KeyMarker = km
			in.VersionIdMarker = mi
		}

		out, err := cli.s3.ListObjectVersions(cli.GetContext(), &in)

		if err != nil {
			return cli.GetError(err)
		} else if out == nil {
			return libhlp.ErrorResponse.Error(nil)
		} else if okm {
			okm = fmd(Metadata{
				Versions:      len(out.Versions),
				DeleteMarkers: len(out.DeleteMarkers),
			})
		}

		if okv {
			for _, o := range out.Versions {
				if !okv {
					break
				} else if o.Key == nil || len(*o.Key) < 1 {
					continue
				} else if o.VersionId == nil || len(*o.VersionId) < 1 {
					continue
				}

				okv = fov(o)
			}
		}

		if okd {
			for _, o := range out.DeleteMarkers {
				if !okd {
					break
				} else if o.Key == nil || len(*o.Key) < 1 {
					continue
				} else if o.VersionId == nil || len(*o.VersionId) < 1 {
					continue
				}

				okd = fod(o)
			}
		}

		if !okm && !okv && !okd {
			return e
		} else if out != nil && out.IsTruncated != nil && *out.IsTruncated {
			km = out.NextKeyMarker
			mi = out.NextVersionIdMarker
		} else {
			return e
		}
	}
}

func (cli *client) VersionGet(object, version string) (*sdksss.GetObjectOutput, error) {
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

func (cli *client) VersionHead(object, version string) (*sdksss.HeadObjectOutput, error) {
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

func (cli *client) VersionSize(object, version string) (size int64, err error) {
	var (
		h *sdksss.HeadObjectOutput
	)

	if h, err = cli.VersionHead(object, version); err != nil {
		return
	} else if h != nil && h.ContentLength != nil {
		return *h.ContentLength, nil
	} else {
		return 0, nil
	}
}

func (cli *client) VersionDelete(check bool, object, version string) error {
	return cli.VersionDeleteLock(check, object, version, false)
}

func (cli *client) VersionDeleteLock(check bool, object, version string, byPassGovernance bool) error {
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

	if byPassGovernance {
		in.BypassGovernanceRetention = sdkaws.Bool(true)
	}

	_, err := cli.s3.DeleteObject(cli.GetContext(), &in)

	if !check && err != nil && strings.Contains(err.Error(), "api error NoSuchKey") {
		return nil
	} else if err != nil {
		return cli.GetError(err)
	}

	return nil
}

func (cli *client) VersionCopy(source, version, destination string) error {
	return cli.VersionCopyBucket(*cli.GetBucketAws(), source, version, *cli.GetBucketAws(), destination)
}

func (cli *client) VersionCopyBucket(bucketSource, source, version, bucketDestination, destination string) error {
	in := sdksss.CopyObjectInput{
		Bucket: sdkaws.String(bucketDestination),
		Key:    sdkaws.String(destination),
	}

	if version != "" {
		in.CopySource = sdkaws.String(path.Join(bucketSource, source) + "?versionId=" + version)
	} else {
		in.CopySource = sdkaws.String(path.Join(bucketSource, source))
	}

	_, err := cli.s3.CopyObject(cli.GetContext(), &in)

	if err != nil {
		return cli.GetError(err)
	}

	return nil
}
