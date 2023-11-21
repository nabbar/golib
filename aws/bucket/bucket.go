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

package bucket

import (
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) Check() error {
	out, err := cli.s3.HeadBucket(cli.GetContext(), &sdksss.HeadBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return libhlp.ErrorBucketNotFound.Error(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return nil
}

func (cli *client) Create(RegionConstraint string) error {
	return cli._create(RegionConstraint, false)
}

func (cli *client) CreateWithLock(RegionConstraint string) error {
	return cli._create(RegionConstraint, true)
}

func (cli *client) _create(RegionConstraint string, lockEnable bool) error {
	in := &sdksss.CreateBucketInput{
		Bucket:                    cli.GetBucketAws(),
		CreateBucketConfiguration: &sdkstp.CreateBucketConfiguration{},
	}

	if RegionConstraint != "" {
		in.CreateBucketConfiguration.LocationConstraint = sdkstp.BucketLocationConstraint(RegionConstraint)
	} else {
		in.CreateBucketConfiguration.LocationConstraint = sdkstp.BucketLocationConstraint(cli.GetRegion())
	}

	if lockEnable {
		in.ObjectLockEnabledForBucket = sdkaws.Bool(true)
	}

	out, err := cli.s3.CreateBucket(cli.GetContext(), in)

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || len(*out.Location) == 0 {
		return libhlp.ErrorResponse.Error(nil)
	}

	return nil

}

func (cli *client) Delete() error {
	_, err := cli.s3.DeleteBucket(cli.GetContext(), &sdksss.DeleteBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	return cli.GetError(err)
}

func (cli *client) List() ([]sdkstp.Bucket, error) {
	out, err := cli.s3.ListBuckets(cli.GetContext(), nil)

	if err != nil {
		return make([]sdkstp.Bucket, 0), cli.GetError(err)
	} else if out == nil || out.Buckets == nil {
		return make([]sdkstp.Bucket, 0), libhlp.ErrorAwsEmpty.Error(nil)
	}

	return out.Buckets, nil
}

func (cli *client) Walk(f WalkFunc) error {
	out, err := cli.s3.ListBuckets(cli.GetContext(), nil)

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || out.Buckets == nil {
		return libhlp.ErrorAwsEmpty.Error(nil)
	}

	var e error
	for _, b := range out.Buckets {
		if b.Name == nil || b.CreationDate == nil || len(*b.Name) < 3 || b.CreationDate.IsZero() {
			continue
		}
		if f != nil {
			e = f(e, b)
		}
	}

	return e
}
