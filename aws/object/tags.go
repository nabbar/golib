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
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (cli *client) GetTags(object, version string) ([]sdktps.Tag, error) {
	in := sdksss.GetObjectTaggingInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, err := cli.s3.GetObjectTagging(cli.GetContext(), &in)

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out.TagSet, nil
}

func (cli *client) SetTags(object, version string, tags ...sdktps.Tag) error {
	in := sdksss.PutObjectTaggingInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		Tagging: &sdktps.Tagging{
			TagSet: tags,
		},
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	_, err := cli.s3.PutObjectTagging(cli.GetContext(), &in)

	if err != nil {
		return cli.GetError(err)
	}

	return nil
}
