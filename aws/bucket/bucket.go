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
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) Check() liberr.Error {
	out, err := cli.s3.HeadBucket(cli.GetContext(), &sdksss.HeadBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return libhlp.ErrorBucketNotFound.ErrorParent(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return nil
}

func (cli *client) Create() liberr.Error {
	out, err := cli.s3.CreateBucket(cli.GetContext(), &sdksss.CreateBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || len(*out.Location) == 0 {
		return libhlp.ErrorResponse.Error(nil)
	}

	return cli.GetError(err)
}

func (cli *client) Delete() liberr.Error {
	_, err := cli.s3.DeleteBucket(cli.GetContext(), &sdksss.DeleteBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	return cli.GetError(err)
}

func (cli *client) List() ([]*sdkstp.Bucket, liberr.Error) {
	out, err := cli.s3.ListBuckets(cli.GetContext(), nil)

	if err != nil {
		return make([]*sdkstp.Bucket, 0), cli.GetError(err)
	} else if out == nil || out.Buckets == nil {
		return make([]*sdkstp.Bucket, 0), libhlp.ErrorAwsEmpty.Error(nil)
	}

	return out.Buckets, nil
}

func (cli *client) SetVersioning(state bool) liberr.Error {
	var status sdkstp.BucketVersioningStatus = libhlp.STATE_ENABLED
	if !state {
		status = libhlp.STATE_SUSPENDED
	}

	_, err := cli.s3.PutBucketVersioning(cli.GetContext(), &sdksss.PutBucketVersioningInput{
		Bucket: cli.GetBucketAws(),
		VersioningConfiguration: &sdkstp.VersioningConfiguration{
			Status: status,
		},
	})

	return cli.GetError(err)
}

func (cli *client) GetVersioning() (string, liberr.Error) {
	out, err := cli.s3.GetBucketVersioning(cli.GetContext(), &sdksss.GetBucketVersioningInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return "", cli.GetError(err)
	} else if out == nil {
		return "", libhlp.ErrorResponse.Error(nil)
	}

	// MarshalValue always return error as nil
	return string(out.Status), nil
}

func (cli *client) EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) liberr.Error {
	var status sdkstp.ReplicationRuleStatus = libhlp.STATE_ENABLED

	_, err := cli.s3.PutBucketReplication(cli.GetContext(), &sdksss.PutBucketReplicationInput{
		Bucket: cli.GetBucketAws(),
		ReplicationConfiguration: &sdkstp.ReplicationConfiguration{
			Role: sdkaws.String(srcRoleARN + "," + dstRoleARN),
			Rules: []*sdkstp.ReplicationRule{
				{
					Destination: &sdkstp.Destination{
						Bucket: sdkaws.String("arn:aws:s3:::" + dstBucketName),
					},
					Status: status,
					Prefix: sdkaws.String(""),
				},
			},
		},
	})

	return cli.GetError(err)
}

func (cli *client) DeleteReplication() liberr.Error {
	_, err := cli.s3.DeleteBucketReplication(cli.GetContext(), &sdksss.DeleteBucketReplicationInput{
		Bucket: cli.GetBucketAws(),
	})

	return cli.GetError(err)
}
