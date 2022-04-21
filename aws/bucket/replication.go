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
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) LoadReplication() (*sdkstp.ReplicationConfiguration, liberr.Error) {
	in := sdksss.GetBucketReplicationInput{
		Bucket: cli.GetBucketAws(),
	}

	out, err := cli.s3.GetBucketReplication(cli.GetContext(), &in)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out.ReplicationConfiguration, nil
	}
}

func (cli *client) EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) liberr.Error {
	var status sdkstp.ReplicationRuleStatus = libhlp.STATE_ENABLED

	_, err := cli.s3.PutBucketReplication(cli.GetContext(), &sdksss.PutBucketReplicationInput{
		Bucket: cli.GetBucketAws(),
		ReplicationConfiguration: &sdkstp.ReplicationConfiguration{
			Role: sdkaws.String(srcRoleARN + "," + dstRoleARN),
			Rules: []sdkstp.ReplicationRule{
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
