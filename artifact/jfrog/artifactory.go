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

package jfrog

import (
	"context"
	"net/http"
	"net/url"

	libart "github.com/nabbar/golib/artifact"
	artcli "github.com/nabbar/golib/artifact/client"
)

// NewArtifactory creates a JFrog Artifactory artifact client using the Storage API.
// Version extraction is performed via regex matching on artifact file names.
//
// Parameters:
//   - ctx: Request context for API calls
//   - Do: HTTP client Do function (e.g., http.Client.Do) for custom authentication/transport
//   - uri: Artifactory base URL (e.g., "https://artifactory.example.com")
//   - releaseRegex: Regex pattern to match artifacts and extract versions (must have at least one capture group)
//   - releaseGroup: Capture group index (1-based) that contains the version string
//   - reposPath: Repository path segments (e.g., "repo-name", "path", "to", "artifacts")
//
// The regex pattern must include a capture group for version extraction:
//   - Pattern: `myapp-(\d+\.\d+\.\d+)\.tar\.gz` extracts "1.2.3" from "myapp-1.2.3.tar.gz"
//   - Pattern: `release-v(\d+\.\d+\.\d+)-linux\.zip` extracts "2.1.0" from "release-v2.1.0-linux.zip"
//
// Returns a client implementing the artifact.Client interface for:
//   - Listing releases by scanning repository files
//   - Version extraction via regex
//   - Direct artifact downloads
//
// Example:
//
//	ctx := context.Background()
//	httpClient := &http.Client{}
//	client, err := NewArtifactory(
//	    ctx,
//	    httpClient.Do,
//	    "https://artifactory.example.com",
//	    `myapp-(\d+\.\d+\.\d+)\.tar\.gz`,  // Regex with version capture
//	    1,                                   // Group 1 contains version
//	    "releases", "myapp",                 // Path: releases/myapp/
//	)
func NewArtifactory(ctx context.Context, Do func(req *http.Request) (*http.Response, error), uri, releaseRegex string, releaseGroup int, reposPath ...string) (libart.Client, error) {
	if u, e := url.Parse(uri); e != nil {
		return nil, ErrorURLParse.Error(e)
	} else {
		a := &art{
			Helper:   artcli.Helper{},
			Do:       Do,
			ctx:      ctx,
			endpoint: u,
			path:     reposPath,
			group:    releaseGroup,
			regex:    releaseRegex,
		}
		// no more needed
		//a.Helper.F = a.ListReleases

		return a, nil
	}
}
