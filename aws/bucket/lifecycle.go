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

package bucket

import (
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) GetLifeCycle() ([]sdkstp.LifecycleRule, error) {
	out, err := cli.s3.GetBucketLifecycleConfiguration(cli.GetContext(), &sdksss.GetBucketLifecycleConfigurationInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return nil, libhlp.ErrorBucketNotFound.Error(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return out.Rules, nil
}

func (cli *client) SetLifeCycle(rules ...sdkstp.LifecycleRule) error {
	out, err := cli.s3.PutBucketLifecycleConfiguration(cli.GetContext(), &sdksss.PutBucketLifecycleConfigurationInput{
		Bucket: cli.GetBucketAws(),
		LifecycleConfiguration: &sdkstp.BucketLifecycleConfiguration{
			Rules: rules,
		},
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return libhlp.ErrorBucketNotFound.Error(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return nil
}

func (cli *client) GetLock() (*sdkstp.ObjectLockConfiguration, error) {
	out, err := cli.s3.GetObjectLockConfiguration(cli.GetContext(), &sdksss.GetObjectLockConfigurationInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return nil, libhlp.ErrorBucketNotFound.Error(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return out.ObjectLockConfiguration, nil
}

func (cli *client) SetLock(cfg sdkstp.ObjectLockConfiguration, token string) error {
	in := &sdksss.PutObjectLockConfigurationInput{
		Bucket:                  cli.GetBucketAws(),
		ObjectLockConfiguration: &cfg,
	}

	if len(token) > 0 {
		in.Token = sdkaws.String(token)
	}

	out, err := cli.s3.PutObjectLockConfiguration(cli.GetContext(), in)

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		//nolint #goerr113
		return libhlp.ErrorBucketNotFound.Error(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return nil
}
