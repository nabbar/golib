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
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) GetRetention(object, version string) (*sdktps.ObjectLockRetention, liberr.Error) {
	in := sdksss.GetObjectRetentionInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, err := cli.s3.GetObjectRetention(cli.GetContext(), &in)

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out.Retention, nil
}

func (cli *client) SetRetention(object, version string, retentionUntil time.Time) liberr.Error {
	in := sdksss.PutObjectRetentionInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		Retention: &sdktps.ObjectLockRetention{
			RetainUntilDate: sdkaws.Time(retentionUntil),
		},
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	_, err := cli.s3.PutObjectRetention(cli.GetContext(), &in)

	if err != nil {
		return cli.GetError(err)
	}

	return nil
}
