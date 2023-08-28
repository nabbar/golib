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
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) GetCORS() ([]sdkstp.CORSRule, error) {
	out, err := cli.s3.GetBucketCors(cli.GetContext(), &sdksss.GetBucketCorsInput{
		Bucket: cli.GetBucketAws(),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else if out.CORSRules == nil || len(out.CORSRules) < 1 {
		return make([]sdkstp.CORSRule, 0), nil
	}

	// MarshalValue always return error as nil
	return out.CORSRules, nil
}

func (cli *client) SetCORS(cors []sdkstp.CORSRule) error {
	_, err := cli.s3.PutBucketCors(cli.GetContext(), &sdksss.PutBucketCorsInput{
		Bucket: cli.GetBucketAws(),
		CORSConfiguration: &sdkstp.CORSConfiguration{
			CORSRules: cors,
		},
	})

	return cli.GetError(err)
}
