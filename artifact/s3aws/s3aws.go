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

package s3aws

import (
	"context"
	"net/http"

	"github.com/nabbar/golib/artifact"
	"github.com/nabbar/golib/artifact/client"
	"github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
)

func NewS3AWS(ctx context.Context, cfg aws.Config, httpcli *http.Client, forceModePath bool, releaseRegex string, releaseGroup int) (cli artifact.Client, err errors.Error) {
	var (
		c aws.AWS
		e errors.Error
	)

	if c, e = aws.New(ctx, cfg, httpcli); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	if forceModePath {
		e = c.ForcePathStyle(true)
	} else {
		e = c.ForcePathStyle(false)
	}

	if e != nil {
		return nil, e
	}

	o := &s3awsModel{
		ClientHelper: client.ClientHelper{},
		c:            c,
		x:            ctx,
		regex:        releaseRegex,
		group:        releaseGroup,
	}

	o.ClientHelper.F = o.ListReleases

	return o, nil
}
