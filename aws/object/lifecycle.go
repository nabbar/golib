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
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (cli *client) GetRetention(object, version string) (until time.Time, mode string, err error) {
	until = time.Time{}
	mode = ""

	in := sdksss.GetObjectRetentionInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, e := cli.s3.GetObjectRetention(cli.GetContext(), &in)

	if e != nil {
		return until, mode, cli.GetError(e)
	}

	until = *out.Retention.RetainUntilDate

	switch out.Retention.Mode {
	case sdktps.ObjectLockRetentionModeGovernance:
		mode = string(sdktps.ObjectLockRetentionModeGovernance)
	case sdktps.ObjectLockRetentionModeCompliance:
		mode = string(sdktps.ObjectLockRetentionModeCompliance)
	}

	return until, mode, nil
}

func (cli *client) SetRetention(object, version string, bypass bool, until time.Time, mode string) error {
	in := sdksss.PutObjectRetentionInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		Retention: &sdktps.ObjectLockRetention{
			RetainUntilDate: sdkaws.Time(until),
		},
	}

	if bypass {
		in.BypassGovernanceRetention = sdkaws.Bool(true)
	}

	switch {
	case strings.EqualFold(mode, string(sdktps.ObjectLockRetentionModeGovernance)):
		in.Retention.Mode = sdktps.ObjectLockRetentionModeGovernance
	case strings.EqualFold(mode, string(sdktps.ObjectLockRetentionModeCompliance)):
		in.Retention.Mode = sdktps.ObjectLockRetentionModeCompliance
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

func (cli *client) GetLegalHold(object, version string) (sdktps.ObjectLockLegalHoldStatus, error) {
	in := sdksss.GetObjectLegalHoldInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, err := cli.s3.GetObjectLegalHold(cli.GetContext(), &in)

	if err != nil {
		return sdktps.ObjectLockLegalHoldStatusOff, cli.GetError(err)
	}

	return out.LegalHold.Status, nil
}

func (cli *client) SetLegalHold(object, version string, flag sdktps.ObjectLockLegalHoldStatus) error {
	in := sdksss.PutObjectLegalHoldInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		LegalHold: &sdktps.ObjectLockLegalHold{
			Status: flag,
		},
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	_, err := cli.s3.PutObjectLegalHold(cli.GetContext(), &in)

	if err != nil {
		return cli.GetError(err)
	}

	return nil
}
