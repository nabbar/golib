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

	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
	libaws "github.com/nabbar/golib/aws"
)

// NewS3AWS creates an AWS S3 artifact client for managing releases stored in S3 buckets.
// Version extraction is performed via regex matching on S3 object keys (file paths).
//
// Parameters:
//   - ctx: Request context for AWS API calls
//   - cfg: AWS configuration (credentials, region, bucket) from github.com/nabbar/golib/aws
//   - httpcli: Optional HTTP client for custom transport/timeout configuration
//   - forceModePath: If true, uses path-style URLs (bucket.s3.region.amazonaws.com/key);
//     if false, uses virtual-hosted-style URLs (bucket-name.s3.region.amazonaws.com/key)
//   - releaseRegex: Regex pattern to match S3 object keys and extract versions (must have at least one capture group)
//   - releaseGroup: Capture group index (1-based) that contains the version string
//
// The regex pattern must include a capture group for version extraction:
//   - Pattern: `releases/myapp-v(\d+\.\d+\.\d+)\.zip` extracts "1.2.3" from "releases/myapp-v1.2.3.zip"
//   - Pattern: `artifacts/(\d+\.\d+\.\d+)/package\.tar\.gz` extracts "2.1.0" from "artifacts/2.1.0/package.tar.gz"
//
// Returns a client implementing the artifact.Client interface for:
//   - Listing releases by scanning S3 bucket objects
//   - Version extraction via regex on object keys
//   - Direct S3 downloads with presigned URLs
//
// Example:
//
//	ctx := context.Background()
//	awsConfig := libaws.Config{
//	    Region: "us-east-1",
//	    Bucket: "my-releases-bucket",
//	}
//	client, err := NewS3AWS(
//	    ctx,
//	    awsConfig,
//	    nil,                                    // Use default HTTP client
//	    false,                                  // Use virtual-hosted-style
//	    `releases/myapp-v(\d+\.\d+\.\d+)\.zip`, // Regex with version capture
//	    1,                                      // Group 1 contains version
//	)
func NewS3AWS(ctx context.Context, cfg libaws.Config, httpcli *http.Client, forceModePath bool, releaseRegex string, releaseGroup int) (cli libart.Client, err error) {
	var (
		c libaws.AWS
		e error
	)

	if c, e = libaws.New(ctx, cfg, httpcli); e != nil {
		return nil, ErrorClientInit.Error(e)
	}

	if forceModePath {
		e = c.ForcePathStyle(ctx, true)
	} else {
		e = c.ForcePathStyle(ctx, false)
	}

	if e != nil {
		return nil, e
	}

	o := &s3awsModel{
		Helper: artcli.Helper{},
		c:      c,
		x:      ctx,
		regex:  releaseRegex,
		group:  releaseGroup,
	}
	// no more needed
	// o.Helper.F = o.ListReleases

	return o, nil
}
